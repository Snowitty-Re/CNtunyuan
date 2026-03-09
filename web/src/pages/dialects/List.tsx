import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Input, Select, Tag, Space, Popconfirm, message, Card, Row, Col } from 'antd'
import { PlusOutlined, StarOutlined, StarFilled } from '@ant-design/icons'
import { dialectApi } from '@/api/dialect'
import { usePermission } from '@/hooks/usePermission'
import type { Dialect, DialectQuery } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const typeMap: Record<string, string> = {
  phrase: '短语', story: '故事', song: '歌曲', daily: '日常', other: '其他',
}

const statusMap: Record<string, { label: string; color: string }> = {
  active: { label: '已发布', color: 'green' },
  inactive: { label: '已下架', color: 'default' },
  pending: { label: '待审核', color: 'orange' },
}

export default function DialectList() {
  const [data, setData] = useState<Dialect[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState<DialectQuery>({ page: 1, page_size: 10 })
  const navigate = useNavigate()
  const { isAdmin, isManager } = usePermission()

  const fetchData = async () => {
    setLoading(true)
    try {
      const res = await dialectApi.list(query)
      setData(res.data.data.list || [])
      setTotal(res.data.data.total)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [query])

  const handleDelete = async (id: string) => {
    await dialectApi.delete(id)
    message.success('删除成功')
    fetchData()
  }

  const handleFeature = async (id: string, featured: boolean) => {
    if (featured) {
      await dialectApi.unfeature(id)
      message.success('已取消精选')
    } else {
      await dialectApi.feature(id)
      message.success('已设为精选')
    }
    fetchData()
  }

  const handleApprove = async (id: string) => {
    await dialectApi.updateStatus(id, 'active')
    message.success('审核通过')
    fetchData()
  }

  const columns: ColumnsType<Dialect> = [
    { title: '标题', dataIndex: 'title', ellipsis: true },
    { title: '类型', dataIndex: 'dialect_type', width: 80, render: (v: string) => <Tag>{typeMap[v] || v}</Tag> },
    { title: '地区', width: 120, render: (_: unknown, r: Dialect) => [r.province, r.city].filter(Boolean).join(' ') || r.region || '-' },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: string) => <Tag color={statusMap[v]?.color}>{statusMap[v]?.label || v}</Tag> },
    { title: '播放', dataIndex: 'play_count', width: 70 },
    { title: '点赞', dataIndex: 'like_count', width: 70 },
    { title: '精选', dataIndex: 'is_featured', width: 60, render: (v: boolean) => v ? <StarFilled style={{ color: '#faad14' }} /> : <StarOutlined style={{ color: '#d9d9d9' }} /> },
    { title: '创建时间', dataIndex: 'created_at', width: 150, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
    {
      title: '操作', width: 240, fixed: 'right',
      render: (_: unknown, r: Dialect) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/dialects/${r.id}/edit`)}>编辑</Button>
          {isAdmin && (
            <Button type="link" size="small" onClick={() => handleFeature(r.id, r.is_featured)}>
              {r.is_featured ? '取消精选' : '设为精选'}
            </Button>
          )}
          {isManager && r.status === 'pending' && (
            <Button type="link" size="small" style={{ color: '#52c41a' }} onClick={() => handleApprove(r.id)}>审核通过</Button>
          )}
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>方言管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/dialects/new')}>上传方言</Button>
      </div>

      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={[12, 12]}>
          <Col flex="auto">
            <Input.Search placeholder="搜索标题" allowClear style={{ width: 240 }} onSearch={(v) => setQuery((q) => ({ ...q, keyword: v, page: 1 }))} />
          </Col>
          <Col>
            <Select placeholder="类型" allowClear style={{ width: 100 }} onChange={(v) => setQuery((q) => ({ ...q, type: v, page: 1 }))}>
              {Object.entries(typeMap).map(([k, v]) => <Select.Option key={k} value={k}>{v}</Select.Option>)}
            </Select>
          </Col>
          <Col>
            <Select placeholder="状态" allowClear style={{ width: 110 }} onChange={(v) => setQuery((q) => ({ ...q, status: v, page: 1 }))}>
              {Object.entries(statusMap).map(([k, v]) => <Select.Option key={k} value={k}>{v.label}</Select.Option>)}
            </Select>
          </Col>
        </Row>
      </Card>

      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        scroll={{ x: 1000 }}
        pagination={{
          current: query.page, pageSize: query.page_size, total,
          showSizeChanger: true, showTotal: (t) => `共 ${t} 条`,
          onChange: (page, pageSize) => setQuery((q) => ({ ...q, page, page_size: pageSize })),
        }}
      />
    </div>
  )
}
