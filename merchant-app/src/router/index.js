import { createRouter, createWebHashHistory } from 'vue-router'
import TablesView from '../views/TablesView.vue'
import OrdersView from '../views/OrdersView.vue'
import ChefsView from '../views/ChefsView.vue'
import DishesView from '../views/DishesView.vue'
import CategoriesView from '../views/CategoriesView.vue'
import RecipesView from '../views/RecipesView.vue'
import TasksView from '../views/TasksView.vue'
import ConfigView from '../views/ConfigView.vue'
import WaitersView from '../views/WaitersView.vue'
import DeliveryTasksView from '../views/DeliveryTasksView.vue'

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', redirect: '/tables' },
    { path: '/tables', component: TablesView },
    { path: '/orders', component: OrdersView },
    { path: '/chefs', component: ChefsView },
    { path: '/dishes', component: DishesView },
    { path: '/categories', component: CategoriesView },
    { path: '/recipes', component: RecipesView },
    { path: '/tasks', component: TasksView },
    { path: '/waiters', component: WaitersView },
    { path: '/delivery-tasks', component: DeliveryTasksView },
    { path: '/config', component: ConfigView },
  ]
})
