import { useEffect, useState } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, message, Card } from 'antd'
import { PlusOutlined, PlayCircleOutlined, LikeOutlined, DeleteOutlined, SoundOutlined } from '@ant-design/icons'
import { dialectApi } from '../services/dialect'
import type { Dialect } from '../types'

const { TextArea } = Input

const Dialects = () => {
  const [data, setData] = useState<Dialect[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [playModalVisible, setPlayModalVisible] = useState(false)
  const [selectedDialect, setSelectedDialect] = useState<Dialect | null>(null)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [form] = Form.useForm()

  useEffect(() => {
    fetchData()
  }, [pagination.current, pagination.pageSize])

  const fetchData = async () => {
    setLoading(true)
    try {
      const result = await dialectApi.getList({
        page: pagination.current,
        page_size: pagination.pageSize,
      })
      setData(result.list)
      setPagination({ ...pagination, total: result.total })
    } finally {
      setLoading(false)
    }
  }

  const handlePlay = (record: Dialect) => {
    setSelectedDialect(record)
    setPlayModalVisible(true)
    dialectApi.play(record.id)
  }

  const handleLike = async (record: Dialect) => {
    await dialectApi.like(record.id)
    message.success('点赞成功')
    fetchData()
  }

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后无法恢复，是否确认？',
      onOk: async () => {
        await dialectApi.delete(id)
        message.success('删除成功')
        fetchData()
      },
    })
  }

  const handleSubmit = async (values: any) => {
    await dialectApi.create(values)
    message.success('创建成功')
    setModalVisible(false)
    fetchData()
  }

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const columns = [
    { title: '标题', dataIndex: 'title' },
    {
      title: '地区',
      render: (_: any, record: Dialect) => `${record.province} ${record.city} ${record.district}`,
    },
    {
      title: '时长',
      dataIndex: 'duration',
      render: (duration: number) => formatDuration(duration),
    },
    {
      title: '播放/点赞',
      render: (_: any, record: Dialect) => (
        <Space>
          <Tag icon={<PlayCircleOutlined />} color="blue">{record.play_count}</Tag>
          <Tag icon={<LikeOutlined />} color="pink">{record.like_count}</Tag>
        </Space>
      ),
    },
    {
      title: '采集人',
      dataIndex: ['collector', 'nickname'],
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '正常' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      render: (_: any, record: Dialect) => (
        <Space>
          <Button icon={<PlayCircleOutlined />} onClick={() => handlePlay(record)}>
            播放
          </Button>
          <Button icon={<LikeOutlined />} onClick={() => handleLike(record)}>
            点赞
          </Button>
          <Button danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setModalVisible(true) }}>
          新增方言录音
        </Button>
      </div>

      <Card style={{ marginBottom: 16 }}>
        <p>💡 方言录音要求：时长15-20秒，包含地区特征明显的语音内容</p>
      </Card>

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

      {/* 播放Modal */}
      <Modal
        title="播放方言"
        open={playModalVisible}
        onCancel={() => setPlayModalVisible(false)}
        footer={null}
      >
        {selectedDialect && (
          <div style={{ textAlign: 'center', padding: 24 }}>
            <SoundOutlined style={{ fontSize: 64, color: '#1890ff' }} />
            <h3 style={{ marginTop: 16 }}>{selectedDialect.title}</h3>
            <p>{selectedDialect.description}</p>
            <audio controls src={selectedDialect.audio_url} style={{ width: '100%', marginTop: 16 }} />
            <div style={{ marginTop: 16, color: '#888' }}>
              <span>📍 {selectedDialect.province} {selectedDialect.city}</span>
              <span style={{ marginLeft: 16 }}>⏱️ {formatDuration(selectedDialect.duration)}</span>
            </div>
          </div>
        )}
      </Modal>

      {/* 新增Modal */}
      <Modal
        title="新增方言录音"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true }]}>
            <Input placeholder="请输入方言标题" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <TextArea rows={3} placeholder="描述方言内容、背景等" />
          </Form.Item>
          <Form.Item name="audio_url" label="音频URL" rules={[{ required: true }]}>
            <Input placeholder="音频文件地址" />
          </Form.Item>
          <Form.Item name="duration" label="时长(秒)" rules={[{ required: true }]}>
            <Input type="number" min={15} max={20} placeholder="15-20秒" />
          </Form.Item>
          <Form.Item name="province" label="省">
            <Input />
          </Form.Item>
          <Form.Item name="city" label="市">
            <Input />
          </Form.Item>
          <Form.Item name="district" label="区">
            <Input />
          </Form.Item>
          <Form.Item name="address" label="详细地址">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Dialects
