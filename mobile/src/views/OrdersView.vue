<template>
  <div class="orders-page">
    <h2 class="page-title">我的订单</h2>
    <div v-if="orders.length === 0" class="empty-state">
      <p>暂无订单</p>
      <router-link :to="`/table/${tableStore.tableId}`" class="go-order-btn">去点餐</router-link>
    </div>
    <div v-for="order in orders" :key="order.id" class="order-card">
      <div class="order-header">
        <span class="order-id">订单 #{{ order.id }}</span>
        <span :class="['order-status', order.status]">{{ statusLabel(order.status) }}</span>
      </div>
      <div class="order-items">
        <div v-for="item in order.items" :key="item.id" class="order-item-row">
          <span>{{ item.dish_name }} × {{ item.quantity }}</span>
          <span>¥{{ item.subtotal.toFixed(2) }}</span>
        </div>
      </div>
      <div class="order-footer">
        <span>共 {{ order.items?.length || 0 }} 种菜品</span>
        <span class="total">合计：¥{{ order.total_amount.toFixed(2) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useTableStore } from '../stores/table.js'

const tableStore = useTableStore()
const orders = ref([])

const statusMap = {
  pending: '待确认', confirmed: '已确认', cooking: '制作中',
  served: '已上菜', paid: '已结账', cancelled: '已取消'
}
function statusLabel(s) { return statusMap[s] || s }

onMounted(async () => {
  if (tableStore.tableId) {
    const res = await axios.get(`/api/orders?table_id=${tableStore.tableId}`)
    orders.value = res.data
  }
})
</script>

<style scoped>
.orders-page { padding: 16px; }
.page-title { font-size: 20px; font-weight: bold; margin-bottom: 16px; }
.empty-state { text-align: center; padding: 60px 0; color: #999; }
.go-order-btn {
  display: inline-block; margin-top: 16px;
  background: #e74c3c; color: white; padding: 10px 24px; border-radius: 20px; text-decoration: none;
}
.order-card {
  background: white; border-radius: 12px; padding: 16px; margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.order-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.order-id { font-weight: bold; color: #333; }
.order-status { padding: 4px 10px; border-radius: 12px; font-size: 12px; }
.order-status.pending { background: #fff3cd; color: #856404; }
.order-status.confirmed { background: #d1ecf1; color: #0c5460; }
.order-status.cooking { background: #cce5ff; color: #004085; }
.order-status.served { background: #d4edda; color: #155724; }
.order-status.paid { background: #f8f9fa; color: #6c757d; }
.order-status.cancelled { background: #f8d7da; color: #721c24; }
.order-item-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 14px; border-bottom: 1px solid #f5f5f5; }
.order-footer { display: flex; justify-content: space-between; margin-top: 12px; font-size: 14px; color: #666; }
.total { color: #e74c3c; font-weight: bold; font-size: 15px; }
</style>
