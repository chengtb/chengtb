<template>
  <div>
    <h1 class="page-title">订单管理</h1>
    
    <!-- Status Tabs -->
    <div class="tabs">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        :class="['tab', { active: activeTab === tab.value }]"
        @click="activeTab = tab.value; loadOrders()"
      >{{ tab.label }}</button>
    </div>

    <!-- Orders Table -->
    <div class="orders-table-wrapper">
      <table class="orders-table">
        <thead>
          <tr>
            <th>订单号</th><th>桌台</th><th>下单时间</th><th>菜品数</th><th>金额</th><th>状态</th><th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="order in orders" :key="order.id">
            <td>#{{ order.id }}</td>
            <td>{{ order.table?.number }}</td>
            <td>{{ formatTime(order.created_at) }}</td>
            <td>{{ order.items?.length || 0 }}种</td>
            <td>¥{{ order.total_amount.toFixed(2) }}</td>
            <td><span :class="['status-badge', order.status]">{{ orderStatusLabel(order.status) }}</span></td>
            <td class="actions">
              <button @click="viewOrder(order)" class="btn-view">详情</button>
              <button @click="openDiscount(order)" class="btn-discount">折扣</button>
              <button v-if="order.status !== 'paid' && order.status !== 'cancelled'" @click="changeStatus(order)" class="btn-next">下一步</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Order Detail Modal -->
    <div v-if="detailOrder" class="modal-overlay" @click.self="detailOrder = null">
      <div class="modal">
        <div class="modal-header">
          <h2>订单 #{{ detailOrder.id }} 详情</h2>
          <button @click="detailOrder = null" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <p>桌台：{{ detailOrder.table?.name }}</p>
          <p>状态：{{ orderStatusLabel(detailOrder.status) }}</p>
          <table class="detail-table">
            <thead><tr><th>菜品</th><th>单价</th><th>数量</th><th>小计</th></tr></thead>
            <tbody>
              <tr v-for="item in detailOrder.items" :key="item.id">
                <td>{{ item.dish_name }}</td>
                <td>¥{{ item.dish_price.toFixed(2) }}</td>
                <td>× {{ item.quantity }}</td>
                <td>¥{{ item.subtotal.toFixed(2) }}</td>
              </tr>
            </tbody>
          </table>
          <div class="detail-total">合计：¥{{ detailOrder.total_amount.toFixed(2) }}</div>
        </div>
      </div>
    </div>

    <!-- Discount Modal -->
    <div v-if="discountOrder" class="modal-overlay" @click.self="discountOrder = null">
      <div class="modal small-modal">
        <div class="modal-header">
          <h2>优惠折扣 - 订单 #{{ discountOrder.id }}</h2>
          <button @click="discountOrder = null" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <p>原价：¥{{ discountOrder.total_amount.toFixed(2) }}</p>
          <div class="discount-form">
            <label>折扣类型：
              <select v-model="discountType">
                <option value="fixed">固定金额</option>
                <option value="percent">百分比</option>
              </select>
            </label>
            <label>
              {{ discountType === 'fixed' ? '减免金额 (¥)' : '折扣 (%)' }}：
              <input type="number" v-model="discountValue" min="0" />
            </label>
            <p v-if="discountType === 'percent'">优惠后：¥{{ (discountOrder.total_amount * (1 - discountValue/100)).toFixed(2) }}</p>
            <p v-else>优惠后：¥{{ Math.max(0, discountOrder.total_amount - discountValue).toFixed(2) }}</p>
          </div>
          <button @click="applyDiscount" class="submit-btn">应用折扣</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const orders = ref([])
const activeTab = ref('')
const detailOrder = ref(null)
const discountOrder = ref(null)
const discountType = ref('fixed')
const discountValue = ref(0)

const tabs = [
  { label: '全部', value: '' },
  { label: '待确认', value: 'pending' },
  { label: '已确认', value: 'confirmed' },
  { label: '制作中', value: 'cooking' },
  { label: '已上菜', value: 'served' },
  { label: '已结账', value: 'paid' }
]

