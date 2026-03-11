import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Form, Input, Select, Spin, Row, Col, message, Space } from 'antd';
import { EditOutlined, LockOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import RoleTag from '@/components/RoleTag';
import StatusTag from '@/components/StatusTag';
import { getProfile, updateProfile } from '@/api/user';
import { formatDateTime } from '@/utils/format';
import { GENDER_MAP } from '@/utils/constants';
import type { UserProfileResponse } from '@/types';

export default function ProfilePage() {
  const navigate = useNavigate();
  const [profile, setProfile] = useState<UserProfileResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => {
    getProfile().then(setProfile).finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    const values = await form.validateFields();
    try {
      const res = await updateProfile(values);
      setProfile(res);
      setEditing(false);
      message.success('更新成功');
    } catch { /* interceptor */ }
  };

  if (loading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;
  if (!profile) return null;

  return (
    <PageContainer
      title="个人资料"
      breadcrumb={[{ title: '首页' }, { title: '个人资料' }]}
      extra={
        <Space>
          <Button icon={<LockOutlined />} onClick={() => navigate('/profile/change-password')}>修改密码</Button>
          {!editing && <Button type="primary" icon={<EditOutlined />} onClick={() => { form.setFieldsValue(profile); setEditing(true); }}>编辑</Button>}
        </Space>
      }
    >
      {!editing ? (
        <Card>
          <Descriptions bordered column={{ xs: 1, sm: 2 }}>
            <Descriptions.Item label="昵称">{profile.nickname}</Descriptions.Item>
            <Descriptions.Item label="手机号">{profile.phone}</Descriptions.Item>
            <Descriptions.Item label="邮箱">{profile.email || '-'}</Descriptions.Item>
            <Descriptions.Item label="角色"><RoleTag role={profile.role} /></Descriptions.Item>
            <Descriptions.Item label="状态"><StatusTag status={profile.status} /></Descriptions.Item>
            <Descriptions.Item label="组织">{profile.org_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="真实姓名">{profile.real_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="性别">{GENDER_MAP[profile.gender] || profile.gender || '-'}</Descriptions.Item>
            <Descriptions.Item label="身份证号">{profile.id_card || '-'}</Descriptions.Item>
            <Descriptions.Item label="地址">{profile.address || '-'}</Descriptions.Item>
            <Descriptions.Item label="紧急联系人">{profile.emergency || '-'}</Descriptions.Item>
            <Descriptions.Item label="紧急联系电话">{profile.emergency_tel || '-'}</Descriptions.Item>
            <Descriptions.Item label="个人简介" span={2}>{profile.introduction || '-'}</Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDateTime(profile.created_at)}</Descriptions.Item>
            <Descriptions.Item label="最后登录">{formatDateTime(profile.last_login_at)}</Descriptions.Item>
          </Descriptions>
        </Card>
      ) : (
        <Card>
          <Form form={form} layout="vertical" style={{ maxWidth: 600 }}>
            <Row gutter={16}>
              <Col xs={24} sm={12}><Form.Item name="nickname" label="昵称"><Input /></Form.Item></Col>
              <Col xs={24} sm={12}><Form.Item name="email" label="邮箱"><Input /></Form.Item></Col>
              <Col xs={24} sm={12}><Form.Item name="real_name" label="真实姓名"><Input /></Form.Item></Col>
              <Col xs={24} sm={12}>
                <Form.Item name="gender" label="性别">
                  <Select options={Object.entries(GENDER_MAP).map(([k, v]) => ({ label: v, value: k }))} allowClear />
                </Form.Item>
              </Col>
              <Col xs={24} sm={12}><Form.Item name="id_card" label="身份证号"><Input /></Form.Item></Col>
              <Col xs={24} sm={12}><Form.Item name="address" label="地址"><Input /></Form.Item></Col>
              <Col xs={24} sm={12}><Form.Item name="emergency" label="紧急联系人"><Input /></Form.Item></Col>
              <Col xs={24} sm={12}><Form.Item name="emergency_tel" label="紧急联系电话"><Input /></Form.Item></Col>
              <Col xs={24}><Form.Item name="introduction" label="个人简介"><Input.TextArea rows={3} /></Form.Item></Col>
            </Row>
            <Space>
              <Button type="primary" onClick={handleSave}>保存</Button>
              <Button onClick={() => setEditing(false)}>取消</Button>
            </Space>
          </Form>
        </Card>
      )}
    </PageContainer>
  );
}
