import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Button, Input, Tag, Space, Avatar, Dropdown, Modal, message, Select } from 'antd';
import { PlusOutlined, MoreOutlined, EyeOutlined, EditOutlined, DeleteOutlined, UserOutlined } from '@ant-design/icons';
import type { MissingPerson } from '@/types';
import { http } from '@/utils/request';
import { usePermission } from '@/utils/permission';
import dayjs from 'dayjs';

export default function CasesPage() {
  const navigate = useNavigate();
  const { isManager } = usePermission();
  const [data, setData] = useState<MissingPerson[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  useEffect(() => {
    fetchData();
  }, [pagination.current, pagination.pageSize, statusFilter]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res: any = await http.get('/missing-persons', {
        params: {
          page: pagination.current,
          page_size: pagination.pageSize,
          keyword: searchText || undefined,
          status: statusFilter || undefined,
        },
      });
      setData(res.list || []);
      setPagination((prev) => ({ ...prev, total: res.total || 0 }));
    } catch (error) {
      console.error('获取案件列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    setPagination((prev) => ({ ...prev, current: 1 }));
    fetchData();
  };

  const handleDelete = (record: MissingPerson) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除案件 "${record.name}" 吗？此操作不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.delete(`/missing-persons/${record.id}`);
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
      title: '案件信息',
      dataIndex: 'name',
      key: 'name',
      render: (_: any, record: MissingPerson) => (
        <Space>
          <Avatar
            size={40}
            icon={<UserOutlined />}
            src={record.photos?.[0]?.url}
            style={{ backgroundColor: '#e67e22' }}
          />
          <div>
            <div style={{ fontWeight: 500, color: '#1f2329' }}>{record.name}</div>
            <div style={{ color: '#8f959e', fontSize: 12 }}>{record.case_no}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '性别',
      dataIndex: 'gender',
      key: 'gender',
      width: 80,
      render: (gender: string) => (
        <Tag color={gender === 'male' ? 'blue' : 'pink'} style={{ fontSize: 13 }}>
          {gender === 'male' ? '男' : '女'}
        </Tag>
      ),
    },
    {
      title: '年龄',
      dataIndex: 'age',
      key: 'age',
      width: 80,
      render: (age?: number) => age ? `${age}岁` : '未知',
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
      title: '走失时间',
      dataIndex: 'missing_time',
      key: 'missing_time',
      render: (time: string) =>
        time ? dayjs(time).format('MM-DD HH:mm') : '-',
    },
    {
      title: '走失地点',
      dataIndex: 'missing_location',
      key: 'missing_location',
      ellipsis: true,
      render: (text: string) => <span style={{ color: '#646a73' }}>{text || '-'}</span>,
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      fixed: 'right' as const,
      render: (_: any, record: MissingPerson) => (
        <Dropdown
          menu={{
            items: [
              {
                key: 'view',
                icon: <EyeOutlined />,
                label: '查看详情',
                onClick: () => navigate(`/cases/${record.id}`),
              },
              ...(isManager
                ? [
                    {
                      key: 'edit',
                      icon: <EditOutlined />,
                      label: '编辑',
                      onClick: () => navigate(`/cases/${record.id}/edit`),
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
      title="寻人案件管理"
      extra={
        isManager && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/cases/create')}
            style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
          >
            发布寻人
          </Button>
        )
      }
      bordered={false}
    >
      <Space style={{ marginBottom: 16 }}>
        <Input.Search
          placeholder="搜索姓名、案件编号"
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          onSearch={handleSearch}
          style={{ width: 280 }}
          allowClear
        />
        <Select
          placeholder="筛选状态"
          value={statusFilter || undefined}
          onChange={setStatusFilter}
          style={{ width: 140 }}
          allowClear
        >
          <Select.Option value="missing">失踪中</Select.Option>
          <Select.Option value="searching">寻找中</Select.Option>
          <Select.Option value="found">已找到</Select.Option>
          <Select.Option value="reunited">已团圆</Select.Option>
          <Select.Option value="closed">已结案</Select.Option>
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

function getStatusColor(status: string) {
  const colors: Record<string, string> = {
    missing: 'red',
    searching: 'orange',
    found: 'blue',
    reunited: 'green',
    closed: 'gray',
  };
  return colors[status] || 'default';
}

function getStatusLabel(status: string) {
  const labels: Record<string, string> = {
    missing: '失踪中',
    searching: '寻找中',
    found: '已找到',
    reunited: '已团圆',
    closed: '已结案',
  };
  return labels[status] || status;
}
