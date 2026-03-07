import { useState, useEffect } from 'react';
import { Card, Form, Input, Button, Avatar, Upload, message, Tabs, Descriptions, Tag } from 'antd';
import { UserOutlined, CameraOutlined, LockOutlined, SafetyOutlined, SaveOutlined } from '@ant-design/icons';
import { useAuthStore } from '@/stores/auth';
import { http } from '@/utils/request';
import { getRoleLabel, getRoleColor } from '@/utils/permission';
import dayjs from 'dayjs';


export default function ProfilePage() {
  const { user } = useAuthStore();
  const [form] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [activeTab, setActiveTab] = useState('1');

  useEffect(() => {
    if (user) {
      fetchProfile();
    }
  }, []);

  const fetchProfile = async () => {
    try {
      const res: any = await http.get('/auth/me');
      form.setFieldsValue({
        nickname: res.nickname,
        real_name: res.real_name,
        phone: res.phone,
        email: res.email,
        avatar: res.avatar,
      });
    } catch (error) {
      message.error('获取个人信息失败');
    }
  };

  const handleUpdateProfile = async (values: any) => {
    if (!user?.id) return;
    setLoading(true);
    try {
      await http.put(`/users/${user.id}`, values);
      message.success('个人信息更新成功');
      fetchProfile();
    } catch (error) {
      message.error('更新失败');
    } finally {
      setLoading(false);
    }
  };

  const handleChangePassword = async (values: any) => {
    setLoading(true);
    try {
      await http.put('/profile/password', {
        old_password: values.old_password,
        new_password: values.new_password,
      });
      message.success('密码修改成功');
      passwordForm.resetFields();
    } catch (error: any) {
      message.error(error?.message || '密码修改失败');
    } finally {
      setLoading(false);
    }
  };

  const handleAvatarUpload = async (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', 'image');

    setUploading(true);
    try {
      const res: any = await http.post('/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      
      // 更新用户头像
      if (user?.id) {
        await http.put(`/users/${user.id}`, { avatar: res.url });
        message.success('头像更新成功');
        fetchProfile();
      }
    } catch (error) {
      message.error('上传失败');
    } finally {
      setUploading(false);
    }
    return false;
  };

  return (
    <div>
      {/* 个人信息卡片 */}
      <Card style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 24 }}>
          <div style={{ position: 'relative' }}>
            <Avatar
              size={100}
              icon={<UserOutlined />}
              src={user?.avatar}
              style={{ backgroundColor: '#e67e22' }}
            />
            <Upload
              accept="image/*"
              beforeUpload={handleAvatarUpload}
              showUploadList={false}
            >
              <Button
                type="primary"
                shape="circle"
                icon={<CameraOutlined />}
                size="small"
                loading={uploading}
                style={{
                  position: 'absolute',
                  bottom: 0,
                  right: 0,
                  backgroundColor: '#e67e22',
                  borderColor: '#e67e22',
                }}
              />
            </Upload>
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 24, fontWeight: 600, marginBottom: 8 }}>
              {user?.nickname}
              <Tag
                color={getRoleColor(user?.role || '')}
                style={{ marginLeft: 12, fontSize: 13 }}
              >
                {getRoleLabel(user?.role || '')}
              </Tag>
            </div>
            <div style={{ color: '#646a73', marginBottom: 4 }}>
              用户名: {user?.real_name || '未设置'}
            </div>
            <div style={{ color: '#8f959e' }}>
              手机号: {user?.phone}
            </div>
          </div>
        </div>
      </Card>

      {/* 标签页内容 */}
      <Card>
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={[
            {
              key: '1',
              label: '基本信息',
              children: (
                <Form
                  form={form}
                  layout="vertical"
                  onFinish={handleUpdateProfile}
                  autoComplete="off"
                  style={{ maxWidth: 600 }}
                >
                  <Form.Item
                    label="昵称"
                    name="nickname"
                    rules={[{ required: true, message: '请输入昵称' }]}
                  >
                    <Input prefix={<UserOutlined />} placeholder="请输入昵称" />
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
                    <Input prefix={<SafetyOutlined />} placeholder="请输入手机号" disabled />
                  </Form.Item>

                  <Form.Item
                    label="邮箱"
                    name="email"
                    rules={[{ type: 'email', message: '请输入正确的邮箱' }]}
                  >
                    <Input placeholder="请输入邮箱" />
                  </Form.Item>

                  <Form.Item style={{ marginTop: 24 }}>
                    <Button
                      type="primary"
                      htmlType="submit"
                      icon={<SaveOutlined />}
                      loading={loading}
                      style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
                    >
                      保存修改
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: '2',
              label: '修改密码',
              children: (
                <Form
                  form={passwordForm}
                  layout="vertical"
                  onFinish={handleChangePassword}
                  autoComplete="off"
                  style={{ maxWidth: 600 }}
                >
                  <Form.Item
                    label="当前密码"
                    name="old_password"
                    rules={[{ required: true, message: '请输入当前密码' }]}
                  >
                    <Input.Password
                      prefix={<LockOutlined />}
                      placeholder="请输入当前密码"
                    />
                  </Form.Item>

                  <Form.Item
                    label="新密码"
                    name="new_password"
                    rules={[
                      { required: true, message: '请输入新密码' },
                      { min: 6, message: '密码至少6位' },
                    ]}
                  >
                    <Input.Password
                      prefix={<LockOutlined />}
                      placeholder="请输入新密码"
                    />
                  </Form.Item>

                  <Form.Item
                    label="确认新密码"
                    name="confirm_password"
                    rules={[
                      { required: true, message: '请确认新密码' },
                      ({ getFieldValue }) => ({
                        validator(_, value) {
                          if (!value || getFieldValue('new_password') === value) {
                            return Promise.resolve();
                          }
                          return Promise.reject(new Error('两次输入的密码不一致'));
                        },
                      }),
                    ]}
                  >
                    <Input.Password
                      prefix={<LockOutlined />}
                      placeholder="请再次输入新密码"
                    />
                  </Form.Item>

                  <Form.Item style={{ marginTop: 24 }}>
                    <Button
                      type="primary"
                      htmlType="submit"
                      icon={<SaveOutlined />}
                      loading={loading}
                      style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
                    >
                      修改密码
                    </Button>
                  </Form.Item>
                </Form>
              ),
            },
            {
              key: '3',
              label: '账号信息',
              children: (
                <Descriptions column={1} labelStyle={{ width: 120 }}>
                  <Descriptions.Item label="用户ID">{user?.id}</Descriptions.Item>
                  <Descriptions.Item label="角色">
                    <Tag color={getRoleColor(user?.role || '')}>{getRoleLabel(user?.role || '')}</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="账号状态">
                    <Tag color={user?.status === 'active' ? 'success' : 'error'}>
                      {user?.status === 'active' ? '正常' : '禁用'}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="所属组织">
                    {user?.org?.name || '未分配'}
                  </Descriptions.Item>
                  <Descriptions.Item label="注册时间">
                    {user?.created_at ? dayjs(user.created_at).format('YYYY-MM-DD HH:mm:ss') : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="最后登录">
                    {user?.last_login ? dayjs(user.last_login).format('YYYY-MM-DD HH:mm:ss') : '从未登录'}
                  </Descriptions.Item>
                </Descriptions>
              ),
            },
          ]}
        />
      </Card>
    </div>
  );
}
