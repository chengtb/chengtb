import { createRouter, createWebHistory } from 'vue-router'
import KitchenView from '../views/KitchenView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [{ path: '/', component: KitchenView }]
})

export default router
