<template>
  <div class="app">
    <!-- Header -->
    <header class="app-header">
      <div class="app-header__inner">
        <div class="app-header__brand">
          <span class="brand-icon">⚙️</span>
          <span class="brand-name">企业整改流程可视化平台</span>
        </div>
        <div class="app-header__actions">
          <span class="header-time">{{ currentTime }}</span>
        </div>
      </div>
    </header>

    <main class="app-main">
      <!-- Process Selector -->
      <div class="process-selector-bar">
        <div class="selector-label">选择流程：</div>
        <div class="selector-tabs">
          <button
            v-for="proc in processes"
            :key="proc.id"
            class="selector-tab"
            :class="{ 'selector-tab--active': selectedProcessId === proc.id }"
            @click="selectProcess(proc.id)"
          >
            <span class="tab-priority" :class="`priority--${proc.priority}`">●</span>
            <span class="tab-title">{{ proc.title }}</span>
            <span class="tab-type">{{ proc.type }}</span>
          </button>
        </div>
      </div>

      <div v-if="currentProcess" class="content-area">
        <!-- Process Header Card -->
        <div class="process-header-card">
          <div class="process-header-card__left">
            <div class="process-title-row">
              <h2 class="process-title">{{ currentProcess.title }}</h2>
              <span class="process-type-tag">{{ currentProcess.type }}</span>
              <span class="priority-badge" :class="`priority-badge--${currentProcess.priority}`">
                {{ priorityText(currentProcess.priority) }}
              </span>
            </div>
            <p class="process-desc">{{ currentProcess.description }}</p>
            <div class="process-meta">
              <span>创建时间：{{ currentProcess.createdAt }}</span>
              <span>最后更新：{{ currentProcess.updatedAt }}</span>
              <span>流程编号：{{ currentProcess.id }}</span>
            </div>
          </div>
          <div class="process-header-card__right">
            <div class="progress-ring-wrapper">
              <svg class="progress-ring" width="80" height="80" viewBox="0 0 80 80">
                <circle cx="40" cy="40" r="32" fill="none" stroke="#e4e7ed" stroke-width="6" />
                <circle
                  cx="40" cy="40" r="32"
                  fill="none"
                  stroke="#409eff"
                  stroke-width="6"
                  stroke-linecap="round"
                  stroke-dasharray="201"
                  :stroke-dashoffset="201 - (201 * completionRate)"
                  transform="rotate(-90 40 40)"
                  style="transition: stroke-dashoffset 0.6s ease"
                />
                <text x="40" y="45" text-anchor="middle" font-size="16" font-weight="700" fill="#303133">
                  {{ Math.round(completionRate * 100) }}%
                </text>
              </svg>
              <div class="progress-label">完成进度</div>
            </div>
          </div>
        </div>

        <!-- Tab Navigation -->
        <div class="view-tabs">
          <button
            class="view-tab"
            :class="{ 'view-tab--active': activeView === 'flow' }"
            @click="activeView = 'flow'"
          >
            🔄 流程图视图
          </button>
          <button
            class="view-tab"
            :class="{ 'view-tab--active': activeView === 'timeline' }"
            @click="activeView = 'timeline'"
          >
            📋 时间线视图
          </button>
        </div>

        <div class="two-column">
          <!-- Left: Flow or Timeline -->
          <div class="column-main">
            <div class="panel">
              <div class="panel__header">
                <span class="panel__title">
                  {{ activeView === 'flow' ? '整改流程图' : '流程时间线' }}
                </span>
                <span class="panel__subtitle">
                  共 {{ currentProcess.steps.length }} 个节点，当前处于「{{ currentStepName }}」阶段
                </span>
              </div>
              <div class="panel__body">
                <ProcessFlow
                  v-if="activeView === 'flow'"
                  :process="currentProcess"
                  :selected-step-id="selectedStepId"
                  @select-step="selectStep"
                />
                <ProcessTimeline
                  v-else
                  :process="currentProcess"
                  :selected-step-id="selectedStepId"
                  @select-step="selectStep"
                />
              </div>
            </div>
          </div>

          <!-- Right: Step Detail -->
          <div class="column-aside">
            <div class="panel">
              <div class="panel__header">
                <span class="panel__title">节点详情</span>
              </div>
              <div class="panel__body">
                <StepDetail :step="selectedStep" />
              </div>
            </div>

            <!-- Stats Card -->
            <div class="stats-card">
              <div class="stats-card__title">流程统计</div>
              <div class="stats-grid">
                <div class="stat-item stat-item--completed">
                  <div class="stat-num">{{ stepCounts.completed }}</div>
                  <div class="stat-label">已完成</div>
                </div>
                <div class="stat-item stat-item--active">
                  <div class="stat-num">{{ stepCounts.active }}</div>
                  <div class="stat-label">进行中</div>
                </div>
                <div class="stat-item stat-item--pending">
                  <div class="stat-num">{{ stepCounts.pending }}</div>
                  <div class="stat-label">待处理</div>
                </div>
                <div class="stat-item stat-item--rejected">
                  <div class="stat-num">{{ stepCounts.rejected }}</div>
                  <div class="stat-label">已驳回</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { Process } from './types/process'
