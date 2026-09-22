package singleinstance

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestAcquireRejectsHealthySecondInstance(t *testing.T) {
	uniqueID := fmt.Sprintf("unigodesktop-test-%d", time.Now().UnixNano())
	first, err := Acquire(uniqueID)
	if err != nil {
		t.Fatalf("acquire first instance: %v", err)
	}
	defer first.Close()

	second, err := Acquire(uniqueID)
	if !errors.Is(err, ErrAlreadyRunning) {
		if second != nil {
			second.Close()
		}
		t.Fatalf("expected ErrAlreadyRunning, got %v", err)
	}
}
