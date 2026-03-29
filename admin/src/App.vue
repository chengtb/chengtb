<template>
  <div class="admin-layout">
    <aside class="sidebar">
      <div class="logo">🍜 餐厅管理</div>
      <nav>
        <router-link to="/tables" class="nav-link">🪑 桌台总览</router-link>
        <router-link to="/orders" class="nav-link">📋 订单管理</router-link>
        <router-link to="/menu" class="nav-link">🍽️ 菜单管理</router-link>
      </nav>
    </aside>
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.admin-layout { display: flex; min-height: 100vh; }
.sidebar {
  width: 220px; background: #1a1a2e; color: white;
  display: flex; flex-direction: column; flex-shrink: 0;
}
.logo { padding: 24px 20px; font-size: 20px; font-weight: bold; border-bottom: 1px solid #333; }
nav { padding: 16px 0; }
.nav-link {
  display: block; padding: 14px 20px; color: #aaa; text-decoration: none;
  font-size: 15px; transition: all 0.2s;
}
.nav-link:hover { background: #16213e; color: white; }
.nav-link.router-link-active { background: #e74c3c; color: white; }
.main-content { flex: 1; padding: 24px; overflow: auto; }
</style>
