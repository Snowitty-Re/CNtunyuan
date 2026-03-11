import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Table, Card, Input, Select, Button, Space, Row, Col, Modal, Form, message } from 'antd';
import { PlusOutlined, SearchOutlined, ReloadOutlined, ApartmentOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import OrgTreeSelect from '@/components/OrgTreeSelect';
import ConfirmButton from '@/components/ConfirmButton';
import { getOrganizations, createOrganization, updateOrganization, deleteOrganization } from '@/api/organization';
import { usePagination } from '@/hooks/usePagination';
import { formatDateTime } from '@/utils/format';
import { ORG_TYPE_MAP } from '@/utils/constants';
import type { OrganizationResponse, CreateOrganizationRequest } from '@/types';

export default function OrganizationsPage() {
  const navigate = useNavigate();
  const { page, pageSize, pagination, setTotal, onChange } = usePagination();
  const [data, setData] = useState<OrganizationResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [editingOrg, setEditingOrg] = useState<OrganizationResponse | null>(null);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getOrganizations({ page, page_size: pageSize, keyword, type: typeFilter });
      setData(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, keyword, typeFilter, setTotal]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const openModal = (org?: OrganizationResponse) => {
    setEditingOrg(org || null);
    form.resetFields();
    if (org) form.setFieldsValue(org);
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    const values = await form.validateFields();
    try {
      if (editingOrg) {
        await updateOrganization(editingOrg.id, values);
        message.success('更新成功');
      } else {
        await createOrganization(values as CreateOrganizationRequest);
        message.success('创建成功');
      }
      setModalOpen(false);
      fetchData();
    } catch { /* interceptor */ }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteOrganization(id);
      message.success('删除成功');
      fetchData();
    } catch { /* interceptor */ }
  };

  const columns = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    { title: '编码', dataIndex: 'code', key: 'code' },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      render: (v: string) => ORG_TYPE_MAP[v] || v,
    },
    { title: '层级', dataIndex: 'level', key: 'level' },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (v: string) => <StatusTag status={v} />,
    },
    { title: '联系人', dataIndex: 'contact_name', key: 'contact_name' },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => formatDateTime(v),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: OrganizationResponse) => (
        <Space>
          <Button type="link" size="small" onClick={() => openModal(record)}>编辑</Button>
          <ConfirmButton type="link" size="small" danger title="确定删除?" onConfirm={() => handleDelete(record.id)}>
            删除
          </ConfirmButton>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer
      title="组织管理"
      breadcrumb={[{ title: '首页' }, { title: '组织管理' }]}
      extra={
        <Space>
          <Button icon={<ApartmentOutlined />} onClick={() => navigate('/organizations/tree')}>
            树形视图
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => openModal()}>
            新建组织
          </Button>
        </Space>
      }
    >
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}>
            <Input placeholder="搜索名称" prefix={<SearchOutlined />} value={keyword} onChange={(e) => setKeyword(e.target.value)} onPressEnter={() => onChange(1, pageSize)} />
          </Col>
          <Col xs={12} sm={6}>
            <Select placeholder="类型" allowClear style={{ width: '100%' }} value={typeFilter || undefined} onChange={(v) => setTypeFilter(v || '')}
              options={Object.entries(ORG_TYPE_MAP).map(([k, v]) => ({ label: v, value: k }))}
            />
          </Col>
          <Col>
            <Space>
              <Button type="primary" icon={<SearchOutlined />} onClick={() => onChange(1, pageSize)}>搜索</Button>
              <Button icon={<ReloadOutlined />} onClick={() => { setKeyword(''); setTypeFilter(''); onChange(1, pageSize); }}>重置</Button>
            </Space>
          </Col>
        </Row>
      </Card>
      <Card>
        <Table rowKey="id" columns={columns} dataSource={data} loading={loading} pagination={pagination} scroll={{ x: 800 }} />
      </Card>

      <Modal title={editingOrg ? '编辑组织' : '新建组织'} open={modalOpen} onOk={handleSubmit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true }]}><Input disabled={!!editingOrg} /></Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: !editingOrg }]}>
            <Select options={Object.entries(ORG_TYPE_MAP).map(([k, v]) => ({ label: v, value: k }))} disabled={!!editingOrg} />
          </Form.Item>
          {!editingOrg && (
            <Form.Item name="parent_id" label="上级组织"><OrgTreeSelect /></Form.Item>
          )}
          <Form.Item name="description" label="描述"><Input.TextArea rows={2} /></Form.Item>
          <Form.Item name="address" label="地址"><Input /></Form.Item>
          <Form.Item name="contact_name" label="联系人"><Input /></Form.Item>
          <Form.Item name="contact_phone" label="联系电话"><Input /></Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
}
