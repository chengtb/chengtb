<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>菜品管理</h2>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加菜品</v-btn>
    </div>

    <v-data-table :headers="headers" :items="dishes" :loading="loading" item-value="id">
      <template #item.image_url="{ item }">
        <v-img v-if="item.image_url" :src="item.image_url" width="48" height="48" cover rounded />
        <v-icon v-else>mdi-image-off</v-icon>
      </template>
      <template #item.price="{ item }">
        ¥{{ (item.price / 100).toFixed(2) }}
      </template>
      <template #item.is_available="{ item }">
        <v-switch :model-value="item.is_available" density="compact" hide-details @change="toggleAvailable(item)" />
      </template>
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-pencil" variant="text" @click="openEdit(item)" />
        <v-btn size="small" icon="mdi-delete" variant="text" color="error" @click="deleteDish(item)" />
      </template>
    </v-data-table>

    <v-dialog v-model="dialog" max-width="480">
      <v-card>
        <v-card-title>{{ editing ? '编辑菜品' : '添加菜品' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.name" label="菜品名称" variant="outlined" class="mb-3" />
          <v-text-field v-model.number="form.price" label="价格(分)" type="number" variant="outlined" class="mb-3" />
          <v-text-field v-model="form.image_url" label="图片URL" variant="outlined" class="mb-3" />
          <v-select
            v-model="form.category_id"
            :items="categories"
            item-title="name"
            item-value="id"
            label="分类"
            variant="outlined"
            class="mb-3"
          />
          <v-switch v-model="form.is_available" label="可供应" density="compact" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dialog = false">取消</v-btn>
          <v-btn color="primary" @click="saveDish">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const dishes = ref([])
const categories = ref([])
const loading = ref(true)
const dialog = ref(false)
const editing = ref(false)
const form = ref({ id: null, name: '', price: 0, image_url: '', category_id: null, is_available: true })

const headers = [
  { title: '图片', key: 'image_url', sortable: false, width: 70 },
  { title: '名称', key: 'name' },
  { title: '价格', key: 'price', width: 100 },
  { title: '分类', key: 'category_name', width: 120 },
  { title: '供应', key: 'is_available', width: 80 },
  { title: '操作', key: 'actions', sortable: false, width: 120 },
]

async function fetchDishes() {
  loading.value = true
  try {
    const res = await api.get('/dishes')
    dishes.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editing.value = false
  form.value = { id: null, name: '', price: 0, image_url: '', category_id: null, is_available: true }
  dialog.value = true
}

function openEdit(dish) {
  editing.value = true
  form.value = { ...dish }
  dialog.value = true
}

async function saveDish() {
  try {
    if (editing.value) {
      await api.put(`/dishes/${form.value.id}`, form.value)
    } else {
      await api.post('/dishes', form.value)
    }
    dialog.value = false
    await fetchDishes()
  } catch {}
}

async function deleteDish(dish) {
  if (!confirm(`确认删除菜品 ${dish.name}?`)) return
  try {
    await api.delete(`/dishes/${dish.id}`)
    await fetchDishes()
  } catch {}
}

async function toggleAvailable(dish) {
  try {
    await api.put(`/dishes/${dish.id}`, { ...dish, is_available: !dish.is_available })
    await fetchDishes()
  } catch {}
}

onMounted(async () => {
  await fetchDishes()
  try {
    const catRes = await api.get('/categories')
    categories.value = catRes.data || []
  } catch {}
})
</script>
