import { createRouter, createWebHistory } from 'vue-router'
import MenuView from '../views/MenuView.vue'
import OrdersView from '../views/OrdersView.vue'

const routes = [
  { path: '/', name: 'menu', component: MenuView },
  { path: '/orders', name: 'orders', component: OrdersView },
]

const router = createRouter({
  history: createWebHistory('/customer/'),
  routes,
})

export default router
