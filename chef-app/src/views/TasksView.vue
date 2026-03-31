<template>
  <v-container>
    <div class="d-flex align-center mb-4 gap-2">
      <h2>我的任务</h2>
      <v-spacer />
      <v-btn-toggle v-model="tab" mandatory density="compact" color="primary">
        <v-btn value="pending">待处理</v-btn>
        <v-btn value="done">已完成</v-btn>
      </v-btn-toggle>
    </div>

    <div v-if="loading" class="text-center pa-8">
      <v-progress-circular indeterminate size="48" color="primary" />
    </div>
    <div v-else>
      <v-empty-state v-if="filteredTasks.length === 0" title="暂无任务" icon="mdi-chef-hat" />
      <v-row>
        <v-col v-for="task in filteredTasks" :key="task.task_id" cols="12" sm="6" md="4">
          <v-card :color="task.status === 'pending' ? '' : 'grey-lighten-3'" elevation="2">
            <v-card-title class="text-h6">
              {{ task.recipe?.name || task.recipe_id }}
              <v-chip size="small" class="ml-2" color="primary">x{{ task.total_portion }}</v-chip>
            </v-card-title>
            <v-card-subtitle>{{ task.dish_name || '' }}</v-card-subtitle>
            <v-card-text>
              <div class="d-flex align-center gap-2 mb-1">
                <v-icon size="small">mdi-table-chair</v-icon>
                <span>桌号: {{ (task.table_ids || []).join(', ') || '未知' }}</span>
              </div>
              <div class="d-flex align-center gap-2 mb-1">
                <v-icon size="small">mdi-clock-outline</v-icon>
                <span>{{ formatTime(task.created_at) }}</span>
              </div>
              <div v-if="task.notes" class="d-flex align-center gap-2">
                <v-icon size="small">mdi-note-text</v-icon>
                <span>{{ task.notes }}</span>
              </div>
              <v-chip size="small" :color="statusColor(task.status)" class="mt-2">
                {{ statusLabel(task.status) }}
              </v-chip>
            </v-card-text>
            <v-card-actions v-if="task.status === 'pending' || task.status === 'cooking'">
              <v-spacer />
              <v-btn
                color="success"
                variant="elevated"
                prepend-icon="mdi-check-circle"
                @click="completeTask(task)"
                :loading="completing === task.task_id"
              >
                完成烹饪
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>
    </div>

    <v-snackbar v-model="snackbar" color="success" timeout="3000">任务已完成！</v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import api from '../api'

const tasks = ref([])
const loading = ref(true)
const tab = ref('pending')
const completing = ref(null)
const snackbar = ref(false)
let timer = null

const filteredTasks = computed(() => tasks.value.filter(t => {
  if (tab.value === 'pending') return t.status === 'pending' || t.status === 'cooking'
  return t.status === 'done'
}))

const statusLabelMap = { pending: '待烹饪', cooking: '烹饪中', done: '已完成' }
const statusColorMap = { pending: 'warning', cooking: 'info', done: 'success' }

function statusLabel(s) { return statusLabelMap[s] || s }
function statusColor(s) { return statusColorMap[s] || 'default' }
function formatTime(ts) { return ts ? new Date(ts).toLocaleString('zh-CN') : '' }

async function fetchTasks() {
  try {
    const res = await api.get('/tasks')
    tasks.value = res.data || []
  } catch {
    tasks.value = []
  } finally {
    loading.value = false
  }
}

async function completeTask(task) {
  completing.value = task.task_id
  try {
    await api.put(`/tasks/${task.task_id}/complete`)
    snackbar.value = true
    await fetchTasks()
  } catch {} finally {
    completing.value = null
  }
}

onMounted(() => {
  fetchTasks()
  timer = setInterval(fetchTasks, 10000)
})

onUnmounted(() => clearInterval(timer))
</script>
