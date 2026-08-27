// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen_test

import (
	"context"
	"fmt"
	"log"

	"github.com/lin-snow/ech0/pkg/busen"
)

func Example() {
	type UserCreated struct {
		Email string
	}

	b := busen.New()

	unsubscribe, err := b.Subscribe(func(_ context.Context, event busen.Event[UserCreated]) error {
		fmt.Println("welcome", event.Value.Email)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(context.Background(), UserCreated{Email: "hello@example.com"}); err != nil {
		log.Fatal(err)
	}

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleBus_SubscribeTopic() {
	b := busen.New()

	unsubscribe, err := b.SubscribeTopic("orders.>", func(_ context.Context, event busen.Event[string]) error {
		fmt.Printf("%s=%s\n", event.Topic, event.Value)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(context.Background(), "created", busen.WithTopic("orders.eu.created")); err != nil {
		log.Fatal(err)
	}

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleBus_SubscribeTopics() {
	b := busen.New()

	unsubscribe, err := b.SubscribeTopics([]string{"orders.created", "orders.updated"}, func(_ context.Context, event busen.Event[string]) error {
		fmt.Printf("%s=%s\n", event.Topic, event.Value)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(context.Background(), "created", busen.WithTopic("orders.created")); err != nil {
		log.Fatal(err)
	}
	if err := b.Publish(context.Background(), "updated", busen.WithTopic("orders.updated")); err != nil {
		log.Fatal(err)
	}

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleAsync() {
	type JobQueued struct {
		ID string
	}

	b := busen.New()
	done := make(chan struct{})

	unsubscribe, err := b.Subscribe(func(_ context.Context, event busen.Event[JobQueued]) error {
		fmt.Println("processed", event.Value.ID)
		close(done)
		return nil
	}, busen.Async(), busen.WithBuffer(1))
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(context.Background(), JobQueued{ID: "job-42"}); err != nil {
		log.Fatal(err)
	}

	<-done

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleWithKey() {
	type UserCreated struct {
		ID string
	}

	b := busen.New()
	done := make(chan struct{}, 2)

	unsubscribe, err := b.Subscribe(func(_ context.Context, event busen.Event[UserCreated]) error {
		fmt.Printf("%s:%s\n", event.Key, event.Value.ID)
		done <- struct{}{}
		return nil
	}, busen.Async(), busen.WithParallelism(2), busen.WithBuffer(4))
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(context.Background(), UserCreated{ID: "1"}, busen.WithKey("tenant-a")); err != nil {
		log.Fatal(err)
	}
	if err := b.Publish(context.Background(), UserCreated{ID: "2"}, busen.WithKey("tenant-a")); err != nil {
		log.Fatal(err)
	}

	<-done
	<-done

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleBus_Use() {
	type AuditEvent struct {
		Action string
	}

	b := busen.New()
	if err := b.Use(func(next busen.Next) busen.Next {
		return func(ctx context.Context, dispatch busen.Dispatch) error {
			if dispatch.Headers == nil {
				dispatch.Headers = make(map[string]string, 1)
			}
			dispatch.Headers["source"] = "middleware"
			return next(ctx, dispatch)
		}
	}); err != nil {
		log.Fatal(err)
	}

	unsubscribe, err := b.Subscribe(func(_ context.Context, event busen.Event[AuditEvent]) error {
		fmt.Printf("%s from %s\n", event.Value.Action, event.Headers["source"])
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(
		context.Background(),
		AuditEvent{Action: "saved"},
		busen.WithHeaders(map[string]string{"request-id": "req-1"}),
	); err != nil {
		log.Fatal(err)
	}

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleWithHooks() {
	type UserCreated struct {
		ID string
	}

	b := busen.New(busen.WithHooks(busen.Hooks{
		OnPublishDone: func(info busen.PublishDone) {
			fmt.Printf("matched=%d delivered=%d\n", info.MatchedSubscribers, info.DeliveredSubscribers)
		},
	}))

	unsubscribe, err := b.Subscribe(func(_ context.Context, event busen.Event[UserCreated]) error {
		fmt.Println("handled", event.Value.ID)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(context.Background(), UserCreated{ID: "u-1"}); err != nil {
		log.Fatal(err)
	}

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleWithMetadataBuilder() {
	type OrderCreated struct {
		ID string
	}

	b := busen.New(
		busen.WithMetadataBuilder(func(input busen.PublishMetadataInput) map[string]string {
			return map[string]string{
				"source": "billing",
			}
		}),
	)

	unsubscribe, err := b.Subscribe(func(_ context.Context, event busen.Event[OrderCreated]) error {
		fmt.Printf("%s from %s\n", event.Value.ID, event.Meta["source"])
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(
		context.Background(),
		OrderCreated{ID: "o-1"},
		busen.WithMetadata(map[string]string{"trace_id": "tr-1"}),
	); err != nil {
		log.Fatal(err)
	}

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleBus_UseObserver() {
	type OrderCreated struct {
		ID string
	}

	b := busen.New()
	if err := b.UseObserver(
		func(_ context.Context, obs busen.Observation) {
			fmt.Printf("observe %s %v\n", obs.Topic, obs.EventType)
		},
		busen.ObserveTopic("orders.>"),
	); err != nil {
		log.Fatal(err)
	}

	unsubscribe, err := b.SubscribeTopic("orders.>", func(_ context.Context, event busen.Event[OrderCreated]) error {
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	if err := b.Publish(context.Background(), OrderCreated{ID: "o-1"}, busen.WithTopic("orders.created")); err != nil {
		log.Fatal(err)
	}

	if err := b.Close(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func ExampleBus_Shutdown() {
	type Job struct {
		ID int
	}

	b := busen.New()
	unsubscribe, err := b.Subscribe(func(_ context.Context, _ busen.Event[Job]) error {
		return nil
	}, busen.Async(), busen.WithBuffer(8))
	if err != nil {
		log.Fatal(err)
	}
	defer unsubscribe()

	_ = b.Publish(context.Background(), Job{ID: 1})

	result, err := b.Shutdown(context.Background(), busen.ShutdownDrain)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Mode == busen.ShutdownDrain, result.Completed)
}
