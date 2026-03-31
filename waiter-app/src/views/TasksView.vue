<template>
  <v-container class="pa-3" style="max-width: 600px; margin: 0 auto;">
    <!-- Header bar -->
    <div class="d-flex align-center mb-3 gap-2">
      <v-btn-toggle v-model="tab" mandatory density="compact" color="orange-darken-2" rounded="pill">
        <v-btn value="pending" size="small">
          <v-badge :content="pendingCount" color="error" inline>待领取</v-badge>
        </v-btn>
        <v-btn value="delivering" size="small">配送中</v-btn>
        <v-btn value="done" size="small">已完成</v-btn>
      </v-btn-toggle>
      <v-spacer />
      <v-btn icon="mdi-refresh" size="small" variant="tonal" @click="fetchTasks" :loading="loading" />
    </div>

    <!-- Area filter chip (only for pending) -->
    <div v-if="tab === 'pending' && waiterArea" class="mb-3">
      <v-chip color="orange-darken-2" size="small" prepend-icon="mdi-map-marker">
        区域: {{ waiterArea }}
      </v-chip>
    </div>

    <div v-if="loading && tasks.length === 0" class="text-center pa-12">
      <v-progress-circular indeterminate color="orange-darken-2" size="56" />
    </div>

    <v-empty-state
      v-else-if="filteredTasks.length === 0"
      :title="emptyText"
      icon="mdi-food-takeout-box"
      class="mt-8"
    />

    <!-- Task cards -->
    <div v-else>
      <v-card
        v-for="task in filteredTasks"
        :key="task.task_id"
        class="mb-4"
        elevation="3"
        rounded="xl"
        :color="cardColor(task.status)"
      >
        <!-- Table number header – large & prominent -->
        <v-card-title class="d-flex align-center pa-4 pb-2">
          <v-icon size="28" color="orange-darken-2" class="mr-2">mdi-table-chair</v-icon>
          <span class="text-h5 font-weight-black text-orange-darken-2">
            桌 {{ tableLabel(task) }}
          </span>
          <v-spacer />
          <v-chip size="small" :color="statusColor(task.status)" class="font-weight-bold">
            {{ statusLabel(task.status) }}
          </v-chip>
        </v-card-title>

        <v-divider />

        <v-card-text class="pa-4">
          <!-- Dish info -->
          <div class="d-flex align-center gap-2 mb-2">
            <v-icon size="20" color="grey">mdi-food</v-icon>
            <span class="text-body-1 font-weight-medium">
              {{ task.cooking_task?.dish?.name || task.cooking_task?.recipe?.name || '—' }}
            </span>
            <v-chip size="x-small" color="primary" class="ml-auto">
              x{{ task.cooking_task?.total_portion }}
            </v-chip>
          </div>

          <!-- Note / special remark highlight -->
          <div v-if="task.cooking_task?.note" class="d-flex align-center gap-2 mb-2">
            <v-chip color="red" size="small" prepend-icon="mdi-alert-circle" variant="tonal">
              {{ task.cooking_task.note }}
            </v-chip>
          </div>

          <!-- Timing -->
          <div class="d-flex align-center gap-2 mb-1 text-caption text-grey">
            <v-icon size="16">mdi-clock-outline</v-icon>
            出餐时间: {{ formatTime(task.created_at) }}
          </div>
          <div v-if="task.pickup_time?.Valid" class="d-flex align-center gap-2 text-caption text-grey">
            <v-icon size="16">mdi-hand-pointing-right</v-icon>
            领取时间: {{ formatTime(task.pickup_time?.Time) }}
          </div>

          <!-- Reject reason -->
          <v-alert
            v-if="task.reject_reason"
            type="error"
            variant="tonal"
            density="compact"
            class="mt-2"
            :text="task.reject_reason"
          />

          <!-- Waiter info (for delivering tasks) -->
          <div v-if="task.waiter && tab === 'delivering'" class="d-flex align-center gap-1 mt-2 text-caption text-grey">
            <v-icon size="16">mdi-account</v-icon>
            {{ task.waiter.name }}
          </div>
        </v-card-text>

        <!-- Action buttons -->
        <v-card-actions class="pa-4 pt-0 gap-2">
          <!-- Pickup button (pending) -->
          <v-btn
            v-if="task.status === 'pending'"
            color="orange-darken-2"
            variant="elevated"
            size="large"
            block
            rounded="lg"
            prepend-icon="mdi-hand-back-right"
            :loading="actionLoading === task.task_id + '_pickup'"
            @click="pickupTask(task)"
          >
            领取任务
          </v-btn>

          <!-- Confirm delivery button (delivering – only my tasks) -->
          <template v-if="task.status === 'delivering' && String(task.waiter_id) === waiterId">
            <v-btn
              color="success"
              variant="elevated"
              size="large"
              rounded="lg"
              class="flex-grow-1"
              prepend-icon="mdi-check-circle"
              :loading="actionLoading === task.task_id + '_deliver'"
              @click="confirmDelivery(task)"
            >
              确认上菜
            </v-btn>
            <v-btn
              color="error"
              variant="outlined"
              size="large"
              rounded="lg"
              prepend-icon="mdi-undo"
              @click="openRejectDialog(task)"
            >
              退回
            </v-btn>
          </template>
        </v-card-actions>
      </v-card>
    </div>

    <!-- Reject dialog -->
    <v-dialog v-model="rejectDialog" max-width="380">
      <v-card rounded="xl">
        <v-card-title class="pa-5 text-h6">退回后厨</v-card-title>
        <v-card-text>
          <v-select
            v-model="rejectReason"
            :items="rejectReasons"
            label="退回原因"
            variant="outlined"
            density="comfortable"
          />
        </v-card-text>
        <v-card-actions class="px-5 pb-5">
          <v-spacer />
          <v-btn variant="text" @click="rejectDialog = false">取消</v-btn>
          <v-btn
            color="error"
            variant="elevated"
            rounded="lg"
            :disabled="!rejectReason"
            :loading="actionLoading === rejectTarget?.task_id + '_reject'"
            @click="confirmReject"
          >
            确认退回
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-snackbar v-model="snackbar.show" :color="snackbar.color" timeout="3000" location="top">
      {{ snackbar.text }}
    </v-snackbar>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import api from '../api'

