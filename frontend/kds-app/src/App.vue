<template>
  <v-app>
    <v-app-bar color="grey-darken-4" density="compact">
      <v-app-bar-title>
        <v-icon>mdi-chef-hat</v-icon>
        后厨管理 (KDS)
      </v-app-bar-title>
      <template #append>
        <v-badge :content="newTaskCount" color="error" v-if="newTaskCount > 0">
          <v-icon>mdi-bell</v-icon>
        </v-badge>
        <v-chip class="ml-2" color="info" size="small">
          {{ currentTime }}
        </v-chip>
      </template>
    </v-app-bar>

    <v-main>
      <router-view />
    </v-main>

    <v-snackbar v-model="showNotification" color="warning" timeout="3000" location="top">
      {{ notificationText }}
    </v-snackbar>
  </v-app>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const newTaskCount = ref(0)
const showNotification = ref(false)
const notificationText = ref('')
const currentTime = ref('')
let ws = null
let timer = null

function updateTime() {
  currentTime.value = new Date().toLocaleTimeString()
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)

  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${protocol}//${location.host}/ws?type=kds`)

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.type === 'new_tasks' || msg.type === 'task_assigned') {
        newTaskCount.value++
        notificationText.value = '新任务到达！'
        showNotification.value = true
        // Play notification sound
        try { new Audio('/notification.mp3').play() } catch {}
      }
    } catch (e) {
      console.error('WS parse error', e)
    }
  }

  ws.onclose = () => {
    setTimeout(() => location.reload(), 5000)
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  if (ws) ws.close()
})
</script>
