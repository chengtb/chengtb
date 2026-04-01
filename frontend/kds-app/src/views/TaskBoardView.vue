<template>
  <v-container fluid>
    <v-row>
      <!-- Pending Tasks Column -->
      <v-col cols="12" md="3">
        <v-card color="orange-darken-4">
          <v-card-title class="text-center">
            <v-icon>mdi-clock-outline</v-icon>
            待处理
            <v-badge :content="pendingTasks.length" inline />
          </v-card-title>
          <v-card-text>
            <task-card
              v-for="task in pendingTasks"
              :key="task.id"
              :task="task"
              @complete="handleComplete"
            />
            <v-alert v-if="pendingTasks.length === 0" type="info" variant="tonal" class="mt-2">
              暂无待处理任务
            </v-alert>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- Assigned Tasks Column -->
      <v-col cols="12" md="3">
        <v-card color="blue-darken-4">
          <v-card-title class="text-center">
            <v-icon>mdi-account-check</v-icon>
            已分配
            <v-badge :content="assignedTasks.length" inline />
          </v-card-title>
          <v-card-text>
            <task-card
              v-for="task in assignedTasks"
              :key="task.id"
              :task="task"
              @complete="handleComplete"
            />
            <v-alert v-if="assignedTasks.length === 0" type="info" variant="tonal" class="mt-2">
              暂无已分配任务
            </v-alert>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- Cooking Tasks Column -->
      <v-col cols="12" md="3">
        <v-card color="red-darken-4">
          <v-card-title class="text-center">
            <v-icon>mdi-fire</v-icon>
            烹饪中
            <v-badge :content="cookingTasks.length" inline />
          </v-card-title>
          <v-card-text>
            <task-card
              v-for="task in cookingTasks"
              :key="task.id"
              :task="task"
              @complete="handleComplete"
            />
            <v-alert v-if="cookingTasks.length === 0" type="info" variant="tonal" class="mt-2">
              暂无烹饪中任务
            </v-alert>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- Completed Tasks Column -->
      <v-col cols="12" md="3">
        <v-card color="green-darken-4">
          <v-card-title class="text-center">
            <v-icon>mdi-check-circle</v-icon>
            已完成
            <v-badge :content="completedTasks.length" inline />
          </v-card-title>
          <v-card-text>
            <task-card
              v-for="task in completedTasks"
              :key="task.id"
              :task="task"
              :show-actions="false"
            />
            <v-alert v-if="completedTasks.length === 0" type="info" variant="tonal" class="mt-2">
              暂无已完成任务
            </v-alert>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { ref, computed, onMounted, onUnmounted, defineComponent, h } from 'vue'
import { listTasks, completeTask } from '../api/index.js'

// Inline TaskCard component
const TaskCard = defineComponent({
  name: 'TaskCard',
  props: {
    task: { type: Object, required: true },
    showActions: { type: Boolean, default: true },
  },
  emits: ['complete'],
  setup(props, { emit }) {
    const elapsedTime = ref('')
    let timer = null

    function updateElapsed() {
      const created = new Date(props.task.created_at)
      const now = new Date()
      const diff = Math.floor((now - created) / 1000)
      const mins = Math.floor(diff / 60)
      const secs = diff % 60
      elapsedTime.value = `${mins}:${secs.toString().padStart(2, '0')}`
    }

    onMounted(() => {
      updateElapsed()
      timer = setInterval(updateElapsed, 1000)
    })

    onUnmounted(() => {
      if (timer) clearInterval(timer)
    })

    return () => h('div', { class: 'task-card mb-2' }, [
      h('div', { class: 'v-card v-card--variant-elevated pa-3' }, [
        h('div', { class: 'd-flex justify-space-between align-center' }, [
          h('span', { class: 'text-caption' }, `#${props.task.batch_no || props.task.id}`),
          h('span', { class: 'text-caption font-weight-bold' }, elapsedTime.value),
        ]),
        h('div', { class: 'text-h6 mt-1' }, props.task.recipe?.product?.name || `任务 #${props.task.id}`),
        h('div', { class: 'd-flex justify-space-between align-center mt-1' }, [
          h('span', null, `数量: ${props.task.target_quantity}`),
          props.task.type === 'merged'
            ? h('span', { class: 'v-chip v-chip--size-x-small bg-warning' }, '合单')
            : null,
        ]),
        props.task.task_items?.length
          ? h('div', { class: 'text-caption mt-1' },
              props.task.task_items.map(ti =>
                h('div', null, `桌 ${ti.table?.table_number || ti.table_id}`)
              )
            )
          : null,
        props.task.chef
          ? h('div', { class: 'text-caption mt-1' }, `厨师: ${props.task.chef.name}`)
          : null,
        props.showActions && props.task.status !== 'completed'
          ? h('div', { class: 'mt-2' }, [
              h('button', {
                class: 'v-btn v-btn--size-small v-btn--variant-flat bg-success',
                style: 'width: 100%',
                onClick: () => emit('complete', props.task.id),
              }, '完成'),
            ])
          : null,
      ]),
    ])
  },
})

export default defineComponent({
  name: 'TaskBoardView',
  components: { TaskCard },
  setup() {
    const tasks = ref([])
    const loading = ref(false)
    let refreshTimer = null

    const pendingTasks = computed(() => tasks.value.filter(t => t.status === 'pending'))
    const assignedTasks = computed(() => tasks.value.filter(t => t.status === 'assigned'))
    const cookingTasks = computed(() => tasks.value.filter(t => t.status === 'cooking'))
    const completedTasks = computed(() => tasks.value.filter(t => t.status === 'completed'))

    async function loadTasks() {
      loading.value = true
      try {
        const res = await listTasks({ page: 1, page_size: 100 })
        tasks.value = res.data?.tasks || []
      } catch (e) {
        console.error('Load tasks failed', e)
      }
      loading.value = false
    }

    async function handleComplete(taskId) {
      try {
        await completeTask(taskId)
        await loadTasks()
      } catch (e) {
        console.error('Complete task failed', e)
      }
    }

    onMounted(() => {
      loadTasks()
      refreshTimer = setInterval(loadTasks, 10000) // Auto-refresh every 10s
    })

    onUnmounted(() => {
      if (refreshTimer) clearInterval(refreshTimer)
    })

    return {
      pendingTasks,
      assignedTasks,
      cookingTasks,
      completedTasks,
      handleComplete,
    }
  },
})
</script>

<style scoped>
.task-card .v-card {
  border-left: 4px solid currentColor;
}
</style>
