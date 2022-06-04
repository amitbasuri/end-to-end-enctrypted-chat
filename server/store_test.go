package main

import (
	"sync"
	"testing"
)

func TestStore(t *testing.T) {
	store := NewInMem()
	wg := sync.WaitGroup{}
	n := 100
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.AddMessage("alice", SenderMessage{
				To:      "bob",
				Message: []byte("hey bob"),
			})
		}()
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.AddMessage("bob", SenderMessage{
				To:      "alice",
				Message: []byte("hey alice"),
			})
		}()
	}

	totalAliceMessages := 0
	aliceMessages := store.GetUserMessages("alice")
	totalAliceMessages += len(aliceMessages)

	totalBobMessages := 0
	bobMessages := store.GetUserMessages("bob")
	totalBobMessages += len(bobMessages)

	wg.Wait()

	aliceMessages = store.GetUserMessages("alice")
	totalAliceMessages += len(aliceMessages)
	if totalAliceMessages != n {
		t.Errorf("alice should have received %d messages, got %d", n, totalAliceMessages)
	}

	bobMessages = store.GetUserMessages("bob")
	totalBobMessages += len(bobMessages)
	if totalBobMessages != n {
		t.Errorf("bob should have received %d messages, got %d\"", n, totalBobMessages)
	}

}
