import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Descriptions, Tag, Button, Space, Card, Timeline, Modal, Form, Input, Slider, message, Spin, Divider, Progress } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { taskApi } from '@/api/task'
import { usePermission } from '@/hooks/usePermission'
import type { Task, TaskLog } from '@/types'
import dayjs from 'dayjs'

const statusMap: Record<string, { label: string; color: string }> = {
  draft: { label: '草稿', color: 'default' }, pending: { label: '待分配', color: 'gold' },
  assigned: { label: '已分配', color: 'cyan' }, processing: { label: '进行中', color: 'blue' },
  completed: { label: '已完成', color: 'green' }, cancelled: { label: '已取消', color: 'red' },
  overdue: { label: '已逾期', color: 'magenta' },
}
const priorityMap: Record<string, { label: string; color: string }> = {
  low: { label: '低', color: 'default' }, medium: { label: '中', color: 'blue' },
  high: { label: '高', color: 'orange' }, urgent: { label: '紧急', color: 'red' },
}

export default function TaskDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { isManager } = usePermission()
  const [data, setData] = useState<Task | null>(null)
  const [logs, setLogs] = useState<TaskLog[]>([])
  const [loading, setLoading] = useState(true)
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

  const handleStart = async () => {
    await taskApi.start(id!)
    message.success('已开始处理')
    fetchData()
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
  if (!data) return <div>未找到任务</div>

  const st = statusMap[data.status]
  const pr = priorityMap[data.priority]
  const canStart = ['pending', 'assigned'].includes(data.status)
  const canComplete = data.status === 'processing'
  const canCancel = ['processing', 'assigned', 'pending'].includes(data.status)

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tasks')}>返回</Button>
          <h2 style={{ margin: 0 }}>任务详情</h2>
        </Space>
        <Space>
          {canStart && <Button type="primary" onClick={handleStart}>开始处理</Button>}
          {canComplete && <Button type="primary" onClick={() => setCompleteModal(true)}>完成任务</Button>}
          {canComplete && <Button onClick={() => setProgressModal(true)}>更新进度</Button>}
          {isManager && data.status === 'pending' && <Button onClick={() => setAssignModal(true)}>分配任务</Button>}
          {isManager && canCancel && <Button danger onClick={() => setCancelModal(true)}>取消任务</Button>}
        </Space>
      </div>

      <Card>
        <Descriptions bordered column={{ xs: 1, sm: 2, lg: 3 }}>
          <Descriptions.Item label="标题">{data.title}</Descriptions.Item>
          <Descriptions.Item label="状态"><Tag color={st?.color}>{st?.label}</Tag></Descriptions.Item>
          <Descriptions.Item label="优先级"><Tag color={pr?.color}>{pr?.label}</Tag></Descriptions.Item>
          <Descriptions.Item label="类型">{data.type}</Descriptions.Item>
          <Descriptions.Item label="截止时间">{data.deadline ? dayjs(data.deadline).format('YYYY-MM-DD HH:mm') : '-'}</Descriptions.Item>
          <Descriptions.Item label="进度"><Progress percent={data.progress || 0} size="small" /></Descriptions.Item>
          <Descriptions.Item label="创建人">{data.creator?.nickname || '-'}</Descriptions.Item>
          <Descriptions.Item label="执行人">{data.assignee?.nickname || '-'}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{dayjs(data.created_at).format('YYYY-MM-DD HH:mm')}</Descriptions.Item>
          <Descriptions.Item label="地区">{[data.province, data.city, data.district].filter(Boolean).join(' ') || '-'}</Descriptions.Item>
          <Descriptions.Item label="地址" span={2}>{data.address || '-'}</Descriptions.Item>
          <Descriptions.Item label="描述" span={3}>{data.description || '-'}</Descriptions.Item>
          {data.result && <Descriptions.Item label="结果" span={3}>{data.result}</Descriptions.Item>}
          {data.missing_person && (
            <Descriptions.Item label="关联案件" span={3}>
              <Button type="link" onClick={() => navigate(`/cases/${data.missing_person_id}`)}>{data.missing_person.name} ({data.missing_person.case_no})</Button>
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>

      <Divider />

      <Card title="操作日志">
        {logs.length === 0 ? <div style={{ textAlign: 'center', color: '#8f959e', padding: 24 }}>暂无日志</div> : (
          <Timeline items={logs.map((l) => ({
            children: (
              <div>
                <div style={{ fontWeight: 500 }}>{l.content || l.action}</div>
                <div style={{ color: '#8f959e', fontSize: 12, marginTop: 4 }}>
                  {dayjs(l.created_at).format('YYYY-MM-DD HH:mm')} {l.user?.nickname ? `· ${l.user.nickname}` : ''}
                </div>
              </div>
            ),
          }))} />
        )}
      </Card>

      <Modal title="完成任务" open={completeModal} onCancel={() => setCompleteModal(false)} onOk={() => completeForm.submit()}>
        <Form form={completeForm} onFinish={handleComplete} layout="vertical">
          <Form.Item name="result" label="完成结果" rules={[{ required: true, message: '请输入结果' }]}>
            <Input.TextArea rows={4} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="更新进度" open={progressModal} onCancel={() => setProgressModal(false)} onOk={handleProgress}>
        <div style={{ padding: '16px 0' }}>
          <Slider min={0} max={100} value={progressVal} onChange={setProgressVal} marks={{ 0: '0%', 50: '50%', 100: '100%' }} />
        </div>
      </Modal>

      <Modal title="分配任务" open={assignModal} onCancel={() => setAssignModal(false)} onOk={() => assignForm.submit()}>
        <Form form={assignForm} onFinish={handleAssign} layout="vertical">
          <Form.Item name="assignee_id" label="执行人ID" rules={[{ required: true, message: '请输入执行人ID' }]}>
            <Input placeholder="输入用户ID" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="取消任务" open={cancelModal} onCancel={() => setCancelModal(false)} onOk={() => cancelForm.submit()}>
        <Form form={cancelForm} onFinish={handleCancel} layout="vertical">
          <Form.Item name="reason" label="取消原因" rules={[{ required: true, message: '请输入原因' }]}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
