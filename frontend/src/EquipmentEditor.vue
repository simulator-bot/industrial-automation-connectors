<script setup lang="ts">
import { computed, reactive } from 'vue'
import Drawer from './Drawer.vue'
import { dataTypes, uaTypes } from './catalog'
import type { Catalog, CollectionPlan, DeviceInstance, DeviceModel, Edit, Editor, ModelPoint } from './catalog'

const props = defineProps<{ editor: Editor; catalog: Catalog; saving: boolean; error: string }>()
const emit = defineEmits<{ save: [edit: Edit]; close: [] }>()
const { id, section } = props.editor
const entityName = { models: '设备模型', points: '点位', plans: '采集计划', instances: '设备实例' }[section]
const modelID = props.editor.modelId || props.catalog.models[0]?.id || ''
const oldModel = props.catalog.models.find(item => item.id === id)
const oldPlan = props.catalog.plans.find(item => item.id === id)
const draftID = id ?? crypto.randomUUID()

const deviceModel = reactive<DeviceModel>({
  id: draftID, name: oldModel?.name ?? '', protocol: 'opcua',
  description: oldModel?.description ?? '',
  methods: (oldModel?.methods ?? []).map(method => ({
    ...method, parameters: (method.parameters ?? []).map(parameter => ({ ...parameter })),
  })),
})
const point = reactive<ModelPoint>({
  id: draftID, modelId: modelID, name: '', nodeId: '', dataType: 'string', unit: '',
  ...props.catalog.points.find(item => item.id === id),
})
const plan = reactive<CollectionPlan>({
  id: draftID, modelId: modelID, name: '', mode: 'ON_CHANGE', intervalMillis: 250, deadband: 0,
  ...oldPlan, pointIds: [...(oldPlan?.pointIds ?? [])],
})
const instance = reactive<DeviceInstance>({
  id: draftID, modelId: modelID, planId: '', name: '', endpoint: 'opc.tcp://localhost:4840',
  ...props.catalog.instances.find(item => item.id === id),
})
const modelPoints = computed(() => props.catalog.points.filter(item => item.modelId === plan.modelId))
const modelPlans = computed(() => props.catalog.plans.filter(item => item.modelId === instance.modelId))
if (!id && section === 'instances') instance.planId = modelPlans.value[0]?.id ?? ''

function submit() {
  switch (section) {
    case 'models': emit('save', { section, item: deviceModel }); break
    case 'points': emit('save', { section, item: point }); break
    case 'plans': emit('save', { section, item: plan }); break
    case 'instances': emit('save', { section, item: instance }); break
  }
}

function addMethod() {
  deviceModel.methods.push({
    id: crypto.randomUUID(), name: '', objectNodeId: '', methodNodeId: '', parameters: [],
  })
}
</script>

