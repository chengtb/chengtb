<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>烹饪任务</h2>
      <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchTasks">刷新</v-btn>
    </div>
    <v-row class="mb-3">
      <v-col cols="12" sm="3">
        <v-select
          v-model="filterStatus"
          :items="['pending', 'cooking', 'done']"
          label="状态"
          clearable
          density="compact"
          variant="outlined"
          @update:model-value="fetchTasks"
        />
      </v-col>
      <v-col cols="12" sm="3">
        <v-select
          v-model="filterChef"
          :items="chefs"
          item-title="name"
          item-value="chef_id"
          @update:model-value="fetchTasks"
        />
      </v-col>
    </v-row>

    <v-data-table :headers="headers" :items="tasks" :loading="loading" item-value="id">
      <template #item.status="{ item }">
        <v-chip :color="taskStatusColor(item.status)" size="small">{{ taskStatusLabel(item.status) }}</v-chip>
      </template>
      <template #item.table_ids="{ item }">
        {{ (item.table_ids || []).join(', ') }}
      </template>
      <template #item.created_at="{ item }">{{ formatTime(item.created_at) }}</template>
      <template #item.completed_at="{ item }">{{ formatTime(item.completed_at) }}</template>
    </v-data-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const tasks = ref([])
const chefs = ref([])
const loading = ref(true)
const filterStatus = ref('')
const filterChef = ref(null)

const headers = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '厨师', key: 'chef_name', width: 100 },
  { title: '菜谱', key: 'recipe_name' },
  { title: '份数', key: 'total_portions', width: 80 },
  { title: '桌号', key: 'table_ids', width: 120 },
  { title: '状态', key: 'status', width: 100 },
  { title: '创建时间', key: 'created_at' },
  { title: '完成时间', key: 'completed_at' },
]

const statusLabelMap = { pending: '待烹饪', cooking: '烹饪中', done: '已完成' }
const statusColorMap = { pending: 'warning', cooking: 'info', done: 'success' }

function taskStatusLabel(s) { return statusLabelMap[s] || s }
function taskStatusColor(s) { return statusColorMap[s] || 'default' }
function formatTime(ts) { return ts ? new Date(ts).toLocaleString('zh-CN') : '-' }

async function fetchTasks() {
  loading.value = true
  try {
    const params = {}
    if (filterStatus.value) params.status = filterStatus.value
    if (filterChef.value) params.chef_id = filterChef.value
    const res = await api.get('/tasks', { params })
    tasks.value = (res.data || []).map(t => ({
      ...t,
      chef_name: t.chef?.name || t.chef_id,
      recipe_name: t.recipe?.name || t.recipe_id,
    }))
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await fetchTasks()
  try {
    const res = await api.get('/chefs')
    chefs.value = res.data || []
  } catch {}
})
</script>
