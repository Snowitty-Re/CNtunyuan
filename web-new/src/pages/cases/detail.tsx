import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Descriptions, Tag, Button, Space, Avatar, Timeline, message, Modal } from 'antd';
import { ArrowLeftOutlined, EditOutlined, DeleteOutlined, UserOutlined, EnvironmentOutlined, PhoneOutlined } from '@ant-design/icons';
import type { MissingPerson } from '@/types';
import { http } from '@/utils/request';
import { usePermission } from '@/utils/permission';
import dayjs from 'dayjs';

export default function CaseDetailPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { isManager } = usePermission();
  const [data, setData] = useState<MissingPerson | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDetail();
  }, [id]);

  const fetchDetail = async () => {
    try {
      const res: any = await http.get(`/missing-persons/${id}`);
      setData(res);
    } catch (error) {
      message.error('获取案件详情失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = () => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除此案件吗？此操作不可恢复。',
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.delete(`/missing-persons/${id}`);
          message.success('删除成功');
          navigate('/cases');
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
          <p>案件不存在或已被删除</p>
          <Button onClick={() => navigate('/cases')}>返回列表</Button>
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
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cases')}>
              返回
            </Button>
            <span style={{ fontSize: 18, fontWeight: 600 }}>案件详情</span>
            <Tag color={getStatusColor(data?.status)} style={{ fontSize: 14, padding: '4px 12px' }}>
              {getStatusLabel(data?.status)}
            </Tag>
          </Space>
          {isManager && (
            <Space>
              <Button icon={<EditOutlined />} onClick={() => navigate(`/cases/${id}/edit`)}>
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
      <Card title="基本信息" style={{ marginBottom: 24 }} loading={loading}>
        <div style={{ display: 'flex', gap: 24 }}>
          <Avatar
            size={120}
            icon={<UserOutlined />}
            src={data?.photos?.[0]?.url}
            style={{ backgroundColor: '#e67e22', flexShrink: 0 }}
          />
          <div style={{ flex: 1 }}>
            <Descriptions column={2} labelStyle={{ fontWeight: 500, width: 100 }}>
              <Descriptions.Item label="姓名">{data?.name}</Descriptions.Item>
              <Descriptions.Item label="案件编号">{data?.case_no}</Descriptions.Item>
              <Descriptions.Item label="性别">
                <Tag color={data?.gender === 'male' ? 'blue' : 'pink'}>
                  {data?.gender === 'male' ? '男' : '女'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="年龄">{data?.age ? `${data.age}岁` : '未知'}</Descriptions.Item>
              <Descriptions.Item label="案件类型">
                <Tag color="orange">{getCaseTypeLabel(data?.case_type)}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="走失时间">
                {data?.missing_time ? dayjs(data.missing_time).format('YYYY-MM-DD HH:mm') : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="走失地点" span={2}>
                <EnvironmentOutlined style={{ marginRight: 8, color: '#e67e22' }} />
                {data?.missing_location || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="体貌特征" span={2}>
                {data?.appearance || '-'}
              </Descriptions.Item>
            </Descriptions>
          </div>
        </div>
      </Card>

      {/* 联系人信息 */}
      <Card title="联系人信息" style={{ marginBottom: 24 }} loading={loading}>
        <Descriptions column={2} labelStyle={{ fontWeight: 500, width: 100 }}>
          <Descriptions.Item label="联系人">{data?.contact_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="联系电话">
            {data?.contact_phone ? (
              <span>
                <PhoneOutlined style={{ marginRight: 8, color: '#52c41a' }} />
                {data.contact_phone}
              </span>
            ) : '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 案件进展 */}
      <Card title="案件进展" loading={loading}>
        <Timeline
          items={[
            {
              color: 'green',
              children: (
                <div>
                  <p style={{ fontWeight: 500 }}>案件创建</p>
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
                  <p style={{ fontWeight: 500 }}>最新更新</p>
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

function getStatusColor(status?: string) {
  const colors: Record<string, string> = {
    missing: 'red',
    searching: 'orange',
    found: 'blue',
    reunited: 'green',
    closed: 'gray',
  };
  return colors[status || ''] || 'default';
}

function getStatusLabel(status?: string) {
  const labels: Record<string, string> = {
    missing: '失踪中',
    searching: '寻找中',
    found: '已找到',
    reunited: '已团圆',
    closed: '已结案',
  };
  return labels[status || ''] || status || '-';
}

function getCaseTypeLabel(type?: string) {
  const labels: Record<string, string> = {
    elderly: '老人走失',
    child: '儿童走失',
    adult: '成人走失',
    disability: '残障人士走失',
    other: '其他',
  };
  return labels[type || ''] || type || '-';
}