<template>
  <Drawer :title="`${id ? '编辑' : '新建'}${entityName}`" @close="!saving && emit('close')">
    <form class="drawer-form" @submit.prevent="submit">
      <div class="drawer-body">
        <p v-if="error" class="notice error" role="alert">{{ error }}</p>
        <template v-if="section === 'models'">
          <label>模型名称<input v-model.trim="deviceModel.name" required placeholder="例如，长晶棒检测设备" autofocus></label>
          <label>协议<input value="OPC UA" readonly></label>
          <label>说明 <span class="optional">选填</span><textarea v-model.trim="deviceModel.description" rows="2" /></label>
          <section class="form-section">
            <div class="section-heading">
              <h3>方法</h3>
              <button class="button secondary small" type="button" @click="addMethod">添加方法</button>
            </div>
            <p v-if="!deviceModel.methods.length" class="help">在模型中定义方法后，可以在设备实例中调用测试。</p>
            <fieldset v-for="(method, i) in deviceModel.methods" :key="method.id" class="method-form">
              <div class="section-heading">
                <h4>{{ method.name || '新方法' }}</h4>
                <button class="text-button danger" type="button" @click="deviceModel.methods.splice(i, 1)">移除方法</button>
              </div>
              <label>方法名称<input v-model.trim="method.name" required placeholder="SetString"></label>
              <div class="field-grid">
                <label>对象 NodeID<input v-model.trim="method.objectNodeId" required placeholder="ns=2;i=15007"></label>
                <label>方法 NodeID<input v-model.trim="method.methodNodeId" required placeholder="ns=2;i=17019"></label>
              </div>
              <div class="section-heading">
                <h4>输入参数</h4>
                <button class="text-button" type="button" @click="method.parameters.push({ name: '', type: 'String' })">添加参数</button>
              </div>
              <div v-for="(parameter, j) in method.parameters" :key="j" class="parameter-fields">
                <label>参数 {{ j + 1 }}<input v-model.trim="parameter.name" required></label>
                <label>UA 类型<select v-model="parameter.type"><option v-for="kind in uaTypes" :key="kind">{{ kind }}</option></select></label>
                <button class="icon-button" type="button" :aria-label="`移除参数 ${j + 1}`" @click="method.parameters.splice(j, 1)">×</button>
              </div>
            </fieldset>
          </section>
        </template>

        <template v-else-if="section === 'points'">
          <label>点位名称<input v-model.trim="point.name" required placeholder="例如，晶棒编号" autofocus></label>
          <label>设备模型<select v-model="point.modelId" required>
            <option value="" disabled>选择设备模型</option>
            <option v-for="item in catalog.models" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select></label>
          <label>OPC UA NodeID<input v-model.trim="point.nodeId" required placeholder="ns=2;s=ingotNum" spellcheck="false"></label>
          <div class="field-grid">
            <label>数据类型<select v-model="point.dataType"><option v-for="(label, value) in dataTypes" :key="value" :value="value">{{ label }}</option></select></label>
            <label>单位 <span class="optional">选填</span><input v-model.trim="point.unit" placeholder="例如，mm"></label>
          </div>
        </template>

        <template v-else-if="section === 'plans'">
          <label>计划名称<input v-model.trim="plan.name" required placeholder="例如，检测工位订阅" autofocus></label>
          <label>设备模型<select v-model="plan.modelId" required @change="plan.pointIds = []">
            <option value="" disabled>选择设备模型</option>
            <option v-for="item in catalog.models" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select></label>
          <label>采集模式<input value="数据变化订阅" readonly></label>
          <div class="field-grid">
            <label>发布周期（ms）<input v-model.number="plan.intervalMillis" type="number" min="1" max="3600000" required></label>
            <label>绝对死区<input v-model.number="plan.deadband" type="number" min="0" step="any" required></label>
          </div>
          <p class="help">死区为 0 时使用默认变化通知。正值适用于数值点位。</p>
          <fieldset class="selection-list">
            <legend>订阅点位</legend>
            <p v-if="!modelPoints.length" class="help">请先为该模型配置点位。</p>
            <label v-for="item in modelPoints" :key="item.id" class="check-row">
              <input v-model="plan.pointIds" type="checkbox" :value="item.id">
              <span>{{ item.name }}<small>{{ item.nodeId }}</small></span>
            </label>
          </fieldset>
        </template>

        <template v-else>
          <label>实例名称<input v-model.trim="instance.name" required placeholder="例如，检测工位 01" autofocus></label>
          <label>设备模型<select v-model="instance.modelId" required @change="instance.planId = modelPlans[0]?.id ?? ''">
            <option value="" disabled>选择设备模型</option>
            <option v-for="item in catalog.models" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select></label>
          <label>采集计划<select v-model="instance.planId" required>
            <option value="" disabled>选择该模型的采集计划</option>
            <option v-for="item in modelPlans" :key="item.id" :value="item.id">{{ item.name }}</option>
          </select></label>
          <p v-if="!modelPlans.length" class="help">请先为该模型创建采集计划。</p>
          <label>OPC UA 端点<input v-model.trim="instance.endpoint" required placeholder="opc.tcp://localhost:4840" spellcheck="false"></label>
        </template>
      </div>
      <footer class="drawer-footer">
        <button class="button secondary" type="button" :disabled="saving" @click="emit('close')">取消</button>
        <button class="button primary" type="submit" :disabled="saving || (section === 'plans' && !plan.pointIds.length)">
          {{ saving ? '正在保存…' : '保存' }}
        </button>
      </footer>
    </form>
  </Drawer>
</template>
