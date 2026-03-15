<template>
  <div class="process-timeline">
    <div
      v-for="(step, index) in process.steps"
      :key="step.id"
      class="timeline-item"
      :class="[
        `timeline-item--${step.status}`,
        { 'timeline-item--active': step.id === process.currentStepId },
        { 'timeline-item--selected': selectedStepId === step.id },
      ]"
      @click="$emit('select-step', step.id)"
    >
      <div class="timeline-item__line-top" v-if="index > 0"></div>
      <div class="timeline-item__dot">
        <span v-if="step.status === 'completed'" class="dot-icon">✓</span>
        <span v-else-if="step.status === 'rejected'" class="dot-icon">✕</span>
        <span v-else-if="step.status === 'active'" class="dot-icon dot-icon--spin">⟳</span>
        <span v-else class="dot-icon dot-icon--num">{{ index + 1 }}</span>
      </div>
      <div class="timeline-item__line-bottom" v-if="index < process.steps.length - 1"></div>
      <div class="timeline-item__body">
        <div class="timeline-item__header">
          <span class="timeline-item__name">{{ step.name }}</span>
          <StatusBadge :status="step.status" />
          <span v-if="step.id === process.currentStepId" class="current-tag">当前节点</span>
          <!-- Re-run badge: some other step's returnToStepId points here -->
          <span v-if="getReturnedFromStep(step.id)" class="rerun-tag">↻ 重做中</span>
        </div>
        <div class="timeline-item__meta">
          <span v-if="step.assignee"><b>负责人：</b>{{ step.assignee }}</span>
          <span v-if="step.department"><b>部门：</b>{{ step.department }}</span>
          <span v-if="step.startTime"><b>开始：</b>{{ step.startTime }}</span>
          <span v-if="step.endTime"><b>完成：</b>{{ step.endTime }}</span>
        </div>
        <p v-if="step.comment" class="timeline-item__comment">{{ step.comment }}</p>
        <div v-if="step.subSteps && step.subSteps.length" class="timeline-item__sub">
          <div
            v-for="sub in step.subSteps"
            :key="sub.id"
            class="sub-badge"
            :class="`sub-badge--${sub.status}`"
          >
            <span class="sub-badge__dot"></span>
            {{ sub.name }}
          </div>
        </div>
        <!-- Loop return indicator for rejected steps -->
        <div v-if="step.returnToStepId" class="loop-return-indicator">
          <span class="loop-return-icon">↩</span>
          <span class="loop-return-text">
            已打回至「{{ getStepName(step.returnToStepId) }}」重新整改
          </span>
        </div>
        <!-- Re-run indicator: show who returned to this step -->
        <div v-if="getReturnedFromStep(step.id)" class="loop-rerun-indicator">
          <span class="loop-rerun-icon">↻</span>
          <span class="loop-rerun-text">
            由「{{ getReturnedFromStep(step.id) }}」打回重做
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Process } from '../types/process'
import StatusBadge from './StatusBadge.vue'

const props = defineProps<{
  process: Process
  selectedStepId?: string
}>()

defineEmits<{
  'select-step': [stepId: string]
}>()

function getStepName(stepId: string): string {
  return props.process.steps.find(s => s.id === stepId)?.name ?? stepId
}

/** Returns the name of the step that returned to `stepId`, or null if none. */
function getReturnedFromStep(stepId: string): string | null {
  const source = props.process.steps.find(s => s.returnToStepId === stepId)
  return source ? source.name : null
}
</script>

<style scoped>
.process-timeline {
  padding: 8px 0;
}

.timeline-item {
  display: flex;
  align-items: flex-start;
  gap: 0;
  cursor: pointer;
  position: relative;
}

.timeline-item__line-top,
.timeline-item__line-bottom {
  position: absolute;
  left: 15px;
  width: 2px;
  background: #dcdfe6;
  z-index: 0;
}

.timeline-item__line-top {
  top: 0;
  height: 16px;
}

.timeline-item__line-bottom {
  top: 32px;
  bottom: 0;
  min-height: 16px;
}

.timeline-item--completed .timeline-item__line-top,
.timeline-item--completed .timeline-item__line-bottom {
  background: #67c23a;
}
.timeline-item--active .timeline-item__line-top {
  background: #67c23a;
}

.timeline-item__dot {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  z-index: 1;
  margin-right: 14px;
  margin-top: 8px;
  transition: all 0.2s;
}

.timeline-item--completed .timeline-item__dot {
  background: #67c23a;
  color: #fff;
}

.timeline-item--active .timeline-item__dot {
  background: #409eff;
  color: #fff;
  box-shadow: 0 0 0 4px rgba(64, 158, 255, 0.2);
}

.timeline-item--rejected .timeline-item__dot {
  background: #f56c6c;
  color: #fff;
}

.dot-icon {
  font-size: 14px;
  font-weight: 700;
  color: #fff;
}

.dot-icon--num {
  font-size: 12px;
  color: #909399;
}

