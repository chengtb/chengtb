<template>
  <div class="status-page">
    <van-nav-bar title="订单状态" left-arrow @click-left="$router.back()" />
    <div class="status-content">
      <div v-if="loading" style="text-align:center;padding:40px">
        <van-loading size="32px">加载中...</van-loading>
      </div>
      <template v-else-if="order">
        <van-cell-group title="订单信息" inset style="margin-top:16px">
          <van-cell title="订单号" :value="`#${order.id}`" />
          <van-cell title="桌号" :value="order.table_id" />
          <van-cell title="金额" :value="`¥${(order.total_amount / 100).toFixed(2)}`" />
          <van-cell title="下单时间" :value="formatTime(order.created_at)" />
        </van-cell-group>

        <van-cell-group title="订单状态" inset style="margin-top:16px">
          <div class="timeline">
            <div
              v-for="(step, i) in steps"
              :key="i"
              class="timeline-item"
              :class="{ active: getStepStatus(step.key) === 'active', done: getStepStatus(step.key) === 'done' }"
            >
              <div class="timeline-dot"></div>
              <div class="timeline-content">
                <div class="timeline-title">{{ step.label }}</div>
                <div class="timeline-desc">{{ step.desc }}</div>
              </div>
            </div>
          </div>
        </van-cell-group>

        <van-cell-group title="菜品明细" inset style="margin-top:16px">
          <van-cell
            v-for="item in order.items"
            :key="item.id"
            :title="item.dish?.name || '未知菜品'"
            :label="item.note ? `备注: ${item.note}` : ''"
            :value="`x${item.quantity}  ¥${((item.dish?.price || 0) * item.quantity / 100).toFixed(2)}`"
          />
        </van-cell-group>
      </template>
      <van-empty v-else description="订单不存在" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'

const route = useRoute()
const orderId = route.params.orderId
const order = ref(null)
const loading = ref(true)
let timer = null

const steps = [
  { key: 'pending', label: '等待确认', desc: '订单已提交，等待处理' },
  { key: 'cooking', label: '烹饪中', desc: '厨师正在准备您的餐品' },
  { key: 'done', label: '完成', desc: '餐品已准备好，请享用' },
]

const statusOrder = ['pending', 'cooking', 'done']

function getStepStatus(key) {
  if (!order.value) return ''
  const currentIdx = statusOrder.indexOf(order.value.status)
  const stepIdx = statusOrder.indexOf(key)
  if (stepIdx < currentIdx) return 'done'
  if (stepIdx === currentIdx) return 'active'
  return ''
}

function formatTime(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleString('zh-CN')
}

async function fetchOrder() {
  try {
    const res = await api.get(`/orders/${orderId}`)
    order.value = res.data
  } catch {
    order.value = null
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchOrder()
  timer = setInterval(fetchOrder, 5000)
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>

<style scoped>
.status-page {
  min-height: 100vh;
  background: #f5f5f5;
}
.status-content {
  padding-bottom: 20px;
}
.timeline {
  padding: 16px;
}
.timeline-item {
  display: flex;
  gap: 12px;
  padding-bottom: 20px;
  position: relative;
}
.timeline-item:not(:last-child)::after {
  content: '';
  position: absolute;
  left: 7px;
  top: 16px;
  bottom: 0;
  width: 2px;
  background: #ddd;
}
.timeline-item.done::after { background: #4caf50; }
.timeline-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #ddd;
  flex-shrink: 0;
  margin-top: 2px;
}
.timeline-item.done .timeline-dot { background: #4caf50; }
.timeline-item.active .timeline-dot { background: #ff6900; }
.timeline-title {
  font-weight: 600;
  font-size: 14px;
}
.timeline-item.active .timeline-title { color: #ff6900; }
.timeline-item.done .timeline-title { color: #4caf50; }
.timeline-desc {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}
</style>
