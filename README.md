# Industrial Automation Connectors

工业自动化连接器负责接入 PLC、半导体设备、机器人、仪器、工厂摄像头和其他自动化系统，并向上层提供统一的设备、点位、视觉信号、事件和操作能力。原始视频流通过独立媒体通道接入，视觉算法由连接器外的视觉服务承载。

## 设计文档

- [技术方案与总体设计](docs/architecture/technical-design.md)：Gateway、Process、Frontend 分层，统一设备能力模型、协议适配、订阅与下行控制、动态脚本、Linux/Kubernetes 部署及实施阶段。
- 后续模块设计放在 `docs/modules/`，例如 Gateway、协议适配、设备模型、脚本运行时和前端。

整体系统架构和跨工程职责边界由 `design-docs` 工程维护；本工程维护连接器自身的技术方案、模块接口和实现设计。
