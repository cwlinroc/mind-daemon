package main

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	minddaemonv1 "mind_daemon_dmz/gen/minddaemon/v1"
)

func TestSayHello(t *testing.T) {
	server := &helloServer{}

	res, err := server.SayHello(context.Background(), connect.NewRequest(&minddaemonv1.HelloRequest{
		Name: "tester",
	}))
	if err != nil {
		t.Fatalf("SayHello returned error: %v", err)
	}

	want := "Hello, tester from mind-daemon-dmz"
	if got := res.Msg.GetMessage(); got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestSayHelloUsesDefaultName(t *testing.T) {
	server := &helloServer{}

	res, err := server.SayHello(context.Background(), connect.NewRequest(&minddaemonv1.HelloRequest{}))
	if err != nil {
		t.Fatalf("SayHello returned error: %v", err)
	}

	want := "Hello, local developer from mind-daemon-dmz"
	if got := res.Msg.GetMessage(); got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}
