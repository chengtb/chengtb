<template>
  <div>
    <header class="kitchen-header">
      <span>🍳 厨房操作台</span>
      <span class="time">{{ currentTime }}</span>
    </header>
    <router-view />
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
const currentTime = ref('')
let timer

function updateTime() {
  currentTime.value = new Date().toLocaleTimeString('zh-CN')
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})
onUnmounted(() => clearInterval(timer))
</script>

<style scoped>
.kitchen-header {
  background: #16213e; padding: 16px 24px; font-size: 22px; font-weight: bold;
  display: flex; justify-content: space-between; align-items: center; color: #fff;
  border-bottom: 2px solid #e74c3c;
}
.time { font-size: 20px; font-family: monospace; color: #ffd93d; }
</style>
