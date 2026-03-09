import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Tag, Badge, Space, Popconfirm, message, Card, Input } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { orgApi } from '@/api/organization'
import { usePermission } from '@/hooks/usePermission'
import type { Organization, OrgQuery } from '@/types'
import type { ColumnsType } from 'antd/es/table'

const typeMap: Record<string, { label: string; color: string }> = {
  root: { label: '总部', color: 'red' }, province: { label: '省级', color: 'orange' },
  city: { label: '市级', color: 'blue' }, district: { label: '区级', color: 'cyan' },
  street: { label: '街道', color: 'green' }, community: { label: '社区', color: 'lime' },
  team: { label: '小组', color: 'default' },
}

export default function OrganizationList() {
  const [data, setData] = useState<Organization[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState<OrgQuery>({ page: 1, page_size: 50 })
  const navigate = useNavigate()
  const { isAdmin } = usePermission()

  const fetchData = async () => {
    setLoading(true)
    try {
      const res = await orgApi.list(query)
      setData(res.data.data.list || [])
      setTotal(res.data.data.total)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [query])

  const handleDelete = async (id: string) => {
    await orgApi.delete(id)
    message.success('删除成功')
    fetchData()
  }

  const columns: ColumnsType<Organization> = [
    { title: '名称', dataIndex: 'name', width: 200 },
    { title: '编码', dataIndex: 'code', width: 120 },
    { title: '类型', dataIndex: 'type', width: 80, render: (v: string) => <Tag color={typeMap[v]?.color}>{typeMap[v]?.label || v}</Tag> },
    { title: '联系人', dataIndex: 'contact_name', width: 100 },
    { title: '联系电话', dataIndex: 'contact_phone', width: 130 },
    { title: '状态', dataIndex: 'status', width: 80, render: (v: string) => <Badge status={v === 'active' ? 'success' : 'default'} text={v === 'active' ? '正常' : '禁用'} /> },
    {
      title: '操作', width: 150, fixed: 'right',
      render: (_: unknown, r: Organization) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/organizations/${r.id}/edit`)}>编辑</Button>
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
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>组织管理</h2>
        {isAdmin && <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/organizations/new')}>新建组织</Button>}
      </div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Input.Search placeholder="搜索组织名称" allowClear style={{ width: 300 }} onSearch={(v) => setQuery((q) => ({ ...q, keyword: v, page: 1 }))} />
      </Card>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        scroll={{ x: 900 }}
        pagination={{
          current: query.page, pageSize: query.page_size, total,
          showSizeChanger: true, showTotal: (t) => `共 ${t} 条`,
          onChange: (page, pageSize) => setQuery((q) => ({ ...q, page, page_size: pageSize })),
        }}
      />
    </div>
  )
}