import { sampleProcesses } from './data/sampleData'
import ProcessFlow from './components/ProcessFlow.vue'
import ProcessTimeline from './components/ProcessTimeline.vue'
import StepDetail from './components/StepDetail.vue'

const processes = ref<Process[]>(sampleProcesses)
const selectedProcessId = ref(sampleProcesses[0].id)
const selectedStepId = ref<string | undefined>(sampleProcesses[0].currentStepId)
const activeView = ref<'flow' | 'timeline'>('flow')
const currentTime = ref('')

const currentProcess = computed(() =>
  processes.value.find(p => p.id === selectedProcessId.value)
)

const currentStepName = computed(() => {
  const step = currentProcess.value?.steps.find(
    s => s.id === currentProcess.value?.currentStepId
  )
  return step?.name ?? '—'
})

const selectedStep = computed(() =>
  currentProcess.value?.steps.find(s => s.id === selectedStepId.value)
)

const completionRate = computed(() => {
  const steps = currentProcess.value?.steps ?? []
  const done = steps.filter(s => s.status === 'completed').length
  return steps.length ? done / steps.length : 0
})

const stepCounts = computed(() => {
  const steps = currentProcess.value?.steps ?? []
  return {
    completed: steps.filter(s => s.status === 'completed').length,
    active: steps.filter(s => s.status === 'active').length,
    pending: steps.filter(s => s.status === 'pending').length,
    rejected: steps.filter(s => s.status === 'rejected').length,
  }
})

function selectProcess(id: string) {
  selectedProcessId.value = id
  const proc = processes.value.find(p => p.id === id)
  selectedStepId.value = proc?.currentStepId
}

function selectStep(stepId: string) {
  selectedStepId.value = selectedStepId.value === stepId ? undefined : stepId
}

function priorityText(priority: string) {
  const map: Record<string, string> = {
    low: '低优先级', medium: '中优先级', high: '高优先级', urgent: '紧急',
  }
  return map[priority] ?? priority
}

function updateTime() {
  currentTime.value = new Date().toLocaleString('zh-CN', { hour12: false })
}

let timer: ReturnType<typeof setInterval>
onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})
onUnmounted(() => clearInterval(timer))
</script>

<style>
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
  background: #f0f2f5;
  color: #303133;
  min-height: 100vh;
}

