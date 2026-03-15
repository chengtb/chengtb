<template>
  <div class="step-detail" v-if="step">
    <div class="step-detail__header">
      <div class="step-detail__title">
        <h3>{{ step.name }}</h3>
        <StatusBadge :status="step.status" />
      </div>
      <p class="step-detail__desc">{{ step.description }}</p>
    </div>

    <div class="step-detail__info">
      <div class="info-row" v-if="step.assignee">
        <span class="info-label">负责人</span>
        <span class="info-value">
          <span class="avatar">{{ step.assignee[0] }}</span>
          {{ step.assignee }}
        </span>
      </div>
      <div class="info-row" v-if="step.department">
        <span class="info-label">负责部门</span>
        <span class="info-value">{{ step.department }}</span>
      </div>
      <div class="info-row" v-if="step.startTime">
        <span class="info-label">开始时间</span>
        <span class="info-value">{{ step.startTime }}</span>
      </div>
      <div class="info-row" v-if="step.endTime">
        <span class="info-label">完成时间</span>
        <span class="info-value">{{ step.endTime }}</span>
      </div>
      <div class="info-row" v-if="step.comment">
        <span class="info-label">备注</span>
        <span class="info-value info-value--remark">{{ step.comment }}</span>
      </div>
    </div>

    <!-- Loop: this step was rejected and returned to another step -->
    <div v-if="step.returnToStepId && returnToStepName" class="loop-card loop-card--reject">
      <div class="loop-card__icon">↩</div>
      <div class="loop-card__content">
        <div class="loop-card__title">节点已打回</div>
        <div class="loop-card__desc">
          审核未通过，已退回至「<strong>{{ returnToStepName }}</strong>」节点重新整改。
        </div>
      </div>
    </div>

    <!-- Loop: another step was rejected and returned to this step (re-run) -->
    <div v-if="returnedFromStepName" class="loop-card loop-card--rerun">
      <div class="loop-card__icon">↻</div>
      <div class="loop-card__content">
        <div class="loop-card__title">重做节点</div>
        <div class="loop-card__desc">
          由「<strong>{{ returnedFromStepName }}</strong>」打回，当前节点正在重新执行。
        </div>
      </div>
    </div>

    <div v-if="step.subSteps && step.subSteps.length" class="step-detail__substeps">
      <div class="substeps-title">子步骤</div>
      <div class="substep-list">
        <div
          v-for="sub in step.subSteps"
          :key="sub.id"
          class="substep-item"
          :class="`substep-item--${sub.status}`"
        >
          <div class="substep-item__icon">
            <span v-if="sub.status === 'completed'">✓</span>
            <span v-else-if="sub.status === 'active'" class="spinning">◎</span>
            <span v-else>○</span>
          </div>
          <div class="substep-item__content">
            <span class="substep-item__name">{{ sub.name }}</span>
            <span v-if="sub.assignee" class="substep-item__assignee">{{ sub.assignee }}</span>
            <span v-if="sub.completedAt" class="substep-item__time">{{ sub.completedAt }}</span>
          </div>
          <StatusBadge :status="sub.status" />
        </div>
      </div>
    </div>
  </div>

  <div v-else class="step-detail--empty">
    <div class="empty-icon">🔍</div>
    <p>点击流程节点查看详情</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProcessStep, Process } from '../types/process'
import StatusBadge from './StatusBadge.vue'

const props = defineProps<{
  step?: ProcessStep
  process?: Process
}>()

const returnToStepName = computed(() => {
  if (!props.step?.returnToStepId || !props.process) return null
  return props.process.steps.find(s => s.id === props.step!.returnToStepId)?.name ?? null
})

const returnedFromStepName = computed(() => {
  if (!props.step || !props.process) return null
  const source = props.process.steps.find(s => s.returnToStepId === props.step!.id)
  return source?.name ?? null
})
</script>

<style scoped>
.step-detail {
  padding: 0;
}

.step-detail__header {
  margin-bottom: 16px;
}

.step-detail__title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.step-detail__title h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #303133;
}

.step-detail__desc {
  font-size: 13px;
  color: #606266;
  margin: 0;
  line-height: 1.6;
}

.step-detail__info {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 16px;
}

.info-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.info-label {
  font-size: 12px;
  color: #909399;
  min-width: 60px;
  padding-top: 2px;
  flex-shrink: 0;
}

