package worker

import (
	"context"
	"log"

	"github.com/KrishnaGrg1/hookfire/internal/queue"
	"github.com/KrishnaGrg1/hookfire/internal/store"
)

type Dispatcher struct {
	queue       *queue.Queue
	store       *store.Store
	workerCount int
}

func NewDispatcher(q *queue.Queue, s *store.Store, workerCount int) *Dispatcher {
	return &Dispatcher{
		queue:       q,
		store:       s,
		workerCount: workerCount,
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	log.Printf("Starting %d workers", d.workerCount)

	// Spin up N goroutines — each waits for jobs independently
	for i := 0; i < d.workerCount; i++ {
		go d.runWorker(ctx, i)
	}

	// Wait until app shuts down
	<-ctx.Done()
	log.Println("Workers stopped")
}

func (d *Dispatcher) runWorker(ctx context.Context, id int) {
	deliverer := NewDeliverer(d.store, d.queue)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Blocking wait for a job
			job, err := d.queue.Dequeue(ctx)
			if err != nil {
				log.Printf("Worker %d: dequeue error: %v", id, err)
				continue
			}
			if job == nil {
				continue
			}

			log.Printf("Worker %d: picked up job event=%d endpoint=%d",
				id, job.EventID, job.EndpointID)

			// Deliver in a new goroutine so worker is free immediately
			go deliverer.Deliver(ctx, job)
		}
	}
}
