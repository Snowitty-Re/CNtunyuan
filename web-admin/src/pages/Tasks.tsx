import { useEffect, useState } from 'react'
import { 
  Table, Button, Space, Tag, Modal, Form, Input, Select, DatePicker, 
  Progress, message, Card, Row, Col, Statistic, Popconfirm, Avatar, Timeline, Divider,
  Tabs, Tooltip, List, Descriptions
} from 'antd'
import { 
  PlusOutlined, EditOutlined, DeleteOutlined, CheckCircleOutlined, 
  UserOutlined, EyeOutlined,
  UserAddOutlined, UndoOutlined
} from '@ant-design/icons'
import type { Task, User } from '../types'
import { taskApi } from '../services/task'
import { userApi } from '../services/user'
import dayjs from 'dayjs'

const { Option } = Select
const { TextArea } = Input
const { TabPane } = Tabs

const Tasks = () => {
  const [data, setData] = useState<Task[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [detailModalVisible, setDetailModalVisible] = useState(false)
  const [assignModalVisible, setAssignModalVisible] = useState(false)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [form] = Form.useForm()
  const [assignForm] = Form.useForm()
  const [commentForm] = Form.useForm()
  const [filters, setFilters] = useState({ status: '', priority: '', type: '', keyword: '' })
  const [users, setUsers] = useState<User[]>([])
  const [statistics, setStatistics] = useState<any>({})
  const [logs, setLogs] = useState<any[]>([])
  const [comments, setComments] = useState<any[]>([])

  const priorityColors: Record<string, string> = {
    urgent: 'red',
    high: 'orange',
    normal: 'blue',
    low: 'green',
  }

  const priorityLabels: Record<string, string> = {
    urgent: '紧急',
    high: '高',
    normal: '普通',
    low: '低',
  }

  const statusLabels: Record<string, { color: string; text: string }> = {
    draft: { color: 'default', text: '草稿' },
    pending: { color: 'warning', text: '待分配' },
    assigned: { color: 'processing', text: '已分配' },
    processing: { color: 'blue', text: '进行中' },
    completed: { color: 'success', text: '已完成' },
    cancelled: { color: 'default', text: '已取消' },
  }

  const typeLabels: Record<string, string> = {
    search: '实地寻访',
    call: '电话核实',
    info_collect: '信息收集',
    dialect_record: '方言录制',
    coordination: '协调沟通',
    other: '其他',
  }

  useEffect(() => {
    fetchData()
    fetchStatistics()
    fetchUsers()
  }, [pagination.current, pagination.pageSize, filters])

  const fetchData = async () => {
    setLoading(true)
    try {
      const res = await taskApi.getTasks({
        ...filters,
        page: pagination.current,
        page_size: pagination.pageSize,
      })
      setData(res.list || [])
      setPagination({ ...pagination, total: res.total })
    } catch (error) {
      message.error('获取任务列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchStatistics = async () => {
    try {
      const res = await taskApi.getStatistics()
      setStatistics(res)
    } catch (error) {
      console.error('获取统计失败', error)
    }
  }

  const fetchUsers = async () => {
    try {
      const res = await userApi.getUsers({ page_size: 100 })
      setUsers(res.list || [])
    } catch (error) {
      console.error('获取用户列表失败', error)
    }
  }

  const fetchLogs = async (taskId: string) => {
    try {
      const res = await taskApi.getTaskLogs(taskId)
      setLogs(res || [])
    } catch (error) {
      console.error('获取日志失败', error)
    }
  }

  const fetchComments = async (taskId: string) => {
    try {
      const res = await taskApi.getComments(taskId)
      setComments(res || [])
    } catch (error) {
      console.error('获取评论失败', error)
    }
  }

  const handleCreate = () => {
    setSelectedTask(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: Task) => {
    setSelectedTask(record)
    form.setFieldsValue({
      ...record,
      deadline: record.deadline ? dayjs(record.deadline) : undefined,
    })
    setModalVisible(true)
  }

  const handleSubmit = async (values: any) => {
    try {
      const data = {
        ...values,
        deadline: values.deadline?.toISOString(),
      }

      if (selectedTask) {
        await taskApi.updateTask(selectedTask.id, data)
        message.success('更新成功')
      } else {
        await taskApi.createTask(data)
        message.success('创建成功')
      }
      setModalVisible(false)
      fetchData()
      fetchStatistics()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await taskApi.deleteTask(id)
      message.success('删除成功')
      fetchData()
      fetchStatistics()
    } catch (error: any) {
      message.error(error.message || '删除失败')
    }
  }

  const handleAssign = (record: Task) => {
    setSelectedTask(record)
    assignForm.resetFields()
    setAssignModalVisible(true)
  }

  const handleAssignSubmit = async (values: any) => {
    if (!selectedTask) return
    try {
      await taskApi.assignTask(selectedTask.id, values)
      message.success('分配成功')
      setAssignModalVisible(false)
      fetchData()
      fetchStatistics()
    } catch (error: any) {
      message.error(error.message || '分配失败')
    }
  }

  const handleUnassign = async (record: Task) => {
    try {
      await taskApi.unassignTask(record.id)
      message.success('取消分配成功')
      fetchData()
      fetchStatistics()
    } catch (error: any) {
      message.error(error.message || '取消分配失败')
    }
  }

  const handleComplete = async (record: Task) => {
    try {
      await taskApi.completeTask(record.id, {
        feedback: '任务已完成',
        result: '任务执行完毕',
      })
      message.success('任务已完成')
      fetchData()
      fetchStatistics()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  const handleCancel = async (record: Task) => {
    try {
      await taskApi.cancelTask(record.id, '管理员取消')
      message.success('任务已取消')
      fetchData()
      fetchStatistics()
    } catch (error: any) {
      message.error(error.message || '操作失败')
    }
  }

  const handleViewDetail = (record: Task) => {
    setSelectedTask(record)
    fetchLogs(record.id)
    fetchComments(record.id)
    setDetailModalVisible(true)
  }

  const handleAddComment = async (values: any) => {
    if (!selectedTask) return
    try {
      await taskApi.addComment(selectedTask.id, values)
      message.success('评论添加成功')
      commentForm.resetFields()
      fetchComments(selectedTask.id)
    } catch (error: any) {
      message.error(error.message || '添加评论失败')
    }
  }

  const columns = [
    { title: '任务编号', dataIndex: 'task_no', width: 150 },
    { title: '标题', dataIndex: 'title', ellipsis: true },
    {
      title: '类型',
      dataIndex: 'type',
      width: 100,
      render: (type: string) => typeLabels[type] || type,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 80,
      render: (priority: string) => (
        <Tag color={priorityColors[priority]}>{priorityLabels[priority]}</Tag>
      ),
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
      title: '进度',
      dataIndex: 'progress',
      width: 100,
      render: (progress: number) => <Progress percent={progress} size="small" />,
    },
    {
      title: '负责人',
      dataIndex: 'assignee',
      width: 120,
      render: (assignee: any) => assignee ? (
        <Space>
          <Avatar src={assignee.avatar} icon={<UserOutlined />} size="small" />
          {assignee.nickname || assignee.real_name}
        </Space>
      ) : (
        <Tag color="warning">待分配</Tag>
      ),
    },
    {
      title: '截止时间',
      dataIndex: 'deadline',
      width: 150,
      render: (date: string) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      width: 250,
      fixed: 'right',
      render: (_: any, record: Task) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button size="small" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)} />
          </Tooltip>
          
          {record.status === 'pending' && (
            <Tooltip title="分配">
              <Button size="small" type="primary" icon={<UserAddOutlined />} onClick={() => handleAssign(record)} />
            </Tooltip>
          )}
          
          {record.status === 'assigned' && (
            <Tooltip title="取消分配">
              <Button size="small" icon={<UndoOutlined />} onClick={() => handleUnassign(record)} />
            </Tooltip>
          )}
          
          {(record.status === 'assigned' || record.status === 'processing') && (
            <Tooltip title="完成">
              <Button size="small" type="primary" icon={<CheckCircleOutlined />} onClick={() => handleComplete(record)} />
            </Tooltip>
          )}
          
          <Tooltip title="编辑">
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          
          {record.status !== 'completed' && record.status !== 'cancelled' && (
            <Tooltip title="取消任务">
              <Popconfirm title="确定取消此任务？" onConfirm={() => handleCancel(record)}>
                <Button size="small" danger icon={<DeleteOutlined />} />
              </Popconfirm>
            </Tooltip>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      {/* 统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={4}>
          <Card>
            <Statistic title="总任务" value={statistics.total || 0} />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic title="待分配" value={statistics.pending || 0} valueStyle={{ color: '#faad14' }} />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic title="进行中" value={statistics.processing || 0} valueStyle={{ color: '#1890ff' }} />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic title="已完成" value={statistics.completed || 0} valueStyle={{ color: '#52c41a' }} />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic title="今日新增" value={statistics.today || 0} />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate} block>
              新建任务
            </Button>
          </Card>
        </Col>
      </Row>

      {/* 筛选栏 */}
      <Card style={{ marginBottom: 16 }}>
        <Space wrap>
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
            placeholder="优先级"
            allowClear
            style={{ width: 120 }}
            value={filters.priority || undefined}
            onChange={(value) => setFilters({ ...filters, priority: value })}
          >
            {Object.entries(priorityLabels).map(([key, label]) => (
              <Option key={key} value={key}>{label}</Option>
            ))}
          </Select>
          
          <Select
            placeholder="任务类型"
            allowClear
            style={{ width: 150 }}
            value={filters.type || undefined}
            onChange={(value) => setFilters({ ...filters, type: value })}
          >
            {Object.entries(typeLabels).map(([key, label]) => (
              <Option key={key} value={key}>{label}</Option>
            ))}
          </Select>
          
          <Input.Search
            placeholder="搜索任务标题"
            allowClear
            style={{ width: 250 }}
            value={filters.keyword}
            onChange={(e) => setFilters({ ...filters, keyword: e.target.value })}
            onSearch={() => fetchData()}
          />
        </Space>
      </Card>

      {/* 任务表格 */}
      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        scroll={{ x: 1300 }}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={(p) => setPagination({ ...pagination, current: p.current || 1, pageSize: p.pageSize || 20 })}
      />

      {/* 新建/编辑任务弹窗 */}
      <Modal
        title={selectedTask ? '编辑任务' : '新建任务'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={700}
        destroyOnClose
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item name="title" label="任务标题" rules={[{ required: true, message: '请输入任务标题' }]}>
            <Input placeholder="请输入任务标题" />
          </Form.Item>
          
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="type" label="任务类型" rules={[{ required: true }]}>
                <Select placeholder="选择任务类型">
                  {Object.entries(typeLabels).map(([key, label]) => (
                    <Option key={key} value={key}>{label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="priority" label="优先级" initialValue="normal">
                <Select placeholder="选择优先级">
                  {Object.entries(priorityLabels).map(([key, label]) => (
                    <Option key={key} value={key}>{label}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="任务描述">
            <TextArea rows={3} placeholder="详细描述任务内容和要求" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="deadline" label="截止时间">
                <DatePicker showTime style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="estimated_hours" label="预计工时(小时)">
                <Input type="number" min={0} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="location" label="任务地点">
                <Input placeholder="任务执行地点" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="assignee_id" label="分配给">
                <Select placeholder="选择执行人" allowClear>
                  {users.map((user) => (
                    <Option key={user.id} value={user.id}>{user.nickname || user.real_name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="requirements" label="任务要求">
            <TextArea rows={2} placeholder="具体任务要求和标准" />
          </Form.Item>

          <Form.Item name="notes" label="备注">
            <TextArea rows={2} placeholder="其他备注信息" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 分配任务弹窗 */}
      <Modal
        title="分配任务"
        open={assignModalVisible}
        onCancel={() => setAssignModalVisible(false)}
        onOk={() => assignForm.submit()}
        destroyOnClose
      >
        <Form form={assignForm} onFinish={handleAssignSubmit} layout="vertical">
          <Form.Item name="assignee_id" label="选择执行人" rules={[{ required: true, message: '请选择执行人' }]}>
            <Select placeholder="选择执行人">
              {users.map((user) => (
                <Option key={user.id} value={user.id}>
                  <Space>
                    <Avatar src={user.avatar} icon={<UserOutlined />} size="small" />
                    {user.nickname || user.real_name}
                    <Tag>{user.phone}</Tag>
                  </Space>
                </Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name="comment" label="备注">
            <TextArea rows={2} placeholder="分配说明（可选）" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 任务详情弹窗 */}
      <Modal
        title="任务详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={null}
        width={800}
        destroyOnClose
      >
        {selectedTask && (
          <Tabs defaultActiveKey="info">
            <TabPane tab="基本信息" key="info">
              <Descriptions bordered column={2}>
                <Descriptions.Item label="任务编号">{selectedTask.task_no}</Descriptions.Item>
                <Descriptions.Item label="状态">
                  <Tag color={statusLabels[selectedTask.status]?.color}>
                    {statusLabels[selectedTask.status]?.text}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="类型">{typeLabels[selectedTask.type]}</Descriptions.Item>
                <Descriptions.Item label="优先级">
                  <Tag color={priorityColors[selectedTask.priority]}>
                    {priorityLabels[selectedTask.priority]}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label="创建人">{selectedTask.creator?.nickname}</Descriptions.Item>
                <Descriptions.Item label="负责人">
                  {selectedTask.assignee ? selectedTask.assignee.nickname : <Tag color="warning">待分配</Tag>}
                </Descriptions.Item>
                <Descriptions.Item label="截止时间">
                  {selectedTask.deadline ? dayjs(selectedTask.deadline).format('YYYY-MM-DD HH:mm') : '-'}
                </Descriptions.Item>
                <Descriptions.Item label="预计工时">{selectedTask.estimated_hours || '-'} 小时</Descriptions.Item>
              </Descriptions>
              <Divider />
              <h4>任务描述</h4>
              <p>{selectedTask.description || '无描述'}</p>
              {selectedTask.feedback && (
                <>
                  <Divider />
                  <h4>任务反馈</h4>
                  <p>{selectedTask.feedback}</p>
                </>
              )}
            </TabPane>
            
            <TabPane tab="操作日志" key="logs">
              <Timeline>
                {logs.map((log, index) => (
                  <Timeline.Item key={index}>
                    <p><strong>{log.action}</strong> - {log.user?.nickname}</p>
                    <p>{log.content}</p>
                    <p style={{ color: '#999', fontSize: 12 }}>{dayjs(log.created_at).format('YYYY-MM-DD HH:mm:ss')}</p>
                  </Timeline.Item>
                ))}
              </Timeline>
            </TabPane>
            
            <TabPane tab="评论" key="comments">
              <List
                dataSource={comments}
                renderItem={(item) => (
                  <List.Item>
                    <List.Item.Meta
                      avatar={<Avatar src={item.user?.avatar} icon={<UserOutlined />} />}
                      title={item.user?.nickname}
                      description={item.content}
                    />
                    <div>{dayjs(item.created_at).format('YYYY-MM-DD HH:mm')}</div>
                  </List.Item>
                )}
              />
              <Divider />
              <Form form={commentForm} onFinish={handleAddComment}>
                <Form.Item name="content" rules={[{ required: true, message: '请输入评论内容' }]}>
                  <TextArea rows={3} placeholder="添加评论..." />
                </Form.Item>
                <Button type="primary" htmlType="submit">发表评论</Button>
              </Form>
            </TabPane>
          </Tabs>
        )}
      </Modal>
    </div>
  )
}

export default Tasks
