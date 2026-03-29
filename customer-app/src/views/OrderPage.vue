<template>
  <div class="order-page">
    <van-nav-bar
      :title="`${tableId ? '桌号: ' + tableId : '扫码点餐'}`"
      fixed
    />
    <div class="content">
      <van-row style="height:100%">
        <van-col span="6" class="sidebar">
          <van-sidebar v-model="activeCategory" class="category-sidebar">
            <van-sidebar-item
              v-for="cat in categories"
              :key="cat.category_id"
              :title="cat.name"
              class="category-item"
            />
          </van-sidebar>
        </van-col>
        <van-col span="18" class="dish-list">
          <div v-if="loading" class="loading-wrap">
            <van-loading size="24px" color="#ff6900">加载中...</van-loading>
          </div>
          <template v-else>
            <div
              v-for="dish in filteredDishes"
              :key="dish.dish_id"
              class="dish-card"
              :class="{ 'dish-card-unavailable': !dish.is_available }"
            >
              <div class="dish-img-wrap">
                <van-image
                  v-if="dish.image_url"
                  :src="dish.image_url"
                  width="72"
                  height="72"
                  fit="cover"
                  radius="8"
                />
                <div v-else class="dish-no-image">
                  <van-icon name="shop-o" size="28" color="#ccc" />
                </div>
                <van-tag v-if="dish.special_flag" class="special-tag" type="warning" size="mini">特色</van-tag>
              </div>
              <div class="dish-info">
                <div class="dish-name">{{ dish.name }}</div>
                <div v-if="dish.description" class="dish-desc">{{ dish.description }}</div>
                <div class="dish-footer">
                  <span class="dish-price">
                    <span class="price-symbol">¥</span>{{ (dish.price / 100).toFixed(2) }}
                  </span>
                  <div class="dish-action">
                    <van-stepper
                      v-if="dish.is_available"
                      :model-value="getItemCount(dish.dish_id)"
                      min="0"
                      @change="(val) => handleQuantityChange(dish, val)"
                      integer
                      button-size="22px"
                    />
                    <span v-else class="unavailable">暂停供应</span>
                  </div>
                </div>
              </div>
            </div>
            <van-empty v-if="filteredDishes.length === 0" description="暂无菜品" image-size="80" />
          </template>
        </van-col>
      </van-row>
    </div>

    <!-- Cart Bar -->
    <div v-if="cart.totalCount.value > 0" class="cart-bar" @click="showCart = true">
      <div class="cart-icon-wrap">
        <van-icon name="cart-o" size="28" color="white" />
        <van-badge :content="cart.totalCount.value" position="top-right" />
      </div>
      <span class="cart-amount">¥{{ (cart.totalAmount.value / 100).toFixed(2) }}</span>
      <span class="cart-hint">去结算</span>
    </div>
    <div v-else class="cart-bar cart-bar-empty">
      <div class="cart-icon-wrap">
        <van-icon name="cart-o" size="28" color="rgba(255,255,255,0.5)" />
      </div>
      <span class="cart-hint-empty">购物车空</span>
    </div>

    <!-- Cart Popup -->
    <van-popup v-model:show="showCart" position="bottom" round style="max-height:70vh;overflow-y:auto">
      <div class="cart-popup">
        <div class="cart-popup-header">
          <span class="cart-popup-title">购物车</span>
          <van-button size="small" plain type="danger" @click="showClearConfirm">清空</van-button>
        </div>
        <div v-if="cart.items.value.length === 0" style="padding:20px;text-align:center">
          <van-empty description="购物车空空如也" />
        </div>
        <div v-for="item in cart.items.value" :key="item.id" class="cart-item">
          <span class="cart-item-name">{{ item.dish?.name || '未知菜品' }}</span>
          <span class="cart-item-price">¥{{ ((item.dish?.price || 0) / 100).toFixed(2) }}</span>
          <van-stepper
            :model-value="item.quantity"
            min="0"
            @change="(val) => handleCartItemChange(item, val)"
            integer
            size="small"
          />
        </div>
        <div class="cart-popup-footer">
          <div class="cart-total">合计: <strong>¥{{ (cart.totalAmount.value / 100).toFixed(2) }}</strong></div>
          <van-button type="primary" round block @click="handleSubmit" :loading="submitting">
            提交订单
          </van-button>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast, showConfirmDialog } from 'vant'
import api from '../api'
import { useCart } from '../store/cart'

const route = useRoute()
const router = useRouter()
const tableId = computed(() => route.params.tableId || '')

const categories = ref([])
const dishes = ref([])
const activeCategory = ref(0)
const loading = ref(true)
const showCart = ref(false)
const submitting = ref(false)

const cart = useCart(tableId.value || 1)

const filteredDishes = computed(() => {
  if (categories.value.length === 0) return dishes.value
  const cat = categories.value[activeCategory.value]
  if (!cat) return dishes.value
  return dishes.value.filter(d => d.category_id === cat.category_id)
})

function getItemCount(dishId) {
  const item = cart.items.value.find(i => i.dish_id === dishId)
  return item ? item.quantity : 0
}

