import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Input, Select, Button, Space, Row, Col, DatePicker } from 'antd';
import { SearchOutlined, ReloadOutlined, BarChartOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import { getAuditLogs } from '@/api/audit';
import { usePagination } from '@/hooks/usePagination';
import { formatDateTime } from '@/utils/format';
import { AUDIT_TYPE_MAP, AUDIT_STATUS_MAP } from '@/utils/constants';
import type { AuditLogResponse } from '@/types';

export default function AuditLogsPage() {
  const navigate = useNavigate();
  const { page, pageSize, pagination, setTotal, onChange } = usePagination(20);
  const [data, setData] = useState<AuditLogResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [module, setModule] = useState('');
  const [type, setType] = useState('');
  const [status, setStatus] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getAuditLogs({
        page, page_size: pageSize, keyword,
        module: module || undefined,
        type: type || undefined,
        status: status || undefined,
      });
      setData(res.list || (res as any).items || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword, module, type, status, setTotal]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const columns = [
    { title: '用户', dataIndex: 'username', key: 'username', width: 80 },
    { title: '模块', dataIndex: 'module', key: 'module', width: 100 },
    { title: '操作', dataIndex: 'action', key: 'action', width: 80 },
    {
      title: '类型', dataIndex: 'type', key: 'type', width: 80,
      render: (v: string) => AUDIT_TYPE_MAP[v] || v,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 80,
      render: (v: string) => <StatusTag status={v} />,
    },
    { title: 'IP', dataIndex: 'request_ip', key: 'request_ip', width: 120 },
    {
      title: '耗时', dataIndex: 'duration_ms', key: 'duration_ms', width: 80,
      render: (v: number) => v ? `${v}ms` : '-',
    },
    {
      title: '时间', dataIndex: 'created_at', key: 'created_at', width: 160,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作', key: 'actions', width: 60,
      render: (_: unknown, r: AuditLogResponse) => (
        <Button type="link" size="small" onClick={() => navigate(`/audit/${r.id}`)}>详情</Button>
      ),
    },
  ];

  return (
    <PageContainer
      title="审计日志"
      breadcrumb={[{ title: '首页' }, { title: '审计日志' }]}
      extra={
        <Button icon={<BarChartOutlined />} onClick={() => navigate('/audit/stats')}>统计</Button>
      }
    >
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={6}><Input placeholder="关键词" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => onChange(1, pageSize)} /></Col>
          <Col xs={12} sm={4}><Input placeholder="模块" value={module} onChange={(e) => setModule(e.target.value)} /></Col>
          <Col xs={12} sm={4}>
            <Select placeholder="类型" allowClear style={{ width: '100%' }} value={type || undefined} onChange={(v) => setType(v || '')}
              options={Object.entries(AUDIT_TYPE_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Col>
          <Col xs={12} sm={4}>
            <Select placeholder="状态" allowClear style={{ width: '100%' }} value={status || undefined} onChange={(v) => setStatus(v || '')}
              options={Object.entries(AUDIT_STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Col>
          <Col><Space>
            <Button type="primary" icon={<SearchOutlined />} onClick={() => onChange(1, pageSize)}>搜索</Button>
            <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setModule(''); setType(''); setStatus(''); onChange(1, pageSize); }}>重置</Button>
          </Space></Col>
        </Row>
      </Card>
      <Card><Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={pagination} scroll={{ x: 1000 }} /></Card>
    </PageContainer>
  );
}
