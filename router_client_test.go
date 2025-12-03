package openrouterapigo

import (
	"context"
	"testing"
	"time"
)

func TestChatCompletionsStreamRecv(t *testing.T) {
	out := make(chan Response, 1)
	proc := make(chan interface{}, 1)
	errs := make(chan error, 1)

	stream := &ChatCompletionsStream{
		output:     out,
		processing: proc,
		errs:       errs,
	}

	// Send a processing signal then a response, then close.
	proc <- true
	out <- Response{ID: "resp-1"}
	close(out)
	close(proc)
	close(errs)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	seenProcessing := false
	seenResponse := false
	for i := 0; i < 2; i++ {
		ev, ok := stream.Recv(ctx)
		if !ok {
			t.Fatalf("expected event %d, got stream end", i)
		}
		if ev.Processing {
			seenProcessing = true
		}
		if ev.Response != nil && ev.Response.ID == "resp-1" {
			seenResponse = true
		}
	}
	if !seenProcessing || !seenResponse {
		t.Fatalf("missing events, processing=%v response=%v", seenProcessing, seenResponse)
	}

	if _, ok := stream.Recv(ctx); ok {
		t.Fatalf("expected stream to end")
	}
}
