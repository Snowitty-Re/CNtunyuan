import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Button,
  Form,
  Input,
  Card,
  Tabs,
  message,
  Checkbox,
} from 'antd';
import {
  UserOutlined,
  LockOutlined,
  WechatOutlined,
} from '@ant-design/icons';
import { motion } from 'framer-motion';
import { useAuthStore } from '@/stores/auth';
import request from '@/utils/request';
import './style.css';

interface LoginForm {
  username: string;
  password: string;
  remember: boolean;
}

export default function LoginPage() {
  const navigate = useNavigate();
  const { setToken, setUser } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('password');

  // 账号密码登录
  const handlePasswordLogin = async (values: LoginForm) => {
    setLoading(true);
    try {
      // 使用原始 axios 实例以查看完整响应
      const response: any = await request.post('/auth/admin-login', {
        username: values.username,
        password: values.password,
      });

      console.log('原始响应:', response); // 调试日志
      
      // 响应拦截器返回的是 response.data.data
      const res = response;

      if (res && res.token) {
        setToken(res.token, res.refresh_token || '');
        setUser(res.user);
        message.success('登录成功');
        navigate('/dashboard', { replace: true });
      } else {
        console.error('登录响应缺少token:', res);
        message.error('登录响应数据异常');
      }
    } catch (error: any) {
      console.error('登录失败:', error);
      message.error(error?.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  // 微信扫码登录（模拟）
  const handleWechatLogin = () => {
    message.info('微信登录功能开发中');
  };

  return (
    <div className="login-page">
      {/* 背景装饰 */}
      <div className="login-bg">
        <div className="bg-circle circle-1" />
        <div className="bg-circle circle-2" />
        <div className="bg-circle circle-3" />
      </div>

      {/* 左侧品牌区 */}
      <motion.div
        initial={{ opacity: 0, x: -50 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ duration: 0.6 }}
        className="login-brand"
      >
        <div className="brand-content">
          <div className="brand-logo">
            <div className="logo-icon">团</div>
            <h1 className="logo-text">团圆寻亲</h1>
          </div>
          <p className="brand-desc">
            用科技连接爱心<br />
            让团圆不再遥远
          </p>
          <div className="brand-stats">
            <div className="stat-item">
              <div className="stat-value">10,000+</div>
              <div className="stat-label">志愿者</div>
            </div>
            <div className="stat-item">
              <div className="stat-value">5,000+</div>
              <div className="stat-label">成功案例</div>
            </div>
          </div>
        </div>
      </motion.div>

      {/* 右侧登录表单 */}
      <motion.div
        initial={{ opacity: 0, x: 50 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ duration: 0.6, delay: 0.2 }}
        className="login-form-wrapper"
      >
        <Card className="login-card" bordered={false}>
          <div className="login-header">
            <h2 className="login-title">欢迎登录</h2>
            <p className="login-subtitle">团圆寻亲志愿者系统</p>
          </div>

          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            className="login-tabs"
            centered
          >
            <Tabs.TabPane tab="账号登录" key="password">
              <Form
                name="login"
                initialValues={{ remember: true }}
                onFinish={handlePasswordLogin}
                autoComplete="off"
                layout="vertical"
              >
                <Form.Item
                  name="username"
                  rules={[{ required: true, message: '请输入手机号或用户名' }]}
                >
                  <Input
                    prefix={<UserOutlined />}
                    placeholder="请输入手机号或用户名"
                    size="large"
                  />
                </Form.Item>

                <Form.Item
                  name="password"
                  rules={[{ required: true, message: '请输入密码' }]}
                >
                  <Input.Password
                    prefix={<LockOutlined />}
                    placeholder="请输入密码"
                    size="large"
                  />
                </Form.Item>

                <Form.Item>
                  <div className="flex justify-between items-center">
                    <Form.Item name="remember" valuePropName="checked" noStyle>
                      <Checkbox>记住我</Checkbox>
                    </Form.Item>
                    <a className="text-orange-500 hover:text-orange-600">
                      忘记密码？
                    </a>
                  </div>
                </Form.Item>

                <Form.Item>
                  <Button
                    type="primary"
                    htmlType="submit"
                    size="large"
                    block
                    loading={loading}
                    className="login-btn"
                  >
                    登录
                  </Button>
                </Form.Item>
              </Form>
            </Tabs.TabPane>

            <Tabs.TabPane tab="微信登录" key="wechat">
              <div className="wechat-login">
                <div className="qr-code">
                  <div className="qr-placeholder">
                    <WechatOutlined style={{ fontSize: 64, color: '#07C160' }} />
                  </div>
                  <p className="qr-tip">请使用微信扫一扫登录</p>
                </div>
                <Button
                  type="primary"
                  block
                  size="large"
                  icon={<WechatOutlined />}
                  onClick={handleWechatLogin}
                  style={{ backgroundColor: '#07C160', borderColor: '#07C160' }}
                >
                  唤起微信登录
                </Button>
              </div>
            </Tabs.TabPane>
          </Tabs>

          <div className="login-footer">
            <p className="text-gray-400 text-sm">
              默认账号：13800138000 / admin123
            </p>
          </div>
        </Card>
      </motion.div>
    </div>
  );
}
