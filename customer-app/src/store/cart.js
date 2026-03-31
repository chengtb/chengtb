import { ref, computed } from 'vue'
import api from '../api'

export function useCart(tableId) {
  const session = ref(null)
  const items = ref([])
  const ws = ref(null)

  async function initSession() {
    const res = await api.post('/cart/session', { table_id: parseInt(tableId) })
    session.value = res.data
    await loadItems()
    connectWS()
  }

  async function loadItems() {
    if (!session.value) return
    const res = await api.get(`/cart/session/${session.value.session_id}`)
    items.value = res.data.items || []
  }

  function connectWS() {
    const wsBase = import.meta.env.VITE_WS_BASE || 'ws://localhost:8080'
    ws.value = new WebSocket(`${wsBase}/ws/cart/${session.value.session_id}`)
    ws.value.onmessage = (e) => {
      const data = JSON.parse(e.data)
      if (data.type === 'cart_updated') {
        items.value = data.items || []
      }
    }
    ws.value.onclose = () => {
      setTimeout(connectWS, 3000)
    }
  }

  async function addItem(dishId, quantity = 1, note = '') {
    await api.post(`/cart/session/${session.value.session_id}/item`, {
      dish_id: dishId,
      quantity,
      note,
    })
    await loadItems()
  }

  async function updateItem(itemId, quantity) {
    await api.put(`/cart/session/${session.value.session_id}/item/${itemId}`, { quantity })
    await loadItems()
  }

  async function removeItem(itemId) {
    await api.delete(`/cart/session/${session.value.session_id}/item/${itemId}`)
    await loadItems()
  }

  async function submitOrder() {
    const res = await api.post(`/cart/session/${session.value.session_id}/submit`)
    return res.data
  }

  const totalCount = computed(() => items.value.reduce((sum, i) => sum + i.quantity, 0))
  const totalAmount = computed(() => items.value.reduce((sum, i) => sum + i.quantity * (i.dish?.price || 0), 0))

  return { session, items, totalCount, totalAmount, initSession, addItem, updateItem, removeItem, submitOrder }
}
