import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Row, Col, Spin, Button, Table } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatCard from '@/components/StatCard';
import { getAuditStats, getModuleStats } from '@/api/audit';
import type { AuditLogStats, AuditModuleStat } from '@/types';

export default function AuditStatsPage() {
  const navigate = useNavigate();
  const [stats, setStats] = useState<AuditLogStats | null>(null);
  const [modules, setModules] = useState<AuditModuleStat[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getAuditStats(), getModuleStats()])
      .then(([s, m]) => {
        setStats(s);
        setModules(m || []);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;

  return (
    <PageContainer
      title="审计统计"
      breadcrumb={[{ title: '首页' }, { title: '审计日志' }, { title: '统计' }]}
      extra={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回</Button>}
    >
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={4.8}><StatCard title="总日志数" value={stats?.total_count || 0} /></Col>
        <Col xs={24} sm={12} lg={4.8}><StatCard title="今日数量" value={stats?.today_count || 0} color="#faad14" /></Col>
        <Col xs={24} sm={12} lg={4.8}><StatCard title="登录次数" value={stats?.login_count || 0} color="#52c41a" /></Col>
        <Col xs={24} sm={12} lg={4.8}><StatCard title="操作次数" value={stats?.operation_count || 0} color="#722ed1" /></Col>
        <Col xs={24} sm={12} lg={4.8}><StatCard title="错误次数" value={stats?.error_count || 0} color="#ff4d4f" /></Col>
      </Row>
      <Card title="模块统计" style={{ marginTop: 16 }}>
        <Table
          rowKey="module"
          dataSource={modules}
          pagination={false}
          columns={[
            { title: '模块', dataIndex: 'module', key: 'module' },
            { title: '操作次数', dataIndex: 'count', key: 'count', sorter: (a: AuditModuleStat, b: AuditModuleStat) => a.count - b.count },
          ]}
        />
      </Card>
    </PageContainer>
  );
}
