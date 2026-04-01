import { createRouter, createWebHistory } from 'vue-router'
import TaskBoardView from '../views/TaskBoardView.vue'

const routes = [
  { path: '/', name: 'tasks', component: TaskBoardView },
]

const router = createRouter({
  history: createWebHistory('/kds/'),
  routes,
})

export default router
