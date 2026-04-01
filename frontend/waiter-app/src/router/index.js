import { createRouter, createWebHistory } from 'vue-router'
import OrderListView from '../views/OrderListView.vue'
import TablesView from '../views/TablesView.vue'

const routes = [
  { path: '/', name: 'orders', component: OrderListView },
  { path: '/tables', name: 'tables', component: TablesView },
]

const router = createRouter({
  history: createWebHistory('/waiter/'),
  routes,
})

export default router
