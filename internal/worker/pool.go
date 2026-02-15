package worker

import (
	"log"
	"sync"
	"time"
)

type Job struct {
	Response chan []byte
}

type Pool struct {
	jobQueue chan Job
	wg       sync.WaitGroup
}

func NewPool(workerCount int, queueSize int) *Pool {
	p := &Pool{
		jobQueue: make(chan Job, queueSize),
	}

	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go p.startWorker(i)
	}

	return p
}

func (p *Pool) startWorker(id int) {
	defer p.wg.Done()

	log.Printf("Worker %d started", id)

	for job := range p.jobQueue {
		// Simulate processing logic
		time.Sleep(time.Second*2)
		result := []byte("Processed by worker\n")

		// Send result back to handler
		job.Response <- result
	}
}

func (p *Pool) Submit(job Job) {
	log.Println("Submitting job ...")
	p.jobQueue <- job
	log.Println("Job Submitted ...")
}

func (p *Pool) Shutdown() {
	close(p.jobQueue)
	p.wg.Wait()
}
