import { useEffect, useState } from 'react'
import { Form, Input, Select, Button, Card, message, Spin, Row, Col } from 'antd'
import { userApi } from '@/api/user'
import { useAuthStore } from '@/store/auth'

export default function Profile() {
  const [form] = Form.useForm()
  const [pwdForm] = Form.useForm()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [changingPwd, setChangingPwd] = useState(false)
  const setUser = useAuthStore((s) => s.setUser)

  useEffect(() => {
    userApi.getProfile().then((res) => {
      form.setFieldsValue(res.data.data)
    }).finally(() => setLoading(false))
  }, [])

  const handleSave = async (values: Record<string, string>) => {
    setSaving(true)
    try {
      const res = await userApi.updateProfile({
        nickname: values.nickname, email: values.email,
        real_name: values.real_name, gender: values.gender,
        address: values.address, emergency: values.emergency,
        emergency_tel: values.emergency_tel, introduction: values.introduction,
      })
      setUser(res.data.data)
      message.success('保存成功')
    } finally {
      setSaving(false)
    }
  }

  const handleChangePassword = async (values: { old_password: string; new_password: string }) => {
    setChangingPwd(true)
    try {
      await userApi.changePassword(values)
      message.success('密码修改成功')
      pwdForm.resetFields()
    } finally {
      setChangingPwd(false)
    }
  }

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />

  return (
    <div>
      <h2 style={{ marginBottom: 24, fontSize: 20, fontWeight: 600 }}>个人中心</h2>

      <Row gutter={24}>
        <Col xs={24} lg={14}>
          <Card title="基本信息">
            <Form form={form} layout="vertical" onFinish={handleSave}>
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item name="nickname" label="昵称" rules={[{ required: true, message: '请输入昵称' }]}>
                    <Input />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name="email" label="邮箱"><Input /></Form.Item>
                </Col>
              </Row>
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item name="real_name" label="真实姓名"><Input /></Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name="gender" label="性别">
                    <Select allowClear>
                      <Select.Option value="male">男</Select.Option>
                      <Select.Option value="female">女</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>
              </Row>
              <Form.Item name="address" label="地址"><Input /></Form.Item>
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item name="emergency" label="紧急联系人"><Input /></Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name="emergency_tel" label="紧急联系电话"><Input /></Form.Item>
                </Col>
              </Row>
              <Form.Item name="introduction" label="个人简介"><Input.TextArea rows={3} /></Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={saving}>保存</Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={10}>
          <Card title="修改密码">
            <Form form={pwdForm} layout="vertical" onFinish={handleChangePassword}>
              <Form.Item name="old_password" label="当前密码" rules={[{ required: true, message: '请输入当前密码' }]}>
                <Input.Password />
              </Form.Item>
              <Form.Item name="new_password" label="新密码" rules={[{ required: true, message: '请输入新密码' }, { min: 6, message: '至少6位' }]}>
                <Input.Password />
              </Form.Item>
              <Form.Item
                name="confirm_password" label="确认密码"
                dependencies={['new_password']}
                rules={[
                  { required: true, message: '请确认密码' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('new_password') === value) return Promise.resolve()
                      return Promise.reject(new Error('两次密码不一致'))
                    },
                  }),
                ]}
              >
                <Input.Password />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={changingPwd}>修改密码</Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
