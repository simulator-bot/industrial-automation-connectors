<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import EquipmentEditor from './EquipmentEditor.vue'
import InstanceTest from './InstanceTest.vue'
import { dataTypes, loadCatalog, saveCatalog } from './catalog'
import type { Catalog, Edit, Editor, Section } from './catalog'

const sections = {
  models: {
    title: '设备建模', add: '新建设备模型', empty: '还没有设备模型',
    description: '定义可复用的设备类型、协议和方法。',
    columns: ['模型名称', '协议', '点位', '方法'],
  },
  points: {
    title: '点位配置', add: '添加点位', empty: '当前没有点位',
    description: '为设备模型配置 OPC UA NodeID、数据类型和单位。',
    columns: ['点位名称', '设备模型', 'NodeID', '数据类型', '单位'],
  },
  plans: {
    title: '采集计划', add: '新建采集计划', empty: '当前没有采集计划',
    description: '选择模型点位并配置数据变化订阅的发布周期和死区。',
    columns: ['计划名称', '设备模型', '采集模式', '发布周期', '点位数'],
  },
  instances: {
    title: '设备实例', add: '新建设备实例', empty: '当前没有设备实例',
    description: '将设备模型和采集计划绑定到具体端点，然后测试连接、读取和订阅。',
    columns: ['实例名称', '设备模型', '采集计划', 'OPC UA 端点'],
  },
} satisfies Record<Section, { title: string; add: string; empty: string; description: string; columns: string[] }>

const section = ref<Section>('models')
const metadata = computed(() => sections[section.value])
const catalog = ref<Catalog>({ revision: 0, models: [], points: [], plans: [], instances: [] })
const modelFilter = ref('')
const loaded = ref(false)
const connected = ref(false)
const loading = ref(false)
const saving = ref(false)
const pageError = ref('')
const editorError = ref('')
const feedback = ref('')
const editor = ref<Editor>()
const testingID = ref('')
const testingInstance = computed(() => catalog.value.instances.find(item => item.id === testingID.value))
const modelName = (id: string) => catalog.value.models.find(item => item.id === id)?.name ?? '—'
const matchesModel = (item: { modelId: string }) => !modelFilter.value || item.modelId === modelFilter.value

type Cell = { text: string | number; mono?: boolean; tag?: boolean }
type Row = { id: string; name: string; detail?: string; cells: Cell[] }
const rows = computed<Row[]>(() => {
  switch (section.value) {
    case 'models':
      return catalog.value.models.map(item => ({
        id: item.id, name: item.name, detail: item.description,
        cells: [
          { text: 'OPC UA', tag: true },
          { text: catalog.value.points.filter(point => point.modelId === item.id).length },
          { text: item.methods?.length ?? 0 },
        ],
      }))
    case 'points':
      return catalog.value.points.filter(matchesModel).map(item => ({
        id: item.id, name: item.name,
        cells: [{ text: modelName(item.modelId) }, { text: item.nodeId, mono: true },
          { text: dataTypes[item.dataType] }, { text: item.unit || '—' }],
      }))
    case 'plans':
      return catalog.value.plans.filter(matchesModel).map(item => ({
        id: item.id, name: item.name,
        cells: [{ text: modelName(item.modelId) }, { text: '数据变化订阅' },
          { text: `${item.intervalMillis} ms`, mono: true }, { text: item.pointIds.length }],
      }))
    case 'instances':
      return catalog.value.instances.filter(matchesModel).map(item => ({
        id: item.id, name: item.name,
        cells: [{ text: modelName(item.modelId) },
          { text: catalog.value.plans.find(plan => plan.id === item.planId)?.name ?? '—' },
          { text: item.endpoint, mono: true }],
      }))
  }
})

const prerequisite = computed<{ section: Section; message: string; action: string } | undefined>(() => {
  if (section.value !== 'models' && !catalog.value.models.length) {
    return { section: 'models', message: '先创建设备模型，再配置其点位。', action: '去设备建模' }
  }
  if (section.value === 'plans' && !catalog.value.points.some(matchesModel)) {
    return { section: 'points', message: '配置点位后，才能把它们加入采集计划。', action: '去配置点位' }
  }
  if (section.value === 'instances' && !catalog.value.plans.some(matchesModel)) {
    return { section: 'plans', message: '创建采集计划后，才能将它绑定到设备实例。', action: '去配置采集计划' }
  }
})

function navigate(next: Section) {
  section.value = next
  modelFilter.value = ''
  feedback.value = ''
  pageError.value = ''
}

async function refresh() {
  loading.value = true
  pageError.value = ''
  try {
    catalog.value = await loadCatalog()
    if (!catalog.value.models.some(item => item.id === modelFilter.value)) modelFilter.value = ''
    loaded.value = true
    connected.value = true
  } catch (error) {
    connected.value = false
    pageError.value = (error as Error).message
  } finally {
    loading.value = false
  }
}
onMounted(refresh)

function openEditor(id?: string) {
  editorError.value = ''
  let modelId = modelFilter.value
  if (!modelId && section.value === 'plans') modelId = catalog.value.points[0]?.modelId ?? ''
  if (!modelId && section.value === 'instances') modelId = catalog.value.plans[0]?.modelId ?? ''
  editor.value = { section: section.value, id, modelId }
}

