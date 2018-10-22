package main

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// BackoffBatch containts information needed for the backoff to work
type BackoffBatch struct {
	uuid            uuid.UUID
	shouldIncrement bool
}

// BackoffIntervals holds data about all the intervals
type BackoffIntervals struct {
	StartInterval time.Duration
	Multiplier    float64
	MaxInterval   time.Duration
}

// ExponentialBackoff takes care of exp. backoff for a request type
type ExponentialBackoff struct {
	intervals BackoffIntervals

	// The time that the current backoff lasts to, nil if not in backoff mode
	backoffUntil *time.Time
	// Current amount of increments on the backoff
	increments int
	// Lock to handle async calls to the struct
	mutex *sync.Mutex
	// The batch works as a way to identify which processes calls the Backoff function in an async environment
	// When Backoff() is called it will only increment once per batch
	batchUUID uuid.UUID
}

// NewExponentialBackoff will initialize and create a new ExponentialBackoff
func NewExponentialBackoff(boInvervals BackoffIntervals) *ExponentialBackoff {
	return &ExponentialBackoff{
		intervals: boInvervals,
		mutex:     &sync.Mutex{},
	}
}

// CanProceed returns true if it's no longer in backoff, it also returns a batch that's needed for the Backoff() call
func (back *ExponentialBackoff) CanProceed() (bool, BackoffBatch) {
	back.mutex.Lock()
	canProceed := !back.inBackoff() || back.backoffUntil.Before(time.Now())
	backoffBatch := BackoffBatch{
		uuid:            back.batchUUID,
		shouldIncrement: false,
	}
	if canProceed && back.inBackoff() {
		back.resetTimer()
		backoffBatch = BackoffBatch{
			uuid:            uuid.New(),
			shouldIncrement: true,
		}
		// Set the new batch ID, only one will have the `shouldIncrement` set to true
		back.batchUUID = backoffBatch.uuid
	}
	back.mutex.Unlock()
	return canProceed, backoffBatch
}

// Backoff will wither start a new backoff or increment the backoff time, based on the batch
func (back *ExponentialBackoff) Backoff(batch BackoffBatch) {
	back.mutex.Lock()

	// If it's in backoff and it shouldn't increment
	if back.inBackoff() && !batch.shouldIncrement {
		back.mutex.Unlock()
		return
	}

	back.startOrIncrementBackoff(batch)

	back.mutex.Unlock()
}

func (back *ExponentialBackoff) startOrIncrementBackoff(batch BackoffBatch) {
	if batch.shouldIncrement {
		incrementsMultiplier := back.increments
		if incrementsMultiplier == 0 {
			incrementsMultiplier = 1
		}
		backOffInterval := float64(back.intervals.StartInterval) * back.intervals.Multiplier * float64(incrementsMultiplier)
		if backOffInterval > float64(back.intervals.MaxInterval) {
			backOffInterval = float64(back.intervals.MaxInterval)
		}
		backoffDuration := time.Duration(backOffInterval)
		backoffTime := time.Now().Add(backoffDuration)
		back.backoffUntil = &backoffTime
		back.increments++
	} else {
		backoffTime := time.Now().Add(back.intervals.StartInterval)
		back.backoffUntil = &backoffTime
		back.increments = 0
	}
}

func (back *ExponentialBackoff) inBackoff() bool {
	return back.backoffUntil != nil
}

func (back *ExponentialBackoff) resetTimer() {
	back.backoffUntil = nil
}
