package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	NATS_URL         = "localhost:4222"
	JOBS_STREAM_NAME = "jobs"
)

func main() {
	priority, id := os.Args[1], os.Args[2]
	consumerName := fmt.Sprintf("worker_%s", priority)
	workerName := fmt.Sprintf("worker_%s_%s", priority, id)

	nc, err := nats.Connect(NATS_URL, nats.Name(workerName))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	ctx := context.Background()
	consumer, err := js.CreateOrUpdateConsumer(ctx, JOBS_STREAM_NAME, jetstream.ConsumerConfig{
		Name:        consumerName,
		Durable:     consumerName,
		Description: fmt.Sprintf("WORKER %s", priority),
		MaxDeliver:  4,
		BackOff: []time.Duration{
			5 * time.Second,
			10 * time.Second,
			15 * time.Second,
		},
		FilterSubjects: []string{
			fmt.Sprintf("jobs.%s.*", priority),
		},
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	c, err := consumer.Consume(func(msg jetstream.Msg) {
		meta, err := msg.Metadata()
		if err != nil {
			slog.Error("Error getting metadata", "error", err)
			msg.Nak()
			return
		}

		if priority == "low" {
			slog.Info("Error occur when processing message!")
			return
		}

		slog.Info("Received message", "seq.id", meta.Sequence.Stream)
		time.Sleep(10 * time.Millisecond)
		msg.Ack()
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	<-quit

	c.Stop()
	slog.Info("Worker shutting down gracefully!")
}
