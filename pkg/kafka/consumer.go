package kafka

import (
	"context"
	"sync"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

// Consumer wraps a confluent-kafka consumer along with helpers for run loops
type Consumer struct {
	c       *ckafka.Consumer
	log     SugaredLogger
	perMsgT time.Duration
}

// NewConsumer creates a new consumer with basic config
func NewConsumer(bootstrapServers, groupID, autoOffsetReset string) (*Consumer, error) {
	if autoOffsetReset == "" {
		autoOffsetReset = "earliest"
	}
	c, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          groupID,
		"auto.offset.reset": autoOffsetReset,
	})
	if err != nil {
		return nil, err
	}
	return &Consumer{c: c}, nil
}

func (c *Consumer) WithLogger(l SugaredLogger) *Consumer { c.log = l; return c }

// WithHandleTimeout sets per-message handling timeout (0 means disabled)
func (c *Consumer) WithHandleTimeout(d time.Duration) *Consumer { c.perMsgT = d; return c }

func (c *Consumer) Close() error { return c.c.Close() }

// RunValueLoop runs a worker-pool loop that passes only message value to handler
func (c *Consumer) RunValueLoop(ctx context.Context, topics []string, handle func(context.Context, []byte) error) error {
	defer c.Close()
	if err := c.c.SubscribeTopics(topics, nil); err != nil {
		return err
	}
	const workerCount = 4
	const bufferSize = 128
	type item struct{ v []byte }
	ch := make(chan item, bufferSize)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case it, ok := <-ch:
					if !ok {
						return
					}
					hctx := ctx
					var cancel func()
					if c.perMsgT > 0 {
						hctx, cancel = context.WithTimeout(ctx, c.perMsgT)
					}
					if err := handle(hctx, it.v); err != nil && c.log != nil {
						c.log.Errorw("handler error", "error", err)
					}
					if cancel != nil {
						cancel()
					}
				}
			}
		}()
	}
	for {
		select {
		case <-ctx.Done():
			close(ch)
			wg.Wait()
			return nil
		default:
			msg, err := c.c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if kerr, ok := err.(ckafka.Error); !ok || kerr.Code() != ckafka.ErrTimedOut {
					if c.log != nil {
						c.log.Warnw("kafka read error", "error", err)
					}
				}
				continue
			}
			if msg == nil {
				continue
			}
			select {
			case ch <- item{v: msg.Value}:
			case <-ctx.Done():
				close(ch)
				wg.Wait()
				return nil
			default:
				if c.log != nil {
					c.log.Warnw("dropping message due to full buffer")
				}
			}
		}
	}
}

// RunMetaLoop runs a worker-pool loop that passes topic/partition/offset + value
func (c *Consumer) RunMetaLoop(ctx context.Context, topics []string, handle func(context.Context, string, int32, ckafka.Offset, []byte) error) error {
	defer c.Close()
	if err := c.c.SubscribeTopics(topics, nil); err != nil {
		return err
	}
	const workerCount = 4
	const bufferSize = 128
	type item struct {
		t string
		p int32
		o ckafka.Offset
		v []byte
	}
	ch := make(chan item, bufferSize)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case it, ok := <-ch:
					if !ok {
						return
					}
					hctx := ctx
					var cancel func()
					if c.perMsgT > 0 {
						hctx, cancel = context.WithTimeout(ctx, c.perMsgT)
					}
					if err := handle(hctx, it.t, it.p, it.o, it.v); err != nil && c.log != nil {
						c.log.Errorw("handler error", "topic", it.t, "partition", it.p, "offset", it.o, "error", err)
					}
					if cancel != nil {
						cancel()
					}
				}
			}
		}()
	}
	for {
		select {
		case <-ctx.Done():
			close(ch)
			wg.Wait()
			return nil
		default:
			msg, err := c.c.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if kerr, ok := err.(ckafka.Error); !ok || kerr.Code() != ckafka.ErrTimedOut {
					if c.log != nil {
						c.log.Warnw("kafka read error", "error", err)
					}
				}
				continue
			}
			if msg == nil {
				continue
			}
			t := ""
			if msg.TopicPartition.Topic != nil {
				t = *msg.TopicPartition.Topic
			}
			select {
			case ch <- item{t: t, p: msg.TopicPartition.Partition, o: msg.TopicPartition.Offset, v: msg.Value}:
			case <-ctx.Done():
				close(ch)
				wg.Wait()
				return nil
			default:
				if c.log != nil {
					c.log.Warnw("dropping message due to full buffer", "topic", t, "partition", msg.TopicPartition.Partition, "offset", msg.TopicPartition.Offset)
				}
			}
		}
	}
}
