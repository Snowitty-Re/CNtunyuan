import { useState } from 'react'
import { Card, Button, Input, message, Tabs } from 'antd'
import { WechatOutlined, LockOutlined, UserOutlined } from '@ant-design/icons'
import { useAuthStore } from '../stores/auth'
import { useNavigate } from 'react-router-dom'

const Login = () => {
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('wechat')
  const { setToken, setUser } = useAuthStore()
  const navigate = useNavigate()

  // 模拟微信登录
  const handleWechatLogin = () => {
    setLoading(true)
    // 这里应该调用微信登录API
    setTimeout(() => {
      // 模拟登录成功
      setToken('mock-token', 'mock-refresh-token')
      setUser({
        id: '1',
        nickname: '管理员',
        avatar: '',
        role: 'admin',
      })
      message.success('登录成功')
      navigate('/')
      setLoading(false)
    }, 1000)
  }

  // 模拟账号密码登录
  const handlePasswordLogin = (values: any) => {
    setLoading(true)
    setTimeout(() => {
      setToken('mock-token', 'mock-refresh-token')
      setUser({
        id: '1',
        nickname: values.username,
        avatar: '',
        role: 'admin',
      })
      message.success('登录成功')
      navigate('/')
      setLoading(false)
    }, 1000)
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
              <Button 
                type="primary" 
                size="large" 
                icon={<WechatOutlined />}
                loading={loading}
                onClick={handleWechatLogin}
                style={{ width: '100%', background: '#52c41a', borderColor: '#52c41a' }}
              >
                微信一键登录
              </Button>
              <p style={{ marginTop: 16, color: '#888', fontSize: 12 }}>
                请使用微信扫描二维码登录
              </p>
            </div>
          </Tabs.TabPane>
          
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
                  placeholder="用户名" 
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
          </Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  )
}

export default Login
