<template>
  <div class="app-wrapper">
    <header class="app-header">
      <span class="header-title">{{ tableStore.tableInfo ? tableStore.tableInfo.name : '扫码点餐' }}</span>
    </header>
    <main class="app-main">
      <router-view />
    </main>
    <nav class="bottom-nav">
      <router-link :to="`/table/${tableStore.tableId}`" class="nav-item">
        <span class="nav-icon">🍽️</span>
        <span>点餐</span>
      </router-link>
      <router-link :to="`/table/${tableStore.tableId}/orders`" class="nav-item">
        <span class="nav-icon">📋</span>
        <span>我的订单</span>
      </router-link>
    </nav>
  </div>
</template>

<script setup>
import { useTableStore } from './stores/table.js'
const tableStore = useTableStore()
</script>

<style scoped>
.app-wrapper {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}
.app-header {
  background: #e74c3c;
  color: white;
  padding: 14px 16px;
  font-size: 18px;
  font-weight: bold;
  text-align: center;
  position: sticky;
  top: 0;
  z-index: 100;
}
.app-main {
  flex: 1;
  padding-bottom: 60px;
}
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: 480px;
  background: white;
  border-top: 1px solid #eee;
  display: flex;
  z-index: 100;
}
.nav-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
  text-decoration: none;
  color: #666;
  font-size: 12px;
  gap: 2px;
}
.nav-item.router-link-active {
  color: #e74c3c;
}
.nav-icon {
  font-size: 22px;
}
</style>
