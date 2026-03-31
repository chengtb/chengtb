<template>
  <v-app>
    <v-app-bar elevation="1" color="orange-darken-2">
      <v-toolbar-title>
        <v-icon class="mr-2">mdi-food-takeout-box</v-icon>
        传菜工作台
      </v-toolbar-title>
      <template #append>
        <v-btn v-if="waiterName" variant="text" @click="logout" prepend-icon="mdi-logout">
          {{ waiterName }}
        </v-btn>
      </template>
    </v-app-bar>
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const waiterName = ref(localStorage.getItem('waiter_name') || '')

function logout() {
  localStorage.removeItem('waiter_id')
  localStorage.removeItem('waiter_name')
  localStorage.removeItem('waiter_area')
  waiterName.value = ''
  router.push('/login')
}

onMounted(() => {
  if (!localStorage.getItem('waiter_id')) {
    router.push('/login')
  }
})
</script>
