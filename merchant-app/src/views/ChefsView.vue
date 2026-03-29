<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>厨师管理</h2>
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加厨师</v-btn>
    </div>

    <v-data-table :headers="headers" :items="chefs" :loading="loading" item-value="id">
      <template #item.recipes="{ item }">
        <v-chip v-for="r in (item.recipes || [])" :key="r.id" size="small" class="mr-1">{{ r.name }}</v-chip>
      </template>
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-pencil" variant="text" @click="openEdit(item)" />
        <v-btn size="small" icon="mdi-delete" variant="text" color="error" @click="deleteChef(item)" />
      </template>
    </v-data-table>

    <v-dialog v-model="dialog" max-width="480">
      <v-card>
        <v-card-title>{{ editing ? '编辑厨师' : '添加厨师' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.name" label="姓名" variant="outlined" class="mb-3" />
          <v-text-field v-model.number="form.max_load" label="最大负载" type="number" variant="outlined" class="mb-3" />
          <v-select
            v-model="form.recipe_ids"
            :items="allRecipes"
            item-title="name"
            item-value="id"
            label="擅长菜谱"
            multiple
            chips
            variant="outlined"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dialog = false">取消</v-btn>
          <v-btn color="primary" @click="saveChef">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'

const chefs = ref([])
const loading = ref(true)
const dialog = ref(false)
const editing = ref(false)
const allRecipes = ref([])
const form = ref({ id: null, name: '', max_load: 5, recipe_ids: [] })

const headers = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '姓名', key: 'name' },
  { title: '最大负载', key: 'max_load', width: 100 },
  { title: '当前负载', key: 'current_load', width: 100 },
  { title: '擅长菜谱', key: 'recipes' },
  { title: '操作', key: 'actions', sortable: false, width: 120 },
]

async function fetchChefs() {
  loading.value = true
  try {
    const res = await api.get('/chefs')
    chefs.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editing.value = false
  form.value = { id: null, name: '', max_load: 5, recipe_ids: [] }
  dialog.value = true
}

function openEdit(chef) {
  editing.value = true
  form.value = { id: chef.id, name: chef.name, max_load: chef.max_load, recipe_ids: (chef.recipes || []).map(r => r.id) }
  dialog.value = true
}

async function saveChef() {
  try {
    if (editing.value) {
      await api.put(`/chefs/${form.value.id}`, form.value)
    } else {
      await api.post('/chefs', form.value)
    }
    dialog.value = false
    await fetchChefs()
  } catch {}
}

async function deleteChef(chef) {
  if (!confirm(`确认删除厨师 ${chef.name}?`)) return
  try {
    await api.delete(`/chefs/${chef.id}`)
    await fetchChefs()
  } catch {}
}

onMounted(async () => {
  await fetchChefs()
  try {
    const res = await api.get('/recipes')
    allRecipes.value = res.data || []
  } catch {}
})
</script>
