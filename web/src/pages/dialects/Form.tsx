import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Select, InputNumber, Button, Card, message, Spin, Space } from 'antd'
import { dialectApi } from '@/api/dialect'
import type { CreateDialectRequest } from '@/types'

const typeOptions = [
  { value: 'phrase', label: '短语' }, { value: 'story', label: '故事' },
  { value: 'song', label: '歌曲' }, { value: 'daily', label: '日常' },
  { value: 'other', label: '其他' },
]

export default function DialectForm() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const isEdit = !!id

  useEffect(() => {
    if (id) {
      setLoading(true)
      dialectApi.getById(id).then((res) => form.setFieldsValue(res.data.data)).finally(() => setLoading(false))
    }
  }, [id])

  const onFinish = async (values: Record<string, unknown>) => {
    setSubmitting(true)
    try {
      const payload: CreateDialectRequest = {
        title: values.title as string,
        content: values.content as string,
        dialect_type: values.dialect_type as CreateDialectRequest['dialect_type'],
        region: values.region as string,
        province: values.province as string,
        city: values.city as string,
        audio_url: values.audio_url as string,
        duration: values.duration as number,
        description: values.description as string,
        tags: values.tags as string,
      }
      if (isEdit) {
        await dialectApi.update(id!, payload)
        message.success('更新成功')
      } else {
        await dialectApi.create(payload)
        message.success('创建成功')
      }
      navigate('/dialects')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  return (
    <div>
      <h2 style={{ marginBottom: 24, fontSize: 20, fontWeight: 600 }}>{isEdit ? '编辑方言' : '上传方言'}</h2>
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 600 }}>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="dialect_type" label="类型" rules={[{ required: true, message: '请选择类型' }]}>
            <Select options={typeOptions} />
          </Form.Item>
          <Form.Item name="content" label="内容"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="region" label="地区" rules={[{ required: true, message: '请输入地区' }]}><Input placeholder="如：闽南语、粤语" /></Form.Item>
          <Form.Item name="province" label="省份"><Input /></Form.Item>
          <Form.Item name="city" label="城市"><Input /></Form.Item>
          <Form.Item name="audio_url" label="音频URL" rules={[{ required: true, message: '请输入音频URL' }]}><Input placeholder="上传后的音频文件地址" /></Form.Item>
          <Form.Item name="duration" label="时长(秒)"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="tags" label="标签"><Input placeholder="逗号分隔" /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={submitting}>{isEdit ? '保存' : '创建'}</Button>
              <Button onClick={() => navigate('/dialects')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
