/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/robfig/cron"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "tutorial.kubebuilder.io/project/api/v1"
)

// CronJobReconciler reconciles a CronJob object
type CronJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Clock
}

type realClock struct{}

func (_ realClock) Now() time.Time { return time.Now() }

type Clock interface {
	Now() time.Time
}

type CronJobChildRuns struct {
	activeJobs     []*kbatch.Job
	successfulJobs []*kbatch.Job
	failedJobs     []*kbatch.Job
}

//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=cronjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

var (
	scheduledTimeAnnotation = "batch.tutorial.kubebuilder.io/scheduled-at"
	jobOwnerKey             = ".metadata.controller"
	apiGVStr                = batchv1.GroupVersion.String()
)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var cronJob batchv1.CronJob
	if err := r.Get(ctx, req.NamespacedName, &cronJob); err != nil {
		log.Error(err, "unable to fetch CronJob")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List child runs of the target cron job
	childRuns, err := r.listCronJobChildRuns(ctx, &cronJob)
	if err != nil {
		log.Error(err, "failed to list child runs for CronJob")
		return ctrl.Result{}, err
	}
	// update cron job status
	// status should be able to be reconstituted from the state of the world,
	// so do not read from the status of the root object
	// reconstruct the status every run instead, from scratch
	cronJob.Status.Active = nil
	for _, job := range childRuns.activeJobs {
		jobRef, err := ref.GetReference(r.Scheme, job)
		if err != nil {
			log.Error(err, "unable to make reference to active job", "job", job)
			continue
		}
		// update the cron job status to the list of active runs
		cronJob.Status.Active = append(cronJob.Status.Active, *jobRef)
	}
	log.V(1).Info("job count", "active jobs", len(childRuns.activeJobs), "successful jobs", len(childRuns.successfulJobs), "failed jobs", len(childRuns.failedJobs))
	if err := r.Status().Update(ctx, &cronJob); err != nil {
		log.Error(err, "unable to update CronJob status")
		return ctrl.Result{}, err
	}

	// Clean up older runs of the parent job by history limit
	log.V(1).Info("enforcing failed job history limit", "limit", *cronJob.Spec.FailedJobsHistoryLimit)
	r.enforceHistoryLimit(ctx, cronJob.Spec.FailedJobsHistoryLimit, childRuns.failedJobs)
	log.V(1).Info("enforcing successful job history limit", "limit", *cronJob.Spec.SuccessfulJobsHistoryLimit)
	r.enforceHistoryLimit(ctx, cronJob.Spec.SuccessfulJobsHistoryLimit, childRuns.successfulJobs)

	// start generating next runs
	// skip if the cron job is suspended
	if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
		log.V(1).Info("cronjob suspended, skipping")
		return ctrl.Result{}, nil
	}
	// figure out the next times that we need to create
	// jobs at (or anything we missed).
	missedRun, nextRun, err := getNextSchedule(&cronJob, r.Now())
	if err != nil {
		log.Error(err, "unable to figure out CronJob schedule")
		// we don't really care about requeuing until we get an update that
		// fixes the schedule, so don't return an error
		return ctrl.Result{}, nil
	}
	// generate a new job run
	if newJob, err := r.newJobRun(ctx, &cronJob, childRuns, missedRun, nextRun); err != nil {
		log.Error(err, "failed to generate job run fro CronJob")
		return ctrl.Result{}, err
	} else if newJob == nil {
		// if no new run should be disptached, skip
		log.Info("no new runs will be dispatched at this point")
		return ctrl.Result{}, nil
	} else {
		// otherwise create a new run
		if err := r.Create(ctx, newJob); err != nil {
			log.Error(err, "unable to create Job for CronJob", "job", newJob)
			return ctrl.Result{}, err
		}
		log.V(1).Info("created Job for CronJob run", "job", newJob)
	}
	return ctrl.Result{RequeueAfter: nextRun.Sub(r.Now())}, nil
}

