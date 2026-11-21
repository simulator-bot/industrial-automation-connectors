<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import Drawer from './Drawer.vue'
import { testInstance } from './catalog'
import type { Catalog, DeviceInstance, PointEvent, TestResult } from './catalog'

const props = defineProps<{ instance: DeviceInstance; catalog: Catalog }>()
const emit = defineEmits<{ close: [] }>()
const model = computed(() => props.catalog.models.find(item => item.id === props.instance.modelId))
const plan = computed(() => props.catalog.plans.find(item => item.id === props.instance.planId))
const methods = computed(() => model.value?.methods ?? [])
const tab = ref<'points' | 'methods'>('points')
const methodID = ref(methods.value[0]?.id ?? '')
const method = computed(() => methods.value.find(item => item.id === methodID.value))
const values = ref<string[]>([])
const loading = ref(false)
const result = ref<TestResult>()
const error = ref('')
let abort: AbortController | undefined

watch(method, item => {
  values.value = (item?.parameters ?? []).map(parameter => parameter.type === 'Boolean' ? 'false' : '')
  result.value = undefined
  error.value = ''
}, { immediate: true })

const rows = computed(() => (plan.value?.pointIds ?? []).map(id => ({
  point: props.catalog.points.find(item => item.id === id),
  result: result.value?.points?.find(item => item.pointId === id),
})))

function selectTab(next: typeof tab.value) {
  tab.value = next
  result.value = undefined
  error.value = ''
}

function argumentJSON(type: string, input: string): string {
  if (['String', 'DateTime', 'ByteString'].includes(type)) return JSON.stringify(input)
  const value = input.trim()
  if (type === 'Boolean' && (value === 'true' || value === 'false')) return value
  if (/^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?$/.test(value)) return value
  throw new Error(`请输入有效的 ${type} 参数。`)
}

async function run() {
  if (loading.value) return
  error.value = ''
  result.value = undefined
  const controller = new AbortController()
  abort = controller
  loading.value = true
  try {
    const revision = props.catalog.revision
    let payload = JSON.stringify({ revision })
    if (tab.value === 'methods' && method.value) {
      const args = (method.value.parameters ?? []).map((param, i) =>
        `{"type":${JSON.stringify(param.type)},"value":${argumentJSON(param.type, values.value[i] ?? '')}}`)
      // Retain the original numeric text so 64-bit values do not round in JavaScript.
      payload = `{"revision":${revision},"methodId":${JSON.stringify(method.value.id)},"arguments":[${args.join(',')}]}`
    }
    result.value = await testInstance(props.instance.id, payload, controller.signal)
  } catch (cause) {
    if (!controller.signal.aborted) error.value = (cause as Error).message
  } finally {
    loading.value = false
  }
}

onBeforeUnmount(() => abort?.abort())

function display(value: unknown) {
  if (value === undefined || value === null) return '—'
  return typeof value === 'string' ? value || '""' : JSON.stringify(value)
}

function quality(event?: PointEvent) {
  return event ? { GOOD: '正常', BAD: '异常', UNCERTAIN: '不确定' }[event.quality] : '—'
}

function timestamp(event?: PointEvent) {
  return event?.sourceTimestamp || event?.ingestTimestamp || '—'
}
</script>

