import { useState } from 'react'
import { Card, Button, Input, message, Tabs } from 'antd'
import { WechatOutlined, LockOutlined, UserOutlined } from '@ant-design/icons'
import { useAuthStore } from '../stores/auth'
import { useNavigate } from 'react-router-dom'
import { userApi } from '../services/user'

const Login = () => {
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('password')
  const { setToken, setUser } = useAuthStore()
  const navigate = useNavigate()

  // 账号密码登录
  const handlePasswordLogin = async (values: any) => {
    setLoading(true)
    try {
      const data = await userApi.adminLogin(values.username, values.password)
      setToken(data.token, data.refresh_token)
      
      // 获取用户信息
      const userInfo = await userApi.getCurrentUser()
      setUser(userInfo)
      
      message.success('登录成功')
      navigate('/')
    } catch (error: any) {
      message.error(error.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ 
      height: '100vh', 
      display: 'flex', 
      alignItems: 'center', 
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
    }}>
      <Card style={{ width: 400, borderRadius: 8 }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h1 style={{ margin: 0, color: '#1890ff' }}>团圆寻亲志愿者系统</h1>
          <p style={{ color: '#888', marginTop: 8 }}>后台管理系统</p>
        </div>

        <Tabs activeKey={activeTab} onChange={setActiveTab} centered>
          <Tabs.TabPane tab="账号密码" key="password">
            <form onSubmit={(e) => {
              e.preventDefault()
              const formData = new FormData(e.currentTarget)
              handlePasswordLogin({
                username: formData.get('username'),
                password: formData.get('password'),
              })
            }}>
              <div style={{ marginBottom: 16 }}>
                <Input 
                  name="username"
                  prefix={<UserOutlined />} 
                  placeholder="手机号/用户名" 
                  size="large"
                />
              </div>
              <div style={{ marginBottom: 24 }}>
                <Input.Password 
                  name="password"
                  prefix={<LockOutlined />} 
                  placeholder="密码" 
                  size="large"
                />
              </div>
              <Button 
                type="primary" 
                htmlType="submit"
                size="large" 
                loading={loading}
                style={{ width: '100%' }}
              >
                登录
              </Button>
            </form>
            <p style={{ marginTop: 16, color: '#999', fontSize: 12, textAlign: 'center' }}>
              默认管理员账号: 13800138000 / admin123
            </p>
          </Tabs.TabPane>
          
          <Tabs.TabPane tab="微信登录" key="wechat">
            <div style={{ textAlign: 'center', padding: '40px 0' }}>
              <div style={{ 
                width: 200, 
                height: 200, 
                margin: '0 auto 24px',
                background: '#f5f5f5',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                borderRadius: 8
              }}>
                <WechatOutlined style={{ fontSize: 80, color: '#52c41a' }} />
              </div>
              <p style={{ color: '#999' }}>
                请使用微信小程序扫码登录
              </p>
            </div>
          </Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  )
}

export default Login
