import { reactive } from 'vue'
import wsClient from '../api/websocket'

const state = reactive({
  items: [],
  tableId: null,
})

function addItem(product, specs = [], quantity = 1) {
  const key = `${product.id}_${JSON.stringify(specs.map(s => s.spec_id).sort())}`
  const existing = state.items.find(item => item.key === key)
  
  if (existing) {
    existing.quantity += quantity
  } else {
    let extraPrice = 0
    specs.forEach(s => { extraPrice += (s.extra_price || 0) })
    
    state.items.push({
      key,
      product_id: product.id,
      product_name: product.name,
      product_image: product.image,
      specs,
      spec_detail: specs,
      unit_price: product.price + extraPrice,
      quantity,
      remark: '',
    })
  }
  
  syncCart()
}

function removeItem(key) {
  const index = state.items.findIndex(item => item.key === key)
  if (index > -1) {
    state.items.splice(index, 1)
    syncCart()
  }
}

function updateQuantity(key, quantity) {
  const item = state.items.find(item => item.key === key)
  if (item) {
    if (quantity <= 0) {
      removeItem(key)
    } else {
      item.quantity = quantity
      syncCart()
    }
  }
}

function clearCart() {
  state.items.splice(0, state.items.length)
  syncCart()
}

function getTotal() {
  return state.items.reduce((sum, item) => sum + item.unit_price * item.quantity, 0)
}

function getItemCount() {
  return state.items.reduce((sum, item) => sum + item.quantity, 0)
}

function syncCart() {
  wsClient.send({
    type: 'cart_update',
    table_id: state.tableId,
    data: { items: state.items },
  })
}

function setTableId(id) {
  state.tableId = id
}

// Listen for cart updates from other customers at the same table
wsClient.on('cart_update', (data) => {
  if (data && data.items) {
    state.items.splice(0, state.items.length, ...data.items)
  }
})

export default {
  state,
  addItem,
  removeItem,
  updateQuantity,
  clearCart,
  getTotal,
  getItemCount,
  setTableId,
}
