package worker

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"time"

	db "github.com/KrishnaGrg1/hookfire/internal/db/sqlc"
	"github.com/KrishnaGrg1/hookfire/internal/queue"
	"github.com/KrishnaGrg1/hookfire/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
)

type Deliverer struct {
	store      *store.Store
	queue      *queue.Queue
	httpClient *http.Client
}

func NewDeliverer(s *store.Store, q *queue.Queue) *Deliverer {
	return &Deliverer{
		store: s,
		queue: q,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *Deliverer) Deliver(ctx context.Context, job *queue.Job) {
	// 1. Fetch event from DB
	event, err := d.store.Queries.GetEventByID(ctx, job.EventID)
	if err != nil {
		log.Printf("Deliver: event %d not found", job.EventID)
		return
	}

	// 2. Fetch endpoint from DB
	endpoint, err := d.store.Queries.GetEndpointByID(ctx, job.EndpointID)
	if err != nil {
		log.Printf("Deliver: endpoint %d not found", job.EndpointID)
		return
	}

	// 3. Send HTTP POST to subscriber URL
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.Url,
		bytes.NewReader(event.Payload))
	if err != nil {
		d.saveAttempt(ctx, job, 0, "failed")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hookfire-Event", event.EventType)
	req.Header.Set("X-Hookfire-Delivery", string(rune(job.EventID)))

	resp, err := d.httpClient.Do(req)
	if err != nil {
		log.Printf("Deliver: failed to reach %s: %v", endpoint.Url, err)
		d.handleFailure(ctx, job)
		return
	}
	defer resp.Body.Close()

	// 4. Check if delivery succeeded
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Deliver: success event %d to %s", job.EventID, endpoint.Url)
		d.saveAttempt(ctx, job, resp.StatusCode, "success")
		return
	}

	// Non 2xx = failure
	log.Printf("Deliver: got %d from %s", resp.StatusCode, endpoint.Url)
	d.handleFailure(ctx, job)
}

func (d *Deliverer) handleFailure(ctx context.Context, job *queue.Job) {
	d.saveAttempt(ctx, job, 0, "failed")

	// Give up after 5 attempts
	if job.AttemptNum >= 5 {
		log.Printf("Deliver: event %d dead after 5 attempts", job.EventID)
		return
	}

	// Exponential backoff: 10s, 20s, 40s, 80s
	delay := time.Duration(10*(1<<job.AttemptNum)) * time.Second
	log.Printf("Deliver: retrying event %d in %v", job.EventID, delay)

	time.Sleep(delay)

	nextJob := queue.Job{
		EventID:    job.EventID,
		EndpointID: job.EndpointID,
		AttemptNum: job.AttemptNum + 1,
	}
	d.queue.Enqueue(ctx, nextJob)
}

func (d *Deliverer) saveAttempt(ctx context.Context, job *queue.Job, httpStatus int, status string) {
	d.store.Queries.CreateAttempt(ctx, db.CreateAttemptParams{
		EventID:       job.EventID,
		EndpointID:    job.EndpointID,
		Status:        status,
		HttpStatus:    pgtype.Int4{Int32: int32(httpStatus), Valid: true},
		AttemptNumber: int32(job.AttemptNum),
	})
}
