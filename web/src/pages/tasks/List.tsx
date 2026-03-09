import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Table, Button, Input, Select, Tag, Space, Popconfirm, message, Card, Row, Col, Tabs, Progress } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { taskApi } from '@/api/task'
import { usePermission } from '@/hooks/usePermission'
import type { Task, TaskQuery } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const statusMap: Record<string, { label: string; color: string }> = {
  draft: { label: '草稿', color: 'default' },
  pending: { label: '待分配', color: 'gold' },
  assigned: { label: '已分配', color: 'cyan' },
  processing: { label: '进行中', color: 'blue' },
  completed: { label: '已完成', color: 'green' },
  cancelled: { label: '已取消', color: 'red' },
  overdue: { label: '已逾期', color: 'magenta' },
}

const typeMap: Record<string, string> = {
  search: '搜索', verify: '核实', assist: '协助', follow: '跟进', interview: '访谈', other: '其他',
}

const priorityMap: Record<string, { label: string; color: string }> = {
  low: { label: '低', color: 'default' },
  medium: { label: '中', color: 'blue' },
  high: { label: '高', color: 'orange' },
  urgent: { label: '紧急', color: 'red' },
}

export default function TaskList() {
  const [data, setData] = useState<Task[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState<TaskQuery>({ page: 1, page_size: 10 })
  const [activeTab, setActiveTab] = useState('all')
  const navigate = useNavigate()
  const { isManager } = usePermission()

  const fetchData = async () => {
    setLoading(true)
    try {
      const fetcher = activeTab === 'my' ? taskApi.myTasks : activeTab === 'pending' ? taskApi.pending : taskApi.list
      const res = await fetcher(query)
      setData(res.data.data.list || [])
      setTotal(res.data.data.total)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [query, activeTab])

  const handleDelete = async (id: string) => {
    await taskApi.delete(id)
    message.success('删除成功')
    fetchData()
  }

  const columns: ColumnsType<Task> = [
    { title: '标题', dataIndex: 'title', ellipsis: true },
    { title: '类型', dataIndex: 'type', width: 80, render: (v: string) => <Tag>{typeMap[v] || v}</Tag> },
    { title: '优先级', dataIndex: 'priority', width: 80, render: (v: string) => <Tag color={priorityMap[v]?.color}>{priorityMap[v]?.label || v}</Tag> },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: string) => <Tag color={statusMap[v]?.color}>{statusMap[v]?.label || v}</Tag> },
    { title: '进度', dataIndex: 'progress', width: 120, render: (v: number) => <Progress percent={v || 0} size="small" /> },
    { title: '截止时间', dataIndex: 'deadline', width: 120, render: (v: string) => v ? dayjs(v).format('YYYY-MM-DD') : '-' },
    { title: '执行人', width: 100, render: (_: unknown, r: Task) => r.assignee?.nickname || '-' },
    {
      title: '操作', width: 180, fixed: 'right',
      render: (_: unknown, r: Task) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/tasks/${r.id}`)}>查看</Button>
          <Button type="link" size="small" onClick={() => navigate(`/tasks/${r.id}/edit`)}>编辑</Button>
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
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>任务管理</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/tasks/new')}>新建任务</Button>
      </div>

      <Tabs
        activeKey={activeTab}
        onChange={(k) => { setActiveTab(k); setQuery((q) => ({ ...q, page: 1 })) }}
        items={[
          { key: 'all', label: '全部任务' },
          { key: 'my', label: '我的任务' },
          { key: 'pending', label: '待分配' },
        ]}
      />

      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={[12, 12]}>
          <Col flex="auto">
            <Input.Search placeholder="搜索任务" allowClear style={{ width: 240 }} onSearch={(v) => setQuery((q) => ({ ...q, keyword: v, page: 1 }))} />
          </Col>
          <Col>
            <Select placeholder="状态" allowClear style={{ width: 110 }} onChange={(v) => setQuery((q) => ({ ...q, status: v, page: 1 }))}>
              {Object.entries(statusMap).map(([k, v]) => <Select.Option key={k} value={k}>{v.label}</Select.Option>)}
            </Select>
          </Col>
          <Col>
            <Select placeholder="类型" allowClear style={{ width: 100 }} onChange={(v) => setQuery((q) => ({ ...q, type: v, page: 1 }))}>
              {Object.entries(typeMap).map(([k, v]) => <Select.Option key={k} value={k}>{v}</Select.Option>)}
            </Select>
          </Col>
          <Col>
            <Select placeholder="优先级" allowClear style={{ width: 100 }} onChange={(v) => setQuery((q) => ({ ...q, priority: v, page: 1 }))}>
              {Object.entries(priorityMap).map(([k, v]) => <Select.Option key={k} value={k}>{v.label}</Select.Option>)}
            </Select>
          </Col>
        </Row>
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
