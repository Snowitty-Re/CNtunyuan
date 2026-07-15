'use client'

import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { App, Button, Card, Input, Progress, Select, Space, Table, Tag, Typography } from 'antd'
import Link from 'next/link'
import { useCallback, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { taskService } from '@/services/tasks'
import type { Task } from '@/types/api'

const STATUS_OPTIONS = [
  { value: 'draft', label: '草稿' },
  { value: 'pending', label: '待处理' },
  { value: 'assigned', label: '已分配' },
  { value: 'in_progress', label: '进行中' },
  { value: 'completed', label: '已完成' },
  { value: 'cancelled', label: '已取消' },
  { value: 'overdue', label: '已逾期' },
]

const statusColor: Record<string, string> = {
  draft: 'default',
  pending: 'warning',
  assigned: 'processing',
  in_progress: 'processing',
  completed: 'success',
  cancelled: 'default',
  overdue: 'error',
}

const priorityColor: Record<string, string> = {
  low: 'default',
  normal: 'blue',
  medium: 'blue',
  high: 'orange',
  urgent: 'red',
}

export default function TasksPage() {
  const { ready } = useAuthGuard()
  const { message } = App.useApp()
  const [loading, setLoading] = useState(true)
  const [items, setItems] = useState<Task[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<string | undefined>()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await taskService.list({
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
        status: status || undefined,
      })
      const normalized = listFrom<Task>(data)
      setItems(normalized.list)
      setTotal(normalized.total)
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, keyword, status, message])

  useEffect(() => {
    if (ready) load()
  }, [ready, load])

  if (!ready) {
    return (
      <AppShell>
        <Card loading />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <Space style={{ width: '100%', justifyContent: 'space-between' }} wrap>
          <div>
            <Typography.Title level={4} style={{ margin: 0 }}>
              任务中心
            </Typography.Title>
            <Typography.Text type="secondary">分配、执行与跟进审批</Typography.Text>
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => load()}>
              刷新
            </Button>
            <Link href="/tasks/create">
              <Button type="primary" icon={<PlusOutlined />}>
                新建任务
              </Button>
            </Link>
          </Space>
        </Space>

        <Card size="small">
          <Space wrap style={{ marginBottom: 12 }}>
            <Input.Search
              allowClear
              placeholder="任务标题关键词"
              style={{ width: 240 }}
              onSearch={(v) => {
                setPage(1)
                setKeyword(v.trim())
              }}
            />
            <Select
              allowClear
              placeholder="状态"
              style={{ width: 140 }}
              options={STATUS_OPTIONS}
              value={status}
              onChange={(v) => {
                setPage(1)
                setStatus(v)
              }}
            />
          </Space>

          <Table
            rowKey="id"
            size="small"
            loading={loading}
            dataSource={items}
            pagination={{
              current: page,
              pageSize,
              total,
              showSizeChanger: true,
              onChange: (p, ps) => {
                setPage(p)
                setPageSize(ps)
              },
            }}
            columns={[
              {
                title: '标题',
                dataIndex: 'title',
                render: (title: string, row) => <Link href={`/tasks/${row.id}`}>{title}</Link>,
              },
              {
                title: '状态',
                dataIndex: 'status',
                width: 110,
                render: (s: string) => <Tag color={statusColor[s] || 'default'}>{s || '-'}</Tag>,
              },
              {
                title: '优先级',
                dataIndex: 'priority',
                width: 100,
                render: (p: string) => <Tag color={priorityColor[p] || 'default'}>{p || 'normal'}</Tag>,
              },
              {
                title: '进度',
                dataIndex: 'progress',
                width: 140,
                render: (p?: number) => <Progress percent={Number(p || 0)} size="small" />,
              },
              {
                title: '负责人',
                width: 120,
                render: (_, row) => row.assignee?.nickname || row.assignee?.phone || '-',
              },
              {
                title: '截止',
                dataIndex: 'deadline',
                width: 170,
                render: (t: string) => fmtTime(t),
              },
              {
                title: '操作',
                width: 100,
                render: (_, row) => (
                  <Link href={`/tasks/${row.id}`}>
                    <Button type="link" size="small">
                      详情
                    </Button>
                  </Link>
                ),
              },
            ]}
          />
        </Card>
      </Space>
    </AppShell>
  )
}
