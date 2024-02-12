package main

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "queue-workers/proto"
	"queue-workers/queue"
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	q map[string]queue.Queue
	pb.UnimplementedQueueServer
}

func (s *server) AddTask(ctx context.Context, r *pb.TaskAddRequest) (*pb.TaskAddReply, error) {
	qid := r.GetTask().GetQueue()
	q, ok := s.q[qid]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("requested queue %s is not present", r.GetTask().Queue))
	}
	q.Add(queue.Key(r.GetTask().GetName()), r.GetTask().GetPayload())
	return &pb.TaskAddReply{Msg: fmt.Sprintf("task %s added to queue %s", r.GetTask().Name, r.GetTask().Queue)}, nil
}

func (s *server) WatchQueue(r *pb.TaskWatchRequest, w pb.Queue_WatchQueueServer) error {
	qid := r.Queue
	q, ok := s.q[qid]
	if !ok {
		return status.Errorf(codes.InvalidArgument, fmt.Sprintf("requested queue %s is not present", r.Queue))
	}
	timer := time.Tick(1 * time.Second)
	lastSeen := q.Peep()
	errChan := make(chan error)
	go func() {
		for {
			select {
			case <-timer:
				seen := q.Peep()
				if reflect.DeepEqual(lastSeen, seen) {
					log.Printf("no updates to %+v, waiting ...", seen)
					continue
				}
				k, v := seen.KeyValue()
				pl := v.(string)
				log.Printf("sending event for new item in the queue: %+v", seen)
				if err := w.Send(&pb.TaskWatchReply{Task: &pb.Task{Name: string(k), Payload: pl, Queue: qid}}); err != nil {
					errChan <- status.Errorf(codes.Internal, fmt.Sprintf("failed to stream event: %s", err))
				}
				lastSeen = seen
			case <-w.Context().Done():
				errChan <- status.Errorf(codes.DeadlineExceeded, "client side timeout exceeded")
			}
		}
	}()
	return <-errChan
}

func main() {
	tcpListener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	grpcServ := grpc.NewServer()
	pb.RegisterQueueServer(grpcServ, &server{
		q: map[string]queue.Queue{
			"q1": queue.NewQueue(), "q2": queue.NewQueue(),
		},
	})
	log.Print("server listening at :8000")
	if err := grpcServ.Serve(tcpListener); err != nil {
		log.Fatalf("failed to server grpc: %s", err)
	}
}
