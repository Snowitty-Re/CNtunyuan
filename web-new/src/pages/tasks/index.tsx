import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Button, Tag, Dropdown, Modal, Progress, message, Select, Space } from 'antd';
import { PlusOutlined, MoreOutlined, EyeOutlined, EditOutlined, DeleteOutlined, CheckCircleOutlined } from '@ant-design/icons';
import type { Task } from '@/types';
import { http } from '@/utils/request';
import { usePermission } from '@/utils/permission';
import dayjs from 'dayjs';

export default function TasksPage() {
  const navigate = useNavigate();
  const { isManager } = usePermission();
  const [data, setData] = useState<Task[]>([]);
  const [loading, setLoading] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [priorityFilter, setPriorityFilter] = useState<string>('');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  useEffect(() => {
    fetchData();
  }, [pagination.current, pagination.pageSize, statusFilter, priorityFilter]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res: any = await http.get('/tasks', {
        params: {
          page: pagination.current,
          page_size: pagination.pageSize,
          status: statusFilter || undefined,
          priority: priorityFilter || undefined,
        },
      });
      setData(res.list || []);
      setPagination((prev) => ({ ...prev, total: res.total || 0 }));
    } catch (error) {
      console.error('获取任务列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleComplete = async (record: Task) => {
    Modal.confirm({
      title: '确认完成任务',
      content: `确定要将任务 "${record.title}" 标记为完成吗？`,
      okText: '完成',
      okType: 'primary',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.post(`/tasks/${record.id}/complete`);
          message.success('任务已完成');
          fetchData();
        } catch (error) {
          message.error('操作失败');
        }
      },
    });
  };

  const handleDelete = (record: Task) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除任务 "${record.title}" 吗？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.delete(`/tasks/${record.id}`);
          message.success('删除成功');
          fetchData();
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  const columns = [
    {
      title: '任务标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: Task) => (
        <div>
          <div style={{ fontWeight: 500, color: '#1f2329' }}>{text}</div>
          <div style={{ color: '#8f959e', fontSize: 12 }}>{record.task_no}</div>
        </div>
      ),
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
      width: 90,
      render: (priority: string) => (
        <Tag color={getPriorityColor(priority)} style={{ fontSize: 13 }}>
          {getPriorityLabel(priority)}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={getStatusColor(status)} style={{ fontSize: 13 }}>
          {getStatusLabel(status)}
        </Tag>
      ),
    },
    {
      title: '执行人',
      dataIndex: ['assignee', 'nickname'],
      key: 'assignee',
      width: 120,
      render: (nickname: string) => <span style={{ color: '#646a73' }}>{nickname || '未分配'}</span>,
    },
    {
      title: '进度',
      dataIndex: 'progress',
      key: 'progress',
      width: 150,
      render: (progress: number) => (
        <Progress percent={progress} size="small" strokeColor="#e67e22" trailColor="#f0f0f0" />
      ),
    },
    {
      title: '截止时间',
      dataIndex: 'deadline',
      key: 'deadline',
      width: 160,
      render: (deadline: string) =>
        deadline ? (
          <span style={{ color: dayjs(deadline).isBefore(dayjs()) ? '#f5222d' : '#646a73' }}>
            {dayjs(deadline).format('MM-DD HH:mm')}
          </span>
        ) : '无',
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      fixed: 'right' as const,
      render: (_: any, record: Task) => (
        <Dropdown
          menu={{
            items: [
              {
                key: 'view',
                icon: <EyeOutlined />,
                label: '查看详情',
                onClick: () => navigate(`/tasks/${record.id}`),
              },
              ...(isManager
                ? [
                    {
                      key: 'edit',
                      icon: <EditOutlined />,
                      label: '编辑',
                      onClick: () => navigate(`/tasks/${record.id}/edit`),
                    },
                    {
                      key: 'complete',
                      icon: <CheckCircleOutlined />,
                      label: '标记完成',
                      disabled: record.status === 'completed',
                      onClick: () => handleComplete(record),
                    },
                    {
                      key: 'delete',
                      icon: <DeleteOutlined />,
                      label: <span style={{ color: '#f5222d' }}>删除</span>,
                      onClick: () => handleDelete(record),
                    },
                  ]
                : []),
            ],
          }}
        >
          <Button type="text" icon={<MoreOutlined style={{ fontSize: 18 }} />} style={{ color: '#646a73' }} />
        </Dropdown>
      ),
    },
  ];

  return (
    <Card
      title="任务管理"
      extra={
        isManager && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/tasks/create')}
            style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
          >
            创建任务
          </Button>
        )
      }
      bordered={false}
    >
      <Space style={{ marginBottom: 16 }}>
        <Select
          placeholder="筛选状态"
          value={statusFilter || undefined}
          onChange={setStatusFilter}
          style={{ width: 140 }}
          allowClear
        >
          <Select.Option value="pending">待分配</Select.Option>
          <Select.Option value="assigned">已分配</Select.Option>
          <Select.Option value="processing">进行中</Select.Option>
          <Select.Option value="completed">已完成</Select.Option>
        </Select>
        <Select
          placeholder="筛选优先级"
          value={priorityFilter || undefined}
          onChange={setPriorityFilter}
          style={{ width: 140 }}
          allowClear
        >
          <Select.Option value="urgent">紧急</Select.Option>
          <Select.Option value="high">高</Select.Option>
          <Select.Option value="normal">普通</Select.Option>
          <Select.Option value="low">低</Select.Option>
        </Select>
      </Space>

      <Table
        columns={columns}
        dataSource={data}
        loading={loading}
        rowKey="id"
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={(p) =>
          setPagination({
            current: p.current || 1,
            pageSize: p.pageSize || 10,
            total: pagination.total,
          })
        }
      />
    </Card>
  );
}

function getPriorityColor(priority: string) {
  const colors: Record<string, string> = {
    urgent: 'red',
    high: 'orange',
    normal: 'blue',
    low: 'green',
  };
  return colors[priority] || 'default';
}

function getPriorityLabel(priority: string) {
  const labels: Record<string, string> = {
    urgent: '紧急',
    high: '高',
    normal: '普通',
    low: '低',
  };
  return labels[priority] || priority;
}

function getStatusColor(status: string) {
  const colors: Record<string, string> = {
    draft: 'default',
    pending: 'orange',
    assigned: 'blue',
    processing: 'cyan',
    completed: 'green',
    cancelled: 'gray',
  };
  return colors[status] || 'default';
}

function getStatusLabel(status: string) {
  const labels: Record<string, string> = {
    draft: '草稿',
    pending: '待分配',
    assigned: '已分配',
    processing: '进行中',
    completed: '已完成',
    cancelled: '已取消',
  };
  return labels[status] || status;
}
