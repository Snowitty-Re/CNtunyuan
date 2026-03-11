import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Spin, Space, Timeline, Progress, Modal, Input, InputNumber, message, Row, Col } from 'antd';
import { ArrowLeftOutlined, EditOutlined, PlayCircleOutlined, CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import PageContainer from '@/components/PageContainer';
import StatusTag from '@/components/StatusTag';
import UserSelect from '@/components/UserSelect';
import ConfirmButton from '@/components/ConfirmButton';
import { getTask, getTaskLogs, assignTask, startTask, completeTask, cancelTask, updateTaskProgress, deleteTask } from '@/api/task';
import { usePermission } from '@/hooks/usePermission';
import { formatDateTime } from '@/utils/format';
import { TASK_PRIORITY_MAP } from '@/utils/constants';
import type { TaskResponse, TaskLogResponse } from '@/types';

export default function TaskDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isManager } = usePermission();
  const [task, setTask] = useState<TaskResponse | null>(null);
  const [logs, setLogs] = useState<TaskLogResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [assignModal, setAssignModal] = useState(false);
  const [assigneeId, setAssigneeId] = useState('');
  const [progressModal, setProgressModal] = useState(false);
  const [progressVal, setProgressVal] = useState(0);
  const [completeModal, setCompleteModal] = useState(false);
  const [resultText, setResultText] = useState('');
  const [cancelModal, setCancelModal] = useState(false);
  const [cancelReason, setCancelReason] = useState('');

  const reload = async () => {
    if (!id) return;
    const [t, l] = await Promise.all([getTask(id), getTaskLogs(id)]);
    setTask(t);
    setLogs(l || []);
  };

  useEffect(() => {
    reload().finally(() => setLoading(false));
  }, [id]);

  const handleAssign = async () => {
    if (!id || !assigneeId) return;
    await assignTask(id, { assignee_id: assigneeId });
    message.success('分配成功');
    setAssignModal(false);
    reload();
  };

  const handleStart = async () => {
    if (!id) return;
    await startTask(id);
    message.success('任务已开始');
    reload();
  };

  const handleComplete = async () => {
    if (!id) return;
    await completeTask(id, { result: resultText });
    message.success('任务已完成');
    setCompleteModal(false);
    reload();
  };

  const handleCancel = async () => {
    if (!id) return;
    await cancelTask(id, { reason: cancelReason });
    message.success('任务已取消');
    setCancelModal(false);
    reload();
  };

  const handleProgress = async () => {
    if (!id) return;
    await updateTaskProgress(id, { progress: progressVal });
    message.success('进度已更新');
    setProgressModal(false);
    reload();
  };

  const handleDelete = async () => {
    if (!id) return;
    await deleteTask(id);
    message.success('删除成功');
    navigate('/tasks');
  };

  if (loading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;
  if (!task) return <div>任务不存在</div>;

  const canStart = task.status === 'assigned';
  const canComplete = task.status === 'processing';
  const canAssign = isManager && ['draft', 'pending'].includes(task.status);
  const canCancel = isManager && !['completed', 'cancelled'].includes(task.status);

  return (
    <PageContainer
      title={`任务详情 - ${task.title}`}
      breadcrumb={[{ title: '首页' }, { title: '任务管理' }, { title: '详情' }]}
      extra={
        <Space wrap>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回</Button>
          <Button icon={<EditOutlined />} onClick={() => navigate(`/tasks/${id}/edit`)}>编辑</Button>
          {canAssign && <Button type="primary" onClick={() => setAssignModal(true)}>分配</Button>}
          {canStart && <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleStart}>开始</Button>}
          {canComplete && <Button type="primary" icon={<CheckCircleOutlined />} onClick={() => setCompleteModal(true)}>完成</Button>}
          {task.status === 'processing' && <Button onClick={() => { setProgressVal(task.progress); setProgressModal(true); }}>更新进度</Button>}
          {canCancel && <Button danger icon={<CloseCircleOutlined />} onClick={() => setCancelModal(true)}>取消</Button>}
          {isManager && <ConfirmButton danger title="确定删除?" onConfirm={handleDelete}>删除</ConfirmButton>}
        </Space>
      }
    >
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={16}>
          <Card title="任务信息" style={{ marginBottom: 16 }}>
            <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
              <Descriptions.Item label="标题">{task.title}</Descriptions.Item>
              <Descriptions.Item label="类型">{task.type}</Descriptions.Item>
              <Descriptions.Item label="优先级"><StatusTag status={task.priority} map={TASK_PRIORITY_MAP} /></Descriptions.Item>
              <Descriptions.Item label="状态"><StatusTag status={task.status} /></Descriptions.Item>
              <Descriptions.Item label="进度"><Progress percent={task.progress} size="small" /></Descriptions.Item>
              <Descriptions.Item label="截止日期">{formatDateTime(task.deadline)}</Descriptions.Item>
              <Descriptions.Item label="创建者">{task.creator?.nickname || '-'}</Descriptions.Item>
              <Descriptions.Item label="负责人">{task.assignee?.nickname || '未分配'}</Descriptions.Item>
              <Descriptions.Item label="地点" span={2}>
                {[task.province, task.city, task.district, task.address].filter(Boolean).join(' ') || '-'}
              </Descriptions.Item>
              <Descriptions.Item label="描述" span={2}>{task.description || '-'}</Descriptions.Item>
              {task.result && <Descriptions.Item label="结果" span={2}>{task.result}</Descriptions.Item>}
              {task.missing_person && (
                <Descriptions.Item label="关联案件" span={2}>
                  <Button type="link" onClick={() => navigate(`/missing-persons/${task.missing_person_id}`)}>
                    {task.missing_person.name} ({task.missing_person.case_no})
                  </Button>
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="操作日志">
            {logs.length > 0 ? (
              <Timeline
                items={logs.map((l) => ({
                  children: (
                    <div>
                      <div><strong>{l.action}</strong> {l.content}</div>
                      {l.old_status && <div style={{ fontSize: 12, color: '#999' }}>{l.old_status} → {l.new_status}</div>}
                      <div style={{ fontSize: 12, color: '#999' }}>
                        {l.user?.nickname || '-'} | {formatDateTime(l.created_at)}
                      </div>
                    </div>
                  ),
                }))}
              />
            ) : (
              <div style={{ textAlign: 'center', color: '#999', padding: 20 }}>暂无日志</div>
            )}
          </Card>
        </Col>
      </Row>

      <Modal title="分配任务" open={assignModal} onOk={handleAssign} onCancel={() => setAssignModal(false)}>
        <UserSelect value={assigneeId} onChange={setAssigneeId} placeholder="选择负责人" style={{ width: '100%' }} />
      </Modal>
      <Modal title="更新进度" open={progressModal} onOk={handleProgress} onCancel={() => setProgressModal(false)}>
        <InputNumber min={0} max={100} value={progressVal} onChange={(v) => setProgressVal(v || 0)} addonAfter="%" style={{ width: '100%' }} />
      </Modal>
      <Modal title="完成任务" open={completeModal} onOk={handleComplete} onCancel={() => setCompleteModal(false)}>
        <Input.TextArea value={resultText} onChange={(e) => setResultText(e.target.value)} rows={3} placeholder="任务结果（可选）" />
      </Modal>
      <Modal title="取消任务" open={cancelModal} onOk={handleCancel} onCancel={() => setCancelModal(false)}>
        <Input.TextArea value={cancelReason} onChange={(e) => setCancelReason(e.target.value)} rows={3} placeholder="取消原因（可选）" />
      </Modal>
    </PageContainer>
  );
}
