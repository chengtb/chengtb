<template>
  <v-container max-width="480" class="mt-8">
    <v-card elevation="4" rounded="xl">
      <v-card-title class="text-center pa-6">
        <v-icon size="48" color="orange-darken-2" class="mb-2 d-block">mdi-food-takeout-box</v-icon>
        <div class="text-h5 font-weight-bold">传菜员登录</div>
      </v-card-title>
      <v-card-text class="pa-6">
        <v-text-field
          v-model="waiterName"
          label="请输入您的姓名"
          variant="outlined"
          prepend-inner-icon="mdi-account"
          class="mb-4"
          @keyup.enter="login"
          :error-messages="error"
        />
        <v-btn
          block
          color="orange-darken-2"
          size="x-large"
          :loading="loading"
          :disabled="!waiterName.trim()"
          @click="login"
          rounded="lg"
        >
          登录
        </v-btn>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'

const router = useRouter()
const waiterName = ref('')
const error = ref('')
const loading = ref(false)

async function login() {
  if (!waiterName.value.trim()) return
  loading.value = true
  error.value = ''
  try {
    const res = await api.post('/auth/login', { name: waiterName.value.trim() })
    const waiter = res.data
    localStorage.setItem('waiter_id', waiter.waiter_id)
    localStorage.setItem('waiter_name', waiter.name)
    localStorage.setItem('waiter_area', waiter.area_assigned || '')
    router.push('/tasks')
  } catch {
    error.value = '传菜员不存在，请联系管理员添加账号'
  } finally {
    loading.value = false
  }
}
</script>
