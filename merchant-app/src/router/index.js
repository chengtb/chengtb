import { createRouter, createWebHashHistory } from 'vue-router'
import TablesView from '../views/TablesView.vue'
import OrdersView from '../views/OrdersView.vue'
import ChefsView from '../views/ChefsView.vue'
import DishesView from '../views/DishesView.vue'
import RecipesView from '../views/RecipesView.vue'
import TasksView from '../views/TasksView.vue'
import ConfigView from '../views/ConfigView.vue'

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/tables' },
    { path: '/tables', component: TablesView },
    { path: '/orders', component: OrdersView },
    { path: '/chefs', component: ChefsView },
    { path: '/dishes', component: DishesView },
    { path: '/recipes', component: RecipesView },
    { path: '/tasks', component: TasksView },
    { path: '/config', component: ConfigView },
  ]
})
