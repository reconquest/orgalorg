package main

type (
	terminator  struct{}
	terminators chan terminator
	job         func(terminators)

	jobs chan job
)

func startWorkers(
	amount int,
) (jobs, []terminators) {
	jobsChannel := make(jobs, 0)

	terminatorsList := []terminators{}

	for i := 0; i < amount; i++ {
		terminatorsChannel := make(terminators, 0)
		go worker(jobsChannel, terminatorsChannel)

		terminatorsList = append(terminatorsList, terminatorsChannel)
	}

	return jobsChannel, terminatorsList
}

func stopWorkers(terminatorsList []chan<- struct{}) {
	for _, terminators := range terminatorsList {
		terminators <- struct{}{}
	}
}

func worker(jobsChannel jobs, terminatorsChannel terminators) {
	for {
		select {
		case job := <-jobsChannel:
			job(terminatorsChannel)
		case <-terminatorsChannel:
			return
		}
	}
}
