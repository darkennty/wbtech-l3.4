package kafka

import (
	"context"
	"encoding/json"
	"time"

	"WBTech_L3.4/internal/processor"
	"WBTech_L3.4/internal/repository"
	"github.com/segmentio/kafka-go"
	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type Consumer struct {
	consumer  *wbfkafka.Consumer
	topic     string
	group     string
	repo      *repository.Repository
	processor *processor.Processor
	logger    zlog.Zerolog
}

func NewConsumer(brokers []string, topic, group string, repo *repository.Repository, proc *processor.Processor, logger zlog.Zerolog) (*Consumer, error) {
	consumer := wbfkafka.NewConsumer(brokers, topic, group)

	if group == "" {
		err := consumer.Reader.SetOffset(-1)
		if err != nil {
			return nil, err
		}
	}

	return &Consumer{
		consumer:  consumer,
		topic:     topic,
		group:     group,
		repo:      repo,
		processor: proc,
		logger:    logger,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	msgCh := make(chan kafka.Message)
	strategy := retry.Strategy{
		Attempts: 5,
		Delay:    100 * time.Millisecond,
		Backoff:  2,
	}

	go c.consumer.StartConsuming(ctx, msgCh, strategy)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("shutting down kafka consumer")
			return
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			c.processMessage(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	var task Task
	if err := json.Unmarshal(msg.Value, &task); err != nil {
		c.logger.Error().Err(err).Msg("failed to unmarshal task")
		return
	}

	img, err := c.repo.Image.GetByID(ctx, task.ImageID)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get image")
		return
	}

	img.Status = "processing"
	_ = c.repo.Image.Update(ctx, img)

	watermarkPath, thumbPath, err := c.processor.Process(ctx, img.OriginalPath)
	if err != nil {
		img.Status = "failed"
		img.Error = err.Error()
		_ = c.repo.Image.Update(ctx, img)
		c.logger.Error().Err(err).Msg("image processing failed")
		return
	}

	img.Status = "completed"
	img.WatermarkPath = watermarkPath
	img.ThumbPath = thumbPath
	_ = c.repo.Image.Update(ctx, img)
	c.logger.Info().Str("image_id", task.ImageID).Msg("image processed successfully")
}
