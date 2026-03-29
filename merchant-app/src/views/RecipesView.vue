<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>菜谱管理</h2>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加菜谱</v-btn>
    </div>

    <v-data-table :headers="headers" :items="recipes" :loading="loading" item-value="id" :group-by="[{ key: 'dish_name' }]">
      <template #item.status="{ item }">
        <v-chip :color="item.status === 'active' ? 'success' : 'grey'" size="small">
          {{ item.status === 'active' ? '启用' : '停用' }}
        </v-chip>
      </template>
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-pencil" variant="text" @click="openEdit(item)" />
        <v-btn size="small" icon="mdi-delete" variant="text" color="error" @click="deleteRecipe(item)" />
      </template>
    </v-data-table>

    <v-dialog v-model="dialog" max-width="480">
      <v-card>
        <v-card-title>{{ editing ? '编辑菜谱' : '添加菜谱' }}</v-card-title>
        <v-card-text>
          <v-select
            v-model="form.dish_id"
            :items="dishes"
            item-title="name"
            item-value="id"
            label="菜品"
            variant="outlined"
            class="mb-3"
          />
          <v-text-field v-model="form.name" label="菜谱名称" variant="outlined" class="mb-3" />
          <v-text-field v-model.number="form.portion" label="份量" type="number" variant="outlined" class="mb-3" />
          <v-select
            v-model="form.status"
            :items="[{title:'启用',value:'active'},{title:'停用',value:'inactive'}]"
            item-title="title"
            item-value="value"
            label="状态"
            variant="outlined"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dialog = false">取消</v-btn>
          <v-btn color="primary" @click="saveRecipe">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const recipes = ref([])
const dishes = ref([])
const loading = ref(true)
const dialog = ref(false)
const editing = ref(false)
const form = ref({ id: null, dish_id: null, name: '', portion: 1, status: 'active' })

const headers = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '菜谱名称', key: 'name' },
  { title: '菜品', key: 'dish_name' },
  { title: '份量', key: 'portion', width: 80 },
  { title: '状态', key: 'status', width: 100 },
  { title: '操作', key: 'actions', sortable: false, width: 120 },
]

async function fetchRecipes() {
  loading.value = true
  try {
    const res = await api.get('/recipes')
    recipes.value = (res.data || []).map(r => ({ ...r, dish_name: r.dish?.name || '' }))
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editing.value = false
  form.value = { id: null, dish_id: null, name: '', portion: 1, status: 'active' }
  dialog.value = true
}

function openEdit(recipe) {
  editing.value = true
  form.value = { id: recipe.id, dish_id: recipe.dish_id, name: recipe.name, portion: recipe.portion, status: recipe.status }
  dialog.value = true
}

async function saveRecipe() {
  try {
    if (editing.value) {
      await api.put(`/recipes/${form.value.id}`, form.value)
    } else {
      await api.post('/recipes', form.value)
    }
    dialog.value = false
    await fetchRecipes()
  } catch {}
}

async function deleteRecipe(recipe) {
  if (!confirm(`确认删除菜谱 ${recipe.name}?`)) return
  try {
    await api.delete(`/recipes/${recipe.id}`)
    await fetchRecipes()
  } catch {}
}

onMounted(async () => {
  await fetchRecipes()
  try {
    const res = await api.get('/dishes')
    dishes.value = res.data || []
  } catch {}
})
</script>
