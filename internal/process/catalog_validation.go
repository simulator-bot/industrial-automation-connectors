package process

import (
	"fmt"
	"strings"

	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/model"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol/adapters/opcua"
)

func validateCatalog(data model.EquipmentCatalog) error {
	models := make(map[string]bool, len(data.Models))
	for _, item := range data.Models {
		if err := register(models, item.ID, item.Name, "设备模型"); err != nil {
			return err
		}
		if item.Protocol != opcua.Name {
			return fmt.Errorf("模型「%s」目前只支持 OPC UA", item.Name)
		}
		if err := validateMethods(item); err != nil {
			return err
		}
	}

	points := make(map[string]model.ModelPoint, len(data.Points))
	nodes := make(map[string]map[string]bool)
	pointIDs := make(map[string]bool)
	for _, item := range data.Points {
		if err := register(pointIDs, item.ID, item.Name, "点位"); err != nil {
			return err
		}
		if !models[item.ModelID] {
			return fmt.Errorf("点位「%s」引用的模型不存在", item.Name)
		}
		node, err := opcua.ParseNodeID(item.NodeID)
		if err != nil {
			return fmt.Errorf("点位「%s」的 NodeID 无效：%w", item.Name, err)
		}
		if !validDataType(item.DataType) {
			return fmt.Errorf("点位「%s」的数据类型无效", item.Name)
		}
		if nodes[item.ModelID] == nil {
			nodes[item.ModelID] = make(map[string]bool)
		}
		if nodes[item.ModelID][node.String()] {
			return fmt.Errorf("模型内的点位 NodeID 重复：%s", node)
		}
		nodes[item.ModelID][node.String()] = true
		points[item.ID] = item
	}

	plans := make(map[string]model.CollectionPlan, len(data.Plans))
	planIDs := make(map[string]bool)
	for _, item := range data.Plans {
		if err := register(planIDs, item.ID, item.Name, "采集计划"); err != nil {
			return err
		}
		if !models[item.ModelID] {
			return fmt.Errorf("采集计划「%s」引用的模型不存在", item.Name)
		}
		if err := validatePlan(item, points); err != nil {
			return err
		}
		plans[item.ID] = item
	}

	instanceIDs := make(map[string]bool)
	for _, item := range data.Instances {
		if err := register(instanceIDs, item.ID, item.Name, "设备实例"); err != nil {
			return err
		}
		plan, exists := plans[item.PlanID]
		if !models[item.ModelID] || !exists || plan.ModelID != item.ModelID {
			return fmt.Errorf("设备实例「%s」需要选择与本实例模型一致的采集计划", item.Name)
		}
		if _, err := opcua.SimulatorEndpoint(item.Endpoint); err != nil {
			return fmt.Errorf("设备实例「%s」的端点无效：%w", item.Name, err)
		}
	}
	return nil
}

func register(seen map[string]bool, id, name, label string) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" || strings.Contains(id, "/") {
		return fmt.Errorf("%s的 ID 和名称不能为空，ID 不可包含 /", label)
	}
	if seen[id] {
		return fmt.Errorf("%s ID 重复：%s", label, id)
	}
	seen[id] = true
	return nil
}

func validateMethods(item model.DeviceModel) error {
	ids := make(map[string]bool)
	for _, method := range item.Methods {
		if err := register(ids, method.ID, method.Name, "模型方法"); err != nil {
			return err
		}
		for _, node := range []string{method.ObjectNodeID, method.MethodNodeID} {
			if _, err := opcua.ParseNodeID(node); err != nil {
				return fmt.Errorf("方法「%s」的 NodeID 无效：%w", method.Name, err)
			}
		}
		for _, param := range method.Parameters {
			if strings.TrimSpace(param.Name) == "" || !opcua.SupportsArgumentType(param.Type) {
				return fmt.Errorf("方法「%s」的参数名称或类型无效", method.Name)
			}
		}
	}
	return nil
}

func validatePlan(plan model.CollectionPlan, points map[string]model.ModelPoint) error {
	if plan.Mode != model.SubscriptionOnChange || plan.IntervalMillis < 1 || plan.IntervalMillis > 3600000 || plan.Deadband < 0 {
		return fmt.Errorf("采集计划「%s」需要数据变化订阅模式、1–3600000 ms 发布周期及非负死区", plan.Name)
	}
	if len(plan.PointIDs) == 0 || len(plan.PointIDs) > 200 {
		return fmt.Errorf("采集计划「%s」需要选择 1–200 个点位", plan.Name)
	}
	seen := make(map[string]bool, len(plan.PointIDs))
	for _, id := range plan.PointIDs {
		point, exists := points[id]
		if !exists || point.ModelID != plan.ModelID || seen[id] {
			return fmt.Errorf("采集计划「%s」的点位必须存在、属于该模型且不得重复", plan.Name)
		}
		if plan.Deadband > 0 && point.DataType != model.DataTypeInt && point.DataType != model.DataTypeUint && point.DataType != model.DataTypeFloat {
			return fmt.Errorf("采集计划「%s」的绝对死区需要数值点位", plan.Name)
		}
		seen[id] = true
	}
	return nil
}

func validDataType(kind model.DataType) bool {
	switch kind {
	case model.DataTypeBool, model.DataTypeInt, model.DataTypeUint, model.DataTypeFloat, model.DataTypeString,
		model.DataTypeBytes, model.DataTypeEnum, model.DataTypeTime, model.DataTypeStruct:
		return true
	default:
		return false
	}
}
