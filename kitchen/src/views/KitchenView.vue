<template>
  <div class="kitchen-view">
    <!-- Tabs -->
    <div class="tabs">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        :class="['tab', tab.value, { active: activeTab === tab.value }]"
        @click="activeTab = tab.value"
      >
        {{ tab.label }}
        <span v-if="countByStatus[tab.value] > 0" class="tab-count">{{ countByStatus[tab.value] }}</span>
      </button>
    </div>

    <!-- Task Grid -->
    <div class="tasks-grid">
      <div
        v-for="task in filteredTasks"
        :key="task.id"
        :class="['task-card', task.status]"
      >
        <div class="task-header">
          <span class="table-badge">{{ task.table_number }}桌</span>
          <span class="task-time">{{ formatTime(task.assigned_at) }}</span>
        </div>
        <div class="dish-name">{{ task.dish_name }}</div>
        <div class="dish-qty">× {{ task.quantity }}</div>
        <div class="task-footer">
          <button
            v-if="task.status === 'pending'"
            @click="updateStatus(task, 'cooking')"
            class="action-btn cooking-btn"
          >开始烹饪</button>
          <button
            v-else-if="task.status === 'cooking'"
            @click="updateStatus(task, 'done')"
            class="action-btn done-btn"
          >完成烹饪</button>
          <span v-else class="done-label">✅ 已完成</span>
        </div>
      </div>
    </div>

    <div v-if="filteredTasks.length === 0" class="empty-state">
      <span>暂无{{ currentTabLabel }}任务</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import axios from 'axios'

const tasks = ref([])
const activeTab = ref('pending')
let pollTimer

const tabs = [
  { label: '待处理', value: 'pending' },
  { label: '烹饪中', value: 'cooking' },
  { label: '已完成', value: 'done' }
]

const filteredTasks = computed(() => tasks.value.filter(t => t.status === activeTab.value))

const countByStatus = computed(() => {
  const counts = { pending: 0, cooking: 0, done: 0 }
  tasks.value.forEach(t => { if (counts[t.status] !== undefined) counts[t.status]++ })
  return counts
})

const currentTabLabel = computed(() => tabs.find(t => t.value === activeTab.value)?.label || '')

function formatTime(t) {
  return new Date(t).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

async function loadTasks() {
  const res = await axios.get('/api/cooking-tasks')
  tasks.value = res.data
}

async function updateStatus(task, newStatus) {
  await axios.put(`/api/cooking-tasks/${task.id}/status`, { status: newStatus })
  task.status = newStatus
}

onMounted(() => {
  loadTasks()
  pollTimer = setInterval(loadTasks, 5000)
})
onUnmounted(() => clearInterval(pollTimer))
</script>

<style scoped>
.kitchen-view { padding: 20px; }
.tabs { display: flex; gap: 12px; margin-bottom: 24px; }
.tab {
  padding: 12px 28px; font-size: 18px; font-weight: bold;
  border: 2px solid #444; border-radius: 8px; background: #16213e; color: #aaa; cursor: pointer;
  position: relative;
}
.tab.active.pending { border-color: #f39c12; color: #f39c12; background: rgba(243,156,18,0.1); }
.tab.active.cooking { border-color: #3498db; color: #3498db; background: rgba(52,152,219,0.1); }
.tab.active.done { border-color: #27ae60; color: #27ae60; background: rgba(39,174,96,0.1); }
.tab-count {
  position: absolute; top: -8px; right: -8px;
  background: #e74c3c; color: white; border-radius: 50%;
  width: 22px; height: 22px; display: flex; align-items: center; justify-content: center;
  font-size: 12px;
}
.tasks-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 16px; }
.task-card {
  border-radius: 12px; padding: 20px; border: 2px solid transparent;
  display: flex; flex-direction: column; gap: 12px;
}
.task-card.pending { background: rgba(243,156,18,0.15); border-color: #f39c12; }
.task-card.cooking { background: rgba(52,152,219,0.15); border-color: #3498db; }
.task-card.done { background: rgba(39,174,96,0.15); border-color: #27ae60; opacity: 0.8; }
.task-header { display: flex; justify-content: space-between; align-items: center; }
.table-badge {
  background: #e74c3c; color: white; padding: 4px 12px; border-radius: 20px;
  font-size: 16px; font-weight: bold;
}
.task-time { color: #aaa; font-size: 14px; }
.dish-name { font-size: 24px; font-weight: bold; color: white; }
.dish-qty { font-size: 20px; color: #ffd93d; font-weight: bold; }
.task-footer { margin-top: 4px; }
.action-btn {
  width: 100%; padding: 12px; font-size: 16px; font-weight: bold;
  border: none; border-radius: 8px; cursor: pointer;
}
.cooking-btn { background: #f39c12; color: white; }
.done-btn { background: #27ae60; color: white; }
.done-label { font-size: 16px; color: #27ae60; }
.empty-state { text-align: center; padding: 60px; font-size: 20px; color: #666; }
</style>
