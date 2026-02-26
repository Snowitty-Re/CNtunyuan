import { useEffect, useState } from 'react'
import { Row, Col, Card, Statistic, Table, Tag } from 'antd'
import {
  TeamOutlined,
  SearchOutlined,
  CheckCircleOutlined,
  SoundOutlined,
} from '@ant-design/icons'
import { Pie, Column } from '@ant-design/charts'
import { userApi } from '../services/user'
import { missingPersonApi } from '../services/missing_person'
import { dialectApi } from '../services/dialect'

const Dashboard = () => {
  const [stats, setStats] = useState({
    users: { total: 0, volunteers: 0 },
    cases: { total: 0, resolved: 0, pending: 0 },
    dialects: { total: 0 },
    tasks: { total: 0, completed: 0 },
  })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    setLoading(true)
    try {
      const [userStats, caseStats, dialectStats] = await Promise.all([
        userApi.getStatistics(),
        missingPersonApi.getStatistics(),
        dialectApi.getStatistics(),
      ])
      setStats({
        users: userStats,
        cases: caseStats,
        dialects: dialectStats,
        tasks: { total: 0, completed: 0 },
      })
    } finally {
      setLoading(false)
    }
  }

  const caseStatusData = [
    { type: '已解决', value: stats.cases.resolved || 0 },
    { type: '寻找中', value: stats.cases.pending || 0 },
    { type: '已结案', value: (stats.cases.total || 0) - (stats.cases.resolved || 0) - (stats.cases.pending || 0) },
  ]

  const pieConfig = {
    data: caseStatusData,
    angleField: 'value',
    colorField: 'type',
    radius: 0.8,
    label: {
      type: 'outer',
      content: '{name} {percentage}',
    },
  }

  const monthlyData = [
    { month: '1月', cases: 12, resolved: 8 },
    { month: '2月', cases: 15, resolved: 10 },
    { month: '3月', cases: 18, resolved: 12 },
    { month: '4月', cases: 14, resolved: 11 },
    { month: '5月', cases: 20, resolved: 15 },
    { month: '6月', cases: 22, resolved: 18 },
  ]

  const columnConfig = {
    data: monthlyData,
    xField: 'month',
    yField: 'cases',
    seriesField: 'type',
    isGroup: true,
    columnStyle: {
      radius: [4, 4, 0, 0],
    },
  }

  const recentCases = [
    { id: 1, name: '张三', age: 65, status: 'searching', location: '北京市朝阳区', time: '2024-02-26' },
    { id: 2, name: '李四', age: 8, status: 'found', location: '上海市浦东新区', time: '2024-02-25' },
    { id: 3, name: '王五', age: 72, status: 'resolved', location: '广州市天河区', time: '2024-02-24' },
  ]

  const columns = [
    { title: '姓名', dataIndex: 'name' },
    { title: '年龄', dataIndex: 'age' },
    { title: '走失地点', dataIndex: 'location' },
    { title: '时间', dataIndex: 'time' },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => {
        const statusMap: Record<string, { color: string; text: string }> = {
          searching: { color: 'processing', text: '寻找中' },
          found: { color: 'success', text: '已找到' },
          resolved: { color: 'default', text: '已解决' },
        }
        const { color, text } = statusMap[status] || { color: 'default', text: status }
        return <Tag color={color}>{text}</Tag>
      },
    },
  ]

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic
              title="志愿者总数"
              value={stats.users.total}
              prefix={<TeamOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic
              title="案件总数"
              value={stats.cases.total}
              prefix={<SearchOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic
              title="已解决案件"
              value={stats.cases.resolved}
              valueStyle={{ color: '#3f8600' }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card loading={loading}>
            <Statistic
              title="方言录音"
              value={stats.dialects.total}
              prefix={<SoundOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="案件状态分布">
            <Pie {...pieConfig} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="月度案件趋势">
            <Column {...columnConfig} />
          </Card>
        </Col>
      </Row>

      <Card title="最近案件" style={{ marginTop: 16 }}>
        <Table columns={columns} dataSource={recentCases} rowKey="id" pagination={false} />
      </Card>
    </div>
  )
}

export default Dashboard
