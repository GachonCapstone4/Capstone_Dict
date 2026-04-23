package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"

	"capstone_network_test/internal/models"
)

const (
	defaultHost  = "rabbitmq-headless.rabbitmq"
	defaultPort  = "5672"
	defaultVhost = "/"
	exchangeName = "x.sse.fanout"
)

type Publisher interface {
	Publish(msg models.DiagMessage) error
	Close() error
}

type amqpPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type noopPublisher struct{}

func buildAMQPURL() string {
	user := os.Getenv("RABBITMQ_USER")
	pass := os.Getenv("RABBITMQ_PASS")
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")

	if host == "" {
		host = defaultHost
	}
	if port == "" {
		port = defaultPort
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%s%s", user, pass, host, port, defaultVhost)
}

func NewPublisher() (Publisher, error) {
	url := buildAMQPURL()

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Printf("[MQ] 연결 실패: %v", err)
		return &noopPublisher{}, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return &noopPublisher{}, err
	}

	return &amqpPublisher{conn: conn, channel: ch}, nil
}

func (p *amqpPublisher) Publish(msg models.DiagMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		context.Background(),
		exchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *amqpPublisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}

func (p *noopPublisher) Publish(_ models.DiagMessage) error { return nil }
func (p *noopPublisher) Close() error                       { return nil }
