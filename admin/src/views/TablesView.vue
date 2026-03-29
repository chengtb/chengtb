<template>
  <div>
    <h1 class="page-title">桌台总览</h1>
    <div class="tables-grid">
      <div
        v-for="table in tables"
        :key="table.id"
        :class="['table-card', table.status]"
        @click="selectTable(table)"
      >
        <div class="table-number">{{ table.number }}</div>
        <div class="table-name">{{ table.name }}</div>
        <div class="table-capacity">{{ table.capacity }}人桌</div>
        <div :class="['table-status-badge', table.status]">{{ statusLabel(table.status) }}</div>
      </div>
    </div>

    <!-- Table Detail Modal -->
    <div v-if="selectedTable" class="modal-overlay" @click.self="selectedTable = null">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ selectedTable.name }}</h2>
          <button @click="selectedTable = null" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="table-status-row">
            <span>状态：{{ statusLabel(selectedTable.status) }}</span>
            <a :href="`/api/qrcode/${selectedTable.id}`" target="_blank" class="qr-btn">下载二维码</a>
          </div>
          <div v-if="tableOrders.length > 0">
            <h3 style="margin: 16px 0 8px">当前订单</h3>
            <div v-for="order in tableOrders" :key="order.id" class="order-summary">
              <div class="order-sum-header">
                <span>订单 #{{ order.id }}</span>
                <span :class="['status-badge', order.status]">{{ orderStatusLabel(order.status) }}</span>
              </div>
              <div v-for="item in order.items" :key="item.id" class="order-sum-item">
                {{ item.dish_name }} × {{ item.quantity }} = ¥{{ item.subtotal.toFixed(2) }}
              </div>
              <div class="order-sum-total">合计：¥{{ order.total_amount.toFixed(2) }}</div>
              <button
                v-if="order.status !== 'paid' && order.status !== 'cancelled'"
                @click="markPaid(order)"
                class="pay-btn"
              >结账</button>
            </div>
          </div>
          <div v-else class="empty-orders">当前无订单</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const tables = ref([])
const selectedTable = ref(null)
const tableOrders = ref([])

const statusLabel = s => ({ available: '空闲', occupied: '使用中', reserved: '已预订' }[s] || s)
const orderStatusLabel = s => ({ pending: '待确认', confirmed: '已确认', cooking: '制作中', served: '已上菜', paid: '已结账', cancelled: '已取消' }[s] || s)

async function loadTables() {
  const res = await axios.get('/api/tables')
  tables.value = res.data
}

async function selectTable(table) {
  selectedTable.value = table
  const res = await axios.get(`/api/orders?table_id=${table.id}`)
  tableOrders.value = res.data.filter(o => o.status !== 'paid' && o.status !== 'cancelled')
}

async function markPaid(order) {
  await axios.put(`/api/orders/${order.id}/status`, { status: 'paid' })
  await loadTables()
  await selectTable(selectedTable.value)
}

onMounted(loadTables)
</script>

<style scoped>
.page-title { font-size: 24px; font-weight: bold; margin-bottom: 24px; }
.tables-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 16px; }
.table-card {
  background: white; border-radius: 12px; padding: 20px; text-align: center;
  cursor: pointer; border: 2px solid transparent; transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.table-card:hover { transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0,0,0,0.12); }
.table-card.available { border-color: #27ae60; }
.table-card.occupied { border-color: #e74c3c; }
.table-card.reserved { border-color: #f39c12; }
.table-number { font-size: 28px; font-weight: bold; margin-bottom: 4px; }
.table-name { font-size: 13px; color: #666; margin-bottom: 4px; }
.table-capacity { font-size: 12px; color: #999; margin-bottom: 8px; }
.table-status-badge { padding: 4px 10px; border-radius: 12px; font-size: 12px; font-weight: bold; }
.table-status-badge.available { background: #d4edda; color: #155724; }
.table-status-badge.occupied { background: #f8d7da; color: #721c24; }
.table-status-badge.reserved { background: #fff3cd; color: #856404; }
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100;
  display: flex; align-items: center; justify-content: center;
}
.modal {
  background: white; border-radius: 16px; width: 90%; max-width: 500px;
  max-height: 80vh; display: flex; flex-direction: column;
}
.modal-header {
  padding: 20px; border-bottom: 1px solid #eee;
  display: flex; justify-content: space-between; align-items: center;
}
.modal-header h2 { font-size: 20px; }
.close-btn { background: none; border: none; font-size: 20px; color: #666; }
.modal-body { padding: 20px; overflow-y: auto; }
.table-status-row { display: flex; justify-content: space-between; align-items: center; }
.qr-btn {
  background: #3498db; color: white; padding: 8px 16px; border-radius: 8px;
  text-decoration: none; font-size: 14px;
}
.order-summary { background: #f8f9fa; border-radius: 8px; padding: 16px; margin-bottom: 12px; }
.order-sum-header { display: flex; justify-content: space-between; margin-bottom: 8px; font-weight: bold; }
.status-badge { padding: 2px 8px; border-radius: 10px; font-size: 12px; }
.status-badge.pending { background: #fff3cd; color: #856404; }
.status-badge.cooking { background: #cce5ff; color: #004085; }
.status-badge.served { background: #d4edda; color: #155724; }
.order-sum-item { font-size: 14px; color: #555; padding: 4px 0; }
.order-sum-total { font-weight: bold; color: #e74c3c; margin-top: 8px; border-top: 1px solid #eee; padding-top: 8px; }
.pay-btn {
  background: #27ae60; color: white; border: none; padding: 8px 20px;
  border-radius: 8px; margin-top: 10px; font-size: 14px;
}
.empty-orders { text-align: center; color: #999; padding: 24px; }
</style>
