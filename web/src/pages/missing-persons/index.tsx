import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Input, Select, Button, Space, Row, Col, Tag } from 'antd';
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import UrgencyTag from '@/components/UrgencyTag';
import { getMissingPersons } from '@/api/missingPerson';
import { usePagination } from '@/hooks/usePagination';
import { usePermission } from '@/hooks/usePermission';
import { formatDateTime } from '@/utils/format';
import { MISSING_PERSON_STATUS_MAP, URGENCY_MAP, GENDER_MAP } from '@/utils/constants';
import type { MissingPersonResponse } from '@/types';

export default function MissingPersonsPage() {
  const navigate = useNavigate();
  const { isManager } = usePermission();
  const { page, pageSize, pagination, setTotal, onChange } = usePagination();
  const [data, setData] = useState<MissingPersonResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState('');
  const [urgency, setUrgency] = useState('');
  const [gender, setGender] = useState('');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getMissingPersons({
        page, page_size: pageSize, keyword,
        status: status || undefined,
        urgency_level: urgency || undefined,
        gender: gender || undefined,
      });
      setData(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword, status, urgency, gender, setTotal]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const columns = [
    { title: '案件编号', dataIndex: 'case_no', key: 'case_no', width: 120 },
    { title: '姓名', dataIndex: 'name', key: 'name', width: 80 },
    {
      title: '性别', dataIndex: 'gender', key: 'gender', width: 60,
      render: (v: string) => GENDER_MAP[v] || v,
    },
    { title: '年龄', dataIndex: 'age', key: 'age', width: 60 },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 90,
      render: (v: string) => <StatusTag status={v} />,
    },
    {
      title: '紧急程度', dataIndex: 'urgency', key: 'urgency', width: 90,
      render: (v: string) => <UrgencyTag level={v} />,
    },
    {
      title: '走失地点', key: 'location', width: 150,
      render: (_: unknown, r: MissingPersonResponse) =>
        [r.province, r.city, r.district].filter(Boolean).join(' ') || '-',
    },
    {
      title: '走失时间', dataIndex: 'missing_time', key: 'missing_time', width: 150,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作', key: 'actions', width: 100,
      render: (_: unknown, r: MissingPersonResponse) => (
        <Button type="link" size="small" onClick={() => navigate(`/missing-persons/${r.id}`)}>详情</Button>
      ),
    },
  ];

  return (
    <PageContainer
      title="走失人员"
      breadcrumb={[{ title: '首页' }, { title: '走失人员' }]}
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/missing-persons/new')}>
          登记案件
        </Button>
      }
    >
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8} md={6}>
            <Input placeholder="搜索姓名/案件号" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => onChange(1, pageSize)} />
          </Col>
          <Col xs={12} sm={4}>
            <Select placeholder="状态" allowClear style={{ width: '100%' }} value={status || undefined} onChange={(v) => setStatus(v || '')}
              options={Object.entries(MISSING_PERSON_STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Col>
          <Col xs={12} sm={4}>
            <Select placeholder="紧急程度" allowClear style={{ width: '100%' }} value={urgency || undefined} onChange={(v) => setUrgency(v || '')}
              options={Object.entries(URGENCY_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Col>
          <Col xs={12} sm={4}>
            <Select placeholder="性别" allowClear style={{ width: '100%' }} value={gender || undefined} onChange={(v) => setGender(v || '')}
              options={Object.entries(GENDER_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Col>
          <Col>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => onChange(1, pageSize)}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setStatus(''); setUrgency(''); setGender(''); onChange(1, pageSize); }}>重置</Button>
            </Space>
          </Col>
        </Row>
      </Card>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={pagination} scroll={{ x: 1000 }} />
      </Card>
    </PageContainer>
  );
}
