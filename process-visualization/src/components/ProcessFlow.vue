<template>
  <div class="process-flow">
    <div class="process-flow__scroll-area">
      <!-- Step nodes row -->
      <div class="process-flow__steps" ref="stepsEl">
        <template v-for="(step, index) in process.steps" :key="step.id">
          <div
            :ref="(el) => setNodeRef(el as HTMLElement | null, index)"
            class="step-node"
            :class="[
              `step-node--${step.status}`,
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
            <!-- current step badge -->
            <div v-if="step.id === process.currentStepId" class="step-node__current-tag">当前</div>
            <!-- rejected + returns-to badge -->
            <div v-if="step.returnToStepId" class="step-node__loop-tag">↩ 打回</div>
            <!-- re-run badge: another step returned to this one -->
            <div v-if="isLoopTarget(step.id)" class="step-node__rerun-tag">↻ 重做</div>
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

      <!-- SVG loop arrows drawn below the node row -->
      <div v-if="loopEdges.length > 0" class="process-flow__loop-area">
        <svg
          class="process-flow__loop-svg"
          :width="svgWidth"
          :height="loopSvgHeight"
          xmlns="http://www.w3.org/2000/svg"
        >
          <defs>
            <marker
              id="loop-arrowhead"
              markerWidth="8"
              markerHeight="6"
              refX="8"
              refY="3"
              orient="auto"
            >
              <polygon points="0 0, 8 3, 0 6" fill="#f56c6c" />
            </marker>
          </defs>
          <g
            v-for="(edge, ei) in loopEdges"
            :key="`loop-${edge.fromIdx}-${edge.toIdx}`"
          >
            <!-- Start dot at the rejected node -->
            <circle :cx="edge.fromX" :cy="8" r="4" fill="#f56c6c" opacity="0.8" />
            <!-- Curved dashed arc from rejected node back to target node -->
            <path
              :d="getLoopPath(edge.fromX, edge.toX, ei)"
              fill="none"
              stroke="#f56c6c"
              stroke-width="2"
              stroke-dasharray="6,3"
              marker-end="url(#loop-arrowhead)"
            />
            <!-- "打回" label centered on the arc -->
            <text
              :x="(edge.fromX + edge.toX) / 2"
              :y="loopLabelY(ei)"
              text-anchor="middle"
              font-size="11"
              fill="#f56c6c"
              font-weight="600"
            >↩ 打回</text>
          </g>
        </svg>
      </div>
    </div>

    <div class="process-flow__legend">
      <span class="legend-item"><span class="legend-dot legend-dot--completed"></span>已完成</span>
      <span class="legend-item"><span class="legend-dot legend-dot--active"></span>进行中</span>
      <span class="legend-item"><span class="legend-dot legend-dot--pending"></span>待处理</span>
      <span class="legend-item"><span class="legend-dot legend-dot--rejected"></span>已驳回</span>
      <span class="legend-item legend-item--loop">
        <span class="legend-loop-line"></span>打回路径
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUpdated, nextTick, watch } from 'vue'
import type { Process, StepStatus } from '../types/process'
import StatusBadge from './StatusBadge.vue'

const props = defineProps<{
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

function isLoopTarget(stepId: string): boolean {
  return props.process.steps.some(s => s.returnToStepId === stepId)
}

// ─── Loop edge rendering ─────────────────────────────────────────────────────

const stepsEl = ref<HTMLElement | null>(null)
const nodeEls: (HTMLElement | null)[] = []

function setNodeRef(el: HTMLElement | null, index: number) {
  nodeEls[index] = el
}

/** Center X of each node relative to stepsEl's left edge */
const nodePositions = ref<number[]>([])
const svgWidth = ref(600)

function measurePositions() {
  if (!stepsEl.value) return
  const containerLeft = stepsEl.value.getBoundingClientRect().left
  nodePositions.value = nodeEls.map(el => {
    if (!el) return 0
    const r = el.getBoundingClientRect()
    return r.left - containerLeft + r.width / 2
  })
  svgWidth.value = stepsEl.value.scrollWidth
}

onMounted(() => nextTick(measurePositions))
onUpdated(() => nextTick(measurePositions))
watch(() => props.process.id, () => nextTick(measurePositions))

interface LoopEdge {
  fromIdx: number
  toIdx: number
  fromX: number
  toX: number
}

const loopEdges = computed<LoopEdge[]>(() => {
  return props.process.steps
    .map((step, i) => {
      if (!step.returnToStepId) return null
      const toIdx = props.process.steps.findIndex(s => s.id === step.returnToStepId)
      if (toIdx < 0 || toIdx >= i) return null // only backward edges
      return {
        fromIdx: i,
        toIdx,
        fromX: nodePositions.value[i] ?? 0,
        toX: nodePositions.value[toIdx] ?? 0,
      }
    })
    .filter((e): e is LoopEdge => e !== null)
})

const LOOP_BASE_DEPTH = 58
const LOOP_DEPTH_STEP = 20

const loopSvgHeight = computed(() =>
  loopEdges.value.length === 0
    ? 0
    : LOOP_BASE_DEPTH + (loopEdges.value.length - 1) * LOOP_DEPTH_STEP + 18
)

function getLoopPath(fromX: number, toX: number, edgeIndex: number): string {
  const startY = 8
  const curveBottom = LOOP_BASE_DEPTH + edgeIndex * LOOP_DEPTH_STEP
  return `M ${fromX} ${startY} C ${fromX} ${curveBottom}, ${toX} ${curveBottom}, ${toX} ${startY}`
}

function loopLabelY(edgeIndex: number): number {
  return LOOP_BASE_DEPTH + edgeIndex * LOOP_DEPTH_STEP - 6
}
</script>

<style scoped>
.process-flow {
  padding: 16px 0;
}

.process-flow__scroll-area {
  overflow-x: auto;
}

.process-flow__steps {
  display: flex;
  align-items: flex-start;
  padding: 8px 16px 8px;
  gap: 0;
  width: max-content;
  min-width: 100%;
}

.process-flow__loop-area {
  padding: 0 0 8px 0;
}

.process-flow__loop-svg {
  display: block;
  overflow: visible;
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

/* top-right corner badges */
.step-node__current-tag,
.step-node__loop-tag,
.step-node__rerun-tag {
  position: absolute;
  top: -10px;
  right: -8px;
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 8px;
  font-weight: 600;
  white-space: nowrap;
}

.step-node__current-tag {
  background: #ff6b35;
  color: #fff;
}

.step-node__loop-tag {
  background: #f56c6c;
  color: #fff;
  right: auto;
  left: -8px;
}

.step-node__rerun-tag {
  background: #e6a23c;
  color: #fff;
  right: auto;
  left: -8px;
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
  background: linear-gradient(90deg, #67c23a 50%, #dcdfe6 50%);
}
.step-connector--active .step-connector__arrow {
  color: #dcdfe6;
}

/* Legend */
.process-flow__legend {
  display: flex;
  gap: 16px;
  padding: 8px 16px 0;
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
.legend-dot--active    { background: #409eff; }
.legend-dot--pending   { background: #dcdfe6; }
.legend-dot--rejected  { background: #f56c6c; }

.legend-loop-line {
  display: inline-block;
  width: 20px;
  height: 2px;
  background: repeating-linear-gradient(
    90deg,
    #f56c6c 0 6px,
    transparent 6px 9px
  );
  border-radius: 1px;
  vertical-align: middle;
}
</style>