.dot-icon--spin {
  display: inline-block;
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.timeline-item__body {
  flex: 1;
  padding: 8px 12px 16px;
  margin-bottom: 4px;
  border-radius: 8px;
  transition: all 0.2s;
}

.timeline-item:hover .timeline-item__body,
.timeline-item--selected .timeline-item__body {
  background: #f5f7fa;
}

.timeline-item--active .timeline-item__body {
  background: #ecf5ff;
  border: 1px solid #d9ecff;
}

.timeline-item__header {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 6px;
}

.timeline-item__name {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.current-tag {
  background: #ff6b35;
  color: #fff;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 8px;
  font-weight: 600;
}

.rerun-tag {
  background: #e6a23c;
  color: #fff;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 8px;
  font-weight: 600;
}

.timeline-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 12px;
  color: #606266;
  margin-bottom: 4px;
}

.timeline-item__meta b {
  font-weight: 500;
  color: #909399;
}

.timeline-item__comment {
  font-size: 12px;
  color: #606266;
  margin: 6px 0 0;
  padding: 8px;
  background: rgba(0,0,0,0.03);
  border-radius: 4px;
  line-height: 1.5;
  border-left: 3px solid #dcdfe6;
}

.timeline-item--completed .timeline-item__comment {
  border-left-color: #67c23a;
}

.timeline-item--active .timeline-item__comment {
  border-left-color: #409eff;
}

.timeline-item--rejected .timeline-item__comment {
  border-left-color: #f56c6c;
}

/* Loop return indicator (on rejected step) */
.loop-return-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 6px 10px;
  background: #fff1f0;
  border: 1px dashed #ffa39e;
  border-radius: 6px;
  font-size: 12px;
}

.loop-return-icon {
  font-size: 15px;
  color: #f56c6c;
  flex-shrink: 0;
}

.loop-return-text {
  color: #f56c6c;
  font-weight: 500;
}

/* Re-run indicator (on step being re-done) */
.loop-rerun-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 6px 10px;
  background: #fffbe6;
  border: 1px dashed #ffe58f;
  border-radius: 6px;
  font-size: 12px;
}

.loop-rerun-icon {
  font-size: 15px;
  color: #e6a23c;
  flex-shrink: 0;
}

.loop-rerun-text {
  color: #e6a23c;
  font-weight: 500;
}

.timeline-item__sub {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.sub-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
}

.sub-badge__dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
}

.sub-badge--completed {
  background: #f0f9eb;
  color: #67c23a;
}
.sub-badge--completed .sub-badge__dot { background: #67c23a; }

.sub-badge--active {
  background: #ecf5ff;
  color: #409eff;
}
.sub-badge--active .sub-badge__dot {
  background: #409eff;
  animation: pulse 1.5s infinite;
}

.sub-badge--pending {
  background: #f5f5f5;
  color: #909399;
}
.sub-badge--pending .sub-badge__dot { background: #c0c4cc; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}
</style>


<style scoped>
.process-timeline {
  padding: 8px 0;
}

.timeline-item {
  display: flex;
  align-items: flex-start;
  gap: 0;
  cursor: pointer;
  position: relative;
}

.timeline-item__line-top,
.timeline-item__line-bottom {
  position: absolute;
  left: 15px;
  width: 2px;
  background: #dcdfe6;
  z-index: 0;
}

.timeline-item__line-top {
  top: 0;
  height: 16px;
}

.timeline-item__line-bottom {
  top: 32px;
  bottom: 0;
  min-height: 16px;
}

.timeline-item--completed .timeline-item__line-top,
.timeline-item--completed .timeline-item__line-bottom {
  background: #67c23a;
}
.timeline-item--active .timeline-item__line-top {
  background: #67c23a;
}

.timeline-item__dot {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  z-index: 1;
  margin-right: 14px;
  margin-top: 8px;
  transition: all 0.2s;
}

.timeline-item--completed .timeline-item__dot {
  background: #67c23a;
  color: #fff;
}

.timeline-item--active .timeline-item__dot {
  background: #409eff;
  color: #fff;
  box-shadow: 0 0 0 4px rgba(64, 158, 255, 0.2);
}

.timeline-item--rejected .timeline-item__dot {
  background: #f56c6c;
  color: #fff;
}

.dot-icon {
  font-size: 14px;
  font-weight: 700;
  color: #fff;
}

.dot-icon--num {
  font-size: 12px;
  color: #909399;
}

.dot-icon--spin {
  display: inline-block;
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.timeline-item__body {
  flex: 1;
  padding: 8px 12px 16px;
  margin-bottom: 4px;
  border-radius: 8px;
  transition: all 0.2s;
}

.timeline-item:hover .timeline-item__body,
.timeline-item--selected .timeline-item__body {
  background: #f5f7fa;
}

.timeline-item--active .timeline-item__body {
  background: #ecf5ff;
  border: 1px solid #d9ecff;
}

.timeline-item__header {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 6px;
}

.timeline-item__name {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.current-tag {
  background: #ff6b35;
  color: #fff;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 8px;
  font-weight: 600;
}

.timeline-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 12px;
  color: #606266;
  margin-bottom: 4px;
}

.timeline-item__meta b {
  font-weight: 500;
  color: #909399;
}

.timeline-item__comment {
  font-size: 12px;
  color: #606266;
  margin: 6px 0 0;
  padding: 8px;
  background: rgba(0,0,0,0.03);
  border-radius: 4px;
  line-height: 1.5;
  border-left: 3px solid #dcdfe6;
}

.timeline-item--completed .timeline-item__comment {
  border-left-color: #67c23a;
}

.timeline-item--active .timeline-item__comment {
  border-left-color: #409eff;
}

.timeline-item__sub {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.sub-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 11px;
  font-weight: 500;
}

.sub-badge__dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
}

.sub-badge--completed {
  background: #f0f9eb;
  color: #67c23a;
}
.sub-badge--completed .sub-badge__dot { background: #67c23a; }

.sub-badge--active {
  background: #ecf5ff;
  color: #409eff;
}
.sub-badge--active .sub-badge__dot {
  background: #409eff;
  animation: pulse 1.5s infinite;
}

.sub-badge--pending {
  background: #f5f5f5;
  color: #909399;
}
.sub-badge--pending .sub-badge__dot { background: #c0c4cc; }

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}
</style>