const statusFlow = { pending: 'confirmed', confirmed: 'cooking', cooking: 'served', served: 'paid' }
const orderStatusLabel = s => ({ pending: '待确认', confirmed: '已确认', cooking: '制作中', served: '已上菜', paid: '已结账', cancelled: '已取消' }[s] || s)

function formatTime(t) {
  return new Date(t).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

async function loadOrders() {
  const url = activeTab.value ? `/api/orders?status=${activeTab.value}` : '/api/orders'
  const res = await axios.get(url)
  orders.value = res.data
}

function viewOrder(order) { detailOrder.value = order }

function openDiscount(order) {
  discountOrder.value = order
  discountType.value = 'fixed'
  discountValue.value = 0
}

async function applyDiscount() {
  let amount = 0
  if (discountType.value === 'fixed') {
    amount = parseFloat(discountValue.value)
  } else {
    amount = discountOrder.value.total_amount * (discountValue.value / 100)
  }
  await axios.put(`/api/orders/${discountOrder.value.id}/status`, {
    status: discountOrder.value.status,
    discount_amount: amount
  })
  discountOrder.value = null
  loadOrders()
}

async function changeStatus(order) {
  const next = statusFlow[order.status]
  if (next) {
    await axios.put(`/api/orders/${order.id}/status`, { status: next })
    loadOrders()
  }
}

onMounted(loadOrders)
</script>

<style scoped>
.page-title { font-size: 24px; font-weight: bold; margin-bottom: 20px; }
.tabs { display: flex; gap: 8px; margin-bottom: 20px; }
.tab { padding: 8px 16px; border: 1px solid #ddd; border-radius: 20px; background: white; font-size: 14px; }
.tab.active { background: #e74c3c; color: white; border-color: #e74c3c; }
.orders-table-wrapper { background: white; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
.orders-table { width: 100%; border-collapse: collapse; }
.orders-table th, .orders-table td { padding: 14px 16px; text-align: left; border-bottom: 1px solid #f0f0f0; }
.orders-table th { background: #fafafa; font-weight: 600; color: #555; }
.status-badge { padding: 4px 10px; border-radius: 12px; font-size: 12px; }
.status-badge.pending { background: #fff3cd; color: #856404; }
.status-badge.confirmed { background: #d1ecf1; color: #0c5460; }
.status-badge.cooking { background: #cce5ff; color: #004085; }
.status-badge.served { background: #d4edda; color: #155724; }
.status-badge.paid { background: #f8f9fa; color: #6c757d; }
.status-badge.cancelled { background: #f8d7da; color: #721c24; }
.actions { display: flex; gap: 8px; }
.btn-view, .btn-discount, .btn-next {
  padding: 6px 12px; border: none; border-radius: 6px; font-size: 13px;
}
.btn-view { background: #3498db; color: white; }
.btn-discount { background: #f39c12; color: white; }
.btn-next { background: #27ae60; color: white; }
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal { background: white; border-radius: 16px; width: 90%; max-width: 600px; max-height: 85vh; display: flex; flex-direction: column; }
.small-modal { max-width: 400px; }
.modal-header { padding: 20px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; }
.modal-header h2 { font-size: 18px; }
.close-btn { background: none; border: none; font-size: 20px; color: #666; }
.modal-body { padding: 20px; overflow-y: auto; }
.modal-body p { margin-bottom: 10px; }
.detail-table { width: 100%; border-collapse: collapse; margin: 12px 0; }
.detail-table th, .detail-table td { padding: 10px; border: 1px solid #eee; }
.detail-table th { background: #fafafa; }
.detail-total { font-size: 18px; font-weight: bold; color: #e74c3c; text-align: right; margin-top: 12px; }
.discount-form { display: flex; flex-direction: column; gap: 12px; margin: 16px 0; }
.discount-form label { display: flex; flex-direction: column; gap: 4px; font-size: 14px; }
.discount-form input, .discount-form select { padding: 8px; border: 1px solid #ddd; border-radius: 6px; font-size: 14px; }
.submit-btn { width: 100%; background: #e74c3c; color: white; border: none; padding: 12px; border-radius: 8px; font-size: 16px; }
</style>
