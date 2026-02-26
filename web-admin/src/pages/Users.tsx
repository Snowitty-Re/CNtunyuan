import { useEffect, useState } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, message, Avatar } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { userApi } from '../services/user'
import type { User } from '../types'

const { Option } = Select

const Users = () => {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [form] = Form.useForm()

  useEffect(() => {
    fetchUsers()
  }, [pagination.current, pagination.pageSize])

  const fetchUsers = async () => {
    setLoading(true)
    try {
      const data = await userApi.getUsers({
        page: pagination.current,
        page_size: pagination.pageSize,
      })
      setUsers(data.list)
      setPagination({ ...pagination, total: data.total })
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = (user: User) => {
    setEditingUser(user)
    form.setFieldsValue(user)
    setModalVisible(true)
  }

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后无法恢复，是否确认？',
      onOk: async () => {
        await userApi.deleteUser(id)
        message.success('删除成功')
        fetchUsers()
      },
    })
  }

  const handleSubmit = async (values: any) => {
    if (editingUser) {
      await userApi.updateUser(editingUser.id, values)
      message.success('更新成功')
    }
    setModalVisible(false)
    fetchUsers()
  }

  const columns = [
    {
      title: '头像',
      dataIndex: 'avatar',
      render: (avatar: string, record: User) => (
        <Avatar src={avatar}>{record.nickname?.[0] || record.real_name?.[0]}</Avatar>
      ),
    },
    { title: '昵称', dataIndex: 'nickname' },
    { title: '真实姓名', dataIndex: 'real_name' },
    { title: '手机号', dataIndex: 'phone' },
    {
      title: '角色',
      dataIndex: 'role',
      render: (role: string) => {
        const roleMap: Record<string, { color: string; text: string }> = {
          super_admin: { color: 'red', text: '超级管理员' },
          admin: { color: 'orange', text: '管理员' },
          manager: { color: 'blue', text: '管理者' },
          volunteer: { color: 'green', text: '志愿者' },
        }
        const { color, text } = roleMap[role] || { color: 'default', text: role }
        return <Tag color={color}>{text}</Tag>
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '正常' : '禁用'}
        </Tag>
      ),
    },
    { title: '注册时间', dataIndex: 'created_at' },
    {
      title: '操作',
      render: (_: any, record: User) => (
        <Space>
          <Button icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />}>
          新增志愿者
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={users}
        rowKey="id"
        loading={loading}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={(p) => setPagination({ ...pagination, current: p.current || 1, pageSize: p.pageSize || 20 })}
      />

      <Modal
        title={editingUser ? '编辑用户' : '新增用户'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item name="nickname" label="昵称">
            <Input />
          </Form.Item>
          <Form.Item name="real_name" label="真实姓名">
            <Input />
          </Form.Item>
          <Form.Item name="phone" label="手机号">
            <Input />
          </Form.Item>
          <Form.Item name="role" label="角色">
            <Select>
              <Option value="super_admin">超级管理员</Option>
              <Option value="admin">管理员</Option>
              <Option value="manager">管理者</Option>
              <Option value="volunteer">志愿者</Option>
            </Select>
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select>
              <Option value="active">正常</Option>
              <Option value="inactive">禁用</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Users
