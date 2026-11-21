package opcua

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/model"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol"
)

func (s *session) Subscribe(ctx context.Context, spec protocol.Subscription) (<-chan protocol.Update, error) {
	if spec.Mode != "" && spec.Mode != string(model.SubscriptionOnChange) {
		return nil, fmt.Errorf("OPC UA supports ON_CHANGE subscriptions")
	}
	if spec.Interval != 0 && spec.Interval < time.Millisecond {
		return nil, fmt.Errorf("OPC UA publishing interval must be at least 1 ms")
	}
	if spec.Deadband < 0 || math.IsNaN(spec.Deadband) || math.IsInf(spec.Deadband, 0) {
		return nil, fmt.Errorf("OPC UA absolute deadband must be finite and nonnegative")
	}
	nodes, err := parseNodes(spec.Targets)
	if err != nil {
		return nil, err
	}
	if err := s.Health(ctx); err != nil {
		return nil, err
	}

	setup, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	notices := make(chan *opcua.PublishNotificationData, 64)
	sub, err := s.client.Subscribe(setup, &opcua.SubscriptionParameters{
		Interval: spec.Interval, MaxKeepAliveCount: 10, LifetimeCount: 30,
	}, notices)
	if err != nil {
		return nil, fmt.Errorf("create OPC UA subscription: %w", err)
	}

	items := make([]*ua.MonitoredItemCreateRequest, len(nodes))
	for i, node := range nodes {
		item := opcua.NewMonitoredItemCreateRequestWithDefaults(node, ua.AttributeIDValue, uint32(i+1))
		if spec.Deadband > 0 {
			item.RequestedParameters.Filter = ua.NewExtensionObject(&ua.DataChangeFilter{
				Trigger: ua.DataChangeTriggerStatusValue, DeadbandType: uint32(ua.DeadbandTypeAbsolute), DeadbandValue: spec.Deadband,
			})
		}
		items[i] = item
	}
	res, err := sub.Monitor(setup, ua.TimestampsToReturnSource, items...)
	if err != nil {
		cancelSubscription(sub, notices)
		return nil, fmt.Errorf("monitor OPC UA points: %w", err)
	}
	for i, result := range res.Results {
		if result.StatusCode != ua.StatusOK {
			cancelSubscription(sub, notices)
			return nil, fmt.Errorf("monitor OPC UA NodeID %q: %w", spec.Targets[i], result.StatusCode)
		}
	}

	// Reserve one slot for a terminal error when the consumer falls behind.
	updates := make(chan protocol.Update, 65)
	go s.forwardUpdates(ctx, sub, notices, updates, append([]string(nil), spec.Targets...))
	return updates, nil
}

func cancelSubscription(sub *opcua.Subscription, notices <-chan *opcua.PublishNotificationData) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	_ = sub.Cancel(ctx)
	// Release any SDK publisher that was sending while we removed the subscription.
	for {
		select {
		case <-notices:
		default:
			return
		}
	}
}

func (s *session) forwardUpdates(
	ctx context.Context,
	sub *opcua.Subscription,
	notices <-chan *opcua.PublishNotificationData,
	updates chan<- protocol.Update,
	targets []string,
) {
	defer close(updates)
	defer cancelSubscription(sub, notices)
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case <-s.disconnected:
			queueUpdate(updates, protocol.Update{Err: errors.New("OPC UA session disconnected")})
			return
		case notice := <-notices:
			if notice.Error != nil {
				queueUpdate(updates, protocol.Update{Err: fmt.Errorf("OPC UA subscription: %w", notice.Error)})
				return
			}
			switch data := notice.Value.(type) {
			case *ua.DataChangeNotification:
				receivedAt := time.Now().UTC()
				for _, item := range data.MonitoredItems {
					if item.ClientHandle == 0 || int(item.ClientHandle) > len(targets) {
						continue
					}
					value, err := decodeValue(targets[item.ClientHandle-1], item.Value, receivedAt)
					if !queueUpdate(updates, protocol.Update{Value: value, Err: err}) {
						return
					}
				}
			case *ua.StatusChangeNotification:
				if data.Status != ua.StatusOK {
					queueUpdate(updates, protocol.Update{Err: fmt.Errorf("OPC UA subscription status: %w", data.Status)})
					return
				}
			}
		}
	}
}

func queueUpdate(updates chan<- protocol.Update, update protocol.Update) bool {
	if update.Err == nil && len(updates) == cap(updates)-1 {
		update = protocol.Update{Err: errors.New("OPC UA subscription consumer fell behind")}
	}
	updates <- update
	return update.Err == nil
}
