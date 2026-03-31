<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>菜品分类管理</h2>
      <div class="d-flex gap-2">
        <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchCategories">刷新</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" @click="openAdd">添加分类</v-btn>
      </div>
    </div>

    <v-data-table
      :headers="headers"
      :items="categories"
      :loading="loading"
      item-value="category_id"
      show-expand
      v-model:expanded="expanded"
    >
      <template #item.dishes="{ item }">
        <v-chip size="small" :color="(item.dishes || []).length ? 'primary' : 'default'">
          {{ (item.dishes || []).length }} 个菜品
        </v-chip>
      </template>
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-pencil" variant="text" @click="openEdit(item)" />
        <v-btn size="small" icon="mdi-delete" variant="text" color="error" @click="deleteCategory(item)" />
      </template>
      <template #expanded-row="{ columns, item }">
        <tr>
          <td :colspan="columns.length" class="pa-0">
            <v-sheet class="pa-4 bg-grey-lighten-5">
              <div class="text-subtitle-2 mb-3 text-medium-emphasis">关联菜品</div>
              <div v-if="!(item.dishes || []).length" class="text-body-2 text-disabled">暂无关联菜品</div>
              <v-table v-else density="compact" class="rounded border">
                <thead>
                  <tr>
                    <th>菜品ID</th>
                    <th>名称</th>
                    <th>价格</th>
                    <th>特色</th>
                    <th>状态</th>
                    <th>排序</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(d, idx) in item.dishes" :key="d.dish_id">
                    <td>{{ d.dish_id }}</td>
                    <td>{{ d.name }}</td>
                    <td>¥{{ (d.price / 100).toFixed(2) }}</td>
                    <td>
                      <v-chip v-if="d.special_flag" color="orange" size="x-small">特色</v-chip>
                      <span v-else class="text-disabled">—</span>
                    </td>
                    <td>
                      <v-chip :color="d.is_available ? 'success' : 'grey'" size="x-small">
                        {{ d.is_available ? '上架' : '下架' }}
                      </v-chip>
                    </td>
                    <td style="width:60px">{{ d.sort_order }}</td>
                    <td style="white-space:nowrap">
                      <v-btn
                        icon="mdi-arrow-up"
                        size="x-small"
                        variant="text"
                        :disabled="idx === 0"
                        @click="moveDish(item, idx, -1)"
                      />
                      <v-btn
                        icon="mdi-arrow-down"
                        size="x-small"
                        variant="text"
                        :disabled="idx === item.dishes.length - 1"
                        @click="moveDish(item, idx, 1)"
                      />
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </v-sheet>
          </td>
        </tr>
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
const expanded = ref([])
const form = ref({ category_id: null, name: '', sort_order: 0 })

const headers = [
  { title: 'ID', key: 'category_id', width: 80 },
  { title: '名称', key: 'name' },
  { title: '排序', key: 'sort_order', width: 100 },
  { title: '关联菜品', key: 'dishes', sortable: false },
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

async function moveDish(category, idx, direction) {
  const dishes = category.dishes
  const target = idx + direction
  if (target < 0 || target >= dishes.length) return
  // swap sort_order values
  const a = dishes[idx]
  const b = dishes[target]
  const tmpOrder = a.sort_order
  a.sort_order = b.sort_order
  b.sort_order = tmpOrder
  // if both have same sort_order, spread them out
  if (a.sort_order === b.sort_order) {
    a.sort_order = idx * 10
    b.sort_order = target * 10
  }
  try {
    await Promise.all([
      api.put(`/dishes/${a.dish_id}`, a),
      api.put(`/dishes/${b.dish_id}`, b),
    ])
    await fetchCategories()
  } catch {}
}

onMounted(fetchCategories)
</script>
