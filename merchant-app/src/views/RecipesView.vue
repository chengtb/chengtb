<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>菜谱管理</h2>
      <div class="d-flex gap-2">
        <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchRecipes">刷新</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加菜谱</v-btn>
      </div>
    </div>

    <v-data-table :headers="headers" :items="recipes" :loading="loading" item-value="recipe_id">
      <template #item.is_enabled="{ item }">
        <v-chip :color="item.is_enabled ? 'success' : 'grey'" size="small">
          {{ item.is_enabled ? '启用' : '停用' }}
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
            item-value="dish_id"
            label="菜品"
            variant="outlined"
            class="mb-3"
          />
          <v-text-field v-model="form.name" label="菜谱名称" variant="outlined" class="mb-3" />
          <v-text-field v-model.number="form.portion" label="份量" type="number" variant="outlined" class="mb-3" />
          <v-switch v-model="form.is_enabled" label="启用" density="compact" />
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
const form = ref({ recipe_id: null, dish_id: null, name: '', portion: 1, is_enabled: true })

const headers = [
  { title: 'ID', key: 'recipe_id', width: 60 },
  { title: '菜谱名称', key: 'name' },
  { title: '菜品', key: 'dish_name' },
  { title: '份量', key: 'portion', width: 80 },
  { title: '状态', key: 'is_enabled', width: 100 },
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
  form.value = { recipe_id: null, dish_id: null, name: '', portion: 1, is_enabled: true }
  dialog.value = true
}

function openEdit(recipe) {
  editing.value = true
  form.value = { recipe_id: recipe.recipe_id, dish_id: recipe.dish_id, name: recipe.name, portion: recipe.portion, is_enabled: recipe.is_enabled }
  dialog.value = true
}

async function saveRecipe() {
  try {
    if (editing.value) {
      await api.put(`/recipes/${form.value.recipe_id}`, form.value)
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
    await api.delete(`/recipes/${recipe.recipe_id}`)
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