func (r *CronJobReconciler) newJobRun(ctx context.Context, cronJob *batchv1.CronJob, childRuns *CronJobChildRuns, missedRun, nextRun time.Time) (*kbatch.Job, error) {
	log := log.FromContext(ctx)
	log = log.WithValues("now", r.Now(), "next run", nextRun)
	// no missed run meaning there is nothing to dispatch
	if missedRun.IsZero() {
		log.V(1).Info("no upcoming scheduled times, sleeping until next")
		return nil, nil
	}
	// If we’ve missed a run, and we’re still within the deadline to start it, we’ll need to run a job.
	// make sure we're not too late to start the run
	log = log.WithValues("current run", missedRun)
	tooLate := false
	if cronJob.Spec.StartingDeadlineSeconds != nil {
		tooLate = missedRun.Add(time.Duration(*cronJob.Spec.StartingDeadlineSeconds) * time.Second).Before(r.Now())
	}
	if tooLate {
		log.V(1).Info("missed starting deadline for last run, sleeping till next")
		// TODO(directxman12): events
		return nil, nil
	}
	// figure out how to run this job -- concurrency policy might forbid us from running
	// multiple at the same time...
	if cronJob.Spec.ConcurrencyPolicy == batchv1.ForbidConcurrent && len(childRuns.activeJobs) > 0 {
		log.V(1).Info("concurrency policy blocks concurrent runs, skipping", "num active", len(childRuns.activeJobs))
		return nil, nil
	}
	// ...or instruct us to replace existing ones...
	if cronJob.Spec.ConcurrencyPolicy == batchv1.ReplaceConcurrent {
		for _, activeJob := range childRuns.activeJobs {
			// we don't care if the job was already deleted
			if err := r.Delete(ctx, activeJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
				log.Error(err, "unable to delete active job", "job", activeJob)
				return nil, err
			}
		}
	}
	return r.constructJobForCronJob(cronJob, missedRun)
}

func (r *CronJobReconciler) listCronJobChildRuns(ctx context.Context, cronJob *batchv1.CronJob) (*CronJobChildRuns, error) {
	log := log.FromContext(ctx)
	var childJobs kbatch.JobList
	// list all the child jobs owned by the cron job by the owner key
	if err := r.List(ctx, &childJobs, client.MatchingFields{"jobOwnerKey": cronJob.Name}); err != nil {
		log.Error(err, "unable to list child Jobs")
		return nil, err
	}
	// categorize jobs into different lists
	var activeJobs []*kbatch.Job
	var successfulJobs []*kbatch.Job
	var failedJobs []*kbatch.Job
	// find the last run so we can update the status
	var mostRecentTime *time.Time
	for _, job := range childJobs.Items {
		_, finishedType := isJobFinished(&job)
		switch finishedType {
		case kbatch.JobComplete:
			successfulJobs = append(successfulJobs, &job)
		case kbatch.JobFailed:
			failedJobs = append(failedJobs, &job)
		default:
			activeJobs = append(activeJobs, &job)
		}
		// We'll store the launch time in an annotation, so we'll reconstitute that from
		// the active jobs themselves.
		scheduledTimeForJob, err := getScheduledTimeForJob(&job)
		if err != nil {
			log.Error(err, "unable to parse schedule time for child job", "job", &job)
			continue
		}
		// keep track of the scheduled time of last executed child job
		if scheduledTimeForJob != nil {
			if mostRecentTime == nil {
				mostRecentTime = scheduledTimeForJob
			} else if mostRecentTime.Before(*scheduledTimeForJob) {
				mostRecentTime = scheduledTimeForJob
			}
		}
	}

	// update the most recent run to the status of the parent job
	if mostRecentTime != nil {
		cronJob.Status.LastScheduleTime = &metav1.Time{Time: *mostRecentTime}
	} else {
		cronJob.Status.LastScheduleTime = nil
	}

	return &CronJobChildRuns{activeJobs: activeJobs, failedJobs: failedJobs, successfulJobs: successfulJobs}, nil
}

// enforce job histoyr limit for cron job
func (r *CronJobReconciler) enforceHistoryLimit(ctx context.Context, historyLimit *int32, jobRuns []*kbatch.Job) {
	log := log.FromContext(ctx)
	if historyLimit != nil {
		sort.Slice(jobRuns, func(i, j int) bool {
			if jobRuns[i].Status.StartTime == nil {
				return jobRuns[j].Status.StartTime != nil
			}
			return jobRuns[i].Status.StartTime.Before(
				jobRuns[j].Status.StartTime,
			)
		})
		for i, job := range jobRuns {
			// stop if we've reached the limit
			if int32(i) >= int32(len(jobRuns))-*historyLimit {
				break
			}
			if err := r.Delete(ctx, job, client.PropagationPolicy(
				metav1.DeletePropagationBackground,
			)); client.IgnoreNotFound(err) != nil {
				log.Error(err, "unable to delete old job", "job", job)
			} else {
				log.V(0).Info("deleted old job", "job", job)
			}
		}
	}
}

