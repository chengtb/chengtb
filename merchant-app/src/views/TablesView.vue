<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>餐桌管理</h2>
      <div class="d-flex gap-2">
        <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchTables">刷新</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加餐桌</v-btn>
      </div>
    </div>

    <v-row v-if="loading">
      <v-col cols="12" class="text-center">
        <v-progress-circular indeterminate color="primary" />
      </v-col>
    </v-row>
    <v-row v-else>
      <v-col v-for="table in tables" :key="table.table_id" cols="6" sm="4" md="3" lg="2">
        <v-card
          :color="statusColor(table.status)"
          dark
          class="table-card"
        >
          <v-card-title class="text-center text-h5" style="cursor:pointer" @click="openOrderDetail(table)">
            {{ table.table_no }}
          </v-card-title>
          <v-card-subtitle class="text-center">{{ statusLabel(table.status) }} · {{ table.capacity }}人</v-card-subtitle>
          <v-card-actions class="justify-center pa-1">
            <v-btn size="x-small" icon="mdi-pencil" variant="text" @click.stop="openEdit(table)" />
            <v-btn size="x-small" icon="mdi-delete" variant="text" color="error" @click.stop="deleteTable(table)" />
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
    <v-empty-state v-if="!loading && tables.length === 0" title="暂无餐桌数据" />

    <!-- Add/Edit Dialog -->
    <v-dialog v-model="formDialog" max-width="400">
      <v-card>
        <v-card-title>{{ editing ? '编辑餐桌' : '添加餐桌' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.table_no" label="桌号" variant="outlined" class="mb-3" />
          <v-text-field v-model.number="form.capacity" label="容纳人数" type="number" variant="outlined" class="mb-3" />
          <v-select
            v-model="form.status"
            :items="statusOptions"
            item-title="title"
            item-value="value"
            label="状态"
            variant="outlined"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="formDialog = false">取消</v-btn>
          <v-btn color="primary" @click="saveTable">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Order Detail Dialog -->
    <v-dialog v-model="orderDialog" max-width="600">
      <v-card>
        <v-card-title>桌号 {{ selectedTable?.table_no }} - 订单详情</v-card-title>
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
          <v-btn @click="orderDialog = false">关闭</v-btn>
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
const formDialog = ref(false)
const orderDialog = ref(false)
const editing = ref(false)
const form = ref({ table_id: null, table_no: '', capacity: 4, status: 'idle' })
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
const statusOptions = [
  { title: '空闲', value: 'idle' },
  { title: '点餐中', value: 'ordering' },
  { title: '等待上菜', value: 'waiting' },
  { title: '就餐中', value: 'dining' },
  { title: '待结账', value: 'checkout' },
]

function statusColor(s) { return statusColorMap[s] || 'grey' }
function statusLabel(s) { return statusLabelMap[s] || s }

async function fetchTables() {
  loading.value = true
  try {
    const res = await api.get('/tables')
    tables.value = res.data || []
  } catch {
    tables.value = []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editing.value = false
  form.value = { table_id: null, table_no: '', capacity: 4, status: 'idle' }
  formDialog.value = true
}

function openEdit(table) {
  editing.value = true
  form.value = { table_id: table.table_id, table_no: table.table_no, capacity: table.capacity, status: table.status }
  formDialog.value = true
}

async function saveTable() {
  try {
    if (editing.value) {
      await api.put(`/tables/${form.value.table_id}`, form.value)
    } else {
      await api.post('/tables', form.value)
    }
    formDialog.value = false
    await fetchTables()
  } catch {}
}

async function deleteTable(table) {
  if (!confirm(`确认删除桌号 ${table.table_no}?`)) return
  try {
    await api.delete(`/tables/${table.table_id}`)
    await fetchTables()
  } catch {}
}

async function openOrderDetail(table) {
  selectedTable.value = table
  orderDialog.value = true
  loadingOrders.value = true
  try {
    const res = await api.get(`/tables/${table.table_id}/order`)
    tableOrders.value = res.data ? [res.data] : []
  } catch {
    tableOrders.value = []
  } finally {
    loadingOrders.value = false
  }
}

onMounted(fetchTables)
</script>
