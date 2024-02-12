package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	pb "queue-workers/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	target = flag.String("target", "add", "gRPC to target, one of [add, watch]")
	qid    = flag.String("qid", "q1", "queue to target, one of [q1, q2]")
	key    = flag.String("key", "key", "task key")
	value  = flag.String("value", "value", "task value")
)

type streamEvt struct {
	task *pb.Task
	err  error
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to setup connection: %s", err)
	}
	defer conn.Close()

	ctx, cancelFn := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancelFn()

	cli := pb.NewQueueClient(conn)
	switch *target {
	case "add":
		r, err := cli.AddTask(ctx, &pb.TaskAddRequest{Task: &pb.Task{
			Queue: *qid, Name: *key, Payload: *value,
		}})
		if err != nil {
			log.Fatalf("request failed: %s", err)
		}
		log.Printf("received response: %v", r.Msg)
	case "watch":
		r, err := cli.WatchQueue(ctx, &pb.TaskWatchRequest{Queue: *qid})
		if err != nil {
			log.Fatalf("request failed: %s", err)
		}
		evtChan := make(chan streamEvt)
		log.Printf("estabilishing watch on queue %s", *qid)
		// keep receiving and adding event to chan in the background
		go func() {
			for {
				t, err := r.Recv()
				evtChan <- streamEvt{task: t.Task, err: err}
			}
		}()
		// receive from chan and timeout in the main thread
		for {
			select {
			case e := <-evtChan:
				if e.err != nil {
					if e.err == io.EOF {
						log.Print("end of stream")
						break
					}
					log.Fatalf("failed to stream event: %s", e.err)
				}
				log.Printf("received event from stream: %+v", e.task)
			case <-ctx.Done():
				log.Fatal("timed out waiting for event")
			}
		}
	default:
		log.Fatalf("unsupported target: %s", *target)
	}
}
