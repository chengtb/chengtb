<template>
  <span class="status-badge" :class="`status-badge--${status}`">
    <span class="status-badge__dot"></span>
    <span class="status-badge__text">{{ statusText }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { StepStatus } from '../types/process'

const props = defineProps<{
  status: StepStatus
}>()

const statusText = computed(() => {
  const map: Record<StepStatus, string> = {
    pending: '待处理',
    active: '进行中',
    completed: '已完成',
    rejected: '已驳回',
    skipped: '已跳过',
  }
  return map[props.status]
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.status-badge__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-badge--pending {
  background: #f5f5f5;
  color: #909399;
  border: 1px solid #dcdfe6;
}
.status-badge--pending .status-badge__dot {
  background: #909399;
}

.status-badge--active {
  background: #ecf5ff;
  color: #409eff;
  border: 1px solid #b3d8ff;
}
.status-badge--active .status-badge__dot {
  background: #409eff;
  animation: pulse 1.5s ease-in-out infinite;
}

.status-badge--completed {
  background: #f0f9eb;
  color: #67c23a;
  border: 1px solid #c2e7b0;
}
.status-badge--completed .status-badge__dot {
  background: #67c23a;
}

.status-badge--rejected {
  background: #fef0f0;
  color: #f56c6c;
  border: 1px solid #fbc4c4;
}
.status-badge--rejected .status-badge__dot {
  background: #f56c6c;
}

.status-badge--skipped {
  background: #fdf6ec;
  color: #e6a23c;
  border: 1px solid #f5dab1;
}
.status-badge--skipped .status-badge__dot {
  background: #e6a23c;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.3); }
}
</style>