const tasks = ref([])
const loading = ref(true)
const tab = ref('pending')
const actionLoading = ref(null)
const waiterId = localStorage.getItem('waiter_id') || ''
const waiterArea = localStorage.getItem('waiter_area') || ''

const rejectDialog = ref(false)
const rejectTarget = ref(null)
const rejectReason = ref('')
const rejectReasons = ['菜品损坏', '上错菜', '其他']

const snackbar = ref({ show: false, text: '', color: 'success' })

let pollTimer = null
let ws = null

function showSnack(text, color = 'success') {
  snackbar.value = { show: true, text, color }
}

const pendingCount = computed(() => tasks.value.filter(t => t.status === 'pending').length)

const filteredTasks = computed(() => tasks.value.filter(t => {
  if (tab.value === 'pending') return t.status === 'pending'
  if (tab.value === 'delivering') return t.status === 'delivering'
  return t.status === 'done' || t.status === 'rejected'
}))

const emptyText = computed(() => {
  if (tab.value === 'pending') return '暂无待传菜任务'
  if (tab.value === 'delivering') return '暂无配送中的任务'
  return '暂无已完成的任务'
})

const statusLabelMap = {
  pending: '待领取',
  delivering: '配送中',
  done: '已上菜',
  rejected: '已退回',
}
const statusColorMap = {
  pending: 'warning',
  delivering: 'info',
  done: 'success',
  rejected: 'error',
}
function statusLabel(s) { return statusLabelMap[s] || s }
function statusColor(s) { return statusColorMap[s] || 'default' }
function cardColor(s) {
  if (s === 'done') return 'grey-lighten-4'
  if (s === 'rejected') return 'red-lighten-5'
  return undefined
}

function tableLabel(task) {
  try {
    const ids = JSON.parse(task.table_ids || '[]')
    return ids.join(', ') || '—'
  } catch {
    return '—'
  }
}

function formatTime(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

async function fetchTasks() {
  loading.value = true
  try {
    const res = await api.get('/tasks')
    // Sort: pending first by oldest, then delivering, then done
    const order = { pending: 0, delivering: 1, done: 2, rejected: 3 }
    tasks.value = (res.data || []).sort((a, b) => {
      if (order[a.status] !== order[b.status]) return order[a.status] - order[b.status]
      return new Date(a.created_at) - new Date(b.created_at)
    })
  } catch {
    tasks.value = []
  } finally {
    loading.value = false
  }
}

async function pickupTask(task) {
  actionLoading.value = task.task_id + '_pickup'
  try {
    await api.put(`/tasks/${task.task_id}/pickup`)
    showSnack('已领取任务，请及时送餐！')
    tab.value = 'delivering'
    await fetchTasks()
  } catch (e) {
    const msg = e?.response?.data?.error || '领取失败，任务可能已被他人领取'
    showSnack(msg, 'error')
    await fetchTasks()
  } finally {
    actionLoading.value = null
  }
}

async function confirmDelivery(task) {
  actionLoading.value = task.task_id + '_deliver'
  try {
    await api.put(`/tasks/${task.task_id}/deliver`)
    showSnack('上菜成功！')
    tab.value = 'done'
    await fetchTasks()
  } catch (e) {
    showSnack(e?.response?.data?.error || '操作失败', 'error')
  } finally {
    actionLoading.value = null
  }
}

function openRejectDialog(task) {
  rejectTarget.value = task
  rejectReason.value = ''
  rejectDialog.value = true
}

async function confirmReject() {
  if (!rejectTarget.value || !rejectReason.value) return
  const task = rejectTarget.value
  actionLoading.value = task.task_id + '_reject'
  try {
    await api.put(`/tasks/${task.task_id}/reject`, { reason: rejectReason.value })
    rejectDialog.value = false
    showSnack('已退回后厨')
    tab.value = 'done'
    await fetchTasks()
  } catch (e) {
    showSnack(e?.response?.data?.error || '操作失败', 'error')
  } finally {
    actionLoading.value = null
  }
}

function connectWS() {
  const wsBase = import.meta.env.VITE_WS_BASE || 'ws://localhost:8080'
  const wid = localStorage.getItem('waiter_id')
  if (!wid) return
  ws = new WebSocket(`${wsBase}/ws/waiter/${wid}`)
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.type === 'new_delivery_task') {
        fetchTasks()
        showSnack('新的传菜任务到来！')
      }
    } catch {}
  }
  ws.onclose = () => {
    // Reconnect after 5s
    setTimeout(connectWS, 5000)
  }
}

onMounted(() => {
  fetchTasks()
  pollTimer = setInterval(fetchTasks, 15000)
  connectWS()
})

onUnmounted(() => {
  clearInterval(pollTimer)
  if (ws) ws.close()
})
</script>
