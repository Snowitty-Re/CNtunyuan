import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Descriptions, Tag, Button, Space, Avatar, Timeline, message, Modal, Statistic, Row, Col } from 'antd';
import { ArrowLeftOutlined, EditOutlined, DeleteOutlined, UserOutlined, PhoneOutlined, MailOutlined } from '@ant-design/icons';
import type { User } from '@/types';
import { http } from '@/utils/request';
import { usePermission } from '@/utils/permission';
import dayjs from 'dayjs';

export default function VolunteerDetailPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { isManager } = usePermission();
  const [data, setData] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({
    taskCount: 0,
    completedTasks: 0,
    caseCount: 0,
  });

  useEffect(() => {
    fetchDetail();
    fetchStats();
  }, [id]);

  const fetchDetail = async () => {
    try {
      const res: any = await http.get(`/users/${id}`);
      setData(res);
    } catch (error) {
      message.error('获取志愿者详情失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const res: any = await http.get(`/dashboard/stats`);
      setStats({
        taskCount: res.completed_tasks || 0,
        completedTasks: res.completed_tasks || 0,
        caseCount: 0,
      });
    } catch (error) {
      console.error('获取统计失败', error);
    }
  };

  const handleDelete = () => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除此志愿者吗？此操作不可恢复。',
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.delete(`/users/${id}`);
          message.success('删除成功');
          navigate('/volunteers');
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  if (!data && !loading) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: 60 }}>
          <p>志愿者不存在或已被删除</p>
          <Button onClick={() => navigate('/volunteers')}>返回列表</Button>
        </div>
      </Card>
    );
  }

  return (
    <div>
      {/* 顶部操作栏 */}
      <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: 16 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/volunteers')}>
              返回
            </Button>
            <span style={{ fontSize: 18, fontWeight: 600 }}>志愿者详情</span>
          </Space>
          {isManager && (
            <Space>
              <Button icon={<EditOutlined />} onClick={() => navigate(`/volunteers/${id}/edit`)}>
                编辑
              </Button>
              <Button danger icon={<DeleteOutlined />} onClick={handleDelete}>
                删除
              </Button>
            </Space>
          )}
        </div>
      </Card>

      {/* 基本信息 */}
      <Card style={{ marginBottom: 24 }} loading={loading}>
        <div style={{ display: 'flex', gap: 24, marginBottom: 24 }}>
          <Avatar
            size={100}
            icon={<UserOutlined />}
            src={data?.avatar}
            style={{ backgroundColor: '#e67e22' }}
          />
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 24, fontWeight: 600, marginBottom: 8 }}>
              {data?.nickname}
              <Tag color={getRoleColor(data?.role)} style={{ marginLeft: 12, fontSize: 13 }}>
                {getRoleLabel(data?.role)}
              </Tag>
            </div>
            <div style={{ color: '#646a73', marginBottom: 8 }}>
              {data?.real_name || '未填写真实姓名'}
            </div>
            <div style={{ display: 'flex', gap: 24, color: '#8f959e' }}>
              <span><PhoneOutlined style={{ marginRight: 8 }} />{data?.phone || '-'}</span>
              <span><MailOutlined style={{ marginRight: 8 }} />{data?.email || '-'}</span>
            </div>
          </div>
        </div>

        <Row gutter={24}>
          <Col span={8}>
            <Card bordered bodyStyle={{ padding: 16 }}>
              <Statistic title="参与任务" value={stats.taskCount} />
            </Card>
          </Col>
          <Col span={8}>
            <Card bordered bodyStyle={{ padding: 16 }}>
              <Statistic title="已完成任务" value={stats.completedTasks} />
            </Card>
          </Col>
          <Col span={8}>
            <Card bordered bodyStyle={{ padding: 16 }}>
              <Statistic title="参与案件" value={stats.caseCount} />
            </Card>
          </Col>
        </Row>
      </Card>

      {/* 详细信息 */}
      <Card title="详细信息" style={{ marginBottom: 24 }} loading={loading}>
        <Descriptions column={2} labelStyle={{ fontWeight: 500, width: 100 }}>
          <Descriptions.Item label="所属组织">{data?.org?.name || '未分配'}</Descriptions.Item>
          <Descriptions.Item label="账号状态">
            <Tag color={data?.status === 'active' ? 'success' : 'error'}>
              {data?.status === 'active' ? '正常' : '禁用'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="注册时间">
            {data?.created_at ? dayjs(data.created_at).format('YYYY-MM-DD HH:mm') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="最后登录">
            {data?.last_login ? dayjs(data.last_login).format('YYYY-MM-DD HH:mm') : '从未登录'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 操作记录 */}
      <Card title="操作记录" loading={loading}>
        <Timeline
          items={[
            {
              color: 'green',
              children: (
                <div>
                  <p style={{ fontWeight: 500 }}>账号注册</p>
                  <p style={{ color: '#8f959e', fontSize: 13 }}>
                    {data?.created_at ? dayjs(data.created_at).format('YYYY-MM-DD HH:mm:ss') : '-'}
                  </p>
                </div>
              ),
            },
            {
              color: 'blue',
              children: (
                <div>
                  <p style={{ fontWeight: 500 }}>最后更新</p>
                  <p style={{ color: '#8f959e', fontSize: 13 }}>
                    {data?.updated_at ? dayjs(data.updated_at).format('YYYY-MM-DD HH:mm:ss') : '-'}
                  </p>
                </div>
              ),
            },
          ]}
        />
      </Card>
    </div>
  );
}

function getRoleColor(role?: string) {
  const colors: Record<string, string> = {
    super_admin: 'red',
    admin: 'orange',
    manager: 'blue',
    volunteer: 'green',
  };
  return colors[role || ''] || 'default';
}

function getRoleLabel(role?: string) {
  const labels: Record<string, string> = {
    super_admin: '超级管理员',
    admin: '管理员',
    manager: '组织者',
    volunteer: '志愿者',
  };
  return labels[role || ''] || role || '-';
}
