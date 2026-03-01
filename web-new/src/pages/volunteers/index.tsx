import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Button, Input, Tag, Space, Avatar, Dropdown, Modal, message } from 'antd';
import { PlusOutlined, MoreOutlined, EyeOutlined, EditOutlined, DeleteOutlined, UserOutlined } from '@ant-design/icons';
import type { User } from '@/types';
import { http } from '@/utils/request';
import { usePermission } from '@/utils/permission';

export default function VolunteersPage() {
  const navigate = useNavigate();
  const { isManager } = usePermission();
  const [data, setData] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchText, setSearchText] = useState('');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  useEffect(() => {
    fetchData();
  }, [pagination.current, pagination.pageSize]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const res: any = await http.get('/users', {
        params: {
          page: pagination.current,
          page_size: pagination.pageSize,
          keyword: searchText || undefined,
        },
      });
      setData(res.list || []);
      setPagination((prev) => ({ ...prev, total: res.total || 0 }));
    } catch (error) {
      console.error('获取志愿者列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    setPagination((prev) => ({ ...prev, current: 1 }));
    fetchData();
  };

  const handleDelete = (record: User) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除志愿者 "${record.nickname}" 吗？`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.delete(`/users/${record.id}`);
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
      title: '志愿者',
      dataIndex: 'nickname',
      key: 'nickname',
      render: (_: any, record: User) => (
        <Space>
          <Avatar
            size={40}
            icon={<UserOutlined />}
            src={record.avatar}
            style={{ backgroundColor: '#e67e22' }}
          />
          <div>
            <div style={{ fontWeight: 500, color: '#1f2329' }}>{record.nickname}</div>
            <div style={{ color: '#8f959e', fontSize: 13 }}>{record.real_name || '-'}</div>
          </div>
        </Space>
      ),
    },
    {
      title: '手机号',
      dataIndex: 'phone',
      key: 'phone',
      render: (phone: string) => <span style={{ color: '#646a73' }}>{phone}</span>,
    },
    {
      title: '所属组织',
      dataIndex: ['org', 'name'],
      key: 'org',
      render: (name: string) => <span style={{ color: '#646a73' }}>{name || '未分配'}</span>,
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      width: 110,
      render: (role: string) => (
        <Tag color={getRoleColor(role)} style={{ fontSize: 13 }}>
          {getRoleLabel(role)}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'error'} style={{ fontSize: 13 }}>
          {status === 'active' ? '正常' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '最后登录',
      dataIndex: 'last_login',
      key: 'last_login',
      render: (time: string) =>
        time ? (
          <span style={{ color: '#646a73' }}>
            {new Date(time).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })}
          </span>
        ) : '从未登录',
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      fixed: 'right' as const,
      render: (_: any, record: User) => (
        <Dropdown
          menu={{
            items: [
              {
                key: 'view',
                icon: <EyeOutlined />,
                label: '查看详情',
                onClick: () => navigate(`/volunteers/${record.id}`),
              },
              ...(isManager
                ? [
                    {
                      key: 'edit',
                      icon: <EditOutlined />,
                      label: '编辑',
                      onClick: () => navigate(`/volunteers/${record.id}/edit`),
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
      title="志愿者管理"
      extra={
        isManager && (
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => navigate('/volunteers/create')}
            style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
          >
            添加志愿者
          </Button>
        )
      }
      bordered={false}
    >
      <div style={{ marginBottom: 16 }}>
        <Input.Search
          placeholder="搜索姓名、手机号"
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
          onSearch={handleSearch}
          style={{ width: 300 }}
          allowClear
        />
      </div>

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

function getRoleColor(role: string) {
  const colors: Record<string, string> = {
    super_admin: 'red',
    admin: 'orange',
    manager: 'blue',
    volunteer: 'green',
  };
  return colors[role] || 'default';
}

function getRoleLabel(role: string) {
  const labels: Record<string, string> = {
    super_admin: '超级管理员',
    admin: '管理员',
    manager: '组织者',
    volunteer: '志愿者',
  };
  return labels[role] || role;
}
