'use client'

import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import { App, Button, Card, Input, Select, Space, Table, Tag, Typography } from 'antd'
import Link from 'next/link'
import { useCallback, useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, joinLocation, listFrom } from '@/lib/data'
import { missingPersonService } from '@/services/missingPersons'
import type { MissingPerson } from '@/types/api'

const STATUS_OPTIONS = [
  { value: 'missing', label: '走失中' },
  { value: 'searching', label: '寻访中' },
  { value: 'found', label: '已找到' },
  { value: 'reunited', label: '已团圆' },
  { value: 'closed', label: '已关闭' },
]

const statusColor: Record<string, string> = {
  missing: 'error',
  searching: 'processing',
  found: 'success',
  reunited: 'success',
  closed: 'default',
}

export default function CasesPage() {
  const { ready } = useAuthGuard()
  const { message } = App.useApp()
  const [loading, setLoading] = useState(true)
  const [items, setItems] = useState<MissingPerson[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<string | undefined>()

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await missingPersonService.list({
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
        status: status || undefined,
      })
      const normalized = listFrom<MissingPerson>(data)
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
              案件中心
            </Typography.Title>
            <Typography.Text type="secondary">走失人员登记、状态与线索跟进</Typography.Text>
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => load()}>
              刷新
            </Button>
            <Link href="/cases/create">
              <Button type="primary" icon={<PlusOutlined />}>
                登记案件
              </Button>
            </Link>
          </Space>
        </Space>

        <Card size="small">
          <Space wrap style={{ marginBottom: 12 }}>
            <Input.Search
              allowClear
              placeholder="姓名/联系人关键词"
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
                title: '姓名',
                dataIndex: 'name',
                render: (name: string, row) => <Link href={`/cases/${row.id}`}>{name}</Link>,
              },
              { title: '性别', dataIndex: 'gender', width: 80 },
              {
                title: '状态',
                dataIndex: 'status',
                width: 110,
                render: (s: string) => <Tag color={statusColor[s] || 'default'}>{s || '-'}</Tag>,
              },
              {
                title: '走失时间',
                dataIndex: 'missing_time',
                width: 170,
                render: (t: string) => fmtTime(t),
              },
              {
                title: '地点',
                render: (_, row) => joinLocation(row as unknown as Record<string, unknown>),
              },
              {
                title: '操作',
                width: 100,
                render: (_, row) => (
                  <Link href={`/cases/${row.id}`}>
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
