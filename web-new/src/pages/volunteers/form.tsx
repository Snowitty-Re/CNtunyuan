import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Form, Input, Button, Select, Radio, message, Space } from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import { http } from '@/utils/request';

export default function VolunteerFormPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const [form] = Form.useForm();
  const isEdit = !!id;

  useEffect(() => {
    if (isEdit) {
      fetchDetail();
    }
  }, [id]);

  const fetchDetail = async () => {
    try {
      const res: any = await http.get(`/users/${id}`);
      form.setFieldsValue({
        nickname: res.nickname,
        real_name: res.real_name,
        phone: res.phone,
        email: res.email,
        role: res.role,
        status: res.status,
      });
    } catch (error) {
      message.error('获取志愿者信息失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      if (isEdit) {
        await http.put(`/users/${id}`, values);
        message.success('更新成功');
      } else {
        await http.post('/users', { ...values, password: '123456' });
        message.success('创建成功，初始密码为 123456');
      }
      navigate('/volunteers');
    } catch (error) {
      message.error(isEdit ? '更新失败' : '创建失败');
    }
  };

  return (
    <div>
      <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/volunteers')}>
            返回
          </Button>
          <span style={{ fontSize: 18, fontWeight: 600 }}>
            {isEdit ? '编辑志愿者' : '添加志愿者'}
          </span>
        </Space>
      </Card>

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
          style={{ maxWidth: 600 }}
        >
          <Form.Item
            label="昵称"
            name="nickname"
            rules={[{ required: true, message: '请输入昵称' }]}
          >
            <Input placeholder="请输入昵称" />
          </Form.Item>

          <Form.Item
            label="真实姓名"
            name="real_name"
          >
            <Input placeholder="请输入真实姓名" />
          </Form.Item>

          <Form.Item
            label="手机号"
            name="phone"
            rules={[
              { required: true, message: '请输入手机号' },
              { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' },
            ]}
          >
            <Input placeholder="请输入手机号" disabled={isEdit} />
          </Form.Item>

          <Form.Item
            label="邮箱"
            name="email"
            rules={[{ type: 'email', message: '请输入正确的邮箱' }]}
          >
            <Input placeholder="请输入邮箱" />
          </Form.Item>

          <Form.Item
            label="角色"
            name="role"
            initialValue="volunteer"
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select placeholder="请选择角色">
              <Select.Option value="super_admin">超级管理员</Select.Option>
              <Select.Option value="admin">管理员</Select.Option>
              <Select.Option value="manager">组织者</Select.Option>
              <Select.Option value="volunteer">志愿者</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label="状态"
            name="status"
            initialValue="active"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Radio.Group>
              <Radio value="active">正常</Radio>
              <Radio value="inactive">禁用</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item style={{ marginTop: 24 }}>
            <Space size={16}>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} size="large">
                {isEdit ? '保存修改' : '添加志愿者'}
              </Button>
              <Button onClick={() => navigate('/volunteers')} size="large">
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
