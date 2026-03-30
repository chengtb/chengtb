<template>
  <div>
    <div class="d-flex justify-space-between align-center mb-4">
      <h2>传菜员管理</h2>
      <div class="d-flex gap-2">
        <v-btn prepend-icon="mdi-refresh" variant="tonal" @click="fetchWaiters">刷新</v-btn>
        <v-btn color="orange-darken-2" prepend-icon="mdi-plus" @click="openAdd">添加传菜员</v-btn>
      </div>
    </div>

    <v-data-table
      :headers="headers"
      :items="waiters"
      :loading="loading"
      item-value="waiter_id"
    >
      <template #item.area="{ item }">
        <v-chip v-if="item.area" color="orange-darken-2" variant="tonal" size="small">
          {{ item.area.name }}
        </v-chip>
        <span v-else class="text-grey text-caption">全区域</span>
      </template>
      <template #item.status="{ item }">
        <v-chip :color="item.status === 'online' ? 'success' : 'grey'" size="small">
          {{ item.status === 'online' ? '在线' : '离线' }}
        </v-chip>
      </template>
      <template #item.actions="{ item }">
        <v-btn size="small" icon="mdi-pencil" variant="text" @click="openEdit(item)" />
        <v-btn size="small" icon="mdi-delete" variant="text" color="error" @click="deleteWaiter(item)" />
      </template>
    </v-data-table>

    <v-dialog v-model="dialog" max-width="480">
      <v-card>
        <v-card-title>{{ editing ? '编辑传菜员' : '添加传菜员' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.name" label="姓名" variant="outlined" class="mb-3" />
          <v-text-field v-model="form.phone" label="手机号" variant="outlined" class="mb-3" />
          <v-select
            v-model="form.area_id"
            :items="areaOptions"
            item-title="label"
            item-value="value"
            label="负责区域"
            hint="留空表示负责全区域"
            variant="outlined"
            clearable
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="dialog = false">取消</v-btn>
          <v-btn color="orange-darken-2" @click="saveWaiter" :loading="saving">保存</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../api'

const waiters = ref([])
const areas = ref([])
const loading = ref(true)
const dialog = ref(false)
const editing = ref(false)
const saving = ref(false)
const form = ref({ waiter_id: null, name: '', phone: '', area_id: null })

const areaOptions = computed(() => [
  { label: '全区域（不限）', value: null },
  ...areas.value.map(a => ({ label: a.name, value: a.area_id })),
])

const headers = [
  { title: 'ID', key: 'waiter_id', width: 70 },
  { title: '姓名', key: 'name' },
  { title: '手机号', key: 'phone' },
  { title: '负责区域', key: 'area', sortable: false },
  { title: '状态', key: 'status', width: 100 },
  { title: '操作', key: 'actions', sortable: false, width: 120 },
]

async function fetchWaiters() {
  loading.value = true
  try {
    const res = await api.get('/waiters')
    waiters.value = res.data || []
  } finally {
    loading.value = false
  }
}

function openAdd() {
  editing.value = false
  form.value = { waiter_id: null, name: '', phone: '', area_id: null }
  dialog.value = true
}

function openEdit(waiter) {
  editing.value = true
  form.value = { waiter_id: waiter.waiter_id, name: waiter.name, phone: waiter.phone, area_id: waiter.area_id ?? null }
  dialog.value = true
}

async function saveWaiter() {
  saving.value = true
  try {
    if (editing.value) {
      await api.put(`/waiters/${form.value.waiter_id}`, form.value)
    } else {
      await api.post('/waiters', form.value)
    }
    dialog.value = false
    await fetchWaiters()
  } finally {
    saving.value = false
  }
}

async function deleteWaiter(waiter) {
  if (!confirm(`确认删除传菜员 ${waiter.name}?`)) return
  try {
    await api.delete(`/waiters/${waiter.waiter_id}`)
    await fetchWaiters()
  } catch {}
}

onMounted(async () => {
  await fetchWaiters()
  try {
    const res = await api.get('/areas')
    areas.value = res.data || []
  } catch {}
})
</script>
