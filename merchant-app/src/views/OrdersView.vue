<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>订单列表</h2>
      <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchOrders">刷新</v-btn>
    </div>
    <v-row class="mb-3">
      <v-col cols="12" sm="3">
        <v-select
          v-model="filterStatus"
          :items="statusOptions"
          label="状态筛选"
          clearable
          density="compact"
          variant="outlined"
          @update:model-value="fetchOrders"
        />
      </v-col>
      <v-col cols="12" sm="3">
        <v-text-field
          v-model="filterDate"
          label="日期筛选"
          type="date"
          density="compact"
          variant="outlined"
          clearable
          @update:model-value="fetchOrders"
        />
      </v-col>
    </v-row>

    <v-data-table
      :headers="headers"
      :items="orders"
      :loading="loading"
      item-value="id"
    >
      <template #item.total_amount="{ item }">
        ¥{{ (item.total_amount / 100).toFixed(2) }}
      </template>
      <template #item.status="{ item }">
        <v-chip :color="statusColor(item.status)" size="small">{{ statusLabel(item.status) }}</v-chip>
      </template>
      <template #item.created_at="{ item }">
        {{ formatTime(item.created_at) }}
      </template>
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-eye" variant="text" @click="viewOrder(item)" />
        <v-btn size="small" icon="mdi-update" variant="text" @click="openStatusDialog(item)" />
        <v-btn size="small" icon="mdi-cash" variant="text" @click="recordPayment(item)" />
        <v-btn size="small" icon="mdi-send" variant="text" @click="openDispatch(item)" />
      </template>
    </v-data-table>

    <!-- Order Detail Dialog -->
    <v-dialog v-model="detailDialog" max-width="500">
      <v-card v-if="selectedOrder">
        <v-card-title>订单 #{{ selectedOrder.id }}</v-card-title>
        <v-card-text>
          <v-list density="compact">
            <v-list-item title="桌号" :subtitle="selectedOrder.table_id" />
            <v-list-item title="状态" :subtitle="statusLabel(selectedOrder.status)" />
            <v-list-item title="金额" :subtitle="`¥${(selectedOrder.total_amount/100).toFixed(2)}`" />
            <v-divider class="my-2" />
            <v-list-item v-for="item in selectedOrder.items" :key="item.id"
              :title="item.dish?.name || '未知'"
              :subtitle="`x${item.quantity}  备注: ${item.note || '无'}`"
            />
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-spacer /><v-btn @click="detailDialog = false">关闭</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Status Update Dialog -->
    <v-dialog v-model="statusDialog" max-width="360">
      <v-card>
        <v-card-title>更新订单状态</v-card-title>
        <v-card-text>
          <v-select v-model="newStatus" :items="statusOptions" label="新状态" variant="outlined" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="statusDialog = false">取消</v-btn>
          <v-btn color="primary" @click="updateStatus">确定</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Dispatch Dialog -->
    <v-dialog v-model="dispatchDialog" max-width="400">
      <v-card>
        <v-card-title>派发任务</v-card-title>
        <v-card-text>
          <v-select v-model="selectedChef" :items="chefs" item-title="name" item-value="chef_id" label="选择厨师 (手动)" clearable variant="outlined" class="mb-3" />
          <v-btn block color="success" @click="autoDispatch" :loading="dispatching">自动派发</v-btn>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dispatchDialog = false">取消</v-btn>
          <v-btn color="primary" @click="manualDispatch" :disabled="!selectedChef">手动派发</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const orders = ref([])
const loading = ref(true)
const filterStatus = ref('')
const filterDate = ref('')
const detailDialog = ref(false)
const statusDialog = ref(false)
const dispatchDialog = ref(false)
const selectedOrder = ref(null)
const newStatus = ref('')
const chefs = ref([])
const selectedChef = ref(null)
const dispatching = ref(false)

const headers = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '桌号', key: 'table_id', width: 80 },
  { title: '金额', key: 'total_amount', width: 100 },
  { title: '状态', key: 'status', width: 120 },
  { title: '下单时间', key: 'created_at' },
  { title: '操作', key: 'actions', sortable: false, width: 180 },
]

const statusOptions = ['pending', 'cooking', 'done', 'paid']
const statusLabelMap = { pending: '待处理', cooking: '烹饪中', done: '完成', paid: '已结账' }
const statusColorMap = { pending: 'warning', cooking: 'info', done: 'success', paid: 'grey' }

function statusLabel(s) { return statusLabelMap[s] || s }
function statusColor(s) { return statusColorMap[s] || 'default' }
function formatTime(ts) { return ts ? new Date(ts).toLocaleString('zh-CN') : '' }

async function fetchOrders() {
  loading.value = true
  try {
    const params = {}
    if (filterStatus.value) params.status = filterStatus.value
    if (filterDate.value) params.date = filterDate.value
    const res = await api.get('/orders', { params })
    orders.value = res.data || []
  } catch {
    orders.value = []
  } finally {
    loading.value = false
  }
}

function viewOrder(item) {
  selectedOrder.value = item
  detailDialog.value = true
}

function openStatusDialog(item) {
  selectedOrder.value = item
  newStatus.value = item.status
  statusDialog.value = true
}

async function updateStatus() {
  try {
    await api.put(`/orders/${selectedOrder.value.id}/status`, { status: newStatus.value })
    await fetchOrders()
    statusDialog.value = false
  } catch {}
}

async function recordPayment(item) {
  try {
    await api.post(`/orders/${item.id}/pay`)
    await fetchOrders()
  } catch {}
}

async function openDispatch(item) {
  selectedOrder.value = item
  selectedChef.value = null
  dispatchDialog.value = true
  try {
    const res = await api.get('/chefs')
    chefs.value = res.data || []
  } catch {}
}

async function autoDispatch() {
  dispatching.value = true
  try {
    await api.post(`/orders/${selectedOrder.value.id}/dispatch`)
    dispatchDialog.value = false
  } catch {} finally {
    dispatching.value = false
  }
}

async function manualDispatch() {
  try {
    await api.post(`/orders/${selectedOrder.value.id}/dispatch`, { chef_id: selectedChef.value })
    dispatchDialog.value = false
  } catch {}
}

onMounted(fetchOrders)
</script>
