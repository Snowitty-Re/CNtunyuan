import { useEffect, useState } from 'react'
import { Tree, Card, Button, Space, Modal, Form, Input, Select, message, Tag } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { orgApi } from '../services/organization'
import type { Organization } from '../types'

const { Option } = Select

const Organizations = () => {
  const [orgs, setOrgs] = useState<Organization[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingOrg, setEditingOrg] = useState<Organization | null>(null)
  const [form] = Form.useForm()

  useEffect(() => {
    fetchOrgs()
  }, [])

  const fetchOrgs = async () => {
    setLoading(true)
    try {
      const data = await orgApi.getOrgTree()
      setOrgs(data)
    } finally {
      setLoading(false)
    }
  }

  const buildTreeData = (data: Organization[]): any[] => {
    return data.map((org) => ({
      key: org.id,
      title: (
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span>
            {org.name}
            <Tag color="blue" style={{ marginLeft: 8 }}>
              {org.type === 'root' && '总部'}
              {org.type === 'province' && '省级'}
              {org.type === 'city' && '市级'}
              {org.type === 'district' && '区级'}
              {org.type === 'street' && '街道'}
            </Tag>
            <span style={{ marginLeft: 8, color: '#888' }}>
              志愿者: {org.volunteer_count} | 案件: {org.case_count}
            </span>
          </span>
          <Space>
            <Button size="small" icon={<PlusOutlined />} onClick={() => handleAddChild(org)}>
              添加子组织
            </Button>
            <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(org)}>
              编辑
            </Button>
            <Button size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(org.id)}>
              删除
            </Button>
          </Space>
        </div>
      ),
      children: org.children ? buildTreeData(org.children) : undefined,
    }))
  }

  const handleAddChild = (parent: Organization) => {
    setEditingOrg(null)
    form.setFieldsValue({ parent_id: parent.id })
    setModalVisible(true)
  }

  const handleEdit = (org: Organization) => {
    setEditingOrg(org)
    form.setFieldsValue(org)
    setModalVisible(true)
  }

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后无法恢复，是否确认？',
      onOk: async () => {
        await orgApi.deleteOrg(id)
        message.success('删除成功')
        fetchOrgs()
      },
    })
  }

  const handleSubmit = async (values: any) => {
    if (editingOrg) {
      await orgApi.updateOrg(editingOrg.id, values)
      message.success('更新成功')
    } else {
      await orgApi.createOrg(values)
      message.success('创建成功')
    }
    setModalVisible(false)
    fetchOrgs()
  }

  return (
    <div>
      <Card
        title="组织架构"
        extra={
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditingOrg(null); form.resetFields(); setModalVisible(true) }}>
            新增组织
          </Button>
        }
      >
        <Tree treeData={buildTreeData(orgs)} defaultExpandAll blockNode />
      </Card>

      <Modal
        title={editingOrg ? '编辑组织' : '新增组织'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={600}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item name="name" label="组织名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="组织编码" rules={[{ required: true }]}>
            <Input disabled={!!editingOrg} />
          </Form.Item>
          <Form.Item name="type" label="组织类型" rules={[{ required: true }]}>
            <Select>
              <Option value="root">总部</Option>
              <Option value="province">省级</Option>
              <Option value="city">市级</Option>
              <Option value="district">区级</Option>
              <Option value="street">街道</Option>
            </Select>
          </Form.Item>
          <Form.Item name="province" label="省">
            <Input />
          </Form.Item>
          <Form.Item name="city" label="市">
            <Input />
          </Form.Item>
          <Form.Item name="district" label="区">
            <Input />
          </Form.Item>
          <Form.Item name="street" label="街道">
            <Input />
          </Form.Item>
          <Form.Item name="address" label="详细地址">
            <Input />
          </Form.Item>
          <Form.Item name="contact" label="联系人">
            <Input />
          </Form.Item>
          <Form.Item name="phone" label="联系电话">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Organizations
