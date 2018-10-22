package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setAndTestBackoffOneIncrement(t *testing.T, backoff *ExponentialBackoff) {
	canProceed, batch := backoff.CanProceed()
	assert.Equal(t, true, canProceed)
	assert.Equal(t, false, batch.shouldIncrement)

	backoff.Backoff(batch)
	assert.Equal(t, 0, backoff.increments)

	// Still in the backoff period, so should not increment
	canProceed, newBatch := backoff.CanProceed()
	assert.Equal(t, false, canProceed)
	assert.Equal(t, false, newBatch.shouldIncrement)
	assert.Equal(t, batch.uuid, newBatch.uuid)

	// Fake time interval
	fakeTime := time.Now().Add(-10 * time.Millisecond)
	backoff.backoffUntil = &fakeTime
}

func TestBackoff__NoIncrement(t *testing.T) {
	backoff := NewExponentialBackoff(BackoffIntervals{
		StartInterval: 10 * time.Minute,
		Multiplier:    2,
		MaxInterval:   2 * time.Minute,
	})

	setAndTestBackoffOneIncrement(t, backoff)

	canProceed, newBatch := backoff.CanProceed()
	assert.Equal(t, true, canProceed)
	assert.Equal(t, true, newBatch.shouldIncrement)
}

func TestBackoff__Increment(t *testing.T) {
	backoff := NewExponentialBackoff(BackoffIntervals{
		StartInterval: 10 * time.Minute,
		Multiplier:    2,
		MaxInterval:   2 * time.Minute,
	})

	setAndTestBackoffOneIncrement(t, backoff)

	canProceed, newBatch := backoff.CanProceed()
	assert.Equal(t, true, canProceed)
	assert.Equal(t, true, newBatch.shouldIncrement)

	backoff.Backoff(newBatch)
	assert.Equal(t, 1, backoff.increments)
}

func TestBackoff__IncrementsToMax(t *testing.T) {
	backoff := NewExponentialBackoff(BackoffIntervals{
		StartInterval: 10 * time.Minute,
		Multiplier:    2,
		MaxInterval:   2 * time.Minute,
	})

	setAndTestBackoffOneIncrement(t, backoff)

	for i := 0; i < 50; i++ {
		canProceed, batch := backoff.CanProceed()
		assert.Equal(t, true, canProceed)
		assert.Equal(t, true, batch.shouldIncrement)

		backoff.Backoff(batch)
		assert.Equal(t, i+1, backoff.increments)

		// Fake time interval
		fakeTime := time.Now().Add(-10 * time.Millisecond)
		backoff.backoffUntil = &fakeTime
	}

	assert.Equal(t, 50, backoff.increments)

	// Fake time interval
	fakeTime := time.Now().Add(-10 * time.Millisecond)
	backoff.backoffUntil = &fakeTime

	canProceed, newBatch := backoff.CanProceed()
	assert.Equal(t, true, canProceed)
	assert.Equal(t, true, newBatch.shouldIncrement)
}

func TestBackoff__Batches(t *testing.T) {
	backoff := NewExponentialBackoff(BackoffIntervals{
		StartInterval: 10 * time.Minute,
		Multiplier:    2,
		MaxInterval:   2 * time.Minute,
	})

	for i := 0; i < 50; i++ {
		canProceed, batch := backoff.CanProceed()
		assert.Equal(t, i == 0, canProceed)
		assert.Equal(t, false, batch.shouldIncrement)

		backoff.Backoff(batch)
	}

	assert.Equal(t, 0, backoff.increments)

	// Fake time interval
	fakeTime := time.Now().Add(-10 * time.Millisecond)
	backoff.backoffUntil = &fakeTime

	canProceed, batch := backoff.CanProceed()
	assert.Equal(t, true, canProceed)
	assert.Equal(t, true, batch.shouldIncrement)

	for i := 0; i < 50; i++ {
		canProceed, batch := backoff.CanProceed()
		assert.Equal(t, i == 0, canProceed)
		assert.Equal(t, false, batch.shouldIncrement)

		backoff.Backoff(batch)
	}

	assert.Equal(t, 0, backoff.increments)
}

func TestBackoff__BatchesIncrement(t *testing.T) {
	backoff := NewExponentialBackoff(BackoffIntervals{
		StartInterval: 10 * time.Minute,
		Multiplier:    2,
		MaxInterval:   2 * time.Minute,
	})

	for i := 0; i < 50; i++ {
		canProceed, batch := backoff.CanProceed()
		assert.Equal(t, i == 0, canProceed)
		assert.Equal(t, false, batch.shouldIncrement)

		backoff.Backoff(batch)
	}

	assert.Equal(t, 0, backoff.increments)

	// Fake time interval
	fakeTime := time.Now().Add(-10 * time.Millisecond)
	backoff.backoffUntil = &fakeTime

	canProceed, batch := backoff.CanProceed()
	assert.Equal(t, true, canProceed)
	assert.Equal(t, true, batch.shouldIncrement)

	backoff.Backoff(batch)
	assert.Equal(t, 1, backoff.increments)

	for i := 0; i < 50; i++ {
		canProceed, batch := backoff.CanProceed()
		assert.Equal(t, false, canProceed)
		assert.Equal(t, false, batch.shouldIncrement)

		backoff.Backoff(batch)
	}

	assert.Equal(t, 1, backoff.increments)
}
