package types

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cometbft/cometbft/v2/abci/types"
	"github.com/cometbft/cometbft/v2/libs/log"
	cmtpubsub "github.com/cometbft/cometbft/v2/libs/pubsub"
	"github.com/cometbft/cometbft/v2/libs/service"
)

const defaultCapacity = 0

type EventBusSubscriber interface {
	Subscribe(ctx context.Context, subscriber string, query cmtpubsub.Query, outCapacity ...int) (Subscription, error)
	Unsubscribe(ctx context.Context, subscriber string, query cmtpubsub.Query) error
	UnsubscribeAll(ctx context.Context, subscriber string) error

	NumClients() int
	NumClientSubscriptions(clientID string) int
}

type Subscription interface {
	Out() <-chan cmtpubsub.Message
	Canceled() <-chan struct{}
	Err() error
}

// EventBus is a common bus for all events going through the system. All calls
// are proxied to underlying pubsub server. All events must be published using
// EventBus to ensure correct data types.
type EventBus struct {
	service.BaseService
	pubsub *cmtpubsub.Server
}

// NewEventBus returns a new event bus.
func NewEventBus() *EventBus {
	return NewEventBusWithBufferCapacity(defaultCapacity)
}

// NewEventBusWithBufferCapacity returns a new event bus with the given buffer capacity.
func NewEventBusWithBufferCapacity(cap int) *EventBus {
	// capacity could be exposed later if needed
	pubsub := cmtpubsub.NewServer(cmtpubsub.BufferCapacity(cap))
	b := &EventBus{pubsub: pubsub}
	b.BaseService = *service.NewBaseService(nil, "EventBus", b)
	return b
}

func (b *EventBus) SetLogger(l log.Logger) {
	b.BaseService.SetLogger(l)
	b.pubsub.SetLogger(l.With("module", "pubsub"))
}

func (b *EventBus) OnStart() error {
	return b.pubsub.Start()
}

func (b *EventBus) OnStop() {
	if err := b.pubsub.Stop(); err != nil {
		b.pubsub.Logger.Error("error trying to stop eventBus", "error", err)
	}
}

func (b *EventBus) NumClients() int {
	return b.pubsub.NumClients()
}

func (b *EventBus) NumClientSubscriptions(clientID string) int {
	return b.pubsub.NumClientSubscriptions(clientID)
}

func (b *EventBus) Subscribe(
	ctx context.Context,
	subscriber string,
	query cmtpubsub.Query,
	outCapacity ...int,
) (Subscription, error) {
	return b.pubsub.Subscribe(ctx, subscriber, query, outCapacity...)
}

// SubscribeUnbuffered can be used for a local consensus explorer and synchronous
// testing. Do not use for public facing / untrusted subscriptions!
func (b *EventBus) SubscribeUnbuffered(
	ctx context.Context,
	subscriber string,
	query cmtpubsub.Query,
) (Subscription, error) {
	return b.pubsub.SubscribeUnbuffered(ctx, subscriber, query)
}

func (b *EventBus) Unsubscribe(ctx context.Context, subscriber string, query cmtpubsub.Query) error {
	return b.pubsub.Unsubscribe(ctx, subscriber, query)
}

func (b *EventBus) UnsubscribeAll(ctx context.Context, subscriber string) error {
	return b.pubsub.UnsubscribeAll(ctx, subscriber)
}

func (b *EventBus) Publish(eventType string, eventData TMEventData) error {
	// no explicit deadline for publishing events
	ctx := context.Background()
	return b.pubsub.PublishWithEvents(ctx, eventData, map[string][]string{EventTypeKey: {eventType}})
}

// validateAndStringifyEvents takes a slice of event objects and creates a
// map of stringified events where each key is composed of the event
// type and each of the event's attributes keys in the form of
// "{event.Type}.{attribute.Key}" and the value is each attribute's value.
func (*EventBus) validateAndStringifyEvents(events []types.Event) map[string][]string {
	result := make(map[string][]string)
	for _, event := range events {
		if len(event.Type) == 0 {
			continue
		}
		prefix := event.Type + "."
		for _, attr := range event.Attributes {
			if len(attr.Key) == 0 {
				continue
			}

			compositeTag := prefix + attr.Key
			result[compositeTag] = append(result[compositeTag], attr.Value)
		}
	}

	return result
}

func (b *EventBus) PublishEventNewBlock(data EventDataNewBlock) error {
	// no explicit deadline for publishing events
	ctx := context.Background()
	events := b.validateAndStringifyEvents(data.ResultFinalizeBlock.Events)

	// add predefined new block event
	events[EventTypeKey] = append(events[EventTypeKey], EventNewBlock)

	return b.pubsub.PublishWithEvents(ctx, data, events)
}

