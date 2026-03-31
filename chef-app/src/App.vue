<template>
  <v-app>
    <v-app-bar elevation="1" color="primary">
      <v-toolbar-title>厨师工作台</v-toolbar-title>
      <template #append>
        <v-btn v-if="chefName" variant="text" @click="logout" prepend-icon="mdi-logout">
          {{ chefName }}
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
const chefName = ref(localStorage.getItem('chef_name') || '')

function logout() {
  localStorage.removeItem('chef_id')
  localStorage.removeItem('chef_name')
  chefName.value = ''
  router.push('/login')
}

onMounted(() => {
  if (!localStorage.getItem('chef_id')) {
    router.push('/login')
  }
})
</script>
