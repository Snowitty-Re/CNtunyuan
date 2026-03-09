import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Input, Select, Tag, Badge, Space, Popconfirm, message, Card, Row, Col } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { userApi } from '@/api/user'
import { usePermission } from '@/hooks/usePermission'
import type { User, UserQuery } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import { roleMap, userStatusMap as statusBadge } from '@/constants'
import dayjs from 'dayjs'

export default function UserList() {
  const [data, setData] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState<UserQuery>({ page: 1, page_size: 10 })
  const navigate = useNavigate()
  const { isAdmin, isManager } = usePermission()

  const fetchData = async () => {
    setLoading(true)
    try {
      const res = await userApi.list(query)
      setData(res.data.data.list || [])
      setTotal(res.data.data.total)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [query])

  const handleToggleStatus = async (user: User) => {
    const newStatus = user.status === 'active' ? 'inactive' : 'active'
    await userApi.updateStatus(user.id, newStatus)
    message.success('状态已更新')
    fetchData()
  }

  const handleDelete = async (id: string) => {
    await userApi.delete(id)
    message.success('删除成功')
    fetchData()
  }

  const columns: ColumnsType<User> = [
    { title: '昵称', dataIndex: 'nickname', width: 120 },
    { title: '手机号', dataIndex: 'phone', width: 130 },
    { title: '角色', dataIndex: 'role', width: 110, render: (v: string) => <Tag color={roleMap[v]?.color}>{roleMap[v]?.label || v}</Tag> },
    { title: '状态', dataIndex: 'status', width: 80, render: (v: string) => { const b = statusBadge[v]; return b ? <Badge status={b.status} text={b.text} /> : v } },
    { title: '所属组织', dataIndex: 'org_name', width: 120, render: (v: string) => v || '-' },
    { title: '最后登录', dataIndex: 'last_login_at', width: 150, render: (v: string) => v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-' },
    { title: '创建时间', dataIndex: 'created_at', width: 150, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
    {
      title: '操作', width: 200, fixed: 'right',
      render: (_: unknown, r: User) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/users/${r.id}/edit`)}>编辑</Button>
          {isManager && (
            <Popconfirm title={`确定${r.status === 'active' ? '禁用' : '启用'}？`} onConfirm={() => handleToggleStatus(r)}>
              <Button type="link" size="small">{r.status === 'active' ? '禁用' : '启用'}</Button>
            </Popconfirm>
          )}
          {isAdmin && (
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
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>用户管理</h2>
        {isAdmin && <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/users/new')}>新建用户</Button>}
      </div>

      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={[12, 12]}>
          <Col flex="auto">
            <Input.Search placeholder="搜索昵称、手机号" allowClear style={{ width: 240 }} onSearch={(v) => setQuery((q) => ({ ...q, keyword: v, page: 1 }))} />
          </Col>
          <Col>
            <Select placeholder="角色" allowClear style={{ width: 130 }} onChange={(v) => setQuery((q) => ({ ...q, role: v, page: 1 }))}>
              {Object.entries(roleMap).map(([k, v]) => <Select.Option key={k} value={k}>{v.label}</Select.Option>)}
            </Select>
          </Col>
          <Col>
            <Select placeholder="状态" allowClear style={{ width: 100 }} onChange={(v) => setQuery((q) => ({ ...q, status: v, page: 1 }))}>
              {Object.entries(statusBadge).map(([k, v]) => <Select.Option key={k} value={k}>{v.text}</Select.Option>)}
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
