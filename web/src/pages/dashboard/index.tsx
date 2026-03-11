import { useEffect, useState } from 'react';
import { Row, Col, Card, Table, Spin, Typography } from 'antd';
import {
  UserOutlined,
  TeamOutlined,
  OrderedListOutlined,
  AudioOutlined,
} from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatCard from '@/components/StatCard';
import { getDashboardStats, getTrend } from '@/api/dashboard';
import type { DashboardStats, TrendData } from '@/types';
import { formatDateTime } from '@/utils/format';

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [trend, setTrend] = useState<TrendData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getDashboardStats(), getTrend(7)])
      .then(([s, t]) => {
        setStats(s);
        setTrend(t);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div style={{ textAlign: 'center', paddingTop: 120 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <PageContainer title="工作台" breadcrumb={[{ title: '首页' }, { title: '工作台' }]}>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="用户总数"
            value={stats?.users.total || 0}
            prefix={<UserOutlined />}
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="走失人员"
            value={stats?.missing_persons.total || 0}
            prefix={<TeamOutlined />}
            color="#fa541c"
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="任务总数"
            value={stats?.tasks.total || 0}
            prefix={<OrderedListOutlined />}
            color="#722ed1"
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="方言录音"
            value={stats?.dialects.total || 0}
            prefix={<AudioOutlined />}
            color="#13c2c2"
          />
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <StatCard title="搜寻中" value={stats?.missing_persons.searching || 0} color="#faad14" />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard title="已团圆" value={stats?.missing_persons.reunited || 0} color="#52c41a" />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard title="待处理任务" value={stats?.tasks.pending || 0} color="#faad14" />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard title="已完成任务" value={stats?.tasks.completed || 0} color="#52c41a" />
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={14}>
          <Card title="近7日趋势">
            <Table
              dataSource={trend}
              rowKey="date"
              size="small"
              pagination={false}
              columns={[
                { title: '日期', dataIndex: 'date', key: 'date' },
                { title: '新增案件', dataIndex: 'new_cases', key: 'new_cases' },
                { title: '已解决', dataIndex: 'resolved_cases', key: 'resolved_cases' },
                { title: '新增任务', dataIndex: 'new_tasks', key: 'new_tasks' },
                { title: '完成任务', dataIndex: 'completed_tasks', key: 'completed_tasks' },
                { title: '新增用户', dataIndex: 'new_users', key: 'new_users' },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title="最近活动">
            {stats?.recent_activity && stats.recent_activity.length > 0 ? (
              <Table
                dataSource={stats.recent_activity}
                rowKey="id"
                size="small"
                pagination={false}
                columns={[
                  { title: '用户', dataIndex: 'user_name', key: 'user_name', width: 80 },
                  { title: '操作', dataIndex: 'title', key: 'title' },
                  {
                    title: '时间',
                    dataIndex: 'created_at',
                    key: 'created_at',
                    width: 150,
                    render: (v: string) => formatDateTime(v),
                  },
                ]}
              />
            ) : (
              <Typography.Text type="secondary">暂无最近活动</Typography.Text>
            )}
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
}
