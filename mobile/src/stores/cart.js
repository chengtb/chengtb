import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useCartStore = defineStore('cart', () => {
  const items = ref([])

  const totalCount = computed(() => items.value.reduce((sum, i) => sum + i.quantity, 0))
  const totalPrice = computed(() => items.value.reduce((sum, i) => sum + i.price * i.quantity, 0))

  function addItem(dish) {
    const existing = items.value.find(i => i.id === dish.id)
    if (existing) {
      existing.quantity++
    } else {
      items.value.push({ id: dish.id, name: dish.name, price: dish.price, quantity: 1 })
    }
  }

  function removeItem(dishId) {
    const existing = items.value.find(i => i.id === dishId)
    if (existing) {
      if (existing.quantity > 1) {
        existing.quantity--
      } else {
        items.value = items.value.filter(i => i.id !== dishId)
      }
    }
  }

  function getQuantity(dishId) {
    const item = items.value.find(i => i.id === dishId)
    return item ? item.quantity : 0
  }

  function clearCart() {
    items.value = []
  }

  return { items, totalCount, totalPrice, addItem, removeItem, getQuantity, clearCart }
})
