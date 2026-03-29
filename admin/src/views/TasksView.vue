<template>
  <div>
    <h1 class="page-title">任务看板</h1>

    <!-- Status Tabs -->
    <div class="tabs">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        :class="['tab', { active: activeTab === tab.value }]"
        @click="activeTab = tab.value"
      >
        {{ tab.label }}
        <span v-if="countByStatus[tab.value] > 0" class="tab-count">{{ countByStatus[tab.value] }}</span>
      </button>
    </div>

    <!-- Task Groups -->
    <div v-if="visibleGroups.length > 0">
      <div v-for="group in visibleGroups" :key="group.key" class="group-section">
        <h2 class="group-title">{{ statusLabel(group.key) }} ({{ group.items.length }})</h2>
        <div class="tasks-grid">
          <div
            v-for="task in group.items"
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
      </div>
    </div>

    <div v-else class="empty-state">暂无{{ currentTabLabel }}任务</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import { useGroupBy } from '../composables/group.ts'

const tasks = ref([])
const activeTab = ref('all')
let pollTimer

const tabs = [
  { label: '全部', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '烹饪中', value: 'cooking' },
  { label: '已完成', value: 'done' }
]

const currentTabLabel = computed(() => tabs.find(t => t.value === activeTab.value)?.label || '')

const filteredTasks = computed(() =>
  activeTab.value === 'all' ? tasks.value : tasks.value.filter(t => t.status === activeTab.value)
)

// useGroupBy returns groupBy as a computed array of { key, items } — never a plain object,
// so calling .map() on groupBy.value is always safe.
const { groupBy } = useGroupBy(filteredTasks, 'status')

// Preserve the desired display order for status groups.
const statusOrder = ['pending', 'cooking', 'done']
const visibleGroups = computed(() =>
  groupBy.value
    .slice()
    .sort((a, b) => statusOrder.indexOf(a.key) - statusOrder.indexOf(b.key))
)

const countByStatus = computed(() => {
  const counts = { pending: 0, cooking: 0, done: 0 }
  tasks.value.forEach(t => {
    if (counts[t.status] !== undefined) counts[t.status]++
  })
  return counts
})

const statusLabel = s =>
  ({ pending: '待处理', cooking: '烹饪中', done: '已完成' }[s] || s)

function formatTime(t) {
  return new Date(t).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

async function loadTasks() {
  const res = await axios.get('/api/cooking-tasks')
  tasks.value = Array.isArray(res.data) ? res.data : []
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
.page-title { font-size: 24px; font-weight: bold; margin-bottom: 20px; }
.tabs { display: flex; gap: 8px; margin-bottom: 24px; flex-wrap: wrap; }
.tab {
  padding: 8px 20px; border: 1px solid #ddd; border-radius: 20px;
  background: white; font-size: 14px; cursor: pointer; position: relative;
}
.tab.active { background: #e74c3c; color: white; border-color: #e74c3c; }
.tab-count {
  position: absolute; top: -6px; right: -6px;
  background: #e74c3c; color: white; border-radius: 50%;
  width: 18px; height: 18px; display: flex; align-items: center; justify-content: center;
  font-size: 11px;
}
.tab.active .tab-count { background: white; color: #e74c3c; }
.group-section { margin-bottom: 32px; }
.group-title { font-size: 18px; font-weight: 600; margin-bottom: 14px; color: #333; }
.tasks-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 16px;
}
.task-card {
  border-radius: 12px; padding: 18px; border: 2px solid transparent;
  background: white; box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  display: flex; flex-direction: column; gap: 10px;
}
.task-card.pending { border-color: #f39c12; background: rgba(243,156,18,0.06); }
.task-card.cooking { border-color: #3498db; background: rgba(52,152,219,0.06); }
.task-card.done   { border-color: #27ae60; background: rgba(39,174,96,0.06); opacity: 0.85; }
.task-header { display: flex; justify-content: space-between; align-items: center; }
.table-badge {
  background: #e74c3c; color: white; padding: 3px 10px; border-radius: 20px;
  font-size: 13px; font-weight: bold;
}
.task-time { color: #999; font-size: 13px; }
.dish-name { font-size: 18px; font-weight: bold; color: #222; }
.dish-qty { font-size: 16px; color: #f39c12; font-weight: bold; }
.task-footer { margin-top: 4px; }
.action-btn {
  width: 100%; padding: 10px; font-size: 14px; font-weight: bold;
  border: none; border-radius: 8px; cursor: pointer;
}
.cooking-btn { background: #f39c12; color: white; }
.done-btn { background: #27ae60; color: white; }
.done-label { font-size: 14px; color: #27ae60; }
.empty-state { text-align: center; padding: 60px; font-size: 18px; color: #999; }
</style>
