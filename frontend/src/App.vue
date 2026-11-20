<script setup lang="ts">
import { onMounted, ref } from 'vue'

type EquipmentResponse = { items: Array<{ id: string; name: string }> }

const connected = ref(false)
const equipmentCount = ref<number | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    const response = await fetch('/api/v1/equipment')
    if (!response.ok) throw new Error(`Process API returned ${response.status}`)
    const data = (await response.json()) as EquipmentResponse
    equipmentCount.value = data.items.length
    connected.value = true
  } catch {
    connected.value = false
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <main class="shell">
    <header class="topbar">
      <a class="brand" href="/" aria-label="Connectors 首页">
        <span class="brand-mark">IC</span>
        <span>Industrial Connectors</span>
      </a>
      <span class="environment">LOCAL DEVELOPMENT</span>
    </header>

    <section class="intro">
      <p class="eyebrow">INDUSTRIAL AUTOMATION PLATFORM</p>
      <h1>连接现场设备，<span>统一设备能力。</span></h1>
      <p class="description">设备、点位、订阅和受控操作的统一入口。</p>
    </section>

    <section class="overview" aria-label="平台状态">
      <article class="card service-card">
        <div class="card-heading">
          <div>
            <p class="card-label">PROCESS API</p>
            <h2>业务处理服务</h2>
          </div>
          <span class="status-dot" :class="connected ? 'online' : 'offline'"></span>
        </div>
        <p class="card-value">{{ loading ? '正在连接' : connected ? '运行中' : '未连接' }}</p>
        <p class="card-footnote">设备模型、采集编排与命令控制</p>
      </article>

      <article class="card">
        <p class="card-label">REGISTERED EQUIPMENT</p>
        <h2 class="metric">{{ equipmentCount ?? '—' }}</h2>
        <p class="card-footnote">当前登记的设备</p>
      </article>

      <article class="card">
        <p class="card-label">GATEWAY</p>
        <h2 class="metric muted">等待协议配置</h2>
        <p class="card-footnote">协议适配器将在后续阶段接入</p>
      </article>
    </section>

    <section class="next-step">
      <div class="step-number">01</div>
      <div>
        <p class="card-label">FOUNDATION</p>
        <h2>连接器基础框架</h2>
        <p>Go 服务边界和设备能力模型已建立。下一步可围绕首个试点协议实现连接、点位订阅和状态上报。</p>
      </div>
      <span class="pill">不含视觉接入</span>
    </section>

    <footer>Industrial Automation Connectors <span>·</span> 架构骨架</footer>
  </main>
</template>
