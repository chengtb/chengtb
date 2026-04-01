<template>
  <div class="menu-page">
    <van-nav-bar title="点餐" :border="false" />
    
    <div class="menu-content">
      <van-sidebar v-model="activeCategory" class="category-sidebar">
        <van-sidebar-item
          v-for="cat in categories"
          :key="cat.id"
          :title="cat.name"
          @click="selectCategory(cat)"
        />
      </van-sidebar>

      <div class="product-list">
        <van-pull-refresh v-model="refreshing" @refresh="loadProducts">
          <van-list>
            <div
              v-for="product in products"
              :key="product.id"
              class="product-card"
              @click="showProductDetail(product)"
            >
              <van-image
                :src="product.image || 'https://via.placeholder.com/80'"
                width="80"
                height="80"
                radius="8"
                fit="cover"
              />
              <div class="product-info">
                <div class="product-name">{{ product.name }}</div>
                <div class="product-desc">{{ product.description }}</div>
                <div class="product-bottom">
                  <span class="product-price">¥{{ product.price.toFixed(2) }}</span>
                  <van-stepper
                    :model-value="getCartQuantity(product)"
                    theme="round"
                    min="0"
                    @change="(val) => onQuantityChange(product, val)"
                    @click.stop
                  />
                </div>
              </div>
            </div>
          </van-list>
        </van-pull-refresh>
      </div>
    </div>

    <!-- Cart Bar -->
    <div class="cart-bar" v-if="cart.getItemCount() > 0" @click="showCart = true">
      <van-badge :content="cart.getItemCount()">
        <van-icon name="shopping-cart-o" size="24" color="#fff" />
      </van-badge>
      <span class="cart-total">¥{{ cart.getTotal().toFixed(2) }}</span>
      <van-button type="primary" size="small" round @click.stop="submitOrder">
        提交订单
      </van-button>
    </div>

    <!-- Cart Popup -->
    <van-popup v-model:show="showCart" position="bottom" round>
      <div class="cart-popup">
        <div class="cart-header">
          <span>购物车</span>
          <van-button size="mini" plain @click="cart.clearCart()">清空</van-button>
        </div>
        <div class="cart-items">
          <div v-for="item in cart.state.items" :key="item.key" class="cart-item">
            <div class="cart-item-info">
              <span class="cart-item-name">{{ item.product_name }}</span>
              <span class="cart-item-specs" v-if="item.specs.length">
                {{ item.specs.map(s => s.spec_value || s.value).join(', ') }}
              </span>
              <span class="cart-item-price">¥{{ item.unit_price.toFixed(2) }}</span>
            </div>
            <van-stepper
              v-model="item.quantity"
              min="0"
              @change="(val) => cart.updateQuantity(item.key, val)"
            />
          </div>
        </div>
      </div>
    </van-popup>

    <!-- Product Detail Popup -->
    <van-popup v-model:show="showDetail" position="bottom" round>
      <div class="product-detail" v-if="selectedProduct">
        <van-image
          :src="selectedProduct.image || 'https://via.placeholder.com/200'"
          width="100%"
          height="200"
          fit="cover"
        />
        <div class="detail-info">
          <h3>{{ selectedProduct.name }}</h3>
          <p>{{ selectedProduct.description }}</p>
          <div class="spec-groups" v-if="selectedProduct.specs && selectedProduct.specs.length">
            <div v-for="group in specGroups" :key="group.name" class="spec-group">
              <span class="spec-label">{{ group.name }}：</span>
              <van-radio-group v-model="selectedSpecs[group.name]" direction="horizontal">
                <van-radio
                  v-for="spec in group.options"
                  :key="spec.id"
                  :name="spec.id"
                >
                  {{ spec.spec_value }}
                  <template v-if="spec.extra_price > 0">+¥{{ spec.extra_price }}</template>
                </van-radio>
              </van-radio-group>
            </div>
          </div>
          <div class="detail-bottom">
            <span class="detail-price">¥{{ calculatedPrice.toFixed(2) }}</span>
            <van-button type="primary" round @click="addToCart">加入购物车</van-button>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { showToast, showSuccessToast, showLoadingToast, closeToast } from 'vant'
import { getCategories, getProducts, getProduct, createOrder } from '../api/index.js'
import cart from '../stores/cart.js'
import wsClient from '../api/websocket.js'

const categories = ref([])
const products = ref([])
const activeCategory = ref(0)
const refreshing = ref(false)
const showCart = ref(false)
const showDetail = ref(false)
const selectedProduct = ref(null)
const selectedSpecs = ref({})

// Get table ID from URL query
const urlParams = new URLSearchParams(window.location.search)
const tableId = parseInt(urlParams.get('table_id') || '1')

onMounted(async () => {
  cart.setTableId(tableId)
  wsClient.connect(tableId)
  await loadCategories()
})

