package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"grpc/api"
	"grpc/internal/grpcsync"

	pbErr "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "request missing required metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token provided in request")
)

type greeterSvc struct {
	api.UnimplementedGreeterServer
}

func (s *greeterSvc) SayHello(ctx context.Context, r *api.HelloRequest) (*api.HelloReply, error) {
	log.Printf("received hello request from: %v", r.GetName())
	if !strings.EqualFold(r.GetName(), "will") {
		st := status.New(codes.NotFound, fmt.Sprintf("%s is not found", r.GetName()))
		detailedSt, err := st.WithDetails(&pbErr.ErrorInfo{Reason: "this server will only greet will"})
		if err != nil {
			log.Printf("failed to attach deatils to the error response: %s", err)
			return nil, st.Err()
		}
		return nil, detailedSt.Err()
	}
	return &api.HelloReply{Message: "Hello " + r.GetName()}, nil
}

type echoSvc struct {
	api.UnimplementedEchoServer
}

func (s *echoSvc) BidirectionalStreamingEcho(stream api.Echo_BidirectionalStreamingEchoServer) error {
	log.Printf("starting new bidirectional stream")
	time.Sleep(2 * time.Second)

	// read input stream
	for i := 0; true; i++ {
		if _, err := stream.Recv(); err != nil {
			log.Printf("read %v messages", i)
			if err == io.EOF {
				log.Print("end of stream")
				break
			}
			log.Printf("error reading stream: %s", err)
			return err
		}
	}

	// send output stream
	// make sure we unblock after timeout
	stopEvt := grpcsync.NewEvent()
	sentChan := make(chan struct{})
	go func() {
		for !stopEvt.HasFired() {
			after := time.NewTimer(time.Second)
			select {
			case <-sentChan:
				after.Stop()
			case <-after.C:
				log.Print("event streaming is blocked")
				stopEvt.Fire()
				<-sentChan
			}
		}
	}()

	msgCnt := 0
	for !stopEvt.HasFired() {
		msgCnt++
		if err := stream.Send(&api.EchoResponse{Message: string(make([]byte, 8*1024))}); err != nil {
			log.Printf("error streaming data: %s", err)
			return err
		}
		sentChan <- struct{}{}
	}
	log.Printf("stream ended successfully; %d messages sent", msgCnt)
	return nil
}

func authnMw(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}
	// The keys within metadata.MD are normalized to lowercase.
	if auth, ok := md["authorization"]; !ok || len(auth) != 1 || len(strings.TrimPrefix(auth[0], "Bearer ")) == 0 {
		return nil, errInvalidToken
	}
	return handler(ctx, req)
}

func main() {
	tcpListener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	grpcServ := grpc.NewServer(grpc.UnaryInterceptor(authnMw))
	api.RegisterGreeterServer(grpcServ, &greeterSvc{})
	api.RegisterEchoServer(grpcServ, &echoSvc{})
	log.Print("server listening at :8000")
	if err := grpcServ.Serve(tcpListener); err != nil {
		log.Fatalf("failed to server grpc: %s", err)
	}
}
