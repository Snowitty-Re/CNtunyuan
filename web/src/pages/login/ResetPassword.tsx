import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message, Steps } from 'antd';
import { PhoneOutlined, SafetyOutlined, LockOutlined } from '@ant-design/icons';
import { sendCode, resetPassword } from '@/api/auth';

export default function ResetPasswordPage() {
  const navigate = useNavigate();
  const [step, setStep] = useState(0);
  const [phone, setPhone] = useState('');
  const [loading, setLoading] = useState(false);
  const [countdown, setCountdown] = useState(0);

  const handleSendCode = async (values: { phone: string }) => {
    setLoading(true);
    try {
      await sendCode(values.phone);
      setPhone(values.phone);
      setStep(1);
      message.success('验证码已发送');
      let c = 60;
      setCountdown(c);
      const timer = setInterval(() => {
        c--;
        setCountdown(c);
        if (c <= 0) clearInterval(timer);
      }, 1000);
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  const handleReset = async (values: { code: string; new_password: string }) => {
    setLoading(true);
    try {
      await resetPassword({ phone, code: values.code, new_password: values.new_password });
      message.success('密码重置成功，请登录');
      navigate('/login');
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card style={{ width: 420, borderRadius: 12 }} bodyStyle={{ padding: '40px 32px' }}>
        <Typography.Title level={3} style={{ textAlign: 'center', marginBottom: 24 }}>
          重置密码
        </Typography.Title>
        <Steps
          current={step}
          size="small"
          style={{ marginBottom: 32 }}
          items={[{ title: '验证手机' }, { title: '设置密码' }]}
        />

        {step === 0 ? (
          <Form onFinish={handleSendCode} size="large">
            <Form.Item name="phone" rules={[{ required: true, message: '请输入手机号' }]}>
              <Input prefix={<PhoneOutlined />} placeholder="手机号" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading}>
                发送验证码
              </Button>
            </Form.Item>
          </Form>
        ) : (
          <Form onFinish={handleReset} size="large">
            <Form.Item name="code" rules={[{ required: true, message: '请输入验证码' }]}>
              <Input
                prefix={<SafetyOutlined />}
                placeholder="验证码"
                suffix={countdown > 0 ? `${countdown}s` : null}
              />
            </Form.Item>
            <Form.Item
              name="new_password"
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 8, message: '密码至少8位' },
              ]}
            >
              <Input.Password prefix={<LockOutlined />} placeholder="新密码（至少8位）" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading}>
                重置密码
              </Button>
            </Form.Item>
          </Form>
        )}
        <div style={{ textAlign: 'center' }}>
          <Link to="/login">返回登录</Link>
        </div>
      </Card>
    </div>
  );
}
