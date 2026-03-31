<template>
  <v-app>
    <v-app-bar color="primary" density="compact">
      <v-app-bar-title>服务员工作台</v-app-bar-title>
      <template #append>
        <v-badge :content="pendingCount" color="error" v-if="pendingCount > 0">
          <v-icon>mdi-bell</v-icon>
        </v-badge>
      </template>
    </v-app-bar>

    <v-navigation-drawer permanent>
      <v-list nav>
        <v-list-item prepend-icon="mdi-clipboard-list" title="订单管理" to="/" />
        <v-list-item prepend-icon="mdi-table-furniture" title="餐桌概览" to="/tables" />
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <router-view />
    </v-main>

    <v-snackbar v-model="showNotification" color="info" timeout="5000" location="top right">
      {{ notificationText }}
    </v-snackbar>
  </v-app>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const pendingCount = ref(0)
const showNotification = ref(false)
const notificationText = ref('')

let ws = null

onMounted(() => {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${protocol}//${location.host}/ws?type=waiter`)

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'new_order') {
        pendingCount.value++
        notificationText.value = `新订单！桌号：${msg.table_id}`
        showNotification.value = true
      } else if (msg.type === 'dish_ready') {
        notificationText.value = '有菜品已完成，请取餐上桌'
        showNotification.value = true
      }
    } catch (e) {
      console.error('WS parse error', e)
    }
  }

  ws.onclose = () => {
    setTimeout(() => location.reload(), 5000)
  }
})
</script>
