import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Spin, Typography } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import { getAuditLog } from '@/api/audit';
import { formatDateTime } from '@/utils/format';
import { AUDIT_TYPE_MAP } from '@/utils/constants';
import type { AuditLogResponse } from '@/types';

export default function AuditDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [log, setLog] = useState<AuditLogResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    getAuditLog(id).then(setLog).finally(() => setLoading(false));
  }, [id]);

  if (loading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;
  if (!log) return <div>日志不存在</div>;

  return (
    <PageContainer
      title="审计日志详情"
      breadcrumb={[{ title: '首页' }, { title: '审计日志' }, { title: '详情' }]}
      extra={<Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回</Button>}
    >
      <Card>
        <Descriptions bordered column={{ xs: 1, sm: 2 }}>
          <Descriptions.Item label="用户">{log.username}</Descriptions.Item>
          <Descriptions.Item label="角色">{log.user_role}</Descriptions.Item>
          <Descriptions.Item label="模块">{log.module}</Descriptions.Item>
          <Descriptions.Item label="操作">{log.action}</Descriptions.Item>
          <Descriptions.Item label="类型">{AUDIT_TYPE_MAP[log.type] || log.type}</Descriptions.Item>
          <Descriptions.Item label="状态"><StatusTag status={log.status} /></Descriptions.Item>
          <Descriptions.Item label="请求方法">{log.request_method}</Descriptions.Item>
          <Descriptions.Item label="请求URL">{log.request_url}</Descriptions.Item>
          <Descriptions.Item label="IP">{log.request_ip}</Descriptions.Item>
          <Descriptions.Item label="耗时">{log.duration_ms ? `${log.duration_ms}ms` : '-'}</Descriptions.Item>
          <Descriptions.Item label="响应码">{log.response_code}</Descriptions.Item>
          <Descriptions.Item label="Trace ID">{log.trace_id || '-'}</Descriptions.Item>
          <Descriptions.Item label="时间" span={2}>{formatDateTime(log.created_at)}</Descriptions.Item>
          <Descriptions.Item label="描述" span={2}>{log.description || '-'}</Descriptions.Item>
          {log.error_message && (
            <Descriptions.Item label="错误信息" span={2}>
              <Typography.Text type="danger">{log.error_message}</Typography.Text>
            </Descriptions.Item>
          )}
        </Descriptions>
      </Card>
    </PageContainer>
  );
}
