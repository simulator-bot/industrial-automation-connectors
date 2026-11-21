package gateway

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"time"

	v1 "github.com/simulator-bot/industrial-automation-connectors/internal/contracts/v1"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol"
)

const observationWindow = 4 * time.Second

type opcuaTests struct {
	gatewayID string
	adapter   protocol.Adapter
	slots     chan struct{}
}

func (t opcuaTests) handle(w http.ResponseWriter, r *http.Request) {
	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	var req v1.OPCUATestRequest
	if err := decodeJSON(http.MaxBytesReader(w, r.Body, 256<<10), &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := opcuaConfig{
		Endpoint: req.Endpoint, ConnectionID: req.DeviceID, DeviceID: req.DeviceID,
		IntervalMillis: req.IntervalMillis, Deadband: req.Deadband, Points: req.Points,
	}
	if req.Method != nil {
		cfg.Methods = map[string]v1.OPCUAMethod{"test": req.Method.OPCUAMethod}
	}
	pilot, err := newOPCUAPilot(t.gatewayID, cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	select {
	case t.slots <- struct{}{}:
		defer func() { <-t.slots }()
	default:
		http.Error(w, "Gateway 正在执行其他设备测试，请稍后重试", http.StatusTooManyRequests)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	result := v1.OPCUATestResult{}
	if err := pilot.connect(ctx, t.adapter); err != nil {
		result.Error = fmt.Sprintf("连接失败：%v", err)
		writeJSON(w, result)
		return
	}
	defer pilot.session.Close()
	result.Connected = true
	if req.Method != nil {
		var err error
		result.MethodResult, err = pilot.callMethod(ctx, req.Method.OPCUAMethod, req.Method.Arguments)
		if err != nil {
			result.Error = fmt.Sprintf("方法调用失败：%v", err)
		}
	} else {
		pilot.testPoints(ctx, &result)
	}
	writeJSON(w, result)
}

func (p *opcuaPilot) testPoints(ctx context.Context, result *v1.OPCUATestResult) {
	values, err := p.session.Read(ctx, p.targets)
	if err != nil {
		result.Error = fmt.Sprintf("读取点位失败：%v", err)
		return
	}
	result.ReadCompleted = true
	result.Points = make([]v1.OPCUATestPoint, 0, len(values))
	index := make(map[string]int, len(values))
	for _, value := range values {
		event := p.pointEvent(value)
		index[value.TargetID] = len(result.Points)
		result.Points = append(result.Points, v1.OPCUATestPoint{PointID: event.PointID, Read: &event})
	}
	subscription, cancel := context.WithCancel(ctx)
	defer cancel()
	updates, err := p.subscribe(subscription)
	if err != nil {
		result.Error = fmt.Sprintf("创建订阅失败：%v", err)
		return
	}
	result.SubscriptionOpened = true
	timer := time.NewTimer(observationWindow)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			result.Error = fmt.Sprintf("订阅测试结束：%v", ctx.Err())
			return
		case <-timer.C:
			return
		case update, ok := <-updates:
			if !ok {
				result.Error = "订阅已断开"
				return
			}
			if update.Err != nil {
				result.Error = update.Err.Error()
				return
			}
			point := &result.Points[index[update.Value.TargetID]]
			event := p.pointEvent(update.Value)
			point.Latest = &event
			point.Notifications++
		}
	}
}
