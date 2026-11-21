export const dataTypes = {
  bool: '布尔', int: '整数', uint: '无符号整数', float: '浮点数',
  string: '字符串', bytes: '字节串', enum: '枚举', time: '时间', struct: '结构体',
}

export const uaTypes = [
  'Boolean', 'SByte', 'Byte', 'Int16', 'UInt16', 'Int32', 'UInt32',
  'Int64', 'UInt64', 'Float', 'Double', 'String', 'DateTime', 'ByteString',
]

export interface ModelMethod {
  id: string
  name: string
  objectNodeId: string
  methodNodeId: string
  parameters: { name: string; type: string }[]
}

export interface DeviceModel {
  id: string
  name: string
  protocol: 'opcua'
  description: string
  methods: ModelMethod[]
}

export interface ModelPoint {
  id: string
  modelId: string
  name: string
  nodeId: string
  dataType: keyof typeof dataTypes
  unit: string
}

export interface CollectionPlan {
  id: string
  modelId: string
  name: string
  mode: 'ON_CHANGE'
  pointIds: string[]
  intervalMillis: number
  deadband: number
}

export interface DeviceInstance {
  id: string
  name: string
  modelId: string
  planId: string
  endpoint: string
}

export interface Catalog {
  revision: number
  models: DeviceModel[]
  points: ModelPoint[]
  plans: CollectionPlan[]
  instances: DeviceInstance[]
}

export type Section = Exclude<keyof Catalog, 'revision'>
export type Editor = { section: Section; id?: string; modelId?: string }
export type Edit = {
  [K in Section]: { section: K; item: Catalog[K][number] }
}[Section]

export interface PointEvent {
  value: unknown
  quality: 'GOOD' | 'BAD' | 'UNCERTAIN'
  protocolStatus?: string
  sourceTimestamp?: string
  ingestTimestamp: string
}

export interface TestResult {
  connected: boolean
  readCompleted: boolean
  subscriptionOpened: boolean
  points?: { pointId: string; read?: PointEvent; latest?: PointEvent; notifications: number }[]
  methodResult?: { statusCode: string; outputArguments: { type: string; value: unknown }[] }
  error?: string
}

const endpoint = '/api/v1/equipment'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(endpoint + path, { cache: 'no-store', ...options })
  if (!response.ok) {
    const message = (await response.text()).trim()
    if (response.status === 404 && path === '/configuration') {
      throw new Error('设备配置服务不可用，请用 make run-process 启动 Process。')
    }
    throw new Error(message || `请求失败（HTTP ${response.status}）`)
  }
  return await response.json() as T
}

export function loadCatalog() {
  return request<Catalog>('/configuration')
}

export function saveCatalog(data: Catalog) {
  return request<Catalog>('/configuration', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
}

export function testInstance(id: string, body: string, signal: AbortSignal) {
  return request<TestResult>(`/instances/${encodeURIComponent(id)}/test`, {
    method: 'POST', signal,
    headers: { 'Content-Type': 'application/json' },
    body,
  })
}
