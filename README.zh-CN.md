# 工业自动化连接器

[English](README.md) | **简体中文**

工业自动化连接器负责接入 PLC、半导体设备、机器人、仪器、工厂摄像头和其他自动化系统，并向上层提供统一的设备、点位、视觉信号、事件和操作能力。原始视频流通过独立媒体通道接入，视觉算法由连接器外的视觉服务承载。

## 设计文档

- [技术方案与总体设计](docs/architecture/technical-design.zh-CN.md)：Gateway、Process、Frontend 分层，统一设备能力模型、协议与视觉源适配、受限规则 DSL、Linux/Kubernetes 部署及实施阶段。
- 后续模块设计放在 `docs/modules/`，例如 Gateway、协议适配、设备模型、规则运行时和前端。

整体系统架构和跨工程职责边界由 `design-docs` 工程维护；本工程维护连接器自身的技术方案、模块接口和实现设计。

## 当前实现

本仓库已初始化 Go 模块和三类运行入口：

- `cmd/gateway`：设备协议会话和适配器运行时入口。
- `cmd/process`：设备模型管理 API 与业务处理入口。
- `cmd/rule-worker`：与设备网络隔离的规则执行进程边界（当前为空闲骨架）。
- `frontend`：Vue 3 + TypeScript + Vite 运维界面骨架。
- `internal/equipment`、`internal/protocol`、`internal/contracts/v1`：统一能力模型、适配器 SPI 和事件/命令契约。
- `deployments/kubernetes/base`：Linux 容器及 Kubernetes 初始部署模板。

Gateway 已支持连接本地 OPC UA 模拟器、推送原生监视项通知和调用已配置的方法。Process 仍使用内存设备仓储，仅提供健康检查和空设备列表。规则 Worker 尚未加载 CEL。

### 本地运行

```sh
make run-process
# 新终端中：
make run-gateway
```

Process 默认监听 `:8080`，Gateway 默认监听 `:8081`。前端开发环境：

```sh
cd frontend
npm install
npm run dev
```

生产部署模板见 `deployments/kubernetes/README.md`。模块边界见 `docs/modules/README.md`。

### 本地 OPC UA 模拟器

[示例配置](deployments/examples/opcua-simulator.json) 对应监听 4840 端口的长晶棒检测模拟器。接入其他模拟器时，按实际节点修改端点、点位 NodeID 和方法别名。Gateway 和模拟器通过回环地址连接，共用一个匿名 OPC UA 会话，安全策略为 `None`。端点里的 `localhost` 按 IPv4 连接。仅监听 IPv6 时请填 `[::1]`。

```sh
GATEWAY_HTTP_ADDR=127.0.0.1:8081 \
GATEWAY_OPCUA_CONFIG=deployments/examples/opcua-simulator.json make run-gateway

# 另开终端，接收初始值以及后续值或质量变化。
curl -N http://127.0.0.1:8081/api/v1/gateway/opcua/events

# 调用模拟器的 SetString 写入测试文本。
curl -fsS http://127.0.0.1:8081/api/v1/gateway/opcua/methods/set-string \
  -H 'Content-Type: application/json' \
  -d '{"arguments":[{"type":"UInt32","value":1},{"type":"UInt32","value":0},{"type":"String","value":"CONNECTORS-OPCUA-TEST"}]}'
```

每个 SSE 客户端打开一个 OPC UA 数据变更订阅，断开后释放。点位事件包含质量、原始 UA 状态码、接收时间，以及服务端提供的源时间。`intervalMillis` 指定请求的发布周期。可选的 `deadband` 为所有配置点位设置绝对死区，适用于服务端支持该过滤器的数值点位。收到终止性的 `subscription-error` 后可重新打开事件流。如果 OPC UA 会话已断开，需要先重启 Gateway。试点暂未提供事件补发。

方法参数按顺序传入，并显式声明 `UInt32`、`String` 等 UA 标量类型。无入参时使用 `{"arguments":[]}`。响应包含 Call 状态码和带类型的出参，业务结果依据出参中的 `Succeeded` 等字段判断。Gateway 对 Call 只发送一次。Process 编排、权限校验和命令记录留待后续接入。
