<template>
  <v-container max-width="400" class="mt-10">
    <v-card>
      <v-card-title class="text-center pa-4">厨师登录</v-card-title>
      <v-card-text>
        <v-text-field
          v-model="chefId"
          label="厨师ID"
          variant="outlined"
          type="number"
          @update:model-value="fetchChefInfo"
          class="mb-3"
        />
        <div v-if="chefName" class="text-center mb-3">
          <v-chip color="success" prepend-icon="mdi-account">{{ chefName }}</v-chip>
        </div>
        <div v-if="error" class="text-error text-center mb-3">{{ error }}</div>
        <v-btn
          block
          color="primary"
          :disabled="!chefName"
          @click="login"
          size="large"
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
const chefId = ref('')
const chefName = ref('')
const error = ref('')

async function fetchChefInfo() {
  if (!chefId.value) { chefName.value = ''; return }
  try {
    const res = await api.get(`/chefs/${chefId.value}`)
    chefName.value = res.data.name || ''
    error.value = ''
  } catch {
    chefName.value = ''
    error.value = '厨师不存在'
  }
}

function login() {
  localStorage.setItem('chef_id', chefId.value)
  localStorage.setItem('chef_name', chefName.value)
  router.push('/tasks')
}
</script>
