import { useEffect, useState } from 'react'
import { 
  Table, Button, Space, Tag, Modal, Form, Input, Select, Card, Row, Col, 
  message, Popconfirm, Steps, Divider, Timeline, Badge, Tooltip, Empty
} from 'antd'
import { 
  PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, PlayCircleOutlined,
  ArrowUpOutlined, ArrowDownOutlined, CheckCircleOutlined, CloseCircleOutlined,
  ClockCircleOutlined, ProfileOutlined, NodeExpandOutlined
} from '@ant-design/icons'
import { workflowApi } from '../services/workflow'
import dayjs from 'dayjs'

const { Option } = Select
const { Step } = Steps
const { TextArea } = Input

interface Workflow {
  id: string
  name: string
  code: string
  description: string
  type: string
  status: string
  version: number
  is_default: boolean
  created_at: string
  steps?: WorkflowStep[]
}

interface WorkflowStep {
  id: string
  name: string
  description: string
  step_order: number
  step_type: string
  assignee_type: string
  assignee_role: string
}

const Workflows = () => {
  const [workflows, setWorkflows] = useState<Workflow[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [stepModalVisible, setStepModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(null)
  const [selectedStep, setSelectedStep] = useState<WorkflowStep | null>(null)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [form] = Form.useForm()
  const [stepForm] = Form.useForm()
  const [filters, setFilters] = useState({ status: '', type: '' })

  const statusLabels: Record<string, { color: string; text: string }> = {
    draft: { color: 'default', text: '草稿' },
    active: { color: 'success', text: '激活' },
    inactive: { color: 'default', text: '停用' },
  }

  const typeLabels: Record<string, string> = {
    approval: '审批流程',
    task: '任务流程',
    issue: '问题处理',
    other: '其他',
  }

  const stepTypeLabels: Record<string, string> = {
    start: '开始',
    approval: '审批',
    task: '任务',
    notify: '通知',
    end: '结束',
  }

  useEffect(() => {
    fetchData()
  }, [pagination.current, pagination.pageSize, filters])

  const fetchData = async () => {
    setLoading(true)
    try {
      const res = await workflowApi.getWorkflows({
        ...filters,
        page: pagination.current,
        page_size: pagination.pageSize,
      })
      setWorkflows(res.list || [])
      setPagination({ ...pagination, total: res.total })
    } catch (error) {
      message.error('获取工作流列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCreate = () => {
    setSelectedWorkflow(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: Workflow) => {
    setSelectedWorkflow(record)
    form.setFieldsValue(record)
    setModalVisible(true)
  }

  const handleSubmit = async (values: any) => {
    try {
      if (selectedWorkflow) {
        await workflowApi.updateWorkflow(selectedWorkflow.id, values)
        message.success('更新成功')
      } else {
        await workflowApi.createWorkflow(values)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchData()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await workflowApi.deleteWorkflow(id)
      message.success('删除成功')
      fetchData()
    } catch (error: any) {
      message.error(error.message || '删除失败')
    }
  }

  const handleViewDetail = async (record: Workflow) => {
    try {
      const res = await workflowApi.getWorkflow(record.id)
      setSelectedWorkflow(res)
      setDetailModalVisible(true)
    } catch (error) {
      message.error('获取详情失败')
    }
  }

  const handleAddStep = (workflow: Workflow) => {
    setSelectedWorkflow(workflow)
    setSelectedStep(null)
    stepForm.resetFields()
    setStepModalVisible(true)
  }

  const handleEditStep = (workflow: Workflow, step: WorkflowStep) => {
    setSelectedWorkflow(workflow)
    setSelectedStep(step)
    stepForm.setFieldsValue(step)
    setStepModalVisible(true)
  }

  const handleStepSubmit = async (values: any) => {
    if (!selectedWorkflow) return
    try {
      if (selectedStep) {
        await workflowApi.updateStep(selectedWorkflow.id, selectedStep.id, values)
        message.success('更新步骤成功')
      } else {
        await workflowApi.createStep(selectedWorkflow.id, values)
        message.success('创建步骤成功')
      }
      setStepModalVisible(false)
      fetchData()
      if (detailModalVisible) {
        const res = await workflowApi.getWorkflow(selectedWorkflow.id)
        setSelectedWorkflow(res)
      }
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  const handleDeleteStep = async (workflowId: string, stepId: string) => {
    try {
      await workflowApi.deleteStep(workflowId, stepId)
      message.success('删除步骤成功')
      fetchData()
      if (detailModalVisible) {
        const res = await workflowApi.getWorkflow(workflowId)
        setSelectedWorkflow(res)
      }
    } catch (error: any) {
      message.error(error.message || '删除失败')
    }
  }

  const handleActivate = async (record: Workflow) => {
    try {
      await workflowApi.updateWorkflow(record.id, { status: 'active' })
      message.success('工作流已激活')
      fetchData()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  const handleDeactivate = async (record: Workflow) => {
    try {
      await workflowApi.updateWorkflow(record.id, { status: 'inactive' })
      message.success('工作流已停用')
      fetchData()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  const columns = [
    { title: '工作流名称', dataIndex: 'name', width: 200 },
    { title: '编码', dataIndex: 'code', width: 150 },
    { 
      title: '类型', 
      dataIndex: 'type', 
      width: 120,
      render: (type: string) => typeLabels[type] || type 
    },
    { 
      title: '状态', 
      dataIndex: 'status', 
      width: 100,
      render: (status: string) => {
        const { color, text } = statusLabels[status] || { color: 'default', text: status }
        return <Tag color={color}>{text}</Tag>
      },
    },
    { 
      title: '默认', 
      dataIndex: 'is_default', 
      width: 80,
      render: (isDefault: boolean) => isDefault ? <Tag color="blue">是</Tag> : '否',
    },
    { title: '版本', dataIndex: 'version', width: 80 },
    { 
      title: '创建时间', 
      dataIndex: 'created_at', 
      width: 180,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      width: 350,
      fixed: 'right',
      render: (_: any, record: Workflow) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)} />
          </Tooltip>
          
          <Tooltip title="添加步骤">
            <Button size="small" icon={<NodeExpandOutlined />} onClick={() => handleAddStep(record)} />
          </Tooltip>
          
          {record.status === 'draft' && (
            <Tooltip title="激活">
              <Button size="small" type="primary" icon={<PlayCircleOutlined />} onClick={() => handleActivate(record)}>
                激活
              </Button>
            </Tooltip>
          )}
          
          {record.status === 'active' && (
            <Tooltip title="停用">
              <Button size="small" danger icon={<CloseCircleOutlined />} onClick={() => handleDeactivate(record)}>
                停用
              </Button>
            </Tooltip>
          )}
          
          <Tooltip title="编辑">
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          
          <Tooltip title="删除">
            <Popconfirm title="确定删除此工作流？" onConfirm={() => handleDelete(record.id)}>
              <Button size="small" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <Select
                placeholder="状态筛选"
                allowClear
                style={{ width: 120 }}
                value={filters.status || undefined}
                onChange={(value) => setFilters({ ...filters, status: value })}
              >
                {Object.entries(statusLabels).map(([key, { text }]) => (
                  <Option key={key} value={key}>{text}</Option>
                ))}
              </Select>
              
              <Select
                placeholder="类型筛选"
                allowClear
                style={{ width: 150 }}
                value={filters.type || undefined}
                onChange={(value) => setFilters({ ...filters, type: value })}
              >
                {Object.entries(typeLabels).map(([key, label]) => (
                  <Option key={key} value={key}>{label}</Option>
                ))}
              </Select>
            </Space>
          </Col>
          <Col>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              新建工作流
            </Button>
          </Col>
        </Row>
      </Card>

      <Table
        columns={columns}
        dataSource={workflows}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1100 }}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={(p) => setPagination({ ...pagination, current: p.current || 1, pageSize: p.pageSize || 20 })}
      />

      {/* 新建/编辑工作流弹窗 */}
      <Modal
        title={selectedWorkflow ? '编辑工作流' : '新建工作流'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
        destroyOnClose
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item name="name" label="工作流名称" rules={[{ required: true, message: '请输入工作流名称' }]}>
            <Input placeholder="请输入工作流名称" />
          </Form.Item>
          
          <Form.Item name="code" label="工作流编码" rules={[{ required: true, message: '请输入工作流编码' }]}>
            <Input placeholder="请输入唯一编码，如：approval_workflow" disabled={!!selectedWorkflow} />
          </Form.Item>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="type" label="工作流类型" rules={[{ required: true }]} initialValue="approval">
                <Select placeholder="选择类型">
                  {Object.entries(typeLabels).map(([key, label]) => (
                    <Option key={key} value={key}>{label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="status" label="状态" initialValue="draft">
                <Select placeholder="选择状态">
                  {Object.entries(statusLabels).map(([key, { text }]) => (
                    <Option key={key} value={key}>{text}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="描述">
            <TextArea rows={3} placeholder="工作流描述" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 步骤编辑弹窗 */}
      <Modal
        title={selectedStep ? '编辑步骤' : '添加步骤'}
        open={stepModalVisible}
        onCancel={() => setStepModalVisible(false)}
        onOk={() => stepForm.submit()}
        width={600}
        destroyOnClose
      >
        <Form form={stepForm} onFinish={handleStepSubmit} layout="vertical">
          <Form.Item name="name" label="步骤名称" rules={[{ required: true, message: '请输入步骤名称' }]}>
            <Input placeholder="如：部门经理审批" />
          </Form.Item>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="step_type" label="步骤类型" rules={[{ required: true }]} initialValue="approval">
                <Select placeholder="选择类型">
                  {Object.entries(stepTypeLabels).map(([key, label]) => (
                    <Option key={key} value={key}>{label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="duration" label="预计时长(小时)" initialValue={24}>
                <Input type="number" min={0} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="assignee_type" label="分配类型" initialValue="manual">
                <Select placeholder="选择分配类型">
                  <Option value="manual">手动分配</Option>
                  <Option value="auto">自动分配</Option>
                  <Option value="role">按角色</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="assignee_role" label="分配角色">
                <Select placeholder="选择角色" allowClear>
                  <Option value="super_admin">超级管理员</Option>
                  <Option value="admin">管理员</Option>
                  <Option value="manager">管理者</Option>
                  <Option value="volunteer">志愿者</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="步骤描述">
            <TextArea rows={2} placeholder="步骤详细描述" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 工作流详情弹窗 */}
      <Modal
        title="工作流详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={900}
        destroyOnClose
      >
        {selectedWorkflow && (
          <div>
            <Card title="基本信息" style={{ marginBottom: 16 }}>
              <Row gutter={16}>
                <Col span={8}>
                  <p><strong>名称：</strong>{selectedWorkflow.name}</p>
                </Col>
                <Col span={8}>
                  <p><strong>编码：</strong>{selectedWorkflow.code}</p>
                </Col>
                <Col span={8}>
                  <p><strong>类型：</strong>{typeLabels[selectedWorkflow.type] || selectedWorkflow.type}</p>
                </Col>
              </Row>
              <Row gutter={16}>
                <Col span={8}>
                  <p>
                    <strong>状态：</strong>
                    <Tag color={statusLabels[selectedWorkflow.status]?.color}>
                      {statusLabels[selectedWorkflow.status]?.text}
                    </Tag>
                  </p>
                </Col>
                <Col span={8}>
                  <p><strong>版本：</strong>{selectedWorkflow.version}</p>
                </Col>
                <Col span={8}>
                  <p><strong>默认：</strong>{selectedWorkflow.is_default ? <Tag color="blue">是</Tag> : '否'}</p>
                </Col>
              </Row>
              {selectedWorkflow.description && (
                <Row>
                  <Col span={24}>
                    <p><strong>描述：</strong>{selectedWorkflow.description}</p>
                  </Col>
                </Row>
              )}
            </Card>

            <Card 
              title="流程步骤" 
              extra={
                <Button type="primary" icon={<PlusOutlined />} onClick={() => handleAddStep(selectedWorkflow)}>
                  添加步骤
                </Button>
              }
            >
              {selectedWorkflow.steps && selectedWorkflow.steps.length > 0 ? (
                <Steps direction="vertical" current={-1}>
                  {selectedWorkflow.steps.map((step, index) => (
                    <Step
                      key={step.id}
                      title={
                        <Space>
                          <span>{step.name}</span>
                          <Tag size="small">{stepTypeLabels[step.step_type] || step.step_type}</Tag>
                          {step.assignee_role && (
                            <Tag color="blue" size="small">{step.assignee_role}</Tag>
                          )}
                        </Space>
                      }
                      description={
                        <div style={{ marginTop: 8 }}>
                          {step.description && <p>{step.description}</p>}
                          <Space size="small">
                            <Tooltip title="编辑">
                              <Button size="small" icon={<EditOutlined />} onClick={() => handleEditStep(selectedWorkflow, step)} />
                            </Tooltip>
                            <Tooltip title="删除">
                              <Popconfirm title="确定删除此步骤？" onConfirm={() => handleDeleteStep(selectedWorkflow.id, step.id)}>
                                <Button size="small" danger icon={<DeleteOutlined />} />
                              </Popconfirm>
                            </Tooltip>
                          </Space>
                        </div>
                      }
                    />
                  ))}
                </Steps>
              ) : (
                <Empty description="暂无步骤，请点击「添加步骤」按钮创建" />
              )}
            </Card>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default Workflows
