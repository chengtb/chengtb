<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>菜品分类管理</h2>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加分类</v-btn>
    </div>

    <v-data-table :headers="headers" :items="categories" :loading="loading" item-value="category_id">
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-pencil" variant="text" @click="openEdit(item)" />
        <v-btn size="small" icon="mdi-delete" variant="text" color="error" @click="deleteCategory(item)" />
      </template>
    </v-data-table>

    <v-dialog v-model="dialog" max-width="420">
      <v-card>
        <v-card-title>{{ editing ? '编辑分类' : '添加分类' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.name" label="分类名称" variant="outlined" class="mb-3" />
          <v-text-field v-model.number="form.sort_order" label="排序" type="number" variant="outlined" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dialog = false">取消</v-btn>
          <v-btn color="primary" @click="saveCategory">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const categories = ref([])
const loading = ref(true)
const dialog = ref(false)
const editing = ref(false)
const form = ref({ category_id: null, name: '', sort_order: 0 })

const headers = [
  { title: 'ID', key: 'category_id', width: 80 },
  { title: '名称', key: 'name' },
  { title: '排序', key: 'sort_order', width: 100 },
  { title: '操作', key: 'actions', sortable: false, width: 120 },
]

async function fetchCategories() {
  loading.value = true
  try {
    const res = await api.get('/categories')
    categories.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editing.value = false
  form.value = { category_id: null, name: '', sort_order: 0 }
  dialog.value = true
}

function openEdit(cat) {
  editing.value = true
  form.value = { ...cat }
  dialog.value = true
}

async function saveCategory() {
  try {
    if (editing.value) {
      await api.put(`/categories/${form.value.category_id}`, form.value)
    } else {
      await api.post('/categories', form.value)
    }
    dialog.value = false
    await fetchCategories()
  } catch {}
}

async function deleteCategory(cat) {
  if (!confirm(`确认删除分类 ${cat.name}?`)) return
  try {
    await api.delete(`/categories/${cat.category_id}`)
    await fetchCategories()
  } catch {}
}

onMounted(fetchCategories)
</script>
