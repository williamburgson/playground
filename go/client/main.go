package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	"grpc/api"
	"grpc/internal/grpcsync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	name   = flag.String("name", "will", "Name to greet")
	target = flag.String("target", "hello", "gRPC to target, one of [hello, echo]")
)

// implements PerRPCCredentials interface, used to inject dummy token to requests
type staticTokenProvider struct{}

func (s *staticTokenProvider) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer some-token"}, nil
}

func (s *staticTokenProvider) RequireTransportSecurity() bool {
	return false
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial(
		"localhost:8000",
		grpc.WithPerRPCCredentials(&staticTokenProvider{}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to setup connection: %s", err)
	}
	defer conn.Close()

	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	switch *target {
	case "hello":
		cli := api.NewGreeterClient(conn)
		r, err := cli.SayHello(ctx, &api.HelloRequest{Name: *name})
		if err != nil {
			log.Fatalf("request failed: %s", err)
		}
		log.Printf("received response: %v", r.GetMessage())
	case "echo":
		cli := api.NewEchoClient(conn)
		stream, err := cli.BidirectionalStreamingEcho(ctx)
		if err != nil {
			log.Fatalf("failed to create stream: %s", err)
		}

		log.Print("starting new stream")
		// First we will send data on the stream until we cannot send any more.  We
		// detect this by not seeing a message sent 1s after the last sent message.
		stopSending := grpcsync.NewEvent()
		sentOne := make(chan struct{})
		go func() {
			i := 0
			for !stopSending.HasFired() {
				i++
				if err := stream.Send(&api.EchoRequest{Message: string(make([]byte, 8*1024))}); err != nil {
					log.Fatalf("Error sending data: %v", err)
				}
				sentOne <- struct{}{}
			}
			log.Printf("sent %v messages.", i)
			stream.CloseSend()
		}()

		for !stopSending.HasFired() {
			after := time.NewTimer(time.Second)
			select {
			case <-sentOne:
				after.Stop()
			case <-after.C:
				log.Printf("sending is blocked.")
				stopSending.Fire()
				<-sentOne
			}
		}

		// wait 2 seconds before reading from the stream, to give the
		// server an opportunity to block while sending its responses.
		time.Sleep(2 * time.Second)

		// read all the data sent by the server to allow it to unblock.
		for i := 0; true; i++ {
			if _, err := stream.Recv(); err != nil {
				log.Printf("read %v messages", i)
				if err == io.EOF {
					log.Printf("stream ended successfully.")
					return
				}
				log.Fatalf("error receiving data: %v", err)
			}
		}
	default:
		log.Fatalf("unsupported target: %s", *target)
	}
}
