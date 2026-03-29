import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import OrdersView from '../views/OrdersView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/table/1' },
    { path: '/table/:tableId', component: HomeView },
    { path: '/table/:tableId/orders', component: OrdersView }
  ]
})

export default router