<template>
  <Drawer title="实例测试" wide @close="emit('close')">
    <form class="drawer-form" @submit.prevent="run">
      <div class="drawer-body test-body">
        <div>
          <h3 class="instance-title">{{ instance.name }}</h3>
          <dl class="instance-details">
            <div><dt>设备模型</dt><dd>{{ model?.name }}</dd></div>
            <div><dt>采集计划</dt><dd>{{ plan?.name }}</dd></div>
            <div><dt>连接端点</dt><dd class="mono">{{ instance.endpoint }}</dd></div>
          </dl>
        </div>

        <div v-if="methods.length" class="segments" aria-label="测试类型">
          <button type="button" :aria-pressed="tab === 'points'" :disabled="loading" @click="selectTab('points')">采集测试</button>
          <button type="button" :aria-pressed="tab === 'methods'" :disabled="loading" @click="selectTab('methods')">方法调用</button>
        </div>

        <p v-if="loading" class="notice info" role="status">
          {{ tab === 'points' ? '正在连接、读取点位并观察 4 秒订阅通知…' : '正在连接并调用已配置的方法…' }}
        </p>
        <p v-if="error || result?.error" class="notice error" role="alert">{{ error || result?.error }}</p>

        <template v-if="tab === 'points'">
          <div v-if="result" class="test-summary" aria-live="polite">
            <span class="tag" :class="result.connected ? 'good' : 'bad'">{{ result.connected ? '已连接' : '连接失败' }}</span>
            <span class="tag" :class="{ good: result.readCompleted }">{{ result.readCompleted ? '点位读取已完成' : '读取未完成' }}</span>
            <span class="tag" :class="{ good: result.subscriptionOpened }">{{ result.subscriptionOpened ? '已建立订阅' : '订阅未建立' }}</span>
          </div>
          <p v-else-if="!loading" class="help">测试会按该实例的采集计划读取一次点位，随后观察 4 秒订阅通知。</p>
          <div class="table-frame">
            <table class="test-table">
              <thead><tr><th scope="col">点位</th><th scope="col">读取值</th><th scope="col">最新订阅值</th>
                <th scope="col">通知</th><th scope="col">质量</th></tr></thead>
              <tbody>
                <tr v-for="row in rows" :key="row.point?.id">
                  <td><strong>{{ row.point?.name }}</strong><small class="mono">{{ row.point?.nodeId }}</small></td>
                  <td class="mono">{{ display(row.result?.read?.value) }}</td>
                  <td class="mono">{{ display(row.result?.latest?.value) }}</td>
                  <td>{{ row.result?.notifications ?? '—' }}</td>
                  <td>
                    <span class="tag" :class="(row.result?.latest ?? row.result?.read)?.quality.toLowerCase()">
                      {{ quality(row.result?.latest ?? row.result?.read) }}
                    </span>
                    <small v-if="row.result?.read" class="mono" :title="timestamp(row.result.latest ?? row.result.read)">
                      {{ (row.result.latest ?? row.result.read).protocolStatus }}
                    </small>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div v-if="result?.points?.length" class="timestamp-list">
            <h4>最新数据时间</h4>
            <dl v-for="row in rows" :key="row.point?.id">
              <dt>{{ row.point?.name }}</dt>
              <dd class="mono">{{ timestamp(row.result?.latest ?? row.result?.read) }}</dd>
            </dl>
          </div>
        </template>

        <template v-else>
          <label>测试方法<select v-model="methodID" :disabled="loading">
            <option v-for="item in methods" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select></label>
          <p v-if="method" class="help mono">{{ method.objectNodeId }} / {{ method.methodNodeId }}</p>
          <label v-for="(param, i) in method?.parameters" :key="i">
            {{ param.name }} <span class="optional">{{ param.type }}</span>
            <select v-if="param.type === 'Boolean'" v-model="values[i]" :disabled="loading">
              <option value="true">true</option><option value="false">false</option>
            </select>
            <input v-else v-model="values[i]" :required="param.type !== 'String'" :disabled="loading"
              :placeholder="param.type === 'DateTime' ? '2026-01-01T00:00:00Z' : param.type === 'ByteString' ? 'Base64' : undefined">
          </label>
          <p v-if="method" class="help">点击调用会执行模拟器中的方法。</p>
          <div v-if="result?.methodResult" class="method-result">
            <h4>方法返回</h4>
            <pre>{{ JSON.stringify(result.methodResult, null, 2) }}</pre>
          </div>
        </template>
      </div>
      <footer class="drawer-footer">
        <button class="button secondary" type="button" @click="emit('close')">关闭</button>
        <button class="button primary" type="submit" :disabled="loading || (tab === 'methods' && !method)">
          {{ loading ? '测试中…' : tab === 'points' ? '测试采集' : '调用方法' }}
        </button>
      </footer>
    </form>
  </Drawer>
</template>
