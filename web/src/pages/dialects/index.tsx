import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Input, Select, Button, Space, Row, Col, Tag } from 'antd';
import { PlusOutlined, SearchOutlined, ReloadOutlined, PlayCircleOutlined, HeartOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import { getDialects } from '@/api/dialect';
import { usePagination } from '@/hooks/usePagination';
import { formatDateTime, formatDuration } from '@/utils/format';
import { DIALECT_STATUS_MAP } from '@/utils/constants';
import type { DialectResponse } from '@/types';

export default function DialectsPage() {
  const navigate = useNavigate();
  const { page, pageSize, pagination, setTotal, onChange } = usePagination();
  const [data, setData] = useState<DialectResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState('');
  const [region, setRegion] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getDialects({ page, page_size: pageSize, keyword, status: status || undefined, region: region || undefined });
      setData(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword, status, region, setTotal]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const columns = [
    { title: '标题', dataIndex: 'title', key: 'title', ellipsis: true },
    { title: '地区', dataIndex: 'region', key: 'region', width: 100 },
    {
      title: '时长', dataIndex: 'duration', key: 'duration', width: 80,
      render: (v: number) => formatDuration(v),
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 90,
      render: (v: string) => <StatusTag status={v} />,
    },
    {
      title: '精选', dataIndex: 'is_featured', key: 'is_featured', width: 60,
      render: (v: boolean) => v ? <Tag color="gold">精选</Tag> : null,
    },
    {
      title: '播放', dataIndex: 'play_count', key: 'play_count', width: 70,
      render: (v: number) => <span><PlayCircleOutlined /> {v}</span>,
    },
    {
      title: '点赞', dataIndex: 'like_count', key: 'like_count', width: 70,
      render: (v: number) => <span><HeartOutlined /> {v}</span>,
    },
    {
      title: '上传者', key: 'uploader', width: 80,
      render: (_: unknown, r: DialectResponse) => r.uploader?.nickname || '-',
    },
    {
      title: '操作', key: 'actions', width: 80,
      render: (_: unknown, r: DialectResponse) => (
        <Button type="link" size="small" onClick={() => navigate(`/dialects/${r.id}`)}>详情</Button>
      ),
    },
  ];

  return (
    <PageContainer
      title="方言库"
      breadcrumb={[{ title: '首页' }, { title: '方言库' }]}
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/dialects/new')}>上传方言</Button>
      }
    >
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}><Input placeholder="搜索标题" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => onChange(1, pageSize)} /></Col>
          <Col xs={12} sm={4}><Select placeholder="状态" allowClear style={{ width: '100%' }} value={status || undefined} onChange={(v) => setStatus(v || '')} options={Object.entries(DIALECT_STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))} /></Col>
          <Col xs={12} sm={4}><Input placeholder="地区" value={region} onChange={(e) => setRegion(e.target.value)} /></Col>
          <Col><Space>
            <Button type="primary" icon={<SearchOutlined />} onClick={() => onChange(1, pageSize)}>搜索</Button>
            <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setStatus(''); setRegion(''); onChange(1, pageSize); }}>重置</Button>
          </Space></Col>
        </Row>
      </Card>
      <Card><Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={pagination} scroll={{ x: 900 }} /></Card>
    </PageContainer>
  );
}
