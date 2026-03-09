import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Select, DatePicker, Button, Card, Row, Col, message, Spin, Space } from 'antd'
import { taskApi } from '@/api/task'
import { missingPersonApi } from '@/api/missingPerson'
import type { CreateTaskRequest, MissingPerson } from '@/types'
import dayjs from 'dayjs'

const typeOptions = [
  { value: 'search', label: '搜索' }, { value: 'verify', label: '核实' },
  { value: 'assist', label: '协助' }, { value: 'follow', label: '跟进' },
  { value: 'interview', label: '访谈' }, { value: 'other', label: '其他' },
]
const priorityOptions = [
  { value: 'low', label: '低' }, { value: 'medium', label: '中' },
  { value: 'high', label: '高' }, { value: 'urgent', label: '紧急' },
]

export default function TaskForm() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [cases, setCases] = useState<MissingPerson[]>([])
  const [caseSearching, setCaseSearching] = useState(false)
  const isEdit = !!id

  useEffect(() => {
    // Preload active cases
    missingPersonApi.list({ page: 1, page_size: 50 }).then((res) => {
      setCases(res.data.data?.list || [])
    }).catch(() => {})

    if (id) {
      setLoading(true)
      taskApi.getById(id).then((res) => {
        const d = res.data.data
        form.setFieldsValue({ ...d, deadline: d.deadline ? dayjs(d.deadline) : undefined })
      }).finally(() => setLoading(false))
    }
  }, [id])

  const handleCaseSearch = async (value: string) => {
    if (!value || value.length < 1) return
    setCaseSearching(true)
    try {
      const res = await missingPersonApi.list({ page: 1, page_size: 20, keyword: value })
      setCases(res.data.data?.list || [])
    } finally {
      setCaseSearching(false)
    }
  }

  const onFinish = async (values: Record<string, unknown>) => {
    setSubmitting(true)
    try {
      const payload: CreateTaskRequest = {
        title: values.title as string,
        description: values.description as string,
        type: values.type as CreateTaskRequest['type'],
        priority: values.priority as CreateTaskRequest['priority'],
        deadline: values.deadline ? (values.deadline as dayjs.Dayjs).toISOString() : undefined,
        missing_person_id: values.missing_person_id as string,
        province: values.province as string,
        city: values.city as string,
        district: values.district as string,
        address: values.address as string,
      }
      if (isEdit) {
        await taskApi.update(id!, payload)
        message.success('更新成功')
      } else {
        await taskApi.create(payload)
        message.success('创建成功')
      }
      navigate('/tasks')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  return (
    <div>
      <h2 style={{ marginBottom: 24, fontSize: 20, fontWeight: 600 }}>{isEdit ? '编辑任务' : '新建任务'}</h2>
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 800 }} initialValues={{ priority: 'medium' }}>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input placeholder="简要描述任务内容" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="type" label="类型" rules={[{ required: true, message: '请选择类型' }]}>
                <Select options={typeOptions} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="priority" label="优先级">
                <Select options={priorityOptions} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="deadline" label="截止时间">
                <DatePicker showTime style={{ width: '100%' }} disabledDate={(d) => d.isBefore(dayjs(), 'day')} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={4} placeholder="详细描述任务要求和注意事项" />
          </Form.Item>
          <Form.Item name="missing_person_id" label="关联案件">
            <Select
              showSearch
              allowClear
              placeholder="搜索并选择关联的案件（可选）"
              filterOption={false}
              onSearch={handleCaseSearch}
              loading={caseSearching}
              options={cases.map((c) => ({
                value: c.id,
                label: `${c.name} - ${c.case_no} (${c.status === 'missing' ? '走失' : c.status})`,
              }))}
            />
          </Form.Item>
          <Row gutter={16}>
            <Col span={8}><Form.Item name="province" label="省份"><Input /></Form.Item></Col>
            <Col span={8}><Form.Item name="city" label="城市"><Input /></Form.Item></Col>
            <Col span={8}><Form.Item name="district" label="区县"><Input /></Form.Item></Col>
          </Row>
          <Form.Item name="address" label="详细地址"><Input /></Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={submitting}>{isEdit ? '保存' : '创建'}</Button>
              <Button onClick={() => navigate('/tasks')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