.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Header */
.app-header {
  background: linear-gradient(135deg, #1a237e 0%, #283593 60%, #3949ab 100%);
  color: #fff;
  padding: 0 24px;
  height: 56px;
  display: flex;
  align-items: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
  position: sticky;
  top: 0;
  z-index: 100;
}

.app-header__inner {
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.app-header__brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-icon { font-size: 22px; }

.brand-name {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.header-time {
  font-size: 13px;
  opacity: 0.8;
  font-variant-numeric: tabular-nums;
}

/* Main */
.app-main {
  flex: 1;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  padding: 20px 20px 40px;
}

/* Process Selector */
.process-selector-bar {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.selector-label {
  font-size: 14px;
  color: #606266;
  padding-top: 10px;
  white-space: nowrap;
}

.selector-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.selector-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1.5px solid #dcdfe6;
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  color: #606266;
  transition: all 0.2s;
}

.selector-tab:hover {
  border-color: #409eff;
  color: #409eff;
}

.selector-tab--active {
  border-color: #409eff;
  background: #ecf5ff;
  color: #409eff;
  font-weight: 600;
}

.tab-title {
  font-weight: 600;
  color: inherit;
}

.tab-type {
  font-size: 11px;
  opacity: 0.7;
}

.tab-priority {
  font-size: 10px;
}

.priority--low { color: #909399; }
.priority--medium { color: #e6a23c; }
.priority--high { color: #f56c6c; }
.priority--urgent { color: #f56c6c; animation: blink 1s step-end infinite; }

@keyframes blink {
  50% { opacity: 0; }
}

/* Process Header Card */
.process-header-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px 24px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
}

.process-header-card__left {
  flex: 1;
}

.process-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.process-title {
  font-size: 20px;
  font-weight: 700;
  color: #1a237e;
}

.process-type-tag {
  background: #ecf5ff;
  color: #409eff;
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 12px;
  font-weight: 600;
  border: 1px solid #d9ecff;
}

.priority-badge {
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 12px;
  font-weight: 600;
}

.priority-badge--low { background: #f5f5f5; color: #909399; }
.priority-badge--medium { background: #fdf6ec; color: #e6a23c; border: 1px solid #faecd8; }
.priority-badge--high { background: #fef0f0; color: #f56c6c; border: 1px solid #fde2e2; }
.priority-badge--urgent { background: #f56c6c; color: #fff; animation: none; }

.process-desc {
  font-size: 13px;
  color: #606266;
  line-height: 1.6;
  margin-bottom: 10px;
}

.process-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #909399;
  flex-wrap: wrap;
}

.progress-ring-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.progress-ring {
  filter: drop-shadow(0 2px 4px rgba(64,158,255,0.2));
}

.progress-label {
  font-size: 12px;
  color: #909399;
}

/* View Tabs */
.view-tabs {
  display: flex;
  gap: 2px;
  margin-bottom: 12px;
  background: #fff;
  border-radius: 8px;
  padding: 4px;
  width: fit-content;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

.view-tab {
  padding: 7px 16px;
  border: none;
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  font-size: 13px;
  color: #606266;
  transition: all 0.2s;
  font-weight: 500;
}

.view-tab:hover {
  background: #f5f7fa;
}

.view-tab--active {
  background: #409eff;
  color: #fff;
  font-weight: 600;
}

/* Two Column Layout */
.two-column {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.column-main {
  flex: 1;
  min-width: 0;
}

.column-aside {
  width: 320px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Panel */
.panel {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  overflow: hidden;
}

.panel__header {
  padding: 14px 20px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
}

.panel__title {
  font-size: 15px;
  font-weight: 700;
  color: #303133;
}

.panel__subtitle {
  font-size: 12px;
  color: #909399;
}

.panel__body {
  padding: 16px;
}

/* Stats Card */
.stats-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  padding: 16px;
}

.stats-card__title {
  font-size: 14px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 12px;
}

.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.stat-item {
  padding: 12px;
  border-radius: 8px;
  text-align: center;
}

.stat-item--completed { background: #f0f9eb; }
.stat-item--active { background: #ecf5ff; }
.stat-item--pending { background: #f5f5f5; }
.stat-item--rejected { background: #fef0f0; }

.stat-num {
  font-size: 24px;
  font-weight: 800;
  line-height: 1.2;
}

.stat-item--completed .stat-num { color: #67c23a; }
.stat-item--active .stat-num { color: #409eff; }
.stat-item--pending .stat-num { color: #909399; }
.stat-item--rejected .stat-num { color: #f56c6c; }

.stat-label {
  font-size: 11px;
  color: #909399;
  margin-top: 2px;
}

/* Responsive */
@media (max-width: 900px) {
  .two-column {
    flex-direction: column;
  }
  .column-aside {
    width: 100%;
  }
}
</style>

