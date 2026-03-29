import { createRouter, createWebHistory } from 'vue-router'
import TablesView from '../views/TablesView.vue'
import OrdersView from '../views/OrdersView.vue'
import MenuView from '../views/MenuView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/tables' },
    { path: '/tables', component: TablesView },
    { path: '/orders', component: OrdersView },
    { path: '/menu', component: MenuView }
  ]
})

export default router
