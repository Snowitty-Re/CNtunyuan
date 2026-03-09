import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Row, Col, Card, Statistic, Button, Space, Spin } from 'antd'
import {
  AlertOutlined, TeamOutlined, ScheduleOutlined, AudioOutlined,
  PlusOutlined, SearchOutlined,
} from '@ant-design/icons'
import { dashboardApi } from '@/api/dashboard'
import type { DashboardStats } from '@/types'

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    dashboardApi.getStats().then((res) => setStats(res.data.data)).finally(() => setLoading(false))
  }, [])

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  const s = stats
  return (
    <div>
      <h2 style={{ marginBottom: 24, fontSize: 20, fontWeight: 600 }}>工作台</h2>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable>
            <Statistic title="案件总数" value={s?.missing_persons.total ?? 0} prefix={<AlertOutlined style={{ color: '#e67e22' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              走失 {s?.missing_persons.missing ?? 0} / 搜寻中 {s?.missing_persons.searching ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable>
            <Statistic title="已团圆" value={(s?.missing_persons.found ?? 0) + (s?.missing_persons.reunited ?? 0)} prefix={<SearchOutlined style={{ color: '#52c41a' }} />} valueStyle={{ color: '#52c41a' }} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              已找到 {s?.missing_persons.found ?? 0} / 已团圆 {s?.missing_persons.reunited ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable>
            <Statistic title="任务" value={s?.tasks.total ?? 0} prefix={<ScheduleOutlined style={{ color: '#1890ff' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              待处理 {s?.tasks.pending ?? 0} / 进行中 {s?.tasks.processing ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable>
            <Statistic title="用户" value={s?.users.total ?? 0} prefix={<TeamOutlined style={{ color: '#722ed1' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              今日新增 {s?.users.new_today ?? 0}
            </div>
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable>
            <Statistic title="方言样本" value={s?.dialects.total ?? 0} prefix={<AudioOutlined style={{ color: '#fa8c16' }} />} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              精选 {s?.dialects.featured ?? 0} / 播放 {s?.dialects.plays ?? 0}
            </div>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card hoverable>
            <Statistic title="任务完成" value={s?.tasks.completed ?? 0} valueStyle={{ color: '#52c41a' }} />
            <div style={{ marginTop: 8, fontSize: 13, color: '#8f959e' }}>
              逾期 {s?.tasks.overdue ?? 0}
            </div>
          </Card>
        </Col>
      </Row>

      <Card style={{ marginTop: 24 }} title="快捷操作">
        <Space wrap>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/cases/new')}>新建案件</Button>
          <Button icon={<PlusOutlined />} onClick={() => navigate('/tasks/new')}>新建任务</Button>
          <Button icon={<AlertOutlined />} onClick={() => navigate('/cases')}>查看案件</Button>
          <Button icon={<ScheduleOutlined />} onClick={() => navigate('/tasks')}>查看任务</Button>
        </Space>
      </Card>
    </div>
  )
}
