'use client'

import { ReloadOutlined } from '@ant-design/icons'
import { App, Button, Card, Col, Empty, List, Progress, Row, Segmented, Space, Statistic, Table, Typography } from 'antd'
import Link from 'next/link'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { fmtTime, listFrom } from '@/lib/data'
import { resolveWorkbench, workbenchLabel } from '@/lib/nav'
import { dashboardService } from '@/services/dashboard'
import { missingPersonService } from '@/services/missingPersons'
import { taskService } from '@/services/tasks'

type KPIs = {
  totalCases: number
  resolvedCases: number
  totalUsers: number
  totalTasks: number
}

type TrendRow = { date: string; tasks: number; cases: number; resolved: number }

export default function DashboardPage() {
  const { ready, user } = useAuthGuard()
  const { message } = App.useApp()
  const [loading, setLoading] = useState(true)
  const [kpi, setKpi] = useState<KPIs>({ totalCases: 0, resolvedCases: 0, totalUsers: 0, totalTasks: 0 })
  const [pendingTasks, setPendingTasks] = useState<Array<{ id: string; title: string; deadline?: string }>>([])
  const [activeCases, setActiveCases] = useState<Array<{ id: string; name: string; missing_time?: string }>>([])
  const [trendDays, setTrendDays] = useState(7)
  const [trendRows, setTrendRows] = useState<TrendRow[]>([])

  const workbench = resolveWorkbench(user)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [stats, overview, tasksData, casesData, trendData] = await Promise.all([
        dashboardService.stats(),
        dashboardService.overview(),
        taskService.list({ page: 1, page_size: 5, status: 'pending' }),
        missingPersonService.list({ page: 1, page_size: 5, status: 'searching' }),
        dashboardService.trend(trendDays),
      ])
      const totalCases = Number(stats?.missing_persons?.total || overview?.total_cases || 0)
      const resolvedCases = Number(
        (stats?.missing_persons?.found || 0) + (stats?.missing_persons?.reunited || 0) || overview?.resolved_cases || 0,
      )
      const totalUsers = Number(stats?.users?.total || overview?.total_users || 0)
      const totalTasks = Number(stats?.tasks?.total || overview?.total_tasks || 0)
      setKpi({ totalCases, resolvedCases, totalUsers, totalTasks })

      const t = listFrom<{ id: string; title?: string; deadline?: string; created_at?: string }>(tasksData).list
      const c = listFrom<{ id: string; name?: string; missing_time?: string; created_at?: string }>(casesData).list
      const trendParsed = listFrom<Record<string, unknown>>(trendData)
      const trendList =
        trendParsed.list.length > 0
          ? trendParsed.list
          : Array.isArray((trendData as { data?: unknown[] })?.data)
            ? ((trendData as { data: Record<string, unknown>[] }).data)
            : []

      setPendingTasks(t.map((x) => ({ id: x.id, title: x.title || '未命名任务', deadline: x.deadline || x.created_at })))
      setActiveCases(c.map((x) => ({ id: x.id, name: x.name || '未命名案件', missing_time: x.missing_time || x.created_at })))
      setTrendRows(
        trendList.map((r) => ({
          date: String(r.date || r.day || r.time || '-'),
          tasks: Number(r.tasks || r.task_count || 0),
          cases: Number(r.cases || r.case_count || r.missing_count || 0),
          resolved: Number(r.resolved || r.resolved_count || r.found_count || 0),
        })),
      )
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败')
      setPendingTasks([])
      setActiveCases([])
      setTrendRows([])
    } finally {
      setLoading(false)
    }
  }, [trendDays, message])

  useEffect(() => {
    if (ready) load()
  }, [ready, load])

  const closureRate = useMemo(() => {
    if (!kpi.totalCases) return 0
    return Math.min(100, Math.round((kpi.resolvedCases / kpi.totalCases) * 100))
  }, [kpi.resolvedCases, kpi.totalCases])

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
              {workbenchLabel(workbench)}
            </Typography.Title>
            <Typography.Text type="secondary">案件、任务与协作概览</Typography.Text>
          </div>
          <Button icon={<ReloadOutlined />} onClick={() => load()} loading={loading}>
            刷新
          </Button>
        </Space>

        <Row gutter={[16, 16]}>
          <Col xs={12} md={6}>
            <Card size="small" loading={loading}>
              <Statistic title="走失人员" value={kpi.totalCases} />
            </Card>
          </Col>
          <Col xs={12} md={6}>
            <Card size="small" loading={loading}>
              <Statistic title="已找到 / 团圆" value={kpi.resolvedCases} suffix={<Typography.Text type="secondary">闭环 {closureRate}%</Typography.Text>} />
            </Card>
          </Col>
          <Col xs={12} md={6}>
            <Card size="small" loading={loading}>
              <Statistic title="平台人员" value={kpi.totalUsers} />
            </Card>
          </Col>
          <Col xs={12} md={6}>
            <Card size="small" loading={loading}>
              <Statistic title="任务总数" value={kpi.totalTasks} />
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col xs={24} lg={14}>
            <Card
              size="small"
              title="趋势分析"
              extra={
                <Segmented
                  size="small"
                  value={trendDays}
                  onChange={(v) => setTrendDays(Number(v))}
                  options={[
                    { label: '近7天', value: 7 },
                    { label: '近30天', value: 30 },
                  ]}
                />
              }
              loading={loading}
            >
              {trendRows.length === 0 ? (
                <Empty description="暂无趋势数据" />
              ) : (
                <Table
                  size="small"
                  rowKey="date"
                  pagination={false}
                  dataSource={trendRows}
                  columns={[
                    { title: '日期', dataIndex: 'date' },
                    { title: '任务', dataIndex: 'tasks', width: 80 },
                    { title: '案件', dataIndex: 'cases', width: 80 },
                    { title: '闭环', dataIndex: 'resolved', width: 80 },
                  ]}
                />
              )}
            </Card>
          </Col>
          <Col xs={24} lg={10}>
            <Card size="small" title="协作热度" loading={loading}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <Typography.Text>案件闭环率</Typography.Text>
                  <Progress percent={closureRate} status="active" strokeColor="#d97706" />
                </div>
                <div>
                  <Typography.Text>任务规模</Typography.Text>
                  <Progress percent={Math.min(100, kpi.totalTasks * 6)} showInfo={false} strokeColor="#e76f51" />
                </div>
                <div>
                  <Typography.Text>人员参与</Typography.Text>
                  <Progress percent={Math.min(100, kpi.totalUsers * 8)} showInfo={false} strokeColor="#15803d" />
                </div>
              </Space>
            </Card>
          </Col>
        </Row>

        <Row gutter={[16, 16]}>
          <Col xs={24} md={12}>
            <Card
              size="small"
              title="待处理任务"
              extra={<Link href="/tasks">任务中心</Link>}
              loading={loading}
            >
              {pendingTasks.length === 0 ? (
                <Empty description="暂无待处理任务" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              ) : (
                <List
                  size="small"
                  dataSource={pendingTasks}
                  renderItem={(t) => (
                    <List.Item>
                      <Link href={`/tasks/${t.id}`}>{t.title}</Link>
                      <Typography.Text type="secondary">{fmtTime(t.deadline)}</Typography.Text>
                    </List.Item>
                  )}
                />
              )}
            </Card>
          </Col>
          <Col xs={24} md={12}>
            <Card
              size="small"
              title="进行中案件"
              extra={<Link href="/cases">案件中心</Link>}
              loading={loading}
            >
              {activeCases.length === 0 ? (
                <Empty description="暂无进行中案件" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              ) : (
                <List
                  size="small"
                  dataSource={activeCases}
                  renderItem={(c) => (
                    <List.Item>
                      <Link href={`/cases/${c.id}`}>{c.name}</Link>
                      <Typography.Text type="secondary">{fmtTime(c.missing_time)}</Typography.Text>
                    </List.Item>
                  )}
                />
              )}
            </Card>
          </Col>
        </Row>
      </Space>
    </AppShell>
  )
}
