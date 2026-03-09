import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Form, Input, Select, InputNumber, Button, Card, TreeSelect, message, Spin, Space } from 'antd'
import { orgApi } from '@/api/organization'
import type { Organization } from '@/types'

const typeOptions = [
  { value: 'root', label: '总部' }, { value: 'province', label: '省级' },
  { value: 'city', label: '市级' }, { value: 'district', label: '区级' },
  { value: 'street', label: '街道' }, { value: 'community', label: '社区' },
  { value: 'team', label: '小组' },
]

function buildTreeData(orgs: Organization[]): any[] {
  return orgs.map((org) => ({
    title: org.name,
    value: org.id,
    key: org.id,
    children: org.children ? buildTreeData(org.children) : undefined,
  }))
}

export default function OrganizationForm() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [orgTree, setOrgTree] = useState<any[]>([])
  const isEdit = !!id

  useEffect(() => {
    // Load org tree for parent selection
    if (!isEdit) {
      orgApi.getTree().then((res) => {
        const data = res.data.data
        setOrgTree(buildTreeData(Array.isArray(data) ? data : [data]))
      }).catch(() => {})
    }
    if (id) {
      setLoading(true)
      orgApi.getById(id).then((res) => form.setFieldsValue(res.data.data)).finally(() => setLoading(false))
    }
  }, [id])

  const onFinish = async (values: Record<string, unknown>) => {
    setSubmitting(true)
    try {
      if (isEdit) {
        await orgApi.update(id!, {
          name: values.name as string, code: values.code as string,
          description: values.description as string, address: values.address as string,
          contact_name: values.contact_name as string, contact_phone: values.contact_phone as string,
          sort_order: values.sort_order as number,
        })
        message.success('更新成功')
      } else {
        await orgApi.create({
          name: values.name as string, code: values.code as string,
          type: values.type as any, parent_id: values.parent_id as string,
          description: values.description as string, address: values.address as string,
          contact_name: values.contact_name as string, contact_phone: values.contact_phone as string,
          sort_order: values.sort_order as number,
        })
        message.success('创建成功')
      }
      navigate('/organizations')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  return (
    <div>
      <h2 style={{ marginBottom: 24, fontSize: 20, fontWeight: 600 }}>{isEdit ? '编辑组织' : '新建组织'}</h2>
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 500 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true, message: '请输入编码' }]}>
            <Input placeholder="组织唯一编码" />
          </Form.Item>
          {!isEdit && (
            <>
              <Form.Item name="type" label="类型" rules={[{ required: true, message: '请选择类型' }]}>
                <Select options={typeOptions} />
              </Form.Item>
              <Form.Item name="parent_id" label="上级组织">
                <TreeSelect
                  treeData={orgTree}
                  placeholder="选择上级组织（总部级可不选）"
                  allowClear
                  showSearch
                  treeNodeFilterProp="title"
                  dropdownStyle={{ maxHeight: 400, overflow: 'auto' }}
                />
              </Form.Item>
            </>
          )}
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="address" label="地址"><Input /></Form.Item>
          <Form.Item name="contact_name" label="联系人"><Input /></Form.Item>
          <Form.Item name="contact_phone" label="联系电话"><Input /></Form.Item>
          <Form.Item name="sort_order" label="排序"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={submitting}>{isEdit ? '保存' : '创建'}</Button>
              <Button onClick={() => navigate('/organizations')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
