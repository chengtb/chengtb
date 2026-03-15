<template>
  <div class="process-flow">
    <div class="process-flow__steps">
      <template v-for="(step, index) in process.steps" :key="step.id">
        <div
          class="step-node"
          :class="[
            `step-node--${step.status}`,
            { 'step-node--active': step.id === process.currentStepId },
            { 'step-node--selected': selectedStepId === step.id },
          ]"
          @click="$emit('select-step', step.id)"
        >
          <div class="step-node__index">
            <span v-if="step.status === 'completed'" class="step-icon step-icon--check">✓</span>
            <span v-else-if="step.status === 'rejected'" class="step-icon step-icon--reject">✕</span>
            <span v-else class="step-icon">{{ index + 1 }}</span>
          </div>
          <div class="step-node__content">
            <div class="step-node__name">{{ step.name }}</div>
            <div class="step-node__dept">{{ step.department || '—' }}</div>
            <StatusBadge :status="step.status" />
          </div>
          <div v-if="step.id === process.currentStepId" class="step-node__current-tag">
            当前
          </div>
        </div>
        <div
          v-if="index < process.steps.length - 1"
          class="step-connector"
          :class="`step-connector--${getConnectorStatus(step.status)}`"
        >
          <div class="step-connector__line"></div>
          <div class="step-connector__arrow">›</div>
        </div>
      </template>
    </div>

    <div class="process-flow__legend">
      <span class="legend-item"><span class="legend-dot legend-dot--completed"></span>已完成</span>
      <span class="legend-item"><span class="legend-dot legend-dot--active"></span>进行中</span>
      <span class="legend-item"><span class="legend-dot legend-dot--pending"></span>待处理</span>
      <span class="legend-item"><span class="legend-dot legend-dot--rejected"></span>已驳回</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Process, StepStatus } from '../types/process'
import StatusBadge from './StatusBadge.vue'

defineProps<{
  process: Process
  selectedStepId?: string
}>()

defineEmits<{
  'select-step': [stepId: string]
}>()

function getConnectorStatus(prevStatus: StepStatus): string {
  if (prevStatus === 'completed') return 'completed'
  if (prevStatus === 'active') return 'active'
  return 'pending'
}
</script>

<style scoped>
.process-flow {
  padding: 16px 0;
}

.process-flow__steps {
  display: flex;
  align-items: flex-start;
  overflow-x: auto;
  padding: 8px 16px 16px;
  gap: 0;
}

.step-node {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 14px 12px;
  min-width: 120px;
  max-width: 140px;
  border-radius: 10px;
  border: 2px solid #e4e7ed;
  background: #fff;
  cursor: pointer;
  transition: all 0.25s ease;
  flex-shrink: 0;
}

.step-node:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.step-node--selected {
  border-color: #409eff !important;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.2) !important;
}

.step-node--completed {
  border-color: #67c23a;
  background: #f0f9eb;
}

.step-node--active {
  border-color: #409eff;
  background: #ecf5ff;
}

.step-node--rejected {
  border-color: #f56c6c;
  background: #fef0f0;
}

.step-node__current-tag {
  position: absolute;
  top: -10px;
  right: -8px;
  background: #ff6b35;
  color: #fff;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 8px;
  font-weight: 600;
}

.step-node__index {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  background: #e4e7ed;
  color: #606266;
  flex-shrink: 0;
}

.step-node--completed .step-node__index {
  background: #67c23a;
  color: #fff;
}

.step-node--active .step-node__index {
  background: #409eff;
  color: #fff;
  animation: ring-pulse 2s ease-in-out infinite;
}

.step-node--rejected .step-node__index {
  background: #f56c6c;
  color: #fff;
}

@keyframes ring-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(64, 158, 255, 0.4); }
  50% { box-shadow: 0 0 0 6px rgba(64, 158, 255, 0); }
}

.step-icon {
  font-size: 16px;
  font-weight: 700;
}

.step-node__content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  text-align: center;
}

.step-node__name {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  line-height: 1.3;
}

.step-node__dept {
  font-size: 11px;
  color: #909399;
}

.step-connector {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  margin-top: 16px;
}

.step-connector__line {
  height: 2px;
  width: 20px;
  background: #dcdfe6;
}

.step-connector__arrow {
  font-size: 20px;
  color: #dcdfe6;
  margin-left: -4px;
  font-weight: 700;
  line-height: 1;
}

.step-connector--completed .step-connector__line {
  background: #67c23a;
}
.step-connector--completed .step-connector__arrow {
  color: #67c23a;
}

.step-connector--active .step-connector__line {
  background: #409eff;
  background: linear-gradient(90deg, #67c23a 50%, #dcdfe6 50%);
}
.step-connector--active .step-connector__arrow {
  color: #dcdfe6;
}

.process-flow__legend {
  display: flex;
  gap: 16px;
  padding: 0 16px;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: #606266;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.legend-dot--completed { background: #67c23a; }
.legend-dot--active { background: #409eff; }
.legend-dot--pending { background: #dcdfe6; }
.legend-dot--rejected { background: #f56c6c; }
</style>
