<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title class="d-flex align-center">
            订单列表
            <v-spacer />
            <v-chip-group v-model="selectedStatus" selected-class="text-primary" mandatory>
              <v-chip value="">全部</v-chip>
              <v-chip value="pending" color="warning">待确认</v-chip>
              <v-chip value="confirmed" color="primary">已确认</v-chip>
              <v-chip value="cooking" color="info">烹饪中</v-chip>
              <v-chip value="completed" color="success">已完成</v-chip>
            </v-chip-group>
          </v-card-title>

          <v-data-table
            :headers="headers"
            :items="orders"
            :loading="loading"
            class="elevation-0"
          >
            <template #item.status="{ item }">
              <v-chip :color="statusColor(item.status)" size="small">
                {{ statusText(item.status) }}
              </v-chip>
            </template>
            <template #item.total_amount="{ item }">
              ¥{{ item.total_amount?.toFixed(2) }}
            </template>
            <template #item.actions="{ item }">
              <v-btn
                v-if="item.status === 'pending'"
                color="primary"
                size="small"
                @click="handleConfirm(item)"
              >
                确认订单
              </v-btn>
              <v-btn size="small" variant="text" @click="viewDetail(item)">
                查看详情
              </v-btn>
            </template>
          </v-data-table>
        </v-card>
      </v-col>
    </v-row>

    <!-- Order Detail Dialog -->
    <v-dialog v-model="showDetail" max-width="600">
      <v-card v-if="currentOrder">
        <v-card-title>订单详情 - {{ currentOrder.order_sn }}</v-card-title>
        <v-card-text>
          <v-table>
            <thead>
              <tr>
                <th>商品</th>
                <th>规格</th>
                <th>单价</th>
                <th>数量</th>
                <th>小计</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in currentOrder.items" :key="item.id">
                <td>{{ item.product?.name || '商品' }}</td>
                <td>{{ formatSpecs(item.spec_detail) }}</td>
                <td>¥{{ item.unit_price?.toFixed(2) }}</td>
                <td>{{ item.quantity }}</td>
                <td>¥{{ (item.unit_price * item.quantity).toFixed(2) }}</td>
                <td>
                  <v-btn
                    v-if="item.status === 'pending'"
                    icon="mdi-delete"
                    size="x-small"
                    color="error"
                    variant="text"
                    @click="refundItem(item)"
                  />
                </td>
              </tr>
            </tbody>
          </v-table>
          <div class="text-right mt-4">
            <strong>合计：¥{{ currentOrder.total_amount?.toFixed(2) }}</strong>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="showDetail = false">关闭</v-btn>
          <v-btn
            v-if="currentOrder.status === 'pending'"
            color="primary"
            @click="handleConfirm(currentOrder)"
          >
            确认订单
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { listOrders, getOrder, confirmOrder } from '../api/index.js'

const orders = ref([])
const loading = ref(false)
const selectedStatus = ref('')
const showDetail = ref(false)
const currentOrder = ref(null)
const WAITER_ID = 1 // TODO: get from auth

const headers = [
  { title: '订单号', key: 'order_sn' },
  { title: '桌号', key: 'table.table_number', value: item => item.table?.table_number || '-' },
  { title: '状态', key: 'status' },
  { title: '金额', key: 'total_amount' },
  { title: '时间', key: 'created_at', value: item => new Date(item.created_at).toLocaleString() },
  { title: '操作', key: 'actions', sortable: false },
]

onMounted(() => loadOrders())

watch(selectedStatus, () => loadOrders())

async function loadOrders() {
  loading.value = true
  try {
    const res = await listOrders({ status: selectedStatus.value, page: 1, page_size: 50 })
    orders.value = res.data?.orders || []
  } catch (e) {
    console.error('Load orders failed', e)
  }
  loading.value = false
}

async function viewDetail(order) {
  try {
    const res = await getOrder(order.id)
    currentOrder.value = res.data
    showDetail.value = true
  } catch (e) {
    console.error('Load detail failed', e)
  }
}

async function handleConfirm(order) {
  try {
    await confirmOrder(order.id, WAITER_ID)
    await loadOrders()
    showDetail.value = false
  } catch (e) {
    console.error('Confirm failed', e)
  }
}

function statusColor(status) {
  const map = { pending: 'warning', confirmed: 'primary', cooking: 'info', completed: 'success', paid: 'success', cancelled: 'error' }
  return map[status] || 'default'
}

function statusText(status) {
  const map = { pending: '待确认', confirmed: '已确认', cooking: '烹饪中', completed: '已完成', paid: '已支付', cancelled: '已取消' }
  return map[status] || status
}

function formatSpecs(specDetail) {
  if (!specDetail) return '-'
  try {
    const specs = typeof specDetail === 'string' ? JSON.parse(specDetail) : specDetail
    return specs.map(s => s.spec_value || s.value).join(', ') || '-'
  } catch {
    return '-'
  }
}

function refundItem(item) {
  // TODO: implement refund dialog
  console.log('Refund item', item)
}
</script>
