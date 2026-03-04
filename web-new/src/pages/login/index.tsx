import { useState, useEffect } from 'react';
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
import axios from 'axios';
import './style.css';

interface LoginForm {
  username: string;
  password: string;
  remember: boolean;
}

export default function LoginPage() {
  const navigate = useNavigate();
  const { setToken, setUser, isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('password');

  // 如果已登录，跳转到工作台
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  // 账号密码登录
  const handlePasswordLogin = async (values: LoginForm) => {
    setLoading(true);
    try {
      const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
      const response: any = await axios.post(`${apiUrl}/auth/admin-login`, {
        username: values.username,
        password: values.password,
      });

      console.log('Axios 原始响应:', response);
      console.log('response.data:', response.data);
      
      // 后端返回格式: { code: 0, message: "success", data: { token, user } }
      const resData = response.data;
      
      if (resData.code === 0 || resData.code === 200) {
        const loginData = resData.data;
        // 后端返回 access_token，前端使用 token
        const token = loginData.access_token || loginData.token;
        if (loginData && token) {
          setToken(token, loginData.refresh_token || '');
          setUser(loginData.user);
          message.success('登录成功');
          navigate('/dashboard', { replace: true });
        } else {
          console.error('登录 data 字段缺少 token:', loginData);
          message.error('登录数据格式错误');
        }
      } else {
        message.error(resData.message || '登录失败');
      }
    } catch (error: any) {
      console.error('登录失败:', error);
      message.error(error?.response?.data?.message || error?.message || '登录失败');
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
