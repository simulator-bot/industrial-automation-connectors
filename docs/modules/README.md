# 模块设计

工程总体架构见 `../architecture/technical-design.zh-CN.md`。本目录维护模块级边界和逐步细化的实现设计。

当前骨架已建立：

- `internal/equipment`：统一设备、点位和函数模型；Process 侧存储接口。
- `internal/protocol`：由 Gateway 管理生命周期的协议适配器接口和注册表。
- `internal/contracts/v1`：Process、Gateway、事件与命令间的版本化契约。
- `internal/process`：管理/API 处理服务入口。
- `internal/gateway`：设备会话和协议适配器运行时入口。
- `internal/rules`、`cmd/rule-worker`：受限 DSL Worker 边界，后续落地 CEL 与动作图。
- `frontend`：Vue 3 运维 UI 骨架。

视觉源、摄像头和媒体处理不在本阶段实现。协议实现应按试点顺序逐个加入 `internal/protocol/adapters/<protocol>`，并单独维护设备兼容记录。
