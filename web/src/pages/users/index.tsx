import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Input, Select, Button, Space, Row, Col, Modal, Form, message } from 'antd';
import { PlusOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import RoleTag from '@/components/RoleTag';
import ConfirmButton from '@/components/ConfirmButton';
import OrgTreeSelect from '@/components/OrgTreeSelect';
import { getUsers, createUser, updateUser, deleteUser, updateUserStatus } from '@/api/user';
import { usePagination } from '@/hooks/usePagination';
import { usePermission } from '@/hooks/usePermission';
import { formatDateTime } from '@/utils/format';
import { USER_ROLE_MAP, USER_STATUS_MAP } from '@/utils/constants';
import type { UserResponse, CreateUserRequest, UserRole } from '@/types';

export default function UsersPage() {
  const navigate = useNavigate();
  const { isAdmin } = usePermission();
  const { page, pageSize, pagination, setTotal, onChange } = usePagination();
  const [data, setData] = useState<UserResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [orgFilter, setOrgFilter] = useState<string>('');
  const [modalOpen, setModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<UserResponse | null>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getUsers({
        page,
        page_size: pageSize,
        keyword,
        role: roleFilter,
        status: statusFilter,
        org_id: orgFilter,
      });
      setData(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword, roleFilter, statusFilter, orgFilter, setTotal]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleSearch = () => {
    onChange(1, pageSize);
  };

  const openModal = (user?: UserResponse) => {
    setEditingUser(user || null);
    form.resetFields();
    if (user) {
      form.setFieldsValue(user);
    }
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    try {
      if (editingUser) {
        await updateUser(editingUser.id, values);
        message.success('更新成功');
      } else {
        await createUser(values as CreateUserRequest);
        message.success('创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch {
      // handled by interceptor
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteUser(id);
      message.success('删除成功');
      fetchData();
    } catch {
      // handled by interceptor
    }
  };

  const handleToggleStatus = async (user: UserResponse) => {
    const newStatus = user.status === 'active' ? 'banned' : 'active';
    try {
      await updateUserStatus(user.id, newStatus);
      message.success('状态更新成功');
      fetchData();
    } catch {
      // handled by interceptor
    }
  };

  const columns = [
    { title: '昵称', dataIndex: 'nickname', key: 'nickname' },
    { title: '手机号', dataIndex: 'phone', key: 'phone' },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => <RoleTag role={role} />,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <StatusTag status={status} />,
    },
    { title: '组织', dataIndex: 'org_name', key: 'org_name' },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: UserResponse) => (
        <Space>
          <Button type="link" size="small" onClick={() => navigate(`/users/${record.id}`)}>
            详情
          </Button>
          {isAdmin && (
            <>
              <Button type="link" size="small" onClick={() => openModal(record)}>
                编辑
              </Button>
              <Button
                type="link"
                size="small"
                onClick={() => handleToggleStatus(record)}
              >
                {record.status === 'active' ? '禁用' : '启用'}
              </Button>
              <ConfirmButton
                type="link"
                size="small"
                danger
                title="确定删除此用户?"
                onConfirm={() => handleDelete(record.id)}
              >
                删除
              </ConfirmButton>
            </>
          )}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer
      title="用户管理"
      breadcrumb={[{ title: '首页' }, { title: '用户管理' }]}
      extra={
        isAdmin ? (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => openModal()}>
            新建用户
          </Button>
        ) : null
      }
    >
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8} md={6}>
            <Input
              placeholder="搜索昵称/手机号"
              prefix={<SearchOutlined />}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onPressEnter={handleSearch}
            />
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              placeholder="角色"
              allowClear
              style={{ width: '100%' }}
              value={roleFilter || undefined}
              onChange={(v) => setRoleFilter(v || '')}
              options={Object.entries(USER_ROLE_MAP).map(([k, v]) => ({ label: v, value: k }))}
            />
          </Col>
          <Col xs={12} sm={6} md={4}>
            <Select
              placeholder="状态"
              allowClear
              style={{ width: '100%' }}
              value={statusFilter || undefined}
              onChange={(v) => setStatusFilter(v || '')}
              options={Object.entries(USER_STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))}
            />
          </Col>
          <Col xs={24} sm={8} md={6}>
            <OrgTreeSelect
              value={orgFilter || undefined}
              onChange={(v) => setOrgFilter(v || '')}
              placeholder="选择组织"
              style={{ width: '100%' }}
            />
          </Col>
          <Col>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={handleSearch}>
                搜索
              </Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => {
                  setKeyword('');
                  setRoleFilter('');
                  setStatusFilter('');
                  setOrgFilter('');
                  onChange(1, pageSize);
                }}
              >
                重置
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      <Card>
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          pagination={pagination}
          scroll={{ x: 800 }}
        />
      </Card>

      <Modal
        title={editingUser ? '编辑用户' : '新建用户'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="nickname" label="昵称" rules={[{ required: true, message: '请输入昵称' }]}>
            <Input />
          </Form.Item>
          {!editingUser && (
            <>
              <Form.Item name="phone" label="手机号" rules={[{ required: true, message: '请输入手机号' }]}>
                <Input />
              </Form.Item>
              <Form.Item
                name="password"
                label="密码"
                rules={[
                  { required: true, message: '请输入密码' },
                  { min: 8, message: '至少8位' },
                ]}
              >
                <Input.Password />
              </Form.Item>
            </>
          )}
          <Form.Item name="email" label="邮箱">
            <Input />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
            <Select
              options={Object.entries(USER_ROLE_MAP).map(([k, v]) => ({ label: v, value: k }))}
            />
          </Form.Item>
          <Form.Item name="org_id" label="组织" rules={[{ required: !editingUser, message: '请选择组织' }]}>
            <OrgTreeSelect />
          </Form.Item>
          {editingUser && (
            <Form.Item name="status" label="状态">
              <Select
                options={Object.entries(USER_STATUS_MAP).map(([k, v]) => ({ label: v, value: k }))}
              />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </PageContainer>
  );
}
