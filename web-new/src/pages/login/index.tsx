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
import { http } from '@/utils/request';
import './style.css';

interface LoginForm {
  phone: string;
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
      const res: any = await http.post('/auth/admin-login', {
        phone: values.phone,
        password: values.password,
      });

      if (res.token) {
        setToken(res.token, res.refresh_token || '');
        setUser(res.user);
        message.success('登录成功');
        navigate('/');
      }
    } catch (error) {
      console.error('登录失败:', error);
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
            用爱心点亮希望，让失散家庭重聚
          </p>
          <div className="brand-stats">
            <div className="stat-item">
              <div className="stat-value">10,000+</div>
              <div className="stat-label">注册志愿者</div>
            </div>
            <div className="stat-item">
              <div className="stat-value">5,000+</div>
              <div className="stat-label">成功寻回</div>
            </div>
            <div className="stat-item">
              <div className="stat-value">98%</div>
              <div className="stat-label">好评率</div>
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
            <h2 className="login-title">欢迎回来</h2>
            <p className="login-subtitle">请登录您的账号</p>
          </div>

          <Tabs
            activeKey={activeTab}
            onChange={setActiveTab}
            centered
            className="login-tabs"
            items={[
              {
                key: 'password',
                label: '账号密码',
                children: (
                  <Form
                    name="login"
                    onFinish={handlePasswordLogin}
                    autoComplete="off"
                    size="large"
                  >
                    <Form.Item
                      name="phone"
                      rules={[
                        { required: true, message: '请输入手机号' },
                        { pattern: /^1[3-9]\d{9}$/, message: '手机号格式错误' },
                      ]}
                    >
                      <Input
                        prefix={<UserOutlined className="text-gray-400" />}
                        placeholder="请输入手机号"
                        maxLength={11}
                      />
                    </Form.Item>

                    <Form.Item
                      name="password"
                      rules={[{ required: true, message: '请输入密码' }]}
                    >
                      <Input.Password
                        prefix={<LockOutlined className="text-gray-400" />}
                        placeholder="请输入密码"
                      />
                    </Form.Item>

                    <Form.Item>
                      <div className="flex items-center justify-between">
                        <Form.Item name="remember" valuePropName="checked" noStyle>
                          <Checkbox>记住我</Checkbox>
                        </Form.Item>
                        <a href="#" className="text-orange-500 hover:text-orange-600">
                          忘记密码？
                        </a>
                      </div>
                    </Form.Item>

                    <Form.Item>
                      <Button
                        type="primary"
                        htmlType="submit"
                        loading={loading}
                        block
                        size="large"
                        className="login-btn"
                      >
                        登 录
                      </Button>
                    </Form.Item>
                  </Form>
                ),
              },
              {
                key: 'wechat',
                label: '微信登录',
                children: (
                  <div className="wechat-login">
                    <div className="qr-code">
                      <div className="qr-placeholder">
                        <WechatOutlined className="text-6xl text-green-500" />
                      </div>
                      <p className="qr-tip">请使用微信扫一扫登录</p>
                    </div>
                    <Button
                      type="primary"
                      block
                      size="large"
                      icon={<WechatOutlined />}
                      onClick={handleWechatLogin}
                      style={{ backgroundColor: '#07c160' }}
                    >
                      微信授权登录
                    </Button>
                  </div>
                ),
              },
            ]}
          />

          <div className="login-footer">
            <p className="text-gray-400 text-sm">
              还没有账号？
              <a href="#" className="text-orange-500 hover:text-orange-600 ml-1">
                联系管理员
              </a>
            </p>
          </div>
        </Card>
      </motion.div>
    </div>
  );
}