func (b *EventBus) PublishEventNewBlockEvents(data EventDataNewBlockEvents) error {
	// no explicit deadline for publishing events
	ctx := context.Background()

	events := b.validateAndStringifyEvents(data.Events)

	// add predefined new block event
	events[EventTypeKey] = append(events[EventTypeKey], EventNewBlockEvents)

	return b.pubsub.PublishWithEvents(ctx, data, events)
}

func (b *EventBus) PublishEventNewBlockHeader(data EventDataNewBlockHeader) error {
	return b.Publish(EventNewBlockHeader, data)
}

func (b *EventBus) PublishEventNewEvidence(evidence EventDataNewEvidence) error {
	return b.Publish(EventNewEvidence, evidence)
}

func (b *EventBus) PublishEventVote(data EventDataVote) error {
	return b.Publish(EventVote, data)
}

func (b *EventBus) PublishEventValidBlock(data EventDataRoundState) error {
	return b.Publish(EventValidBlock, data)
}

func (b *EventBus) PublishEventPendingTx(data EventDataPendingTx) error {
	// no explicit deadline for publishing events
	ctx := context.Background()
	return b.pubsub.PublishWithEvents(ctx, data, map[string][]string{
		EventTypeKey: {EventPendingTx},
		TxHashKey:    {fmt.Sprintf("%X", Tx(data.Tx).Hash())},
	})
}

// PublishEventTx publishes tx event with events from Result. Note it will add
// predefined keys (EventTypeKey, TxHashKey). Existing events with the same keys
// will be overwritten.
func (b *EventBus) PublishEventTx(data EventDataTx) error {
	// no explicit deadline for publishing events
	ctx := context.Background()

	events := b.validateAndStringifyEvents(data.Result.Events)

	// add predefined compositeKeys
	events[EventTypeKey] = append(events[EventTypeKey], EventTx)
	events[TxHashKey] = append(events[TxHashKey], fmt.Sprintf("%X", Tx(data.Tx).Hash()))
	events[TxHeightKey] = append(events[TxHeightKey], strconv.FormatInt(data.Height, 10))

	return b.pubsub.PublishWithEvents(ctx, data, events)
}

func (b *EventBus) PublishEventNewRoundStep(data EventDataRoundState) error {
	return b.Publish(EventNewRoundStep, data)
}

func (b *EventBus) PublishEventTimeoutPropose(data EventDataRoundState) error {
	return b.Publish(EventTimeoutPropose, data)
}

func (b *EventBus) PublishEventTimeoutWait(data EventDataRoundState) error {
	return b.Publish(EventTimeoutWait, data)
}

func (b *EventBus) PublishEventNewRound(data EventDataNewRound) error {
	return b.Publish(EventNewRound, data)
}

func (b *EventBus) PublishEventCompleteProposal(data EventDataCompleteProposal) error {
	return b.Publish(EventCompleteProposal, data)
}

func (b *EventBus) PublishEventPolka(data EventDataRoundState) error {
	return b.Publish(EventPolka, data)
}

func (b *EventBus) PublishEventRelock(data EventDataRoundState) error {
	return b.Publish(EventRelock, data)
}

func (b *EventBus) PublishEventLock(data EventDataRoundState) error {
	return b.Publish(EventLock, data)
}

func (b *EventBus) PublishEventValidatorSetUpdates(data EventDataValidatorSetUpdates) error {
	return b.Publish(EventValidatorSetUpdates, data)
}

// -----------------------------------------------------------------------------.
type NopEventBus struct{}

func (NopEventBus) Subscribe(
	context.Context,
	string,
	cmtpubsub.Query,
	chan<- any,
) error {
	return nil
}

func (NopEventBus) Unsubscribe(context.Context, string, cmtpubsub.Query) error {
	return nil
}

func (NopEventBus) UnsubscribeAll(context.Context, string) error {
	return nil
}

func (NopEventBus) PublishEventNewBlock(EventDataNewBlock) error {
	return nil
}

func (NopEventBus) PublishEventNewBlockHeader(EventDataNewBlockHeader) error {
	return nil
}

func (NopEventBus) PublishEventNewBlockEvents(EventDataNewBlockEvents) error {
	return nil
}

func (NopEventBus) PublishEventNewEvidence(EventDataNewEvidence) error {
	return nil
}

func (NopEventBus) PublishEventVote(EventDataVote) error {
	return nil
}

func (NopEventBus) PublishEventPendingTx(EventDataPendingTx) error {
	return nil
}

func (NopEventBus) PublishEventTx(EventDataTx) error {
	return nil
}

func (NopEventBus) PublishEventNewRoundStep(EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventTimeoutPropose(EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventTimeoutWait(EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventNewRound(EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventCompleteProposal(EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventPolka(EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventRelock(EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventLock(EventDataRoundState) error {
	return nil
}

func (NopEventBus) PublishEventValidatorSetUpdates(EventDataValidatorSetUpdates) error {
	return nil
}
