package tests

import (
	"sync"
	"testing"
)

func TestConcurrency(t *testing.T) {
	ch := make(chan int, 1000)
	var wg sync.WaitGroup

	numGoroutines := 10
	numMessages := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				ch <- start*numMessages + j
			}
		}(i)
	}

	wg.Wait()
	close(ch)

	received := 0
	for range ch {
		received++
	}

	if received != numGoroutines*numMessages {
		t.Errorf("Received %d messages, want %d", received, numGoroutines*numMessages)
	}
}

func TestMutexConcurrency(t *testing.T) {
	var mu sync.Mutex
	counter := 0
	var wg sync.WaitGroup

	numGoroutines := 100
	iterations := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	expected := numGoroutines * iterations
	if counter != expected {
		t.Errorf("Counter = %d, want %d", counter, expected)
	}
}

func TestChannelCommunication(t *testing.T) {
	ch := make(chan string, 10)

	ch <- "hello"
	ch <- "world"

	if len(ch) != 2 {
		t.Errorf("Channel length = %d, want %d", len(ch), 2)
	}

	msg1 := <-ch
	msg2 := <-ch

	if msg1 != "hello" || msg2 != "world" {
		t.Errorf("Received %q, %q, want 'hello', 'world'", msg1, msg2)
	}
}

func TestRaceCondition(t *testing.T) {
	var mu sync.Mutex
	data := make(map[string]int)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "key"
			mu.Lock()
			data[key] = id
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// With proper synchronization, no race conditions
	t.Log("Race condition test completed")
}
