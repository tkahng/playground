package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type EventBus interface {
	Publish(ctx context.Context, event any) error
}
type EventManager interface {
	EventBus() EventBus
	Run(ctx context.Context) error
	AddHandlers(handlers ...cqrs.EventHandler)

	AddHanlder(handler cqrs.EventHandler)
}

type InternalEventManager struct {
	pub message.Publisher
	sub message.Subscriber

	eventBus EventBus

	eventHandlers []cqrs.EventHandler

	eventRouter    *message.Router
	eventProcessor *cqrs.EventProcessor
}

// EventBus implements EventManager.
func (i *InternalEventManager) EventBus() EventBus {
	if i.eventBus == nil {
		panic("event bus is not initialized")
	}
	return i.eventBus
}

// AddHandlers implements EventManager.
func (i *InternalEventManager) AddHandlers(handlers ...cqrs.EventHandler) {
	i.eventHandlers = append(i.eventHandlers, handlers...)
}

func (i *InternalEventManager) AddHanlder(handler cqrs.EventHandler) {
	i.eventHandlers = append(i.eventHandlers, handler)
}

// Run implements EventManager.
func (i *InternalEventManager) Run(ctx context.Context) error {
	if len(i.eventHandlers) == 0 {
		return errors.New("no event handlers")
	}
	err := i.eventProcessor.AddHandlers(i.eventHandlers...)
	if err != nil {
		return err
	}

	return i.eventRouter.Run(ctx)
}

func NewEventManager(logger *slog.Logger) *InternalEventManager {
	loggerAdapter := watermill.NewSlogLogger(logger)
	pub, sub := newPubSub()
	eventBus, err := cqrs.NewEventBusWithConfig(pub, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) { return params.EventName, nil },
		Marshaler:            cqrs.JSONMarshaler{},
		Logger:               loggerAdapter,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create event bus: %w", err))
	}
	eventsRouter, err := message.NewRouter(message.RouterConfig{}, loggerAdapter)
	if err != nil {
		panic(fmt.Errorf("failed to create event router: %w", err))
	}
	eventsRouter.AddMiddleware(middleware.Recoverer)
	eventProcessor, err := cqrs.NewEventProcessorWithConfig(
		eventsRouter,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return params.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return sub, nil
			},
			Marshaler: cqrs.JSONMarshaler{},
			Logger:    loggerAdapter,
		})
	if err != nil {
		panic(fmt.Errorf("failed to create event processor: %w", err))
	}

	return &InternalEventManager{
		pub:            pub,
		sub:            sub,
		eventBus:       eventBus,
		eventProcessor: eventProcessor,
		eventRouter:    eventsRouter,
	}
}

var _ EventManager = (*InternalEventManager)(nil)

func newPubSub() (message.Publisher, message.Subscriber) {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 10000,
		},
		watermill.NopLogger{},
	)
	return pubSub, pubSub
}
