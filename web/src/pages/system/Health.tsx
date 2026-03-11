import { useEffect, useState } from 'react';
import { Card, Row, Col, Spin, Tag, Descriptions, Button } from 'antd';
import { ReloadOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import request from '@/api/request';

interface HealthData {
  status: string;
  timestamp?: string;
  uptime?: string;
  database?: { status: string };
  redis?: { status: string };
  [key: string]: unknown;
}

export default function SystemHealthPage() {
  const [health, setHealth] = useState<HealthData | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchHealth = () => {
    setLoading(true);
    fetch('/api/v1/../health/detailed')
      .then((r) => r.json())
      .then((data) => setHealth(data?.data || data))
      .catch(() => setHealth({ status: 'error' }))
      .finally(() => setLoading(false));
  };

  useEffect(() => { fetchHealth(); }, []);

  const isUp = health?.status === 'ok' || health?.status === 'healthy';

  return (
    <PageContainer
      title="系统监控"
      breadcrumb={[{ title: '首页' }, { title: '系统监控' }]}
      extra={<Button icon={<ReloadOutlined />} onClick={fetchHealth} loading={loading}>刷新</Button>}
    >
      {loading ? (
        <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>
      ) : (
        <>
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={8}>
              <Card>
                <div style={{ textAlign: 'center' }}>
                  {isUp ? (
                    <CheckCircleOutlined style={{ fontSize: 48, color: '#52c41a' }} />
                  ) : (
                    <CloseCircleOutlined style={{ fontSize: 48, color: '#ff4d4f' }} />
                  )}
                  <div style={{ marginTop: 12 }}>
                    <Tag color={isUp ? 'green' : 'red'} style={{ fontSize: 16, padding: '4px 16px' }}>
                      {isUp ? '系统正常' : '系统异常'}
                    </Tag>
                  </div>
                </div>
              </Card>
            </Col>
            <Col xs={24} sm={16}>
              <Card title="系统信息">
                <Descriptions bordered column={1} size="small">
                  <Descriptions.Item label="状态">{health?.status || '-'}</Descriptions.Item>
                  {health?.timestamp && <Descriptions.Item label="时间">{health.timestamp}</Descriptions.Item>}
                  {health?.uptime && <Descriptions.Item label="运行时间">{String(health.uptime)}</Descriptions.Item>}
                  {health?.database && (
                    <Descriptions.Item label="数据库">
                      <Tag color={health.database.status === 'connected' ? 'green' : 'red'}>
                        {health.database.status}
                      </Tag>
                    </Descriptions.Item>
                  )}
                  {health?.redis && (
                    <Descriptions.Item label="Redis">
                      <Tag color={health.redis.status === 'connected' ? 'green' : 'red'}>
                        {health.redis.status}
                      </Tag>
                    </Descriptions.Item>
                  )}
                </Descriptions>
              </Card>
            </Col>
          </Row>
        </>
      )}
    </PageContainer>
  );
}
