<template>
  <v-app>
    <v-navigation-drawer v-model="drawer" permanent>
      <v-list-item
        prepend-icon="mdi-restaurant"
        title="餐厅管理系统"
        nav
      ></v-list-item>
      <v-divider></v-divider>
      <v-list density="compact" nav>
        <v-list-item
          v-for="item in navItems"
          :key="item.to"
          :prepend-icon="item.icon"
          :title="item.title"
          :to="item.to"
          rounded="lg"
        ></v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-app-bar elevation="1">
      <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>餐厅管理系统</v-toolbar-title>
    </v-app-bar>

    <v-main>
      <v-container fluid>
        <router-view />
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup>
import { ref } from 'vue'

const drawer = ref(true)

const navItems = [
  { title: '餐桌概览', icon: 'mdi-table-chair', to: '/tables' },
  { title: '订单列表', icon: 'mdi-clipboard-list', to: '/orders' },
  { title: '厨师管理', icon: 'mdi-chef-hat', to: '/chefs' },
  { title: '菜品管理', icon: 'mdi-food', to: '/dishes' },
  { title: '菜品分类', icon: 'mdi-tag-multiple', to: '/categories' },
  { title: '菜谱管理', icon: 'mdi-book-open-variant', to: '/recipes' },
  { title: '烹饪任务', icon: 'mdi-pot-steam', to: '/tasks' },
  { title: '系统配置', icon: 'mdi-cog', to: '/config' },
]
</script>
