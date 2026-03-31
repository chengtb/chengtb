<template>
  <v-container fluid>
    <v-row>
      <v-col v-for="region in regions" :key="region.id" cols="12" md="6">
        <v-card>
          <v-card-title>{{ region.name }}</v-card-title>
          <v-card-text>
            <v-row>
              <v-col
                v-for="table in region.tables"
                :key="table.id"
                cols="4"
                sm="3"
              >
                <v-card
                  :color="tableColor(table.status)"
                  variant="outlined"
                  class="text-center pa-4"
                  @click="viewTableOrders(table)"
                >
                  <v-icon size="32">mdi-table-furniture</v-icon>
                  <div class="text-subtitle-2 mt-1">{{ table.table_number }}</div>
                  <v-chip :color="tableColor(table.status)" size="x-small" class="mt-1">
                    {{ tableStatusText(table.status) }}
                  </v-chip>
                </v-card>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { listRegions } from '../api/index.js'

const regions = ref([])

onMounted(async () => {
  try {
    const res = await listRegions()
    regions.value = res.data || []
  } catch (e) {
    console.error('Load regions failed', e)
  }
})

function tableColor(status) {
  const map = { free: 'success', occupied: 'warning', reserved: 'info' }
  return map[status] || 'default'
}

function tableStatusText(status) {
  const map = { free: '空闲', occupied: '用餐中', reserved: '已预约' }
  return map[status] || status
}

function viewTableOrders(table) {
  // TODO: show orders for this table
  console.log('View table', table)
}
</script>