async function loadCategories() {
  try {
    const res = await getCategories()
    categories.value = res.data || []
    if (categories.value.length > 0) {
      await loadProductsByCategory(categories.value[0].id)
    }
  } catch (e) {
    showToast('加载分类失败')
  }
}

async function selectCategory(cat) {
  await loadProductsByCategory(cat.id)
}

async function loadProductsByCategory(categoryId) {
  try {
    const res = await getProducts(categoryId)
    products.value = res.data || []
  } catch (e) {
    showToast('加载商品失败')
  }
}

async function loadProducts() {
  if (categories.value.length > 0) {
    await loadProductsByCategory(categories.value[activeCategory.value]?.id)
  }
  refreshing.value = false
}

async function showProductDetail(product) {
  try {
    const res = await getProduct(product.id)
    selectedProduct.value = res.data
    selectedSpecs.value = {}
    showDetail.value = true
  } catch (e) {
    showToast('加载商品详情失败')
  }
}

const specGroups = computed(() => {
  if (!selectedProduct.value?.specs) return []
  const groups = {}
  selectedProduct.value.specs.forEach(spec => {
    if (!groups[spec.spec_name]) {
      groups[spec.spec_name] = { name: spec.spec_name, options: [] }
    }
    groups[spec.spec_name].options.push(spec)
  })
  return Object.values(groups)
})

const calculatedPrice = computed(() => {
  if (!selectedProduct.value) return 0
  let price = selectedProduct.value.price
  Object.values(selectedSpecs.value).forEach(specId => {
    const spec = selectedProduct.value.specs?.find(s => s.id === specId)
    if (spec) price += spec.extra_price
  })
  return price
})

function getCartQuantity(product) {
  return cart.state.items
    .filter(item => item.product_id === product.id)
    .reduce((sum, item) => sum + item.quantity, 0)
}

function onQuantityChange(product, val) {
  const currentQty = getCartQuantity(product)
  if (val > currentQty) {
    cart.addItem(product, [], val - currentQty)
  } else if (val < currentQty) {
    const item = cart.state.items.find(i => i.product_id === product.id)
    if (item) cart.updateQuantity(item.key, val)
  }
}

function addToCart() {
  const specs = Object.entries(selectedSpecs.value).map(([name, specId]) => {
    const spec = selectedProduct.value.specs?.find(s => s.id === specId)
    return spec ? { spec_id: spec.id, spec_value: spec.spec_value, extra_price: spec.extra_price } : null
  }).filter(Boolean)
  
  cart.addItem(selectedProduct.value, specs)
  showDetail.value = false
  showSuccessToast('已加入购物车')
}

async function submitOrder() {
  if (cart.state.items.length === 0) {
    showToast('购物车为空')
    return
  }

  showLoadingToast({ message: '提交中...', forbidClick: true })

  try {
    const items = cart.state.items.map(item => ({
      product_id: item.product_id,
      spec_detail: item.spec_detail.length > 0 ? item.spec_detail : undefined,
      quantity: item.quantity,
      remark: item.remark,
    }))

    await createOrder({
      table_id: tableId,
      items,
    })

    closeToast()
    showSuccessToast('下单成功！')
    cart.clearCart()
    showCart.value = false
  } catch (e) {
    closeToast()
    showToast('下单失败，请重试')
  }
}
</script>

<style scoped>
.menu-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f7f8fa;
}

.menu-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.category-sidebar {
  width: 90px;
  flex-shrink: 0;
}

.product-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.product-card {
  display: flex;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
  margin-bottom: 8px;
}

.product-info {
  flex: 1;
  margin-left: 12px;
  display: flex;
  flex-direction: column;
}

.product-name {
  font-size: 14px;
  font-weight: bold;
  color: #333;
}

.product-desc {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
}

.product-price {
  color: #ee0a24;
  font-size: 16px;
  font-weight: bold;
}

.cart-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 50px;
  background: #333;
  display: flex;
  align-items: center;
  padding: 0 16px;
  z-index: 100;
}

.cart-total {
  color: #fff;
  font-size: 16px;
  font-weight: bold;
  margin-left: 12px;
  flex: 1;
}

.cart-popup {
  padding: 16px;
  max-height: 60vh;
}

.cart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid #eee;
}

.cart-items {
  max-height: 40vh;
  overflow-y: auto;
}

.cart-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f5f5f5;
}

.cart-item-info {
  display: flex;
  flex-direction: column;
}

.cart-item-name {
  font-size: 14px;
  color: #333;
}

.cart-item-specs {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.cart-item-price {
  color: #ee0a24;
  font-size: 14px;
  margin-top: 4px;
}

.product-detail {
  max-height: 80vh;
  overflow-y: auto;
}

.detail-info {
  padding: 16px;
}

.spec-groups {
  margin-top: 16px;
}

.spec-group {
  margin-bottom: 12px;
}

.spec-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
  display: block;
}

.detail-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
}

.detail-price {
  color: #ee0a24;
  font-size: 20px;
  font-weight: bold;
}
</style>
