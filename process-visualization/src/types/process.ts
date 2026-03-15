export type StepStatus = 'pending' | 'active' | 'completed' | 'rejected' | 'skipped'

export interface ProcessStep {
  id: string
  name: string
  description: string
  status: StepStatus
  assignee?: string
  department?: string
  startTime?: string
  endTime?: string
  comment?: string
  subSteps?: ProcessSubStep[]
  /** When this step is rejected it loops back to the step with this id */
  returnToStepId?: string
}

export interface ProcessSubStep {
  id: string
  name: string
  status: StepStatus
  assignee?: string
  completedAt?: string
}

export interface Process {
  id: string
  title: string
  type: string
  description: string
  currentStepId: string
  createdAt: string
  updatedAt: string
  priority: 'low' | 'medium' | 'high' | 'urgent'
  steps: ProcessStep[]
}

export interface RectificationRecord {
  id: string
  processId: string
  stepId: string
  action: string
  operator: string
  department: string
  timestamp: string
  remark?: string
  result: 'pass' | 'reject' | 'pending'
}
