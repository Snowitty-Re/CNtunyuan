import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Row, Col, Card, Statistic, Button, Space, Spin, Table, Tag, Empty } from 'antd'
import {
  AlertOutlined, TeamOutlined, ScheduleOutlined, AudioOutlined,
  PlusOutlined, SearchOutlined, CheckCircleOutlined, ClockCircleOutlined,
  RiseOutlined, FileOutlined,
} from '@ant-design/icons'
import { dashboardApi } from '@/api/dashboard'
import { taskApi } from '@/api/task'
import { usePermission } from '@/hooks/usePermission'
import { taskStatusMap, taskPriorityMap } from '@/constants'
import type { DashboardStats, TrendData, Task } from '@/types'
import dayjs from 'dayjs'

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [trend, setTrend] = useState<TrendData[]>([])
  const [myTasks, setMyTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()
  const { isManager, user } = usePermission()

  useEffect(() => {
    Promise.all([
      dashboardApi.getStats().then((res) => setStats(res.data.data)),
      dashboardApi.getTrend(7).then((res) => setTrend(res.data.data || [])).catch(() => {}),
      taskApi.myTasks({ page: 1, page_size: 5 }).then((res) => setMyTasks(res.data.data?.list || [])).catch(() => {}),
    ]).finally(() => setLoading(false))
  }, [])

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  const s = stats

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>
          {user?.nickname ? `${user.nickname}，你好` : '工作台'}
        </h2>
        <span style={{ color: '#8f959e', fontSize: 13 }}>{dayjs().format('YYYY年MM月DD日 dddd')}</span>
      </div>

      {/* Core Stats */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable onClick={() => navigate('/cases')} style={{ cursor: 'pointer' }}>
            <Statistic title="案件总数" value={s?.missing_persons.total ?? 0} prefix={<AlertOutlined style={{ color: '#e67e22' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              走失 {s?.missing_persons.missing ?? 0} / 搜寻中 {s?.missing_persons.searching ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable onClick={() => navigate('/cases')} style={{ cursor: 'pointer' }}>
            <Statistic title="已团圆" value={(s?.missing_persons.found ?? 0) + (s?.missing_persons.reunited ?? 0)} prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />} valueStyle={{ color: '#52c41a' }} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              已找到 {s?.missing_persons.found ?? 0} / 已团圆 {s?.missing_persons.reunited ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable onClick={() => navigate('/tasks')} style={{ cursor: 'pointer' }}>
            <Statistic title="任务总数" value={s?.tasks.total ?? 0} prefix={<ScheduleOutlined style={{ color: '#1890ff' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              待处理 {s?.tasks.pending ?? 0} / 进行中 {s?.tasks.processing ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable onClick={() => navigate('/users')} style={{ cursor: 'pointer' }}>
            <Statistic title="志愿者" value={s?.users.total ?? 0} prefix={<TeamOutlined style={{ color: '#722ed1' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              今日新增 {s?.users.new_today ?? 0} / 本周 {s?.users.new_week ?? 0}
            </div>
          </Card>
        </Col>
      </Row>

      {/* Secondary Stats */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="方言样本" value={s?.dialects.total ?? 0} prefix={<AudioOutlined style={{ color: '#fa8c16' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              精选 {s?.dialects.featured ?? 0} / 总播放 {s?.dialects.plays ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="任务完成" value={s?.tasks.completed ?? 0} prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />} valueStyle={{ color: '#52c41a' }} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              逾期 <span style={{ color: s?.tasks.overdue ? '#ff4d4f' : undefined }}>{s?.tasks.overdue ?? 0}</span>
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="本周新增案件" value={s?.missing_persons.new_week ?? 0} prefix={<RiseOutlined style={{ color: '#e67e22' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              今日 {s?.missing_persons.new_today ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="文件" value={s?.files.total_count ?? 0} prefix={<FileOutlined style={{ color: '#8f959e' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              组织 {s?.organizations.total ?? 0} 个
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 24 }}>
        {/* My Tasks */}
        <Col xs={24} lg={14}>
          <Card title="我的任务" extra={<Button type="link" onClick={() => navigate('/tasks')}>查看全部</Button>}>
            {myTasks.length === 0 ? (
              <Empty description="暂无任务" />
            ) : (
              <Table
                rowKey="id"
                dataSource={myTasks}
                size="small"
                pagination={false}
                columns={[
                  { title: '标题', dataIndex: 'title', ellipsis: true, render: (v: string, r: Task) => (
                    <Button type="link" style={{ padding: 0 }} onClick={() => navigate(`/tasks/${r.id}`)}>{v}</Button>
                  )},
                  { title: '优先级', dataIndex: 'priority', width: 70, render: (v: string) => {
                    const p = taskPriorityMap[v]; return p ? <Tag color={p.color}>{p.label}</Tag> : v
                  }},
                  { title: '状态', dataIndex: 'status', width: 80, render: (v: string) => {
                    const st = taskStatusMap[v]; return st ? <Tag color={st.color}>{st.label}</Tag> : v
                  }},
                  { title: '截止', dataIndex: 'deadline', width: 100, render: (v: string) => v ? dayjs(v).format('MM-DD HH:mm') : '-' },
                ]}
              />
            )}
          </Card>
        </Col>

        {/* Quick Actions + Trend */}
        <Col xs={24} lg={10}>
          <Card title="快捷操作">
            <Space wrap>
              <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/cases/new')}>新建案件</Button>
              <Button icon={<PlusOutlined />} onClick={() => navigate('/tasks/new')}>新建任务</Button>
              <Button icon={<AlertOutlined />} onClick={() => navigate('/cases')}>查看案件</Button>
              <Button icon={<ScheduleOutlined />} onClick={() => navigate('/tasks')}>查看任务</Button>
              {isManager && <Button icon={<TeamOutlined />} onClick={() => navigate('/users')}>管理用户</Button>}
            </Space>
          </Card>

          {trend.length > 0 && (
            <Card title="近7日趋势" style={{ marginTop: 16 }}>
              <div style={{ fontSize: 13 }}>
                {trend.map((t) => (
                  <div key={t.date} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', borderBottom: '1px solid #f0f0f0' }}>
                    <span style={{ color: '#646a73' }}>{t.date}</span>
                    <Space size={16}>
                      <span>案件 <b style={{ color: '#e67e22' }}>+{t.new_cases}</b></span>
                      <span>团圆 <b style={{ color: '#52c41a' }}>+{t.resolved_cases}</b></span>
                      <span>任务 <b style={{ color: '#1890ff' }}>+{t.new_tasks}</b></span>
                    </Space>
                  </div>
                ))}
              </div>
            </Card>
          )}
        </Col>
      </Row>
    </div>
  )
}
