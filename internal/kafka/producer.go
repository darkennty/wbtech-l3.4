package kafka

import (
	"context"
	"encoding/json"
	"time"

	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type Task struct {
	ImageID      string `json:"image_id"`
	OriginalPath string `json:"original_path"`
}

type Producer struct {
	producer *wbfkafka.Producer
	topic    string
	logger   zlog.Zerolog
}

func NewProducer(brokers []string, topic string, logger zlog.Zerolog) (*Producer, error) {
	producer := wbfkafka.NewProducer(brokers, topic)
	return &Producer{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}, nil
}

func (p *Producer) Send(ctx context.Context, task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	strategy := retry.Strategy{
		Attempts: 3,
		Delay:    500 * time.Millisecond,
		Backoff:  2,
	}
	return p.producer.SendWithRetry(ctx, strategy, []byte(task.ImageID), data)
}
