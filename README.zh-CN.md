# 工业自动化连接器

[English](README.md) | **简体中文**

工业自动化连接器负责接入 PLC、半导体设备、机器人、仪器、工厂摄像头和其他自动化系统，并向上层提供统一的设备、点位、视觉信号、事件和操作能力。原始视频流通过独立媒体通道接入，视觉算法由连接器外的视觉服务承载。

## 设计文档

- [技术方案与总体设计](docs/architecture/technical-design.zh-CN.md)：Gateway、Process、Frontend 分层，统一设备能力模型、协议与视觉源适配、受限规则 DSL、Linux/Kubernetes 部署及实施阶段。
- 后续模块设计放在 `docs/modules/`，例如 Gateway、协议适配、设备模型、规则运行时和前端。

整体系统架构和跨工程职责边界由 `design-docs` 工程维护；本工程维护连接器自身的技术方案、模块接口和实现设计。

## 当前工程骨架

本仓库已初始化 Go 模块和三类运行入口：

- `cmd/gateway`：设备协议会话和适配器运行时入口。
- `cmd/process`：设备模型管理 API 与业务处理入口。
- `cmd/rule-worker`：与设备网络隔离的规则执行进程边界（当前为空闲骨架）。
- `frontend`：Vue 3 + TypeScript + Vite 运维界面骨架。
- `internal/equipment`、`internal/protocol`、`internal/contracts/v1`：统一能力模型、适配器 SPI 和事件/命令契约。
- `deployments/kubernetes/base`：Linux 容器及 Kubernetes 初始部署模板。

当前 Process 使用内存设备仓储，仅提供健康检查和空设备列表；Gateway 尚未注册协议实现，规则 Worker 尚未加载 CEL。该骨架不宣称已具备真实设备采集或下行能力。视觉/摄像头不在本阶段实现。

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
