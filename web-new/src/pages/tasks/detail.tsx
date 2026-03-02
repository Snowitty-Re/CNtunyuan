import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Descriptions, Tag, Button, Space, Timeline, message, Modal, Progress } from 'antd';
import { ArrowLeftOutlined, EditOutlined, DeleteOutlined, CheckCircleOutlined } from '@ant-design/icons';
import type { Task } from '@/types';
import { http } from '@/utils/request';
import { usePermission } from '@/utils/permission';
import dayjs from 'dayjs';

export default function TaskDetailPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const { isManager } = usePermission();
  const [data, setData] = useState<Task | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDetail();
  }, [id]);

  const fetchDetail = async () => {
    try {
      const res: any = await http.get(`/tasks/${id}`);
      setData(res);
    } catch (error) {
      message.error('获取任务详情失败');
    } finally {
      setLoading(false);
    }
  };

  const handleComplete = () => {
    Modal.confirm({
      title: '确认完成任务',
      content: '确定要将此任务标记为完成吗？',
      okText: '完成',
      okType: 'primary',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.post(`/tasks/${id}/complete`);
          message.success('任务已完成');
          fetchDetail();
        } catch (error) {
          message.error('操作失败');
        }
      },
    });
  };

  const handleDelete = () => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除此任务吗？',
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await http.delete(`/tasks/${id}`);
          message.success('删除成功');
          navigate('/tasks');
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
          <p>任务不存在或已被删除</p>
          <Button onClick={() => navigate('/tasks')}>返回列表</Button>
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
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tasks')}>
              返回
            </Button>
            <span style={{ fontSize: 18, fontWeight: 600 }}>任务详情</span>
            <Tag color={getStatusColor(data?.status)} style={{ fontSize: 14, padding: '4px 12px' }}>
              {getStatusLabel(data?.status)}
            </Tag>
          </Space>
          {isManager && (
            <Space>
              {data?.status !== 'completed' && (
                <Button type="primary" icon={<CheckCircleOutlined />} onClick={handleComplete}>
                  标记完成
                </Button>
              )}
              <Button icon={<EditOutlined />} onClick={() => navigate(`/tasks/${id}/edit`)}>
                编辑
              </Button>
              <Button danger icon={<DeleteOutlined />} onClick={handleDelete}>
                删除
              </Button>
            </Space>
          )}
        </div>
      </Card>

      {/* 任务进度 */}
      <Card title="任务进度" style={{ marginBottom: 24 }} loading={loading}>
        <Progress percent={data?.progress || 0} strokeColor="#e67e22" trailColor="#f0f0f0" />
        <div style={{ marginTop: 16, display: 'flex', justifyContent: 'space-between', color: '#646a73' }}>
          <span>当前进度: {data?.progress || 0}%</span>
          <span>截止时间: {data?.deadline ? dayjs(data.deadline).format('YYYY-MM-DD HH:mm') : '无'}</span>
        </div>
      </Card>

      {/* 基本信息 */}
      <Card title="基本信息" style={{ marginBottom: 24 }} loading={loading}>
        <Descriptions column={2} labelStyle={{ fontWeight: 500, width: 100 }}>
          <Descriptions.Item label="任务标题">{data?.title}</Descriptions.Item>
          <Descriptions.Item label="任务编号">{data?.task_no}</Descriptions.Item>
          <Descriptions.Item label="优先级">
            <Tag color={getPriorityColor(data?.priority)}>{getPriorityLabel(data?.priority)}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="任务类型">{data?.type || '-'}</Descriptions.Item>
          <Descriptions.Item label="创建人">{data?.creator?.nickname || '-'}</Descriptions.Item>
          <Descriptions.Item label="执行人">{data?.assignee?.nickname || '未分配'}</Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {data?.created_at ? dayjs(data.created_at).format('YYYY-MM-DD HH:mm') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="截止时间">
            {data?.deadline ? dayjs(data.deadline).format('YYYY-MM-DD HH:mm') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="任务描述" span={2}>
            {data?.description || '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* 关联案件 */}
      {data?.missing_person && (
        <Card title="关联寻人案件" style={{ marginBottom: 24 }} loading={loading}>
          <Descriptions column={2}>
            <Descriptions.Item label="案件编号">{data.missing_person.case_no}</Descriptions.Item>
            <Descriptions.Item label="姓名">{data.missing_person.name}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={getCaseStatusColor(data.missing_person.status)}>
                {getCaseStatusLabel(data.missing_person.status)}
              </Tag>
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      {/* 任务日志 */}
      <Card title="任务日志" loading={loading}>
        <Timeline
          items={[
            {
              color: 'green',
              children: (
                <div>
                  <p style={{ fontWeight: 500 }}>任务创建</p>
                  <p style={{ color: '#8f959e', fontSize: 13 }}>
                    {data?.created_at ? dayjs(data.created_at).format('YYYY-MM-DD HH:mm:ss') : '-'}
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
    draft: 'default',
    pending: 'orange',
    assigned: 'blue',
    processing: 'cyan',
    completed: 'green',
    cancelled: 'gray',
  };
  return colors[status || ''] || 'default';
}

function getStatusLabel(status?: string) {
  const labels: Record<string, string> = {
    draft: '草稿',
    pending: '待分配',
    assigned: '已分配',
    processing: '进行中',
    completed: '已完成',
    cancelled: '已取消',
  };
  return labels[status || ''] || status || '-';
}

function getPriorityColor(priority?: string) {
  const colors: Record<string, string> = {
    urgent: 'red',
    high: 'orange',
    normal: 'blue',
    low: 'green',
  };
  return colors[priority || ''] || 'default';
}

function getPriorityLabel(priority?: string) {
  const labels: Record<string, string> = {
    urgent: '紧急',
    high: '高',
    normal: '普通',
    low: '低',
  };
  return labels[priority || ''] || priority || '-';
}

function getCaseStatusColor(status?: string) {
  const colors: Record<string, string> = {
    missing: 'red',
    searching: 'orange',
    found: 'blue',
    reunited: 'green',
    closed: 'gray',
  };
  return colors[status || ''] || 'default';
}

function getCaseStatusLabel(status?: string) {
  const labels: Record<string, string> = {
    missing: '失踪中',
    searching: '寻找中',
    found: '已找到',
    reunited: '已团圆',
    closed: '已结案',
  };
  return labels[status || ''] || status || '-';
}
