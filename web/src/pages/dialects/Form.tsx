import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Select, InputNumber, Button, Card, Upload, message, Spin, Space } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { dialectApi } from '@/api/dialect'
import { uploadApi } from '@/api/upload'
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
  const [uploading, setUploading] = useState(false)
  const [audioUrl, setAudioUrl] = useState('')
  const isEdit = !!id

  useEffect(() => {
    if (id) {
      setLoading(true)
      dialectApi.getById(id).then((res) => {
        form.setFieldsValue(res.data.data)
        setAudioUrl(res.data.data.audio_url || '')
      }).finally(() => setLoading(false))
    }
  }, [id])

  const handleUpload = async (file: File) => {
    const allowedTypes = ['audio/mpeg', 'audio/wav', 'audio/mp3', 'audio/ogg', 'audio/flac', 'audio/m4a', 'audio/x-m4a']
    if (!allowedTypes.includes(file.type) && !file.name.match(/\.(mp3|wav|ogg|flac|m4a)$/i)) {
      message.error('仅支持 mp3, wav, ogg, flac, m4a 格式的音频文件')
      return false
    }
    if (file.size > 50 * 1024 * 1024) {
      message.error('文件大小不能超过 50MB')
      return false
    }

    setUploading(true)
    try {
      const res = await uploadApi.upload(file)
      const url = res.data.data.url
      setAudioUrl(url)
      form.setFieldsValue({ audio_url: url })
      message.success('音频上传成功')
    } catch {
      message.error('上传失败')
    } finally {
      setUploading(false)
    }
    return false // prevent default upload
  }

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
          <Form.Item name="region" label="方言区域" rules={[{ required: true, message: '请输入方言区域' }]}>
            <Input placeholder="如: 闽南语、粤语、吴语" />
          </Form.Item>
          <Form.Item name="province" label="省份"><Input /></Form.Item>
          <Form.Item name="city" label="城市"><Input /></Form.Item>

          <Form.Item label="音频文件" required>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Upload
                accept=".mp3,.wav,.ogg,.flac,.m4a"
                showUploadList={false}
                beforeUpload={handleUpload}
              >
                <Button icon={<UploadOutlined />} loading={uploading}>
                  {uploading ? '上传中...' : '选择音频文件'}
                </Button>
              </Upload>
              {audioUrl && (
                <div style={{ padding: '8px 12px', background: '#f5f5f5', borderRadius: 6 }}>
                  <audio controls src={audioUrl} style={{ width: '100%' }} />
                </div>
              )}
            </Space>
          </Form.Item>
          <Form.Item name="audio_url" hidden rules={[{ required: true, message: '请上传音频文件' }]}>
            <Input />
          </Form.Item>

          <Form.Item name="duration" label="时长(秒)"><InputNumber min={0} style={{ width: '100%' }} placeholder="自动填充或手动输入" /></Form.Item>
          <Form.Item name="tags" label="标签"><Input placeholder="逗号分隔，如: 闽南语,泉州,日常用语" /></Form.Item>
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
