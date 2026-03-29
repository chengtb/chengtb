<template>
  <div>
    <h2 class="mb-4">系统配置</h2>
    <v-card max-width="600">
      <v-card-text>
        <div v-if="loading" class="text-center pa-4">
          <v-progress-circular indeterminate />
        </div>
        <v-form v-else @submit.prevent="saveConfig">
          <v-switch v-model="config.enable_auto_dispatch" label="启用自动派发" class="mb-2" />
          <v-text-field
            v-model.number="config.time_window_minutes"
            label="时间窗口（分钟）"
            type="number"
            variant="outlined"
            class="mb-3"
          />
          <v-switch v-model="config.allow_combination_split" label="允许组合拆分" class="mb-2" />
          <v-text-field
            v-model.number="config.max_recipe_portion"
            label="最大菜谱份量"
            type="number"
            variant="outlined"
            class="mb-3"
          />
          <v-select
            v-model="config.load_calculation_method"
            :items="[{title:'按任务数',value:'by_tasks'},{title:'按份量',value:'by_portions'}]"
            item-title="title"
            item-value="value"
            label="负载计算方式"
            variant="outlined"
            class="mb-3"
          />
          <v-text-field
            v-model.number="config.chef_max_load_default"
            label="厨师默认最大负载"
            type="number"
            variant="outlined"
            class="mb-3"
          />
          <v-btn type="submit" color="primary" :loading="saving">保存配置</v-btn>
        </v-form>
      </v-card-text>
    </v-card>
    <v-snackbar v-model="snackbar" :color="snackColor" timeout="3000">{{ snackMsg }}</v-snackbar>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const config = ref({
  enable_auto_dispatch: false,
  time_window_minutes: 10,
  allow_combination_split: false,
  max_recipe_portion: 4,
  load_calculation_method: 'by_tasks',
  chef_max_load_default: 5,
})
const loading = ref(true)
const saving = ref(false)
const snackbar = ref(false)
const snackMsg = ref('')
const snackColor = ref('success')

async function fetchConfig() {
  try {
    const res = await api.get('/config')
    config.value = { ...config.value, ...res.data }
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  saving.value = true
  try {
    await api.put('/config', config.value)
    snackMsg.value = '保存成功'
    snackColor.value = 'success'
  } catch {
    snackMsg.value = '保存失败'
    snackColor.value = 'error'
  } finally {
    saving.value = false
    snackbar.value = true
  }
}

onMounted(fetchConfig)
</script>
