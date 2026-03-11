import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Spin, Space } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import RoleTag from '@/components/RoleTag';
import { getUser } from '@/api/user';
import { formatDateTime } from '@/utils/format';
import type { UserResponse } from '@/types';

export default function UserDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [user, setUser] = useState<UserResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    getUser(id)
      .then(setUser)
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;
  }

  if (!user) {
    return <div>用户不存在</div>;
  }

  return (
    <PageContainer
      title="用户详情"
      breadcrumb={[{ title: '首页' }, { title: '用户管理' }, { title: '用户详情' }]}
      extra={
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>
          返回
        </Button>
      }
    >
      <Card>
        <Descriptions column={{ xs: 1, sm: 2, md: 3 }} bordered>
          <Descriptions.Item label="昵称">{user.nickname}</Descriptions.Item>
          <Descriptions.Item label="手机号">{user.phone}</Descriptions.Item>
          <Descriptions.Item label="邮箱">{user.email || '-'}</Descriptions.Item>
          <Descriptions.Item label="角色"><RoleTag role={user.role} /></Descriptions.Item>
          <Descriptions.Item label="状态"><StatusTag status={user.status} /></Descriptions.Item>
          <Descriptions.Item label="所属组织">{user.org_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="最后登录">{formatDateTime(user.last_login_at)}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{formatDateTime(user.created_at)}</Descriptions.Item>
        </Descriptions>
      </Card>
    </PageContainer>
  );
}
