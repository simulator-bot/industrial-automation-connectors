<script setup lang="ts">
import { onMounted, ref } from 'vue'

defineProps<{ title: string; wide?: boolean }>()
const emit = defineEmits<{ close: [] }>()
const dialog = ref<HTMLDialogElement>()
onMounted(() => dialog.value?.showModal())
</script>

<template>
  <dialog ref="dialog" class="drawer" :class="{ 'drawer-wide': wide }" aria-labelledby="drawer-title"
    @cancel.prevent="emit('close')">
    <header class="drawer-header">
      <h2 id="drawer-title">{{ title }}</h2>
      <button class="icon-button" type="button" aria-label="关闭" @click="emit('close')">×</button>
    </header>
    <slot />
  </dialog>
</template>