.info-value {
  font-size: 13px;
  color: #303133;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-value--remark {
  color: #606266;
  font-weight: 400;
  line-height: 1.5;
}

.avatar {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #409eff;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

/* Loop info cards */
.loop-card {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 8px;
  margin-bottom: 14px;
}

.loop-card--reject {
  background: #fff1f0;
  border: 1px dashed #ffa39e;
}

.loop-card--rerun {
  background: #fffbe6;
  border: 1px dashed #ffe58f;
}

.loop-card__icon {
  font-size: 20px;
  flex-shrink: 0;
  line-height: 1.2;
}

.loop-card--reject .loop-card__icon { color: #f56c6c; }
.loop-card--rerun .loop-card__icon  { color: #e6a23c; }

.loop-card__title {
  font-size: 13px;
  font-weight: 700;
  margin-bottom: 3px;
}

.loop-card--reject .loop-card__title { color: #f56c6c; }
.loop-card--rerun .loop-card__title  { color: #e6a23c; }

.loop-card__desc {
  font-size: 12px;
  color: #606266;
  line-height: 1.5;
}

.loop-card__desc strong {
  color: #303133;
}

/* Sub-steps */
.step-detail__substeps {
  margin-top: 16px;
}

.substeps-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
}

.substep-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.substep-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-left: 3px solid #dcdfe6;
  margin-left: 10px;
  position: relative;
  transition: background 0.2s;
}

.substep-item:not(:last-child) {
  padding-bottom: 10px;
}

.substep-item--completed {
  border-left-color: #67c23a;
}

.substep-item--active {
  border-left-color: #409eff;
  background: #f0f7ff;
  border-radius: 0 6px 6px 0;
}

.substep-item__icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  flex-shrink: 0;
}

.substep-item--completed .substep-item__icon { color: #67c23a; }
.substep-item--active .substep-item__icon    { color: #409eff; }
.substep-item--pending .substep-item__icon   { color: #c0c4cc; }

.spinning {
  display: inline-block;
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.substep-item__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.substep-item__name {
  font-size: 13px;
  color: #303133;
  font-weight: 500;
}

.substep-item__assignee,
.substep-item__time {
  font-size: 11px;
  color: #909399;
}

.step-detail--empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #c0c4cc;
  gap: 10px;
}

.empty-icon {
  font-size: 36px;
}

.step-detail--empty p {
  font-size: 13px;
  margin: 0;
}
</style>


<style scoped>
.step-detail {
  padding: 0;
}

.step-detail__header {
  margin-bottom: 16px;
}

.step-detail__title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.step-detail__title h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #303133;
}

.step-detail__desc {
  font-size: 13px;
  color: #606266;
  margin: 0;
  line-height: 1.6;
}

.step-detail__info {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 16px;
}

.info-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.info-label {
  font-size: 12px;
  color: #909399;
  min-width: 60px;
  padding-top: 2px;
  flex-shrink: 0;
}

.info-value {
  font-size: 13px;
  color: #303133;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}

.info-value--remark {
  color: #606266;
  font-weight: 400;
  line-height: 1.5;
}

.avatar {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #409eff;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.step-detail__substeps {
  margin-top: 16px;
}

.substeps-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 10px;
}

.substep-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.substep-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-left: 3px solid #dcdfe6;
  margin-left: 10px;
  position: relative;
  transition: background 0.2s;
}

.substep-item:not(:last-child) {
  padding-bottom: 10px;
}

.substep-item--completed {
  border-left-color: #67c23a;
}

.substep-item--active {
  border-left-color: #409eff;
  background: #f0f7ff;
  border-radius: 0 6px 6px 0;
}

.substep-item__icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  flex-shrink: 0;
}

.substep-item--completed .substep-item__icon {
  color: #67c23a;
}

.substep-item--active .substep-item__icon {
  color: #409eff;
}

.substep-item--pending .substep-item__icon {
  color: #c0c4cc;
}

.spinning {
  display: inline-block;
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.substep-item__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.substep-item__name {
  font-size: 13px;
  color: #303133;
  font-weight: 500;
}

.substep-item__assignee,
.substep-item__time {
  font-size: 11px;
  color: #909399;
}

.step-detail--empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #c0c4cc;
  gap: 10px;
}

.empty-icon {
  font-size: 36px;
}

.step-detail--empty p {
  font-size: 13px;
  margin: 0;
}
</style>
