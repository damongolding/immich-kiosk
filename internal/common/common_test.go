package common

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestShutdown(t *testing.T) {
	c := New()

	returned := make(chan struct{})
	go func() {
		c.Shutdown()
		close(returned)
	}()

	select {
	case <-returned:
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown did not return")
	}

	if err := c.Context().Err(); !errors.Is(err, context.Canceled) {
		t.Errorf("Context().Err() = %v, want context.Canceled", err)
	}

	// The shutdown goroutine has exited (and so stopped listening for signals).
	select {
	case <-c.done:
	default:
		t.Error("shutdown goroutine still running after Shutdown returned")
	}

	// A second call must not block.
	again := make(chan struct{})
	go func() {
		c.Shutdown()
		close(again)
	}()
	select {
	case <-again:
	case <-time.After(5 * time.Second):
		t.Fatal("second Shutdown call did not return")
	}
}
