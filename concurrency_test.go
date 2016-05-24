package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanRunSingleWorker(t *testing.T) {
	test := assert.New(t)

	result := make(chan int, 0)
	jobs, _ := startWorkers(1)

	jobs <- func(_ terminators) {
		result <- 2 * 2
	}

	test.Equal(<-result, 4)
}

func TestCanRunTwoWorkers(t *testing.T) {
	test := assert.New(t)

	result := make(chan int, 0)
	jobs, _ := startWorkers(2)

	jobs <- func(_ terminators) {
		result <- 3 * <-result
	}

	jobs <- func(_ terminators) {
		result <- 2 * 2
	}

	test.Equal(<-result, 12)
}

func TestCanTerminateWorkers(t *testing.T) {
	test := assert.New(t)

	result := make(chan int, 0)
	jobs, terminatorsList := startWorkers(3)

	terminatorsList[0] <- terminator{}
	terminatorsList[1] <- terminator{}

	jobs <- func(_ terminators) {
		result <- 2 * 2
	}

	test.Equal(<-result, 4)
}
