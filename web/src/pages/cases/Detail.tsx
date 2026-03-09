import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Descriptions, Tag, Button, Space, Card, Timeline, Modal, Form, Input, DatePicker, message, Spin, Divider } from 'antd'
import { ArrowLeftOutlined, PlusOutlined } from '@ant-design/icons'
import { missingPersonApi } from '@/api/missingPerson'
import { usePermission } from '@/hooks/usePermission'
import type { MissingPerson, MissingPersonTrack } from '@/types'
import dayjs from 'dayjs'

const statusMap: Record<string, { label: string; color: string }> = {
  missing: { label: '走失', color: 'red' },
  searching: { label: '搜寻中', color: 'orange' },
  found: { label: '已找到', color: 'blue' },
  reunited: { label: '已团圆', color: 'green' },
  closed: { label: '已关闭', color: 'default' },
}

export default function CaseDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { isManager } = usePermission()
  const [data, setData] = useState<MissingPerson | null>(null)
  const [tracks, setTracks] = useState<MissingPersonTrack[]>([])
  const [loading, setLoading] = useState(true)
  const [foundModal, setFoundModal] = useState(false)
  const [trackModal, setTrackModal] = useState(false)
  const [foundForm] = Form.useForm()
  const [trackForm] = Form.useForm()

  const fetchData = async () => {
    if (!id) return
    setLoading(true)
    try {
      const [res, tracksRes] = await Promise.all([
        missingPersonApi.getById(id),
        missingPersonApi.getTracks(id),
      ])
      setData(res.data.data)
      setTracks(Array.isArray(tracksRes.data.data) ? tracksRes.data.data : [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [id])

  const handleMarkFound = async (values: { location: string; note: string }) => {
    if (!id) return
    await missingPersonApi.markFound(id, values)
    message.success('已标记为找到')
    setFoundModal(false)
    foundForm.resetFields()
    fetchData()
  }

  const handleMarkReunited = async () => {
    if (!id) return
    await missingPersonApi.markReunited(id)
    message.success('已标记为团圆')
    fetchData()
  }

  const handleAddTrack = async (values: { location: string; time: dayjs.Dayjs; description: string }) => {
    if (!id) return
    await missingPersonApi.addTrack(id, {
      location: values.location,
      time: values.time.toISOString(),
      description: values.description,
    })
    message.success('线索添加成功')
    setTrackModal(false)
    trackForm.resetFields()
    fetchData()
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />
  if (!data) return <div>未找到案件</div>

  const st = statusMap[data.status]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cases')}>返回</Button>
          <h2 style={{ margin: 0 }}>案件详情</h2>
        </Space>
        <Space>
          <Button onClick={() => navigate(`/cases/${id}/edit`)}>编辑</Button>
          {isManager && data.status === 'missing' && (
            <Button type="primary" onClick={() => setFoundModal(true)}>标记找到</Button>
          )}
          {isManager && data.status === 'found' && (
            <Button type="primary" style={{ background: '#52c41a' }} onClick={handleMarkReunited}>标记团圆</Button>
          )}
        </Space>
      </div>

      <Card>
        <Descriptions bordered column={{ xs: 1, sm: 2, lg: 3 }}>
          <Descriptions.Item label="案件编号">{data.case_no}</Descriptions.Item>
          <Descriptions.Item label="姓名">{data.name}</Descriptions.Item>
          <Descriptions.Item label="状态"><Tag color={st?.color}>{st?.label}</Tag></Descriptions.Item>
          <Descriptions.Item label="性别">{data.gender === 'male' ? '男' : data.gender === 'female' ? '女' : '未知'}</Descriptions.Item>
          <Descriptions.Item label="年龄">{data.age || '-'}</Descriptions.Item>
          <Descriptions.Item label="身高">{data.height ? `${data.height}cm` : '-'}</Descriptions.Item>
          <Descriptions.Item label="体重">{data.weight ? `${data.weight}kg` : '-'}</Descriptions.Item>
          <Descriptions.Item label="走失时间">{data.missing_time ? dayjs(data.missing_time).format('YYYY-MM-DD HH:mm') : '-'}</Descriptions.Item>
          <Descriptions.Item label="紧急程度"><Tag>{data.urgency}</Tag></Descriptions.Item>
          <Descriptions.Item label="省份">{data.province || '-'}</Descriptions.Item>
          <Descriptions.Item label="城市">{data.city || '-'}</Descriptions.Item>
          <Descriptions.Item label="区县">{data.district || '-'}</Descriptions.Item>
          <Descriptions.Item label="详细地址" span={3}>{data.address || '-'}</Descriptions.Item>
          <Descriptions.Item label="衣着特征" span={3}>{data.clothes || '-'}</Descriptions.Item>
          <Descriptions.Item label="体貌特征" span={3}>{data.features || '-'}</Descriptions.Item>
          <Descriptions.Item label="描述" span={3}>{data.description || '-'}</Descriptions.Item>
          <Descriptions.Item label="联系人">{data.contact_name}</Descriptions.Item>
          <Descriptions.Item label="联系电话">{data.contact_phone}</Descriptions.Item>
          <Descriptions.Item label="与走失者关系">{data.contact_rel || '-'}</Descriptions.Item>
          <Descriptions.Item label="浏览次数">{data.views}</Descriptions.Item>
          <Descriptions.Item label="报告人">{data.reporter?.nickname || '-'}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{dayjs(data.created_at).format('YYYY-MM-DD HH:mm')}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Divider />

      <Card title="线索轨迹" extra={<Button icon={<PlusOutlined />} onClick={() => setTrackModal(true)}>添加线索</Button>}>
        {tracks.length === 0 ? (
          <div style={{ textAlign: 'center', color: '#8f959e', padding: 24 }}>暂无线索</div>
        ) : (
          <Timeline
            items={tracks.map((t) => ({
              children: (
                <div>
                  <div style={{ fontWeight: 500 }}>{t.location}</div>
                  <div style={{ color: '#646a73' }}>{t.description}</div>
                  <div style={{ color: '#8f959e', fontSize: 12, marginTop: 4 }}>
                    {dayjs(t.time).format('YYYY-MM-DD HH:mm')} {t.reporter?.nickname ? `· ${t.reporter.nickname}` : ''}
                  </div>
                </div>
              ),
            }))}
          />
        )}
      </Card>

      <Modal title="标记找到" open={foundModal} onCancel={() => setFoundModal(false)} onOk={() => foundForm.submit()} okText="确认">
        <Form form={foundForm} onFinish={handleMarkFound} layout="vertical">
          <Form.Item name="location" label="发现地点" rules={[{ required: true, message: '请输入发现地点' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="note" label="备注">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title="添加线索" open={trackModal} onCancel={() => setTrackModal(false)} onOk={() => trackForm.submit()} okText="提交">
        <Form form={trackForm} onFinish={handleAddTrack} layout="vertical">
          <Form.Item name="location" label="地点" rules={[{ required: true, message: '请输入地点' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="time" label="时间" rules={[{ required: true, message: '请选择时间' }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="description" label="描述" rules={[{ required: true, message: '请输入描述' }]}>
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
