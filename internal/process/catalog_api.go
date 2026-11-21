package process

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	v1 "github.com/simulator-bot/industrial-automation-connectors/internal/contracts/v1"
	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/model"
	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/repository"
)

type catalogAPI struct {
	store      *repository.Catalog
	gatewayURL string
	client     *http.Client
}

func newCatalogAPI(store *repository.Catalog, gatewayURL string) (*catalogAPI, error) {
	api := &catalogAPI{store: store}
	if gatewayURL == "" {
		return api, nil
	}
	endpoint, err := url.Parse(gatewayURL)
	if err != nil || endpoint.Scheme != "http" || !net.ParseIP(endpoint.Hostname()).IsLoopback() ||
		endpoint.User != nil || (endpoint.Path != "" && endpoint.Path != "/") || endpoint.RawQuery != "" || endpoint.Fragment != "" || endpoint.ForceQuery {
		return nil, fmt.Errorf("PROCESS_GATEWAY_URL must be a loopback HTTP origin, such as http://127.0.0.1:8081")
	}
	endpoint.Path = "/api/v1/gateway/opcua/test"
	api.gatewayURL = endpoint.String()
	api.client = &http.Client{
		Timeout:       25 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	return api, nil
}

func (a *catalogAPI) get(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, a.store.Snapshot())
}

func (a *catalogAPI) put(w http.ResponseWriter, r *http.Request) {
	var data model.EquipmentCatalog
	if !readJSON(w, r, &data) {
		return
	}
	if data.Models == nil || data.Points == nil || data.Plans == nil || data.Instances == nil {
		http.Error(w, "保存需要包含模型、点位、计划和实例的完整设备配置", http.StatusBadRequest)
		return
	}
	if err := validateCatalog(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	saved, err := a.store.Replace(data)
	if errors.Is(err, repository.ErrCatalogConflict) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "保存设备配置失败", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

type testInput struct {
	Revision  uint64             `json:"revision"`
	MethodID  string             `json:"methodId,omitempty"`
	Arguments []v1.OPCUAArgument `json:"arguments,omitempty"`
}

func (a *catalogAPI) test(w http.ResponseWriter, r *http.Request) {
	var input testInput
	if !readJSON(w, r, &input) {
		return
	}
	data := a.store.Snapshot()
	if input.Revision != data.Revision {
		http.Error(w, repository.ErrCatalogConflict.Error(), http.StatusConflict)
		return
	}
	var instance *model.DeviceInstance
	for i := range data.Instances {
		if data.Instances[i].ID == r.PathValue("id") {
			instance = &data.Instances[i]
			break
		}
	}
	if instance == nil {
		http.Error(w, "设备实例不存在", http.StatusNotFound)
		return
	}
	if a.client == nil {
		http.Error(w, "请为 Process 配置 PROCESS_GATEWAY_URL 并启动 Gateway", http.StatusServiceUnavailable)
		return
	}
	job, err := buildTestRequest(data, *instance, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload, err := json.Marshal(job)
	if err != nil {
		http.Error(w, "无法生成 Gateway 测试请求", http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, a.gatewayURL, bytes.NewReader(payload))
	if err != nil {
		http.Error(w, "Gateway 测试地址无效", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		http.Error(w, "Gateway 不可用，请确认服务已启动："+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		http.Error(w, fmt.Sprintf("Gateway 返回 %d：%s", resp.StatusCode, strings.TrimSpace(string(message))), http.StatusBadGateway)
		return
	}
	var result v1.OPCUATestResult
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&result); err != nil {
		http.Error(w, "Gateway 测试结果无效", http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func buildTestRequest(data model.EquipmentCatalog, instance model.DeviceInstance, input testInput) (v1.OPCUATestRequest, error) {
	job := v1.OPCUATestRequest{Endpoint: instance.Endpoint, DeviceID: instance.ID}
	if input.MethodID != "" {
		for _, item := range data.Models {
			if item.ID != instance.ModelID {
				continue
			}
			for _, method := range item.Methods {
				if method.ID != input.MethodID {
					continue
				}
				if len(input.Arguments) != len(method.Parameters) {
					return job, fmt.Errorf("方法「%s」需要 %d 个参数", method.Name, len(method.Parameters))
				}
				for i, argument := range input.Arguments {
					if argument.Type != method.Parameters[i].Type {
						return job, fmt.Errorf("参数「%s」需要 %s 类型", method.Parameters[i].Name, method.Parameters[i].Type)
					}
				}
				job.Method = &v1.OPCUAMethodTest{
					OPCUAMethod: v1.OPCUAMethod{ObjectNodeID: method.ObjectNodeID, MethodNodeID: method.MethodNodeID},
					Arguments:   input.Arguments,
				}
				return job, nil
			}
		}
		return job, fmt.Errorf("该设备实例的模型未配置此方法")
	}
	if len(input.Arguments) != 0 {
		return job, fmt.Errorf("请选择要测试的方法")
	}
	for _, plan := range data.Plans {
		if plan.ID != instance.PlanID || plan.ModelID != instance.ModelID {
			continue
		}
		job.IntervalMillis, job.Deadband = plan.IntervalMillis, plan.Deadband
		selected := make(map[string]model.ModelPoint, len(data.Points))
		for _, point := range data.Points {
			if point.ModelID == plan.ModelID {
				selected[point.ID] = point
			}
		}
		for _, id := range plan.PointIDs {
			point, exists := selected[id]
			if !exists {
				return job, fmt.Errorf("采集计划引用的点位不存在")
			}
			job.Points = append(job.Points, v1.OPCUAPoint{ID: point.ID, NodeID: point.NodeID, DataType: string(point.DataType)})
		}
		return job, nil
	}
	return job, fmt.Errorf("设备实例的采集计划不存在")
}

func readJSON(w http.ResponseWriter, r *http.Request, dest any) bool {
	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if contentType != "application/json" {
		http.Error(w, "请求需要 application/json", http.StatusUnsupportedMediaType)
		return false
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		http.Error(w, "JSON 配置无效："+err.Error(), http.StatusBadRequest)
		return false
	}
	if decoder.Decode(new(json.RawMessage)) != io.EOF {
		http.Error(w, "请求需要一个 JSON 文档", http.StatusBadRequest)
		return false
	}
	return true
}