// We need to construct a job based on our CronJob’s template. We’ll copy over the spec from the template and copy some basic object meta.
// Then, we’ll set the “scheduled time” annotation so that we can reconstitute our LastScheduleTime field each reconcile.
// Finally, we’ll need to set an owner reference. This allows the Kubernetes garbage collector to clean up jobs when we delete the CronJob, and allows controller-runtime to figure out which cronjob needs to be reconciled when a given job changes (is added, deleted, completes, etc).
func (r *CronJobReconciler) constructJobForCronJob(cronJob *batchv1.CronJob, scheduledTime time.Time) (*kbatch.Job, error) {
	// We want job names for a given nominal start time to have a deterministic name to avoid the same job being created twice
	name := fmt.Sprintf("%s-%d", cronJob.Name, scheduledTime.Unix())

	job := &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        name,
			Namespace:   cronJob.Namespace,
		},
		Spec: *cronJob.Spec.JobTemplate.Spec.DeepCopy(),
	}
	for k, v := range cronJob.Spec.JobTemplate.Annotations {
		job.Annotations[k] = v
	}
	job.Annotations[scheduledTimeAnnotation] = scheduledTime.Format(time.RFC3339)
	for k, v := range cronJob.Spec.JobTemplate.Labels {
		job.Labels[k] = v
	}
	if err := ctrl.SetControllerReference(cronJob, job, r.Scheme); err != nil {
		return nil, err
	}

	return job, nil
}

// Calculate the next scheduled time using our helpful cron library. We’ll start calculating appropriate times from our last run, or the creation of the CronJob if we can’t find a last run.
// If there are too many missed runs and we don’t have any deadlines set, we’ll bail so that we don’t cause issues on controller restarts or wedges.
// Otherwise, we’ll just return the missed runs (of which we’ll just use the latest), and the next run, so that we can know when it’s time to reconcile again.
func getNextSchedule(cronJob *batchv1.CronJob, now time.Time) (lastMissed time.Time, next time.Time, err error) {
	sched, err := cron.ParseStandard(cronJob.Spec.Schedule)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("unparseable schedule %q: %v", cronJob.Spec.Schedule, err)
	}

	// for optimization purposes, cheat a bit and start from our last observed run time
	// we could reconstitute this here, but there's not much point, since we've
	// just updated it.
	var earliestTime time.Time
	if cronJob.Status.LastScheduleTime != nil {
		earliestTime = cronJob.Status.LastScheduleTime.Time
	} else {
		earliestTime = cronJob.ObjectMeta.CreationTimestamp.Time
	}
	if cronJob.Spec.StartingDeadlineSeconds != nil {
		// controller is not going to schedule anything below this point
		schedulingDeadline := now.Add(-time.Second * time.Duration(*cronJob.Spec.StartingDeadlineSeconds))

		if schedulingDeadline.After(earliestTime) {
			earliestTime = schedulingDeadline
		}
	}
	if earliestTime.After(now) {
		return time.Time{}, sched.Next(now), nil
	}

	starts := 0
	for t := sched.Next(earliestTime); !t.After(now); t = sched.Next(t) {
		lastMissed = t
		// An object might miss several starts. For example, if
		// controller gets wedged on Friday at 5:01pm when everyone has
		// gone home, and someone comes in on Tuesday AM and discovers
		// the problem and restarts the controller, then all the hourly
		// jobs, more than 80 of them for one hourly scheduledJob, should
		// all start running with no further intervention (if the scheduledJob
		// allows concurrency and late starts).
		//
		// However, if there is a bug somewhere, or incorrect clock
		// on controller's server or apiservers (for setting creationTimestamp)
		// then there could be so many missed start times (it could be off
		// by decades or more), that it would eat up all the CPU and memory
		// of this controller. In that case, we want to not try to list
		// all the missed start times.
		starts++
		if starts > 100 {
			// We can't get the most recent times so just return an empty slice
			return time.Time{}, time.Time{}, fmt.Errorf("too many missed start times (> 100); set or decrease .spec.startingDeadlineSeconds or check clock skew")
		}
	}
	return lastMissed, sched.Next(now), nil
}

// checks if job is terminal state (complete or failed)
func isJobFinished(job *kbatch.Job) (bool, kbatch.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if c.Status != corev1.ConditionTrue {
			continue
		}
		if c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed {
			return true, c.Type
		}
	}
	return false, ""
}

// parse job scheduled time
func getScheduledTimeForJob(job *kbatch.Job) (*time.Time, error) {
	raw := job.Annotations[scheduledTimeAnnotation]
	var parsed time.Time
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set up a real clock, since we're not in a test
	if r.Clock == nil {
		r.Clock = realClock{}
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kbatch.Job{}, jobOwnerKey, func(rawObj client.Object) []string {
		// grab the job object, extract the owner...
		job := rawObj.(*kbatch.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		// ...make sure it's a CronJob...
		if owner.APIVersion != apiGVStr || owner.Kind != "CronJob" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.CronJob{}).
		Owns(&kbatch.Job{}).
		Complete(r)
}
