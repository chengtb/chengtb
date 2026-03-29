<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>厨师管理</h2>
      <div class="d-flex gap-2">
        <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchChefs">刷新</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加厨师</v-btn>
      </div>
    </div>

    <v-data-table
      :headers="headers"
      :items="chefs"
      :loading="loading"
      item-value="chef_id"
      show-expand
      v-model:expanded="expanded"
    >
      <template #item.recipes="{ item }">
        <v-chip size="small" :color="(item.recipes || []).length ? 'primary' : 'default'">
          {{ (item.recipes || []).length }} 个菜谱
        </v-chip>
      </template>
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-pencil" variant="text" @click="openEdit(item)" />
        <v-btn size="small" icon="mdi-delete" variant="text" color="error" @click="deleteChef(item)" />
      </template>
      <template #expanded-row="{ columns, item }">
        <tr>
          <td :colspan="columns.length" class="pa-0">
            <v-sheet class="pa-4 bg-grey-lighten-5">
              <div class="text-subtitle-2 mb-3 text-medium-emphasis">擅长菜谱</div>
              <div v-if="!(item.recipes || []).length" class="text-body-2 text-disabled">暂未关联菜谱</div>
              <v-table v-else density="compact" class="rounded border">
                <thead>
                  <tr>
                    <th>菜谱ID</th>
                    <th>菜谱名称</th>
                    <th>份量</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="r in item.recipes" :key="r.recipe_id">
                    <td>{{ r.recipe_id }}</td>
                    <td>{{ r.name }}</td>
                    <td>{{ r.portion }}</td>
                    <td>
                      <v-chip :color="r.is_enabled ? 'success' : 'grey'" size="x-small">
                        {{ r.is_enabled ? '启用' : '停用' }}
                      </v-chip>
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </v-sheet>
          </td>
        </tr>
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
            item-value="recipe_id"
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
const expanded = ref([])
const form = ref({ chef_id: null, name: '', max_load: 5, recipe_ids: [] })

const headers = [
  { title: 'ID', key: 'chef_id', width: 60 },
  { title: '姓名', key: 'name' },
  { title: '最大负载', key: 'max_load', width: 100 },
  { title: '当前负载', key: 'current_load', width: 100 },
  { title: '擅长菜谱', key: 'recipes', sortable: false },
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
  form.value = { chef_id: null, name: '', max_load: 5, recipe_ids: [] }
  dialog.value = true
}

function openEdit(chef) {
  editing.value = true
  form.value = { chef_id: chef.chef_id, name: chef.name, max_load: chef.max_load, recipe_ids: (chef.recipes || []).map(r => r.recipe_id) }
  dialog.value = true
}

async function saveChef() {
  try {
    if (editing.value) {
      await api.put(`/chefs/${form.value.chef_id}`, form.value)
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
    await api.delete(`/chefs/${chef.chef_id}`)
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
