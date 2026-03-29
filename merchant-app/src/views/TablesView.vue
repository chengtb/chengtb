<template>
  <div>
    <h2 class="mb-4">餐桌概览</h2>
    <v-row v-if="loading">
      <v-col cols="12" class="text-center">
        <v-progress-circular indeterminate color="primary" />
      </v-col>
    </v-row>
    <v-row v-else>
      <v-col v-for="table in tables" :key="table.id" cols="6" sm="4" md="3" lg="2">
        <v-card
          :color="statusColor(table.status)"
          dark
          @click="openTable(table)"
          class="table-card"
          style="cursor:pointer"
        >
          <v-card-title class="text-center text-h5">{{ table.table_number }}</v-card-title>
          <v-card-subtitle class="text-center">{{ statusLabel(table.status) }}</v-card-subtitle>
        </v-card>
      </v-col>
    </v-row>
    <v-empty-state v-if="!loading && tables.length === 0" title="暂无餐桌数据" />

    <!-- Order Detail Dialog -->
    <v-dialog v-model="dialog" max-width="600">
      <v-card>
        <v-card-title>桌号 {{ selectedTable?.table_number }} - 订单详情</v-card-title>
        <v-card-text>
          <div v-if="loadingOrders" class="text-center pa-4">
            <v-progress-circular indeterminate />
          </div>
          <v-list v-else>
            <template v-if="tableOrders.length === 0">
              <v-list-item title="暂无订单" />
            </template>
            <v-list-item
              v-for="order in tableOrders"
              :key="order.id"
              :title="`订单 #${order.id}`"
              :subtitle="`金额: ¥${(order.total_amount/100).toFixed(2)}  状态: ${statusLabel(order.status)}`"
            />
          </v-list>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dialog = false">关闭</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const tables = ref([])
const loading = ref(true)
const dialog = ref(false)
const selectedTable = ref(null)
const tableOrders = ref([])
const loadingOrders = ref(false)

const statusColorMap = {
  idle: 'success',
  ordering: 'warning',
  waiting: 'amber',
  dining: 'primary',
  checkout: 'error',
}
const statusLabelMap = {
  idle: '空闲',
  ordering: '点餐中',
  waiting: '等待上菜',
  dining: '就餐中',
  checkout: '待结账',
  pending: '待处理',
  cooking: '烹饪中',
  done: '完成',
  paid: '已结账',
}

function statusColor(s) { return statusColorMap[s] || 'grey' }
function statusLabel(s) { return statusLabelMap[s] || s }

async function openTable(table) {
  selectedTable.value = table
  dialog.value = true
  loadingOrders.value = true
  try {
    const res = await api.get(`/tables/${table.id}/orders`)
    tableOrders.value = res.data || []
  } catch {
    tableOrders.value = []
  } finally {
    loadingOrders.value = false
  }
}

onMounted(async () => {
  try {
    const res = await api.get('/tables')
    tables.value = res.data || []
  } catch {
    tables.value = []
  } finally {
    loading.value = false
  }
})
</script>
