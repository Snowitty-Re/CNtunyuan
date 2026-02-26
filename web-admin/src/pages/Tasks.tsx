import { useState } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, DatePicker, Progress, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, CheckCircleOutlined } from '@ant-design/icons'
import type { Task } from '../types'
import dayjs from 'dayjs'

const { Option } = Select
const { TextArea } = Input
const { RangePicker } = DatePicker

const Tasks = () => {
  const [data, setData] = useState<Task[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [form] = Form.useForm()

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

  const columns = [
    { title: '任务编号', dataIndex: 'task_no' },
    { title: '标题', dataIndex: 'title' },
    {
      title: '类型',
      dataIndex: 'type',
      render: (type: string) => typeLabels[type] || type,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      render: (priority: string) => (
        <Tag color={priorityColors[priority]}>{priorityLabels[priority]}</Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => {
        const { color, text } = statusLabels[status] || { color: 'default', text: status }
        return <Tag color={color}>{text}</Tag>
      },
    },
    {
      title: '进度',
      dataIndex: 'progress',
      render: (progress: number) => <Progress percent={progress} size="small" />,
    },
    {
      title: '负责人',
      dataIndex: ['assignee', 'nickname'],
    },
    {
      title: '截止时间',
      dataIndex: 'deadline',
      render: (date: string) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      render: (_: any, record: Task) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button size="small" icon={<CheckCircleOutlined />} type="primary">
            完成
          </Button>
          <Button size="small" danger icon={<DeleteOutlined />}>
            删除
          </Button>
        </Space>
      ),
    },
  ]

  const handleEdit = (record: Task) => {
    setSelectedTask(record)
    form.setFieldsValue({
      ...record,
      timeRange: record.start_time && record.deadline ? [dayjs(record.start_time), dayjs(record.deadline)] : undefined,
    })
    setModalVisible(true)
  }

  const handleSubmit = async (values: any) => {
    const data = {
      ...values,
      start_time: values.timeRange?.[0]?.toISOString(),
      deadline: values.timeRange?.[1]?.toISOString(),
    }
    delete data.timeRange
    
    if (selectedTask) {
      message.success('更新成功')
    } else {
      message.success('创建成功')
    }
    setModalVisible(false)
  }

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setSelectedTask(null); form.resetFields(); setModalVisible(true) }}>
          新建任务
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={(p) => setPagination({ ...pagination, current: p.current || 1, pageSize: p.pageSize || 20 })}
      />

      <Modal
        title={selectedTask ? '编辑任务' : '新建任务'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={700}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item name="title" label="任务标题" rules={[{ required: true }]}>
            <Input placeholder="请输入任务标题" />
          </Form.Item>
          <Form.Item name="type" label="任务类型" rules={[{ required: true }]}>
            <Select placeholder="选择任务类型">
              <Option value="search">实地寻访</Option>
              <Option value="call">电话核实</Option>
              <Option value="info_collect">信息收集</Option>
              <Option value="dialect_record">方言录制</Option>
              <Option value="coordination">协调沟通</Option>
              <Option value="other">其他</Option>
            </Select>
          </Form.Item>
          <Form.Item name="priority" label="优先级" rules={[{ required: true }]}>
            <Select placeholder="选择优先级">
              <Option value="urgent">紧急</Option>
              <Option value="high">高</Option>
              <Option value="normal">普通</Option>
              <Option value="low">低</Option>
            </Select>
          </Form.Item>
          <Form.Item name="description" label="任务描述">
            <TextArea rows={4} placeholder="详细描述任务内容和要求" />
          </Form.Item>
          <Form.Item name="timeRange" label="起止时间">
            <RangePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="location" label="任务地点">
            <Input placeholder="任务执行地点" />
          </Form.Item>
          <Form.Item name="requirements" label="任务要求">
            <TextArea rows={3} placeholder="具体任务要求和标准" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Tasks
