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
      v-model:expanded="expanded"
      show-expand
      :headers="headers"
      :items="orders"
      :loading="loading"
      item-value="order_id"
    >
      <template #item.total_amount="{ item }">
        ¥{{ Number(item.total_amount).toFixed(2) }}
      </template>
      <template #item.status="{ item }">
        <v-chip :color="orderStatusColor(item.status)" size="small">{{ orderStatusLabel(item.status) }}</v-chip>
      </template>
      <template #item.created_at="{ item }">
        {{ formatTime(item.created_at) }}
      </template>
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-update" variant="text" title="更新状态" @click.stop="openStatusDialog(item)" />
        <v-btn size="small" icon="mdi-cash" variant="text" title="结账" @click.stop="openPaymentDialog(item)" />
      </template>

      <template #expanded-row="{ columns, item }">
        <tr>
          <td :colspan="columns.length" class="pa-4 bg-grey-lighten-5">
            <div class="text-subtitle-2 mb-2 font-weight-bold">菜品明细</div>
            <v-table density="compact" class="rounded border">
              <thead>
                <tr class="bg-grey-lighten-4">
                  <th class="text-left">菜品</th>
                  <th class="text-left">数量</th>
                  <th class="text-left">备注</th>
                  <th class="text-left">状态</th>
                  <th class="text-left">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="dish in item.items" :key="dish.item_id">
                  <td>{{ dish.dish?.name || '未知' }}</td>
                  <td>x{{ dish.quantity }}</td>
                  <td>{{ dish.note || '—' }}</td>
                  <td>
                    <v-chip :color="itemStatusColor(dish.status)" size="x-small">
                      {{ itemStatusLabel(dish.status) }}
                    </v-chip>
                  </td>
                  <td>
                    <v-btn
                      v-if="dish.status === 'pending'"
                      size="x-small"
                      color="primary"
                      variant="tonal"
                      prepend-icon="mdi-send"
                      @click="openItemDispatch(item, dish)"
                    >派发</v-btn>
                    <span v-else class="text-grey text-caption">—</span>
                  </td>
                </tr>
                <tr v-if="!item.items?.length">
                  <td colspan="5" class="text-center text-grey pa-3">暂无菜品</td>
                </tr>
              </tbody>
            </v-table>
          </td>
        </tr>
      </template>
    </v-data-table>

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

    <!-- Payment Dialog -->
    <v-dialog v-model="paymentDialog" max-width="360">
      <v-card>
        <v-card-title>确认结账</v-card-title>
        <v-card-subtitle v-if="selectedOrder">
          订单 #{{ selectedOrder.order_id }} · ¥{{ Number(selectedOrder.total_amount).toFixed(2) }}
        </v-card-subtitle>
        <v-card-text>
          <v-select v-model="paymentMethod" :items="paymentMethods" label="支付方式" variant="outlined" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="paymentDialog = false">取消</v-btn>
          <v-btn color="primary" :disabled="!paymentMethod" @click="recordPayment">确定</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Item Dispatch Dialog -->
    <v-dialog v-model="dispatchDialog" max-width="400">
      <v-card>
        <v-card-title>派发菜品任务</v-card-title>
        <v-card-subtitle v-if="selectedItem">
          {{ selectedItem.dish?.name || '未知' }} x{{ selectedItem.quantity }}
        </v-card-subtitle>
        <v-card-text>
          <v-select
            v-model="selectedChef"
            :items="chefs"
            item-title="name"
            item-value="chef_id"
            label="指定厨师（可选，留空则自动分配）"
            clearable
            variant="outlined"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dispatchDialog = false">取消</v-btn>
          <v-btn color="primary" :loading="dispatching" @click="dispatchItem">派发</v-btn>
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
const expanded = ref([])

// Order-level dialogs
const statusDialog = ref(false)
const paymentDialog = ref(false)
const selectedOrder = ref(null)
const newStatus = ref('')
const paymentMethod = ref('')
const paymentMethods = ['现金', '微信', '支付宝', '刷卡']

// Item-level dispatch
const dispatchDialog = ref(false)
const selectedItem = ref(null)
const chefs = ref([])
const selectedChef = ref(null)
const dispatching = ref(false)

const headers = [
  { title: 'ID', key: 'order_id', width: 60 },
  { title: '桌号', key: 'table_id', width: 80 },
  { title: '金额', key: 'total_amount', width: 110 },
  { title: '状态', key: 'status', width: 120 },
  { title: '下单时间', key: 'created_at' },
  { title: '操作', key: 'actions', sortable: false, width: 100 },
]

const statusOptions = ['pending', 'cooking', 'completed', 'cancelled']
const orderStatusLabelMap = { pending: '待处理', cooking: '烹饪中', completed: '完成', cancelled: '已取消' }
const orderStatusColorMap = { pending: 'warning', cooking: 'info', completed: 'success', cancelled: 'grey' }
const itemStatusLabelMap = { pending: '待派发', dispatched: '已派发', cooking: '烹饪中', done: '已完成' }
const itemStatusColorMap = { pending: 'warning', dispatched: 'blue', cooking: 'info', done: 'success' }

function orderStatusLabel(s) { return orderStatusLabelMap[s] || s }
function orderStatusColor(s) { return orderStatusColorMap[s] || 'default' }
function itemStatusLabel(s) { return itemStatusLabelMap[s] || s }
function itemStatusColor(s) { return itemStatusColorMap[s] || 'default' }
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

function openStatusDialog(item) {
  selectedOrder.value = item
  newStatus.value = item.status
  statusDialog.value = true
}

async function updateStatus() {
  try {
    await api.put(`/orders/${selectedOrder.value.order_id}/status`, { status: newStatus.value })
    await fetchOrders()
    statusDialog.value = false
  } catch {}
}

function openPaymentDialog(item) {
  selectedOrder.value = item
  paymentMethod.value = ''
  paymentDialog.value = true
}

async function recordPayment() {
  try {
    await api.post(`/orders/${selectedOrder.value.order_id}/payment`, { method: paymentMethod.value })
    await fetchOrders()
    paymentDialog.value = false
  } catch {}
}

async function openItemDispatch(order, dish) {
  selectedOrder.value = order
  selectedItem.value = dish
  selectedChef.value = null
  dispatchDialog.value = true
  try {
    const res = await api.get('/chefs')
    chefs.value = res.data || []
  } catch {}
}

async function dispatchItem() {
  dispatching.value = true
  try {
    const body = selectedChef.value ? { chef_id: selectedChef.value } : {}
    await api.post(
      `/orders/${selectedOrder.value.order_id}/items/${selectedItem.value.item_id}/dispatch`,
      body
    )
    await fetchOrders()
    dispatchDialog.value = false
  } catch {} finally {
    dispatching.value = false
  }
}

onMounted(fetchOrders)
</script>

