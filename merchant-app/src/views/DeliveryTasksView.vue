<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>传菜任务</h2>
      <div class="d-flex gap-2">
        <v-select
          v-model="statusFilter"
          :items="statusOptions"
          item-title="label"
          item-value="value"
          label="状态筛选"
          variant="outlined"
          density="compact"
          hide-details
          style="width: 150px"
          @update:model-value="fetchTasks"
        />
        <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchTasks">刷新</v-btn>
      </div>
    </div>

    <v-data-table
      :headers="headers"
      :items="tasks"
      :loading="loading"
      item-value="task_id"
    >
      <template #item.table_ids="{ item }">
        <v-chip color="orange-darken-2" size="small">
          桌 {{ tableLabel(item.table_ids) }}
        </v-chip>
      </template>

      <template #item.dish="{ item }">
        {{ item.cooking_task?.dish?.name || item.cooking_task?.recipe?.name || '—' }}
        <v-chip size="x-small" class="ml-1">x{{ item.cooking_task?.total_portion }}</v-chip>
      </template>

      <template #item.waiter="{ item }">
        <span v-if="item.waiter">{{ item.waiter.name }}</span>
        <span v-else class="text-grey text-caption">未领取</span>
      </template>

      <template #item.status="{ item }">
        <v-chip :color="statusColor(item.status)" size="small">{{ statusLabel(item.status) }}</v-chip>
      </template>

      <template #item.reject_reason="{ item }">
        <span v-if="item.reject_reason" class="text-error text-caption">{{ item.reject_reason }}</span>
        <span v-else>—</span>
      </template>

      <template #item.created_at="{ item }">
        {{ formatTime(item.created_at) }}
      </template>

      <template #item.actions="{ item }">
        <v-btn
          v-if="item.status === 'pending' || item.status === 'delivering'"
          size="small"
          color="error"
          variant="tonal"
          prepend-icon="mdi-undo"
          @click="openReturnDialog(item)"
        >
          退菜
        </v-btn>
      </template>
    </v-data-table>

    <!-- Return dish dialog -->
    <v-dialog v-model="returnDialog" max-width="420">
      <v-card rounded="xl">
        <v-card-title class="pa-5">
          <v-icon color="error" class="mr-2">mdi-undo</v-icon>
          发起退菜指令
        </v-card-title>
        <v-card-text>
          <div class="mb-3 text-body-2">
            传菜任务 #{{ returnTarget?.task_id }} —
            桌 {{ tableLabel(returnTarget?.table_ids) }}
          </div>
          <v-select
            v-model="returnReason"
            :items="returnReasons"
            label="退菜原因"
            variant="outlined"
            density="comfortable"
          />
        </v-card-text>
        <v-card-actions class="px-5 pb-5">
          <v-spacer />
          <v-btn variant="text" @click="returnDialog = false">取消</v-btn>
          <v-btn
            color="error"
            variant="elevated"
            rounded="lg"
            :disabled="!returnReason"
            :loading="returning"
            @click="confirmReturn"
          >
            确认退菜
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-snackbar v-model="snackbar" color="success" timeout="3000">退菜指令已下发</v-snackbar>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const tasks = ref([])
const loading = ref(true)
const statusFilter = ref('')
const returnDialog = ref(false)
const returnTarget = ref(null)
const returnReason = ref('')
const returning = ref(false)
const snackbar = ref(false)

const statusOptions = [
  { label: '全部', value: '' },
  { label: '待领取', value: 'pending' },
  { label: '配送中', value: 'delivering' },
  { label: '已上菜', value: 'done' },
  { label: '已退回', value: 'rejected' },
]

const returnReasons = ['菜品损坏', '上错菜', '顾客退菜', '其他']

const headers = [
  { title: '任务ID', key: 'task_id', width: 80 },
  { title: '餐桌', key: 'table_ids', sortable: false },
  { title: '菜品', key: 'dish', sortable: false },
  { title: '传菜员', key: 'waiter', sortable: false },
  { title: '状态', key: 'status', width: 110 },
  { title: '退回原因', key: 'reject_reason', sortable: false },
  { title: '出餐时间', key: 'created_at', width: 140 },
  { title: '操作', key: 'actions', sortable: false, width: 100 },
]

const statusLabelMap = { pending: '待领取', delivering: '配送中', done: '已上菜', rejected: '已退回' }
const statusColorMap = { pending: 'warning', delivering: 'info', done: 'success', rejected: 'error' }
function statusLabel(s) { return statusLabelMap[s] || s }
function statusColor(s) { return statusColorMap[s] || 'default' }
function tableLabel(raw) {
  try { return JSON.parse(raw || '[]').join(', ') || '—' } catch { return '—' }
}
function formatTime(ts) { return ts ? new Date(ts).toLocaleString('zh-CN') : '' }

async function fetchTasks() {
  loading.value = true
  try {
    const params = statusFilter.value ? { status: statusFilter.value } : {}
    const res = await api.get('/delivery-tasks', { params })
    tasks.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openReturnDialog(task) {
  returnTarget.value = task
  returnReason.value = ''
  returnDialog.value = true
}

async function confirmReturn() {
  if (!returnTarget.value || !returnReason.value) return
  returning.value = true
  try {
    await api.put(`/delivery-tasks/${returnTarget.value.task_id}/return`, { reason: returnReason.value })
    returnDialog.value = false
    snackbar.value = true
    await fetchTasks()
  } finally {
    returning.value = false
  }
}

onMounted(fetchTasks)
</script>
