import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Form, Input, Button, message } from 'antd'
import { UserOutlined, LockOutlined, AlertOutlined } from '@ant-design/icons'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/store/auth'

export default function Login() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)

  const onFinish = async (values: { phone: string; password: string }) => {
    setLoading(true)
    try {
      const res = await authApi.login({ username: values.phone, password: values.password })
      const { access_token, refresh_token, user } = res.data.data
      setAuth(access_token, refresh_token, user)
      message.success('登录成功')
      navigate('/dashboard', { replace: true })
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center',
      background: 'linear-gradient(135deg, #fdf2e9 0%, #f5f7fa 100%)',
    }}>
      <div style={{
        width: 400, padding: 40, background: '#fff', borderRadius: 12,
        boxShadow: '0 4px 24px rgba(0,0,0,0.08)',
      }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <AlertOutlined style={{ fontSize: 48, color: '#e67e22', marginBottom: 12 }} />
          <h1 style={{ fontSize: 24, fontWeight: 600, color: '#1f2329', margin: 0 }}>团圆寻亲</h1>
          <p style={{ color: '#8f959e', marginTop: 4 }}>志愿者管理系统</p>
        </div>
        <Form onFinish={onFinish} size="large" autoComplete="off">
          <Form.Item name="phone" rules={[{ required: true, message: '请输入手机号' }]}>
            <Input prefix={<UserOutlined />} placeholder="手机号" />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }, { min: 6, message: '密码至少6位' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block style={{ height: 44 }}>
              登 录
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  )
}
