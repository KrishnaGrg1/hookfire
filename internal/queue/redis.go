package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const queueKey = "hookfire:jobs"

type Job struct {
	EventID    int64 `json:"event_id"`
	EndpointID int64 `json:"endpoint_id"`
	AttemptNum int   `json:"attempt_num"`
}

type Queue struct {
	client *redis.Client
}

func New(redisURL string) (*Queue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)

	// Verify connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Queue{client: client}, nil
}

func (q *Queue) Enqueue(ctx context.Context, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, queueKey, data).Err()
}

func (q *Queue) Dequeue(ctx context.Context) (*Job, error) {
	result, err := q.client.BRPop(ctx, 0, queueKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// Add this method to your queue/queue.go
func (q *Queue) GetCache(ctx context.Context, key string) (string, error) {
	return q.client.Get(ctx, key).Result()
}

func (q *Queue) SetCache(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return q.client.Set(ctx, key, data, ttl).Err()
}
