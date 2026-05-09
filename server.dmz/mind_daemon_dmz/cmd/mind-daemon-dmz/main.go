package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	minddaemonv1 "mind_daemon_dmz/gen/minddaemon/v1"
	"mind_daemon_dmz/gen/minddaemon/v1/minddaemonv1connect"
)

type helloServer struct {
	minddaemonv1connect.UnimplementedHelloServiceHandler
}

func (s *helloServer) SayHello(
	ctx context.Context,
	req *connect.Request[minddaemonv1.HelloRequest],
) (*connect.Response[minddaemonv1.HelloReply], error) {
	name := req.Msg.GetName()
	if name == "" {
		name = "local developer"
	}

	return connect.NewResponse(&minddaemonv1.HelloReply{
		Message: fmt.Sprintf("Hello, %s from mind-daemon-dmz", name),
	}), nil
}

func main() {
	addr := flag.String("addr", "localhost:8080", "HTTP listen address")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle(minddaemonv1connect.NewHelloServiceHandler(&helloServer{}))

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	server := &http.Server{
		Addr:      *addr,
		Handler:   mux,
		Protocols: protocols,
	}

	log.Printf("mind-daemon-dmz listening on http://%s", *addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
