import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Descriptions, Tag, Button, Space, Card, Timeline, Modal, Form, Input, Select, Slider, message, Spin, Divider, Progress, Empty } from 'antd'
import { ArrowLeftOutlined, ExclamationCircleOutlined } from '@ant-design/icons'
import { taskApi } from '@/api/task'
import { userApi } from '@/api/user'
import { usePermission } from '@/hooks/usePermission'
import { taskStatusMap, taskPriorityMap, taskTypeMap } from '@/constants'
import type { Task, TaskLog, User } from '@/types'
import dayjs from 'dayjs'

export default function TaskDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { isManager } = usePermission()
  const [data, setData] = useState<Task | null>(null)
  const [logs, setLogs] = useState<TaskLog[]>([])
  const [loading, setLoading] = useState(true)
  const [users, setUsers] = useState<User[]>([])
  const [completeModal, setCompleteModal] = useState(false)
  const [progressModal, setProgressModal] = useState(false)
  const [assignModal, setAssignModal] = useState(false)
  const [cancelModal, setCancelModal] = useState(false)
  const [completeForm] = Form.useForm()
  const [assignForm] = Form.useForm()
  const [cancelForm] = Form.useForm()
  const [progressVal, setProgressVal] = useState(0)

  const fetchData = async () => {
    if (!id) return
    setLoading(true)
    try {
      const [res, logsRes] = await Promise.all([taskApi.getById(id), taskApi.getLogs(id)])
      setData(res.data.data)
      setLogs(Array.isArray(logsRes.data.data) ? logsRes.data.data : [])
      setProgressVal(res.data.data.progress || 0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [id])

  // Load users when assign modal opens
  const openAssignModal = async () => {
    setAssignModal(true)
    if (users.length === 0) {
      try {
        const res = await userApi.list({ page: 1, page_size: 100, status: 'active' })
        setUsers(res.data.data.list || [])
      } catch { /* ignore */ }
    }
  }

  const handleStart = () => {
    Modal.confirm({
      title: '开始处理任务',
      icon: <ExclamationCircleOutlined />,
      content: '确定开始处理此任务？任务状态将变为"进行中"。',
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        await taskApi.start(id!)
        message.success('已开始处理')
        fetchData()
      },
    })
  }

  const handleComplete = async (values: { result: string }) => {
    await taskApi.complete(id!, values.result)
    message.success('任务已完成')
    setCompleteModal(false)
    completeForm.resetFields()
    fetchData()
  }

  const handleProgress = async () => {
    await taskApi.updateProgress(id!, progressVal)
    message.success('进度已更新')
    setProgressModal(false)
    fetchData()
  }

  const handleAssign = async (values: { assignee_id: string }) => {
    await taskApi.assign(id!, values.assignee_id)
    message.success('分配成功')
    setAssignModal(false)
    assignForm.resetFields()
    fetchData()
  }

  const handleCancel = async (values: { reason: string }) => {
    await taskApi.cancel(id!, values.reason)
    message.success('任务已取消')
    setCancelModal(false)
    cancelForm.resetFields()
    fetchData()
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!data) return <Empty description="未找到任务" />

  const st = taskStatusMap[data.status]
  const pr = taskPriorityMap[data.priority]
  const canStart = ['pending', 'assigned'].includes(data.status)
  const canComplete = data.status === 'processing'
  const canCancel = ['processing', 'assigned', 'pending'].includes(data.status)

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16, flexWrap: 'wrap', gap: 8 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tasks')}>返回</Button>
          <h2 style={{ margin: 0 }}>任务详情</h2>
          {st && <Tag color={st.color}>{st.label}</Tag>}
        </Space>
        <Space wrap>
          {canStart && <Button type="primary" onClick={handleStart}>开始处理</Button>}
          {canComplete && <Button type="primary" onClick={() => setCompleteModal(true)}>完成任务</Button>}
          {canComplete && <Button onClick={() => setProgressModal(true)}>更新进度</Button>}
          {isManager && (data.status === 'pending' || data.status === 'assigned') && <Button onClick={openAssignModal}>分配任务</Button>}
          {isManager && canCancel && <Button danger onClick={() => setCancelModal(true)}>取消任务</Button>}
          <Button onClick={() => navigate(`/tasks/${id}/edit`)}>编辑</Button>
        </Space>
      </div>

      <Card>
        <Descriptions bordered column={{ xs: 1, sm: 2, lg: 3 }}>
          <Descriptions.Item label="标题">{data.title}</Descriptions.Item>
          <Descriptions.Item label="状态"><Tag color={st?.color}>{st?.label}</Tag></Descriptions.Item>
          <Descriptions.Item label="优先级"><Tag color={pr?.color}>{pr?.label}</Tag></Descriptions.Item>
          <Descriptions.Item label="类型"><Tag>{taskTypeMap[data.type] || data.type}</Tag></Descriptions.Item>
          <Descriptions.Item label="截止时间">
            {data.deadline ? (
              <span style={{ color: dayjs(data.deadline).isBefore(dayjs()) && data.status !== 'completed' ? '#ff4d4f' : undefined }}>
                {dayjs(data.deadline).format('YYYY-MM-DD HH:mm')}
                {dayjs(data.deadline).isBefore(dayjs()) && data.status !== 'completed' && <Tag color="red" style={{ marginLeft: 4 }}>已逾期</Tag>}
              </span>
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="进度"><Progress percent={data.progress || 0} size="small" /></Descriptions.Item>
          <Descriptions.Item label="创建人">{data.creator?.nickname || '-'}</Descriptions.Item>
          <Descriptions.Item label="执行人">{data.assignee?.nickname || <span style={{ color: '#8f959e' }}>未分配</span>}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{dayjs(data.created_at).format('YYYY-MM-DD HH:mm')}</Descriptions.Item>
          {data.started_at && <Descriptions.Item label="开始时间">{dayjs(data.started_at).format('YYYY-MM-DD HH:mm')}</Descriptions.Item>}
          {data.completed_at && <Descriptions.Item label="完成时间">{dayjs(data.completed_at).format('YYYY-MM-DD HH:mm')}</Descriptions.Item>}
          <Descriptions.Item label="地区">{[data.province, data.city, data.district].filter(Boolean).join(' ') || '-'}</Descriptions.Item>
          {data.address && <Descriptions.Item label="地址" span={2}>{data.address}</Descriptions.Item>}
          <Descriptions.Item label="描述" span={3}>{data.description || '-'}</Descriptions.Item>
          {data.result && <Descriptions.Item label="完成结果" span={3}>{data.result}</Descriptions.Item>}
          {data.feedback && <Descriptions.Item label="反馈" span={3}>{data.feedback}</Descriptions.Item>}
          {data.missing_person && (
            <Descriptions.Item label="关联案件" span={3}>
              <Button type="link" style={{ padding: 0 }} onClick={() => navigate(`/cases/${data.missing_person_id}`)}>
                {data.missing_person.name} ({data.missing_person.case_no})
              </Button>
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>

      <Divider />

      <Card title="操作日志">
        {logs.length === 0 ? <Empty description="暂无日志" /> : (
          <Timeline items={logs.map((l) => ({
            children: (
              <div>
                <div style={{ fontWeight: 500 }}>{l.content || l.action}</div>
                {l.old_status && l.new_status && (
                  <div style={{ fontSize: 12, color: '#646a73', marginTop: 2 }}>
                    {taskStatusMap[l.old_status]?.label || l.old_status} → {taskStatusMap[l.new_status]?.label || l.new_status}
                  </div>
                )}
                <div style={{ color: '#8f959e', fontSize: 12, marginTop: 4 }}>
                  {dayjs(l.created_at).format('YYYY-MM-DD HH:mm')} {l.user?.nickname ? `· ${l.user.nickname}` : ''}
                </div>
              </div>
            ),
          }))} />
        )}
      </Card>

      <Modal title="完成任务" open={completeModal} onCancel={() => setCompleteModal(false)} onOk={() => completeForm.submit()} okText="确认完成">
        <Form form={completeForm} onFinish={handleComplete} layout="vertical">
          <Form.Item name="result" label="完成结果" rules={[{ required: true, message: '请输入结果' }]}>
            <Input.TextArea rows={4} placeholder="描述任务完成的结果和情况" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="更新进度" open={progressModal} onCancel={() => setProgressModal(false)} onOk={handleProgress} okText="更新">
        <div style={{ padding: '16px 0' }}>
          <Slider min={0} max={100} value={progressVal} onChange={setProgressVal} marks={{ 0: '0%', 25: '25%', 50: '50%', 75: '75%', 100: '100%' }} />
        </div>
      </Modal>

      <Modal title="分配任务" open={assignModal} onCancel={() => setAssignModal(false)} onOk={() => assignForm.submit()} okText="分配">
        <Form form={assignForm} onFinish={handleAssign} layout="vertical">
          <Form.Item name="assignee_id" label="选择执行人" rules={[{ required: true, message: '请选择执行人' }]}>
            <Select
              showSearch
              placeholder="搜索并选择执行人"
              optionFilterProp="label"
              options={users.map((u) => ({ value: u.id, label: `${u.nickname} (${u.phone})` }))}
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="取消任务" open={cancelModal} onCancel={() => setCancelModal(false)} onOk={() => cancelForm.submit()} okText="确认取消" okButtonProps={{ danger: true }}>
        <Form form={cancelForm} onFinish={handleCancel} layout="vertical">
          <Form.Item name="reason" label="取消原因" rules={[{ required: true, message: '请输入原因' }]}>
            <Input.TextArea rows={3} placeholder="请说明取消此任务的原因" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
