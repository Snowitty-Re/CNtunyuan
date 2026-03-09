import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Select, DatePicker, InputNumber, Button, Card, Upload, Image, Row, Col, message, Spin, Space } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { missingPersonApi } from '@/api/missingPerson'
import { uploadApi } from '@/api/upload'
import { PHONE_RULE } from '@/constants'
import type { CreateMissingPersonRequest } from '@/types'
import dayjs from 'dayjs'

export default function CaseForm() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [photoUrl, setPhotoUrl] = useState('')
  const [uploading, setUploading] = useState(false)
  const isEdit = !!id

  useEffect(() => {
    if (id) {
      setLoading(true)
      missingPersonApi.getById(id).then((res) => {
        const d = res.data.data
        form.setFieldsValue({
          ...d,
          birth_date: d.birth_date ? dayjs(d.birth_date) : undefined,
          missing_time: d.missing_time ? dayjs(d.missing_time) : undefined,
          urgency_level: d.urgency,
        })
        if (d.photo_url) setPhotoUrl(d.photo_url)
      }).finally(() => setLoading(false))
    }
  }, [id])

  const handleUploadPhoto = async (file: File) => {
    if (!file.type.startsWith('image/')) {
      message.error('请上传图片文件')
      return false
    }
    if (file.size > 10 * 1024 * 1024) {
      message.error('图片不能超过 10MB')
      return false
    }
    setUploading(true)
    try {
      const res = await uploadApi.upload(file)
      const url = res.data.data.url
      setPhotoUrl(url)
      form.setFieldsValue({ photo_url: url })
      message.success('照片上传成功')
    } catch {
      message.error('上传失败')
    } finally {
      setUploading(false)
    }
    return false
  }

  const onFinish = async (values: Record<string, unknown>) => {
    setSubmitting(true)
    try {
      const payload: CreateMissingPersonRequest = {
        name: values.name as string,
        gender: values.gender as string,
        birth_date: values.birth_date ? (values.birth_date as dayjs.Dayjs).format('YYYY-MM-DD') : undefined,
        age: values.age as number,
        height: values.height as number,
        weight: values.weight as number,
        description: values.description as string,
        photo_url: values.photo_url as string,
        missing_time: (values.missing_time as dayjs.Dayjs).toISOString(),
        province: values.province as string,
        city: values.city as string,
        district: values.district as string,
        address: values.address as string,
        clothes: values.clothes as string,
        features: values.features as string,
        contact_name: values.contact_name as string,
        contact_phone: values.contact_phone as string,
        contact_rel: values.contact_rel as string,
        alt_contact: values.alt_contact as string,
        urgency_level: values.urgency_level as CreateMissingPersonRequest['urgency_level'],
      }
      if (isEdit) {
        await missingPersonApi.update(id!, payload)
        message.success('更新成功')
      } else {
        await missingPersonApi.create(payload)
        message.success('创建成功')
      }
      navigate('/cases')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  return (
    <div>
      <h2 style={{ marginBottom: 24, fontSize: 20, fontWeight: 600 }}>{isEdit ? '编辑案件' : '新建案件'}</h2>
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 800 }}>
          <Card type="inner" title="基本信息" style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item name="name" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
                  <Input />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="gender" label="性别" rules={[{ required: true, message: '请选择性别' }]}>
                  <Select placeholder="选择性别">
                    <Select.Option value="male">男</Select.Option>
                    <Select.Option value="female">女</Select.Option>
                    <Select.Option value="unknown">未知</Select.Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="age" label="年龄">
                  <InputNumber min={0} max={150} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item name="birth_date" label="出生日期">
                  <DatePicker style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="height" label="身高(cm)">
                  <InputNumber min={0} max={300} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="weight" label="体重(kg)">
                  <InputNumber min={0} max={500} style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>

            <Form.Item label="照片">
              <Space align="start">
                <Upload
                  accept="image/*"
                  showUploadList={false}
                  beforeUpload={handleUploadPhoto}
                >
                  <div style={{
                    width: 104, height: 104, border: '1px dashed #d9d9d9', borderRadius: 8,
                    display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer',
                    background: '#fafafa',
                  }}>
                    {uploading ? <Spin size="small" /> : <PlusOutlined style={{ fontSize: 20, color: '#999' }} />}
                  </div>
                </Upload>
                {photoUrl && <Image src={photoUrl} width={104} height={104} style={{ objectFit: 'cover', borderRadius: 8 }} />}
              </Space>
            </Form.Item>
            <Form.Item name="photo_url" hidden><Input /></Form.Item>
          </Card>

          <Card type="inner" title="走失信息" style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item name="missing_time" label="走失时间" rules={[{ required: true, message: '请选择走失时间' }]}>
                  <DatePicker showTime style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="urgency_level" label="紧急程度" initialValue="medium">
                  <Select placeholder="选择紧急程度">
                    <Select.Option value="low">普通</Select.Option>
                    <Select.Option value="medium">中等</Select.Option>
                    <Select.Option value="high">紧急</Select.Option>
                    <Select.Option value="urgent">特急</Select.Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={8}><Form.Item name="province" label="省份"><Input /></Form.Item></Col>
              <Col span={8}><Form.Item name="city" label="城市"><Input /></Form.Item></Col>
              <Col span={8}><Form.Item name="district" label="区县"><Input /></Form.Item></Col>
            </Row>
            <Form.Item name="address" label="详细地址"><Input placeholder="走失时的详细地址" /></Form.Item>
            <Form.Item name="clothes" label="衣着特征"><Input.TextArea rows={2} placeholder="走失时穿着什么衣服" /></Form.Item>
            <Form.Item name="features" label="体貌特征"><Input.TextArea rows={2} placeholder="如: 左脸有胎记、戴眼镜等" /></Form.Item>
            <Form.Item name="description" label="补充描述"><Input.TextArea rows={3} placeholder="其他需要补充的信息" /></Form.Item>
          </Card>

          <Card type="inner" title="联系信息" style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item name="contact_name" label="联系人" rules={[{ required: true, message: '请输入联系人' }]}>
                  <Input />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="contact_phone" label="联系电话" rules={[{ required: true, message: '请输入联系电话' }, PHONE_RULE]}>
                  <Input maxLength={11} />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="contact_rel" label="与走失者关系"><Input placeholder="如: 父亲、母亲" /></Form.Item>
              </Col>
            </Row>
            <Form.Item name="alt_contact" label="备用联系方式"><Input placeholder="其他联系方式（选填）" /></Form.Item>
          </Card>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={submitting}>
                {isEdit ? '保存' : '创建'}
              </Button>
              <Button onClick={() => navigate('/cases')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