async function handleQuantityChange(dish, val) {
  if (!cart.session.value) {
    try {
      await cart.initSession()
    } catch (e) {
      showToast('初始化会话失败')
      return
    }
  }
  const existing = cart.items.value.find(i => i.dish_id === dish.dish_id)
  if (val === 0 && existing) {
    await cart.removeItem(existing.id)
  } else if (existing) {
    await cart.updateItem(existing.id, val)
  } else if (val > 0) {
    await cart.addItem(dish.dish_id, val)
  }
}

async function handleCartItemChange(item, val) {
  if (val === 0) {
    await cart.removeItem(item.id)
  } else {
    await cart.updateItem(item.id, val)
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    const order = await cart.submitOrder()
    showToast({ type: 'success', message: '订单提交成功!' })
    showCart.value = false
    router.push(`/status/${order.order_id}`)
  } catch (e) {
    showToast('提交失败: ' + (e.response?.data?.error || e.message))
  } finally {
    submitting.value = false
  }
}

async function showClearConfirm() {
  try {
    await showConfirmDialog({ title: '确认清空购物车?', message: '清空后需要重新添加商品' })
    for (const item of [...cart.items.value]) {
      await cart.removeItem(item.id)
    }
  } catch {}
}

onMounted(async () => {
  try {
    const [catRes, dishRes] = await Promise.all([
      api.get('/categories'),
      api.get('/dishes'),
    ])
    categories.value = catRes.data || []
    dishes.value = dishRes.data || []
  } catch (e) {
    showToast('加载菜单失败')
  } finally {
    loading.value = false
  }

  if (tableId.value) {
    try {
      await cart.initSession()
    } catch {}
  }
})
</script>

<style scoped>
.order-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f7f8fa;
}
.content {
  margin-top: 46px;
  flex: 1;
  overflow: hidden;
  height: calc(100vh - 46px - 60px);
  display: flex;
}
/* ── Sidebar ── */
.sidebar {
  overflow-y: auto;
  height: 100%;
  background: #f2f3f5;
}
.category-sidebar {
  width: 100%;
}
:deep(.van-sidebar-item) {
  padding: 14px 10px;
  font-size: 13px;
  line-height: 1.3;
  border-left: 3px solid transparent;
  background: #f2f3f5;
  color: #666;
}
:deep(.van-sidebar-item--select) {
  background: #fff;
  color: #ff6900;
  font-weight: 700;
  border-left-color: #ff6900;
}
/* ── Dish list ── */
.dish-list {
  overflow-y: auto;
  height: 100%;
  padding: 6px 8px;
  background: #fff;
}
.loading-wrap {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}
/* ── Dish card ── */
.dish-card {
  display: flex;
  align-items: flex-start;
  padding: 12px 4px;
  border-bottom: 1px solid #f5f5f5;
  gap: 10px;
}
.dish-card:last-child {
  border-bottom: none;
}
.dish-card-unavailable {
  opacity: 0.5;
}
.dish-img-wrap {
  position: relative;
  flex-shrink: 0;
  width: 72px;
  height: 72px;
}
.dish-no-image {
  width: 72px;
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  border-radius: 8px;
}
.special-tag {
  position: absolute;
  top: 2px;
  left: 2px;
  border-radius: 4px 0;
}
.dish-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.dish-name {
  font-weight: 600;
  font-size: 14px;
  color: #222;
  line-height: 1.3;
}
.dish-desc {
  font-size: 12px;
  color: #999;
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-height: 1.4;
}
.dish-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 4px;
}
.dish-price {
  color: #ff4444;
  font-weight: 700;
  font-size: 15px;
}
.price-symbol {
  font-size: 11px;
  margin-right: 1px;
}
.dish-action {
  flex-shrink: 0;
}
.unavailable {
  font-size: 11px;
  color: #bbb;
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
}
/* ── Cart bar ── */
.cart-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  max-width: 480px;
  margin: 0 auto;
  height: 60px;
  background: linear-gradient(135deg, #3a3a3a 0%, #1a1a1a 100%);
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 12px;
  cursor: pointer;
  z-index: 100;
  box-shadow: 0 -2px 12px rgba(0,0,0,0.15);
}
.cart-bar-empty {
  cursor: default;
}
.cart-icon-wrap {
  position: relative;
  width: 42px;
  height: 42px;
  background: linear-gradient(135deg, #ff8c00 0%, #ff6900 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(255,105,0,0.4);
}
.cart-amount {
  flex: 1;
  color: white;
  font-size: 16px;
  font-weight: 700;
}
.cart-hint {
  color: white;
  font-size: 13px;
  background: #ff6900;
  padding: 6px 14px;
  border-radius: 20px;
}
.cart-hint-empty {
  color: rgba(255,255,255,0.4);
  font-size: 14px;
}
/* ── Cart popup ── */
.cart-popup {
  padding: 16px;
}
.cart-popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f5f5f5;
}
.cart-popup-title {
  font-size: 16px;
  font-weight: 700;
}
.cart-item {
  display: flex;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f5f5f5;
  gap: 8px;
}
.cart-item-name {
  flex: 1;
  font-size: 14px;
  color: #333;
}
.cart-item-price {
  color: #ff4444;
  font-size: 14px;
  font-weight: 600;
  min-width: 60px;
  text-align: right;
}
.cart-popup-footer {
  padding-top: 16px;
}
.cart-total {
  text-align: right;
  font-size: 15px;
  margin-bottom: 12px;
  color: #333;
}
</style>
