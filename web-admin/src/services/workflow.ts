import request from '../utils/request'

export interface WorkflowListParams {
  status?: string
  type?: string
  keyword?: string
  page?: number
  page_size?: number
}

export interface CreateWorkflowRequest {
  name: string
  code: string
  description?: string
  type: string
}

export interface UpdateWorkflowRequest {
  name?: string
  description?: string
  type?: string
  status?: string
  is_default?: boolean
}

export interface CreateStepRequest {
  name: string
  description?: string
  step_type: string
  assignee_type?: string
  assignee_role?: string
  duration?: number
  is_timeout_skip?: boolean
  form_config?: Record<string, any>
  conditions?: Record<string, any>
  actions?: Record<string, any>
}

export interface StartInstanceRequest {
  workflow_id: string
  business_id: string
  business_type: string
  title: string
  form_data?: Record<string, any>
}

export interface ApproveRequest {
  action: 'approve' | 'reject' | 'transfer'
  comment?: string
  form_data?: Record<string, any>
  next_step_id?: string
  transfer_to?: string
}

export const workflowApi = {
  // 获取工作流列表
  getWorkflows: (params?: WorkflowListParams) => 
    request.get('/workflows', { params }),

  // 获取工作流详情
  getWorkflow: (id: string) => 
    request.get(`/workflows/${id}`),

  // 创建工作流
  createWorkflow: (data: CreateWorkflowRequest) => 
    request.post('/workflows', data),

  // 更新工作流
  updateWorkflow: (id: string, data: UpdateWorkflowRequest) => 
    request.put(`/workflows/${id}`, data),

  // 删除工作流
  deleteWorkflow: (id: string) => 
    request.delete(`/workflows/${id}`),

  // 创建工作流步骤
  createStep: (workflowId: string, data: CreateStepRequest) => 
    request.post(`/workflows/${workflowId}/steps`, data),

  // 更新步骤
  updateStep: (workflowId: string, stepId: string, data: CreateStepRequest) => 
    request.put(`/workflows/${workflowId}/steps/${stepId}`, data),

  // 删除步骤
  deleteStep: (workflowId: string, stepId: string) => 
    request.delete(`/workflows/${workflowId}/steps/${stepId}`),

  // 重新排序步骤
  reorderSteps: (workflowId: string, stepIds: string[]) => 
    request.post(`/workflows/${workflowId}/steps/reorder`, { step_ids: stepIds }),

  // 启动工作流实例
  startInstance: (data: StartInstanceRequest) => 
    request.post('/workflow-instances', data),

  // 获取实例详情
  getInstance: (id: string) => 
    request.get(`/workflow-instances/${id}`),

  // 获取实例列表
  getInstances: (params?: { status?: string; workflow_id?: string; starter_id?: string; business_type?: string; page?: number; page_size?: number }) => 
    request.get('/workflow-instances', { params }),

  // 审批
  approve: (id: string, data: ApproveRequest) => 
    request.post(`/workflow-instances/${id}/approve`, data),

  // 取消实例
  cancelInstance: (id: string, reason?: string) => 
    request.post(`/workflow-instances/${id}/cancel`, { reason }),

  // 获取实例历史
  getInstanceHistory: (id: string) => 
    request.get(`/workflow-instances/${id}/history`),

  // 获取我的待办
  getMyInstances: () => 
    request.get('/workflow-instances/my'),
}