function upsert<T extends { id: string }>(items: T[], item: T) {
  return items.some(current => current.id === item.id)
    ? items.map(current => current.id === item.id ? item : current)
    : [...items, item]
}

async function save(edit: Edit) {
  if (saving.value) return
  const next = { ...catalog.value }
  switch (edit.section) {
    case 'models': next.models = upsert(next.models, edit.item); break
    case 'points': next.points = upsert(next.points, edit.item); break
    case 'plans': next.plans = upsert(next.plans, edit.item); break
    case 'instances': next.instances = upsert(next.instances, edit.item); break
  }
  editorError.value = ''
  saving.value = true
  try {
    catalog.value = await saveCatalog(next)
    editor.value = undefined
    feedback.value = `已保存「${edit.item.name}」`
    pageError.value = ''
  } catch (error) {
    editorError.value = (error as Error).message
  } finally {
    saving.value = false
  }
}

async function remove(row: Row) {
  if (saving.value || !window.confirm(`确认删除「${row.name}」？`)) return
  const key = section.value
  pageError.value = ''
  feedback.value = ''
  saving.value = true
  try {
    catalog.value = await saveCatalog({ ...catalog.value, [key]: catalog.value[key].filter(item => item.id !== row.id) })
    feedback.value = `已删除「${row.name}」`
  } catch (error) {
    pageError.value = (error as Error).message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="layout">
    <aside class="sidebar">
      <div class="brand">
        <span class="brand-mark" aria-hidden="true"><i></i><i></i><i></i></span>
        <div><strong>Connectors</strong><span>工业自动化连接器</span></div>
      </div>
      <nav aria-label="设备配置">
        <button v-for="(item, key) in sections" :key="key" type="button" class="nav-item"
          :aria-current="section === key ? 'page' : undefined" @click="navigate(key)">
          <span>{{ item.title }}</span>
          <small v-if="catalog[key].length">{{ catalog[key].length }}</small>
        </button>
      </nav>
      <div class="service-status"><span class="status-light" :class="{ online: connected && !loading }"></span>
        {{ loading ? '正在连接 Process' : connected ? 'Process 已连接' : 'Process 未连接' }}</div>
    </aside>

    <main id="main" class="main">
      <header class="page-header">
        <div><h1>{{ metadata.title }}</h1><p>{{ metadata.description }}</p></div>
        <button class="button primary" type="button" :disabled="!loaded || loading || !!prerequisite"
          @click="openEditor()"><span aria-hidden="true">+</span> {{ metadata.add }}</button>
      </header>
      <p v-if="feedback" class="notice success" role="status">{{ feedback }}</p>
      <div v-if="pageError" class="notice error notice-action" role="alert">
        <span>{{ pageError }}</span><button class="text-button" type="button" @click="refresh">重新加载</button>
      </div>

      <section class="panel" :aria-label="metadata.title">
        <div class="panel-tools" v-if="loaded">
          <span>{{ rows.length }} 项</span>
          <label v-if="section !== 'models' && catalog.models.length" class="filter">
            <span>设备模型</span>
            <select v-model="modelFilter" aria-label="按设备模型筛选">
              <option value="">全部模型</option>
              <option v-for="item in catalog.models" :key="item.id" :value="item.id">{{ item.name }}</option>
            </select>
          </label>
        </div>
        <div class="table-frame" v-if="rows.length">
          <table>
            <thead><tr><th v-for="column in metadata.columns" :key="column" scope="col">{{ column }}</th><th scope="col" class="actions">操作</th></tr></thead>
            <tbody>
              <tr v-for="row in rows" :key="row.id">
                <td class="name-cell"><strong>{{ row.name }}</strong><small v-if="row.detail">{{ row.detail }}</small></td>
                <td v-for="(cell, index) in row.cells" :key="index" :class="{ mono: cell.mono }">
                  <span v-if="cell.tag" class="tag">{{ cell.text }}</span><template v-else>{{ cell.text }}</template>
                </td>
                <td class="actions">
                  <button v-if="section === 'instances'" type="button" class="text-button" @click="testingID = row.id"
                    :aria-label="`测试 ${row.name}`">测试</button>
                  <button type="button" class="text-button neutral" @click="openEditor(row.id)" :aria-label="`编辑 ${row.name}`">编辑</button>
                  <button type="button" class="text-button danger" :disabled="saving" @click="remove(row)"
                    :aria-label="`删除 ${row.name}`">删除</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="empty">
          <strong>{{ loading ? '正在加载配置…' : loaded ? metadata.empty : '设备配置尚未加载' }}</strong>
          <template v-if="loaded && !loading">
            <p v-if="prerequisite">{{ prerequisite.message }}</p>
            <button v-if="prerequisite" class="button secondary" type="button" @click="navigate(prerequisite.section)">{{ prerequisite.action }}</button>
            <button v-else class="button secondary" type="button" @click="openEditor()">{{ metadata.add }}</button>
          </template>
          <button v-else-if="!loading && !pageError" class="button secondary" type="button" @click="refresh">重新加载</button>
        </div>
      </section>
    </main>

    <EquipmentEditor v-if="editor" :editor="editor" :catalog="catalog" :saving="saving" :error="editorError"
      @save="save" @close="editor = undefined" />
    <InstanceTest v-if="testingInstance" :catalog="catalog" :instance="testingInstance" @close="testingID = ''" />
  </div>
</template>
