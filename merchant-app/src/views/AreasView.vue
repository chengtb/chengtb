<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>用餐区域管理</h2>
      <div class="d-flex gap-2">
        <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchAreas">刷新</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加区域</v-btn>
      </div>
    </div>

    <v-data-table
      :headers="headers"
      :items="areas"
      :loading="loading"
      item-value="area_id"
    >
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-pencil" variant="text" @click="openEdit(item)" />
        <v-btn size="small" icon="mdi-delete" variant="text" color="error" @click="deleteArea(item)" />
      </template>
    </v-data-table>

    <v-dialog v-model="dialog" max-width="480">
      <v-card>
        <v-card-title>{{ editing ? '编辑区域' : '添加区域' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.name" label="区域名称" variant="outlined" class="mb-3" />
          <v-text-field v-model="form.description" label="备注说明" variant="outlined" class="mb-3" />
          <v-text-field v-model.number="form.sort_order" label="排序（数字越小越靠前）" type="number" variant="outlined" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dialog = false">取消</v-btn>
          <v-btn color="primary" @click="saveArea" :loading="saving">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const areas = ref([])
const loading = ref(true)
const dialog = ref(false)
const editing = ref(false)
const saving = ref(false)
const form = ref({ area_id: null, name: '', description: '', sort_order: 0 })

const headers = [
  { title: 'ID', key: 'area_id', width: 70 },
  { title: '区域名称', key: 'name' },
  { title: '备注说明', key: 'description' },
  { title: '排序', key: 'sort_order', width: 80 },
  { title: '操作', key: 'actions', sortable: false, width: 120 },
]

async function fetchAreas() {
  loading.value = true
  try {
    const res = await api.get('/areas')
    areas.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editing.value = false
  form.value = { area_id: null, name: '', description: '', sort_order: 0 }
  dialog.value = true
}

function openEdit(area) {
  editing.value = true
  form.value = { ...area }
  dialog.value = true
}

async function saveArea() {
  saving.value = true
  try {
    if (editing.value) {
      await api.put(`/areas/${form.value.area_id}`, form.value)
    } else {
      await api.post('/areas', form.value)
    }
    dialog.value = false
    await fetchAreas()
  } finally {
    saving.value = false
  }
}

async function deleteArea(area) {
  if (!confirm(`确认删除区域「${area.name}」？\n已关联此区域的餐桌和传菜员将不受影响，但不再归属任何区域。`)) return
  try {
    await api.delete(`/areas/${area.area_id}`)
    await fetchAreas()
  } catch {}
}

onMounted(fetchAreas)
</script>
