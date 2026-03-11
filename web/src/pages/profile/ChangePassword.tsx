import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Form, Input, Button, message, Space } from 'antd';
import PageContainer from '@/components/PageContainer';
import { changePassword } from '@/api/user';

export default function ChangePasswordPage() {
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { old_password: string; new_password: string }) => {
    setLoading(true);
    try {
      await changePassword(values);
      message.success('密码修改成功');
      navigate('/profile');
    } catch { /* interceptor */ } finally {
      setLoading(false);
    }
  };

  return (
    <PageContainer
      title="修改密码"
      breadcrumb={[{ title: '首页' }, { title: '个人资料' }, { title: '修改密码' }]}
    >
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 400 }}>
          <Form.Item name="old_password" label="当前密码" rules={[{ required: true, message: '请输入当前密码' }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 8, message: '密码至少8位' },
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="confirm_password"
            label="确认密码"
            dependencies={['new_password']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) return Promise.resolve();
                  return Promise.reject(new Error('两次密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>修改密码</Button>
              <Button onClick={() => navigate(-1)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
}
