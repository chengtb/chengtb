import { createRouter, createWebHashHistory } from 'vue-router'
import OrderPage from '../views/OrderPage.vue'
import OrderStatus from '../views/OrderStatus.vue'

const routes = [
  { path: '/', redirect: '/order' },
  { path: '/order', component: OrderPage },
  { path: '/order/:tableId', component: OrderPage },
  { path: '/status/:orderId', component: OrderStatus },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
