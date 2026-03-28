package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/mempubsub"
)

type Publisher struct {
	topic *pubsub.Topic
}

func NewPublisher(topic *pubsub.Topic) *Publisher {
	return &Publisher{topic: topic}
}

func (p *Publisher) Publish(ctx context.Context, topicName string, data interface{}) error {
	if p.topic == nil {
		log.Printf("Topic not configured, skipping publish to %s", topicName)
		return nil
	}

	msgBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &pubsub.Message{
		Body: msgBytes,
	}

	if err := p.topic.Send(ctx, msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Published message to topic %s", topicName)
	return nil
}

func OpenTopic(ctx context.Context, url string) (*pubsub.Topic, error) {
	topic, err := pubsub.OpenTopic(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to open topic: %w", err)
	}
	return topic, nil
}