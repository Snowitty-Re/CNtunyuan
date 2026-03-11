import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Input, Select, Button, Space, Row, Col, Tabs, Progress } from 'antd';
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import { getTasks, getMyTasks, getPendingTasks, getOverdueTasks } from '@/api/task';
import { usePagination } from '@/hooks/usePagination';
import { formatDateTime } from '@/utils/format';
import { TASK_STATUS_MAP, TASK_PRIORITY_MAP } from '@/utils/constants';
import type { TaskResponse } from '@/types';

type TabKey = 'all' | 'my' | 'pending' | 'overdue';

const fetchMap = { all: getTasks, my: getMyTasks, pending: getPendingTasks, overdue: getOverdueTasks };

export default function TasksPage() {
  const navigate = useNavigate();
  const { page, pageSize, pagination, setTotal, onChange } = usePagination();
  const [data, setData] = useState<TaskResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState('');
  const [priority, setPriority] = useState('');
  const [tab, setTab] = useState<TabKey>('all');

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const fn = fetchMap[tab];
      const res = await fn({ page, page_size: pageSize, keyword, status: status || undefined, priority: priority || undefined });
      setData(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword, status, priority, tab, setTotal]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const columns = [
    { title: '标题', dataIndex: 'title', key: 'title', ellipsis: true },
    { title: '类型', dataIndex: 'type', key: 'type', width: 80 },
    {
      title: '优先级', dataIndex: 'priority', key: 'priority', width: 80,
      render: (v: string) => <StatusTag status={v} map={TASK_PRIORITY_MAP} />,
    },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 90,
      render: (v: string) => <StatusTag status={v} />,
    },
    {
      title: '进度', dataIndex: 'progress', key: 'progress', width: 120,
      render: (v: number) => <Progress percent={v} size="small" />,
    },
    {
      title: '负责人', key: 'assignee', width: 80,
      render: (_: unknown, r: TaskResponse) => r.assignee?.nickname || '-',
    },
    {
      title: '截止日期', dataIndex: 'deadline', key: 'deadline', width: 150,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作', key: 'actions', width: 80,
      render: (_: unknown, r: TaskResponse) => (
        <Button type="link" size="small" onClick={() => navigate(`/tasks/${r.id}`)}>详情</Button>
      ),
    },
  ];

  return (
    <PageContainer
      title="任务管理"
      breadcrumb={[{ title: '首页' }, { title: '任务管理' }]}
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/tasks/new')}>新建任务</Button>
      }
    >
      <Card style={{ marginBottom: 16 }}>
        <Tabs
          activeKey={tab}
          onChange={(k) => { setTab(k as TabKey); onChange(1, pageSize); }}
          items={[
            { key: 'all', label: '全部' },
            { key: 'my', label: '我的任务' },
            { key: 'pending', label: '待处理' },
            { key: 'overdue', label: '已逾期' },
          ]}
        />
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}>
            <Input placeholder="搜索任务" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => onChange(1, pageSize)} />
          </Col>
          <Col xs={12} sm={4}>
            <Select placeholder="状态" allowClear style={{ width: '100%' }} value={status || undefined} onChange={(v) => setStatus(v || '')}
              options={Object.entries(TASK_STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Col>
          <Col xs={12} sm={4}>
            <Select placeholder="优先级" allowClear style={{ width: '100%' }} value={priority || undefined} onChange={(v) => setPriority(v || '')}
              options={Object.entries(TASK_PRIORITY_MAP).map(([k, v]) => ({ label: v, value: k }))} />
          </Col>
          <Col>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => onChange(1, pageSize)}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setStatus(''); setPriority(''); onChange(1, pageSize); }}>重置</Button>
            </Space>
          </Col>
        </Row>
      </Card>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={pagination} scroll={{ x: 900 }} />
      </Card>
    </PageContainer>
  );
}
