<template>
  <div class="home">
    <!-- Category Sidebar -->
    <aside class="category-sidebar">
      <div
        v-for="cat in categories"
        :key="cat.id"
        :class="['cat-item', { active: activeCatId === cat.id }]"
        @click="selectCategory(cat.id)"
      >
        <span class="cat-icon">{{ cat.icon }}</span>
        <span class="cat-name">{{ cat.name }}</span>
      </div>
    </aside>

    <!-- Dish List -->
    <section class="dish-list">
      <div v-for="dish in filteredDishes" :key="dish.id" class="dish-card">
        <div class="dish-image-placeholder">
          <span>{{ dish.name[0] }}</span>
        </div>
        <div class="dish-info">
          <h3 class="dish-name">{{ dish.name }}</h3>
          <p class="dish-desc">{{ dish.description }}</p>
          <div class="dish-footer">
            <span class="dish-price">¥{{ dish.price.toFixed(2) }}</span>
            <div class="qty-control">
              <button v-if="cartStore.getQuantity(dish.id) > 0" @click="cartStore.removeItem(dish.id)" class="qty-btn minus">－</button>
              <span v-if="cartStore.getQuantity(dish.id) > 0" class="qty-num">{{ cartStore.getQuantity(dish.id) }}</span>
              <button @click="cartStore.addItem(dish)" class="qty-btn plus">＋</button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Cart Floating Button -->
    <div v-if="cartStore.totalCount > 0" class="cart-btn" @click="showCart = true">
      <span class="cart-badge">{{ cartStore.totalCount }}</span>
      <span class="cart-icon">🛒</span>
      <span class="cart-total">¥{{ cartStore.totalPrice.toFixed(2) }}</span>
      <span class="cart-action">去结算</span>
    </div>

    <!-- Cart Drawer -->
    <div v-if="showCart" class="cart-overlay" @click.self="showCart = false">
      <div class="cart-drawer">
        <div class="cart-header">
          <h3>购物车</h3>
          <button @click="cartStore.clearCart(); showCart = false" class="clear-btn">清空</button>
        </div>
        <div class="cart-items">
          <div v-for="item in cartStore.items" :key="item.id" class="cart-item">
            <span class="cart-item-name">{{ item.name }}</span>
            <div class="qty-control">
              <button @click="cartStore.removeItem(item.id)" class="qty-btn minus">－</button>
              <span class="qty-num">{{ item.quantity }}</span>
              <button @click="cartStore.addItem(item)" class="qty-btn plus">＋</button>
            </div>
            <span class="cart-item-price">¥{{ (item.price * item.quantity).toFixed(2) }}</span>
          </div>
        </div>
        <div class="cart-footer">
          <span class="cart-total-label">合计：<strong>¥{{ cartStore.totalPrice.toFixed(2) }}</strong></span>
          <button @click="submitOrder" class="submit-btn" :disabled="submitting">
            {{ submitting ? '提交中...' : '提交订单' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { useCartStore } from '../stores/cart.js'
import { useTableStore } from '../stores/table.js'

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()
const tableStore = useTableStore()

const categories = ref([])
const dishes = ref([])
const activeCatId = ref(null)
const showCart = ref(false)
const submitting = ref(false)

const filteredDishes = computed(() =>
  activeCatId.value ? dishes.value.filter(d => d.category_id === activeCatId.value) : dishes.value
)

function selectCategory(id) {
  activeCatId.value = id
}

onMounted(async () => {
  const tableId = route.params.tableId
  await tableStore.setTable(tableId)

  const [catRes, dishRes] = await Promise.all([
    axios.get('/api/categories'),
    axios.get('/api/dishes')
  ])
  categories.value = catRes.data
  dishes.value = dishRes.data
  if (categories.value.length > 0) {
    activeCatId.value = categories.value[0].id
  }
})

async function submitOrder() {
  if (cartStore.items.length === 0) return
  submitting.value = true
  try {
    const payload = {
      table_id: parseInt(tableStore.tableId),
      items: cartStore.items.map(i => ({
        dish_id: i.id,
        quantity: i.quantity,
        note: ''
      }))
    }
    await axios.post('/api/orders', payload)
    cartStore.clearCart()
    showCart.value = false
    alert('订单提交成功！')
    router.push(`/table/${tableStore.tableId}/orders`)
  } catch (e) {
    alert('提交失败：' + (e.response?.data?.error || e.message || '请重试'))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.home {
  display: flex;
  height: calc(100vh - 114px);
  overflow: hidden;
}
.category-sidebar {
  width: 88px;
  background: #f8f8f8;
  overflow-y: auto;
  flex-shrink: 0;
  border-right: 1px solid #eee;
}
.cat-item {
  padding: 16px 8px;
  text-align: center;
  cursor: pointer;
  border-bottom: 1px solid #eee;
  font-size: 12px;
  color: #666;
}
.cat-item.active {
  background: white;
  color: #e74c3c;
  font-weight: bold;
  border-left: 3px solid #e74c3c;
}
.cat-icon { font-size: 20px; display: block; margin-bottom: 4px; }
.dish-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  background: white;
}
.dish-card {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}
.dish-image-placeholder {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #ff6b6b, #ffd93d);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: white;
  font-weight: bold;
  flex-shrink: 0;
}
.dish-info { flex: 1; }
.dish-name { font-size: 15px; font-weight: bold; margin-bottom: 4px; }
.dish-desc { font-size: 12px; color: #999; margin-bottom: 8px; line-height: 1.4; }
.dish-footer { display: flex; align-items: center; justify-content: space-between; }
.dish-price { color: #e74c3c; font-size: 16px; font-weight: bold; }
.qty-control { display: flex; align-items: center; gap: 8px; }
.qty-btn {
  width: 28px; height: 28px; border-radius: 50%; border: none; cursor: pointer;
  font-size: 16px; display: flex; align-items: center; justify-content: center;
}
.qty-btn.plus { background: #e74c3c; color: white; }
.qty-btn.minus { background: white; color: #e74c3c; border: 1px solid #e74c3c; }
.qty-num { font-size: 15px; font-weight: bold; min-width: 20px; text-align: center; }
.cart-btn {
  position: fixed; bottom: 70px; left: 50%; transform: translateX(-50%);
  width: calc(100% - 32px); max-width: 448px;
  background: #333; color: white; border-radius: 24px;
  padding: 12px 20px; display: flex; align-items: center; gap: 8px; cursor: pointer;
  z-index: 50;
}
.cart-badge {
  background: #e74c3c; color: white; border-radius: 50%;
  width: 24px; height: 24px; display: flex; align-items: center; justify-content: center;
  font-size: 12px; font-weight: bold;
}
.cart-icon { font-size: 20px; }
.cart-total { flex: 1; font-size: 16px; font-weight: bold; }
.cart-action { background: #e74c3c; padding: 6px 14px; border-radius: 16px; font-size: 14px; }
.cart-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 200;
  display: flex; align-items: flex-end; justify-content: center;
}
.cart-drawer {
  background: white; border-radius: 16px 16px 0 0;
  width: 100%; max-width: 480px; max-height: 70vh; display: flex; flex-direction: column;
}
.cart-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 16px; border-bottom: 1px solid #eee;
}
.cart-header h3 { font-size: 18px; }
.clear-btn { background: none; border: none; color: #999; cursor: pointer; font-size: 14px; }
.cart-items { flex: 1; overflow-y: auto; padding: 0 16px; }
.cart-item {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 0; border-bottom: 1px solid #f0f0f0;
}
.cart-item-name { flex: 1; font-size: 15px; }
.cart-item-price { color: #e74c3c; font-weight: bold; min-width: 70px; text-align: right; }
.cart-footer {
  padding: 16px; border-top: 1px solid #eee;
  display: flex; justify-content: space-between; align-items: center;
}
.cart-total-label { font-size: 15px; }
.submit-btn {
  background: #e74c3c; color: white; border: none; border-radius: 20px;
  padding: 10px 24px; font-size: 16px; cursor: pointer;
}
.submit-btn:disabled { background: #ccc; }
</style>
