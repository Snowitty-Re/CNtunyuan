import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Select, Button, Card, TreeSelect, message, Spin, Space } from 'antd'
import { userApi } from '@/api/user'
import { orgApi } from '@/api/organization'
import { PHONE_RULE } from '@/constants'
import type { Organization } from '@/types'

const roleOptions = [
  { value: 'volunteer', label: '志愿者' }, { value: 'manager', label: '管理者' },
  { value: 'admin', label: '管理员' }, { value: 'super_admin', label: '超级管理员' },
]
const statusOptions = [
  { value: 'active', label: '正常' }, { value: 'inactive', label: '禁用' }, { value: 'banned', label: '封禁' },
]

function buildTreeData(orgs: Organization[]): any[] {
  return orgs.map((org) => ({
    title: org.name,
    value: org.id,
    key: org.id,
    children: org.children ? buildTreeData(org.children) : undefined,
  }))
}

export default function UserForm() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [orgTree, setOrgTree] = useState<any[]>([])
  const isEdit = !!id

  useEffect(() => {
    // Load org tree
    orgApi.getTree().then((res) => {
      const data = res.data.data
      setOrgTree(buildTreeData(Array.isArray(data) ? data : [data]))
    }).catch(() => {})

    if (id) {
      setLoading(true)
      userApi.getById(id).then((res) => {
        form.setFieldsValue(res.data.data)
      }).finally(() => setLoading(false))
    }
  }, [id])

  const onFinish = async (values: Record<string, string>) => {
    setSubmitting(true)
    try {
      if (isEdit) {
        await userApi.update(id!, { nickname: values.nickname, email: values.email, role: values.role as any, org_id: values.org_id, status: values.status as any })
        message.success('更新成功')
      } else {
        await userApi.create({ nickname: values.nickname, phone: values.phone, password: values.password, email: values.email, role: values.role as any, org_id: values.org_id })
        message.success('创建成功')
      }
      navigate('/users')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  return (
    <div>
      <h2 style={{ marginBottom: 24, fontSize: 20, fontWeight: 600 }}>{isEdit ? '编辑用户' : '新建用户'}</h2>
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 500 }}>
          <Form.Item name="nickname" label="昵称" rules={[{ required: true, message: '请输入昵称' }]}>
            <Input />
          </Form.Item>
          {!isEdit && (
            <>
              <Form.Item name="phone" label="手机号" rules={[{ required: true, message: '请输入手机号' }, PHONE_RULE]}>
                <Input maxLength={11} />
              </Form.Item>
              <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }, { min: 8, message: '至少8位' }]}>
                <Input.Password />
              </Form.Item>
            </>
          )}
          <Form.Item name="email" label="邮箱" rules={[{ type: 'email', message: '请输入正确的邮箱' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
            <Select options={roleOptions} />
          </Form.Item>
          <Form.Item name="org_id" label="所属组织" rules={isEdit ? [] : [{ required: true, message: '请选择组织' }]}>
            <TreeSelect
              treeData={orgTree}
              placeholder="选择所属组织"
              allowClear
              showSearch
              treeNodeFilterProp="title"
              dropdownStyle={{ maxHeight: 400, overflow: 'auto' }}
            />
          </Form.Item>
          {isEdit && (
            <Form.Item name="status" label="状态">
              <Select options={statusOptions} />
            </Form.Item>
          )}
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={submitting}>{isEdit ? '保存' : '创建'}</Button>
              <Button onClick={() => navigate('/users')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
