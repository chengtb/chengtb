<template>
  <div class="orders-page">
    <van-nav-bar title="我的订单" left-arrow @click-left="$router.back()" />
    <van-pull-refresh v-model="refreshing" @refresh="loadOrders">
      <van-empty v-if="orders.length === 0" description="暂无订单" />
      <div v-for="order in orders" :key="order.id" class="order-card">
        <div class="order-header">
          <span class="order-sn">{{ order.order_sn }}</span>
          <van-tag :type="statusType(order.status)">{{ statusText(order.status) }}</van-tag>
        </div>
        <div class="order-items">
          <div v-for="item in order.items" :key="item.id" class="order-item">
            <span>{{ item.product?.name || '商品' }} x{{ item.quantity }}</span>
            <span>¥{{ (item.unit_price * item.quantity).toFixed(2) }}</span>
          </div>
        </div>
        <div class="order-footer">
          <span>合计：¥{{ order.total_amount.toFixed(2) }}</span>
        </div>
      </div>
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showToast } from 'vant'
import { getOrdersByTable } from '../api/index.js'

const orders = ref([])
const refreshing = ref(false)

const urlParams = new URLSearchParams(window.location.search)
const tableId = parseInt(urlParams.get('table_id') || '1')

onMounted(() => loadOrders())

async function loadOrders() {
  try {
    const res = await getOrdersByTable(tableId)
    orders.value = res.data || []
  } catch (e) {
    showToast('加载订单失败')
  }
  refreshing.value = false
}

function statusType(status) {
  const map = { pending: 'warning', confirmed: 'primary', cooking: 'primary', completed: 'success', paid: 'success', cancelled: 'danger' }
  return map[status] || 'default'
}

function statusText(status) {
  const map = { pending: '待确认', confirmed: '已确认', cooking: '烹饪中', completed: '已完成', paid: '已支付', cancelled: '已取消' }
  return map[status] || status
}
</script>

<style scoped>
.orders-page {
  min-height: 100vh;
  background: #f7f8fa;
}

.order-card {
  margin: 8px 12px;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 8px;
  border-bottom: 1px solid #f5f5f5;
}

.order-sn {
  font-size: 12px;
  color: #999;
}

.order-items {
  padding: 8px 0;
}

.order-item {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 14px;
}

.order-footer {
  text-align: right;
  font-weight: bold;
  color: #ee0a24;
  padding-top: 8px;
  border-top: 1px solid #f5f5f5;
}
</style>
