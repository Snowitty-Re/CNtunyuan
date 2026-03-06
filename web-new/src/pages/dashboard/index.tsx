import { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, Button, Tag, List, Progress, Empty, Avatar } from 'antd';
import {
  TeamOutlined,
  SearchOutlined,
  FileTextOutlined,
  SoundOutlined,
  ArrowRightOutlined,
  ClockCircleOutlined,
  UserOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';
import { usePermission } from '@/utils/permission';
import type { DashboardStats, Task, MissingPerson } from '@/types';
import { http } from '@/utils/request';

export default function DashboardPage() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { isManager } = usePermission();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [recentTasks, setRecentTasks] = useState<Task[]>([]);
  const [recentCases, setRecentCases] = useState<MissingPerson[]>([]);

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const statsRes: any = await http.get('/dashboard/stats');
      setStats(statsRes);

      const tasksRes: any = await http.get('/tasks?limit=5');
      setRecentTasks(tasksRes.list || []);

      const casesRes: any = await http.get('/missing-persons?limit=5');
      setRecentCases(casesRes.list || []);
    } catch (error) {
      console.error('获取仪表盘数据失败:', error);
    }
  };

  const quickActions = [
    {
      title: '发布寻人',
      icon: <PlusOutlined />,
      color: '#e67e22',
      path: '/cases/create',
    },
    {
      title: '创建任务',
      icon: <PlusOutlined />,
      color: '#27ae60',
      path: '/tasks/create',
      hide: !isManager,
    },
    {
      title: '录制方言',
      icon: <PlusOutlined />,
      color: '#8e44ad',
      path: '/dialects/create',
    },
  ].filter((item) => !item.hide);

  const statCards = [
    {
      title: '志愿者总数',
      value: stats?.total_volunteers || 0,
      icon: <TeamOutlined style={{ fontSize: 24, color: '#e67e22' }} />,
      bgColor: '#fdf2e9',
      trend: `+${stats?.new_volunteers || 0} 新增`,
    },
    {
      title: '寻人案件',
      value: stats?.total_cases || 0,
      icon: <SearchOutlined style={{ fontSize: 24, color: '#3498db' }} />,
      bgColor: '#ebf5fb',
      trend: `${stats?.resolved_cases || 0} 已解决`,
    },
    {
      title: '进行中的任务',
      value: (stats?.total_tasks || 0) - (stats?.completed_tasks || 0),
      icon: <FileTextOutlined style={{ fontSize: 24, color: '#27ae60' }} />,
      bgColor: '#eafaf1',
      trend: `${stats?.completed_tasks || 0} 已完成`,
    },
    {
      title: '方言样本',
      value: stats?.total_dialects || 0,
      icon: <SoundOutlined style={{ fontSize: 24, color: '#8e44ad' }} />,
      bgColor: '#f5eef8',
      trend: '持续收录',
    },
  ];

  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 6) return '夜深了';
    if (hour < 11) return '早上好';
    if (hour < 14) return '中午好';
    if (hour < 18) return '下午好';
    return '晚上好';
  };

  return (
    <div>
      {/* 欢迎区域 - 简洁温馨 */}
      <Card 
        style={{ 
          marginBottom: 24, 
          background: 'linear-gradient(135deg, #fdf2e9 0%, #fef9e7 100%)',
          border: 'none',
        }}
        bodyStyle={{ padding: 24 }}
      >
        <Row gutter={24} align="middle">
          <Col flex="auto">
            <h1 style={{ 
              margin: '0 0 8px 0', 
              fontSize: 24, 
              fontWeight: 600, 
              color: '#1f2329' 
            }}>
              {getGreeting()}，{user?.nickname || user?.real_name || '志愿者'}！
            </h1>
            <p style={{ margin: 0, color: '#646a73', fontSize: 14 }}>
              今天是 {new Date().toLocaleDateString('zh-CN', { 
                year: 'numeric', 
                month: 'long', 
                day: 'numeric', 
                weekday: 'long' 
              })}
            </p>
          </Col>
          <Col>
            <div style={{ display: 'flex', gap: 12 }}>
              {quickActions.map((action) => (
                <Button
                  key={action.title}
                  type="primary"
                  icon={action.icon}
                  onClick={() => navigate(action.path)}
                  style={{ 
                    backgroundColor: action.color, 
                    borderColor: action.color,
                    height: 36,
                    borderRadius: 6,
                    fontWeight: 500,
                  }}
                >
                  {action.title}
                </Button>
              ))}
            </div>
          </Col>
        </Row>
      </Card>

      {/* 统计卡片 - 简洁干净 */}
      <Row gutter={[24, 24]} style={{ marginBottom: 24 }}>
        {statCards.map((card, index) => (
          <Col xs={24} sm={12} lg={6} key={index}>
            <Card 
              variant="borderless"
              bodyStyle={{ padding: 20 }}
              style={{ 
                borderRadius: 8,
                transition: 'all 0.3s',
                cursor: 'pointer',
              }}
              hoverable
            >
              <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
                <div>
                  <p style={{ 
                    margin: '0 0 8px 0', 
                    color: '#646a73', 
                    fontSize: 14 
                  }}>
                    {card.title}
                  </p>
                  <Statistic
                    value={card.value}
                    valueStyle={{
                      fontSize: 32,
                      fontWeight: 700,
                      color: '#1f2329',
                    }}
                  />
                  <Tag 
                    style={{ 
                      marginTop: 8, 
                      fontSize: 12,
                      background: card.bgColor,
                      border: 'none',
                      color: '#646a73',
                    }}
                  >
                    {card.trend}
                  </Tag>
                </div>
                <div
                  style={{ 
                    width: 48, 
                    height: 48, 
                    borderRadius: 8, 
                    background: card.bgColor,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  {card.icon}
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* 主体内容 */}
      <Row gutter={[24, 24]}>
        {/* 左侧：最近任务 */}
        <Col xs={24} lg={16}>
          <Card
            title="最近任务"
            extra={
              <Button
                type="link"
                onClick={() => navigate('/tasks')}
                style={{ color: '#e67e22', fontWeight: 500 }}
              >
                查看全部 <ArrowRightOutlined />
              </Button>
            }
            bordered={false}
            bodyStyle={{ padding: 0 }}
          >
            {recentTasks.length > 0 ? (
              <List
                dataSource={recentTasks}
                renderItem={(task) => (
                  <List.Item
                    style={{ 
                      padding: '16px 24px', 
                      cursor: 'pointer',
                      borderBottom: '1px solid #f0f0f0',
                    }}
                    onClick={() => navigate(`/tasks/${task.id}`)}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.backgroundColor = '#fafafa';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                    }}
                  >
                    <List.Item.Meta
                      title={
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
                          <span style={{ fontWeight: 500, color: '#1f2329' }}>{task.title}</span>
                          <Tag 
                            color={getPriorityColor(task.priority)}
                            style={{ fontSize: 12, padding: '0 8px' }}
                          >
                            {getPriorityLabel(task.priority)}
                          </Tag>
                        </div>
                      }
                      description={
                        <div style={{ display: 'flex', alignItems: 'center', gap: 16, fontSize: 13, color: '#646a73' }}>
                          <span>创建者: {task.creator?.nickname || '未知'}</span>
                          <span>·</span>
                          <span>
                            <ClockCircleOutlined style={{ marginRight: 4 }} />
                            {task.deadline
                              ? new Date(task.deadline).toLocaleDateString()
                              : '无截止日期'}
                          </span>
                        </div>
                      }
                    />
                    <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                      <Progress
                        percent={task.progress}
                        size="small"
                        style={{ width: 80 }}
                        strokeColor="#e67e22"
                      />
                      <Tag 
                        color={getStatusColor(task.status)}
                        style={{ fontSize: 12, minWidth: 64, textAlign: 'center' }}
                      >
                        {getStatusLabel(task.status)}
                      </Tag>
                    </div>
                  </List.Item>
                )}
              />
            ) : (
              <Empty 
                description="暂无任务" 
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                style={{ padding: '60px 0' }}
              />
            )}
          </Card>
        </Col>

        {/* 右侧：最近案件 */}
        <Col xs={24} lg={8}>
          <Card
            title="最近案件"
            extra={
              <Button
                type="link"
                onClick={() => navigate('/cases')}
                style={{ color: '#e67e22', fontWeight: 500 }}
              >
                更多 <ArrowRightOutlined />
              </Button>
            }
            bordered={false}
            bodyStyle={{ padding: 0 }}
          >
            {recentCases.length > 0 ? (
              <List
                dataSource={recentCases}
                renderItem={(caseItem) => (
                  <List.Item
                    style={{ 
                      padding: '12px 24px', 
                      cursor: 'pointer',
                      borderBottom: '1px solid #f0f0f0',
                    }}
                    onClick={() => navigate(`/cases/${caseItem.id}`)}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.backgroundColor = '#fafafa';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.backgroundColor = 'transparent';
                    }}
                  >
                    <List.Item.Meta
                      avatar={
                        <Avatar
                          size={40}
                          icon={<UserOutlined />}
                          style={{ backgroundColor: '#e67e22' }}
                          src={caseItem.photos?.[0]?.url}
                        />
                      }
                      title={
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <span style={{ fontWeight: 500, color: '#1f2329' }}>{caseItem.name}</span>
                          <Tag 
                            color={caseItem.gender === 'male' ? 'blue' : 'pink'}
                            style={{ fontSize: 12 }}
                          >
                            {caseItem.gender === 'male' ? '男' : '女'}
                          </Tag>
                        </div>
                      }
                      description={
                        <div style={{ fontSize: 13, color: '#646a73' }}>
                          <div>年龄: {caseItem.age || '未知'}岁</div>
                          <div style={{ 
                            overflow: 'hidden', 
                            textOverflow: 'ellipsis', 
                            whiteSpace: 'nowrap',
                            maxWidth: 200,
                          }}>
                            {caseItem.missing_location}
                          </div>
                        </div>
                      }
                    />
                    <Tag 
                      color={getCaseStatusColor(caseItem.status)}
                      style={{ fontSize: 12 }}
                    >
                      {getCaseStatusLabel(caseItem.status)}
                    </Tag>
                  </List.Item>
                )}
              />
            ) : (
              <Empty 
                description="暂无案件" 
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                style={{ padding: '60px 0' }}
              />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
}

// 辅助函数
function getPriorityColor(priority: string) {
  const colors: Record<string, string> = {
    urgent: 'red',
    high: 'orange',
    normal: 'blue',
    low: 'green',
  };
  return colors[priority] || 'default';
}

function getPriorityLabel(priority: string) {
  const labels: Record<string, string> = {
    urgent: '紧急',
    high: '高',
    normal: '普通',
    low: '低',
  };
  return labels[priority] || priority;
}

function getStatusColor(status: string) {
  const colors: Record<string, string> = {
    draft: 'default',
    pending: 'orange',
    assigned: 'blue',
    processing: 'cyan',
    completed: 'green',
    cancelled: 'gray',
  };
  return colors[status] || 'default';
}

function getStatusLabel(status: string) {
  const labels: Record<string, string> = {
    draft: '草稿',
    pending: '待分配',
    assigned: '已分配',
    processing: '进行中',
    completed: '已完成',
    cancelled: '已取消',
  };
  return labels[status] || status;
}

function getCaseStatusColor(status: string) {
  const colors: Record<string, string> = {
    missing: 'red',
    searching: 'orange',
    found: 'blue',
    reunited: 'green',
    closed: 'gray',
  };
  return colors[status] || 'default';
}

function getCaseStatusLabel(status: string) {
  const labels: Record<string, string> = {
    missing: '失踪中',
    searching: '寻找中',
    found: '已找到',
    reunited: '已团圆',
    closed: '已结案',
  };
  return labels[status] || status;
}
