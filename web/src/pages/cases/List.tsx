import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Input, Select, Tag, Space, Popconfirm, message, Card, Row, Col } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { missingPersonApi } from '@/api/missingPerson'
import { usePermission } from '@/hooks/usePermission'
import type { MissingPerson, MissingPersonQuery } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const statusMap: Record<string, { label: string; color: string }> = {
  missing: { label: '走失', color: 'red' },
  searching: { label: '搜寻中', color: 'orange' },
  found: { label: '已找到', color: 'blue' },
  reunited: { label: '已团圆', color: 'green' },
  closed: { label: '已关闭', color: 'default' },
}

const urgencyMap: Record<string, { label: string; color: string }> = {
  low: { label: '普通', color: 'default' },
  medium: { label: '中等', color: 'blue' },
  high: { label: '紧急', color: 'orange' },
  urgent: { label: '特急', color: 'red' },
}

export default function CaseList() {
  const [data, setData] = useState<MissingPerson[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState<MissingPersonQuery>({ page: 1, page_size: 10 })
  const navigate = useNavigate()
  const { isManager } = usePermission()

  const fetchData = async () => {
    setLoading(true)
    try {
      const res = await missingPersonApi.list(query)
      setData(res.data.data.list || [])
      setTotal(res.data.data.total)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [query])

  const handleDelete = async (id: string) => {
    await missingPersonApi.delete(id)
    message.success('删除成功')
    fetchData()
  }

  const columns: ColumnsType<MissingPerson> = [
    { title: '案件编号', dataIndex: 'case_no', width: 140 },
    { title: '姓名', dataIndex: 'name', width: 100 },
    { title: '性别', dataIndex: 'gender', width: 60, render: (v: string) => v === 'male' ? '男' : v === 'female' ? '女' : '未知' },
    { title: '年龄', dataIndex: 'age', width: 60 },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: string) => <Tag color={statusMap[v]?.color}>{statusMap[v]?.label || v}</Tag> },
    { title: '紧急程度', dataIndex: 'urgency', width: 90, render: (v: string) => <Tag color={urgencyMap[v]?.color}>{urgencyMap[v]?.label || v}</Tag> },
    { title: '地区', width: 120, render: (_: unknown, r: MissingPerson) => `${r.province || ''}${r.city || ''}` },
    { title: '走失时间', dataIndex: 'missing_time', width: 120, render: (v: string) => v ? dayjs(v).format('YYYY-MM-DD') : '-' },
    {
      title: '操作', width: 200, fixed: 'right',
      render: (_: unknown, r: MissingPerson) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/cases/${r.id}`)}>查看</Button>
          <Button type="link" size="small" onClick={() => navigate(`/cases/${r.id}/edit`)}>编辑</Button>
          {isManager && (
            <Popconfirm title="确定删除？" onConfirm={() => handleDelete(r.id)}>
              <Button type="link" size="small" danger>删除</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>案件管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/cases/new')}>新建案件</Button>
      </div>

      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={[12, 12]}>
          <Col flex="auto">
            <Input.Search placeholder="搜索姓名、联系人" allowClear onSearch={(v) => setQuery((q) => ({ ...q, keyword: v, page: 1 }))} style={{ width: 240 }} />
          </Col>
          <Col>
            <Select placeholder="状态" allowClear style={{ width: 120 }} onChange={(v) => setQuery((q) => ({ ...q, status: v, page: 1 }))}>
              {Object.entries(statusMap).map(([k, v]) => <Select.Option key={k} value={k}>{v.label}</Select.Option>)}
            </Select>
          </Col>
          <Col>
            <Select placeholder="紧急程度" allowClear style={{ width: 120 }} onChange={(v) => setQuery((q) => ({ ...q, urgency_level: v, page: 1 }))}>
              {Object.entries(urgencyMap).map(([k, v]) => <Select.Option key={k} value={k}>{v.label}</Select.Option>)}
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
