import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Button,
  Card,
  Form,
  Input,
  message,
  Radio,
  Steps,
  Typography,
  Alert,
  Spin,
} from 'antd';
import {
  DatabaseOutlined,
  UserOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import axios from 'axios';
import './style.css';

const { Title, Text } = Typography;

interface DBFormValues {
  db_type: 'postgres' | 'mysql';
  db_host: string;
  db_port: number;
  db_user: string;
  db_password: string;
  db_name: string;
  db_ssl_mode?: string;
  db_charset?: string;
}

interface AdminFormValues {
  admin_phone: string;
  admin_password: string;
  admin_nickname: string;
}

export default function SetupPage() {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [checking, setChecking] = useState(true);
  const [dbForm] = Form.useForm<DBFormValues>();
  const [adminForm] = Form.useForm<AdminFormValues>();
  const [dbConfig, setDbConfig] = useState<DBFormValues | null>(null);
  const apiUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

  // 检查系统是否已初始化
  useEffect(() => {
    checkStatus();
  }, []);

  const checkStatus = async () => {
    try {
      const res = await axios.get(`${apiUrl}/setup/status`);
      if (res.data.data?.initialized) {
        message.info('系统已初始化，跳转到登录页面');
        navigate('/login');
      }
    } catch (error) {
      console.error('检查状态失败:', error);
    } finally {
      setChecking(false);
    }
  };

  // 测试数据库连接
  const testDatabase = async () => {
    try {
      const values = await dbForm.validateFields();
      setLoading(true);
      
      // 转换字段名以匹配后端期望
      const payload = {
        type: values.db_type,
        host: values.db_host,
        port: values.db_port,
        user: values.db_user,
        password: values.db_password,
        database: values.db_name,
        ssl_mode: values.db_ssl_mode || 'disable',
        charset: values.db_charset || 'utf8mb4',
      };
      
      const res = await axios.post(`${apiUrl}/setup/test-db`, payload);
      
      if (res.data.code === 0 || res.data.code === 200) {
        message.success('数据库连接成功');
        setDbConfig(values);
        setCurrentStep(1);
      } else {
        message.error(res.data.message || '连接失败');
      }
    } catch (error: any) {
      console.error('测试连接失败:', error);
      message.error(error?.response?.data?.message || '数据库连接失败');
    } finally {
      setLoading(false);
    }
  };

  // 初始化系统
  const initialize = async () => {
    try {
      const adminValues = await adminForm.validateFields();
      setLoading(true);

      const res = await axios.post(`${apiUrl}/setup/initialize`, {
        ...dbConfig,
        ...adminValues,
      });

      if (res.data.code === 0 || res.data.code === 200) {
        message.success('系统初始化成功！请使用管理员账号登录');
        setCurrentStep(2);
        // 3秒后跳转到登录页
        setTimeout(() => {
          navigate('/login');
        }, 3000);
      } else {
        message.error(res.data.message || '初始化失败');
      }
    } catch (error: any) {
      console.error('初始化失败:', error);
      message.error(error?.response?.data?.message || '初始化失败');
    } finally {
      setLoading(false);
    }
  };

  const steps = [
    {
      title: '数据库配置',
      icon: <DatabaseOutlined />,
    },
    {
      title: '创建管理员',
      icon: <UserOutlined />,
    },
    {
      title: '完成',
      icon: <CheckCircleOutlined />,
    },
  ];

  const dbType = Form.useWatch('db_type', dbForm);

  if (checking) {
    return (
      <div className="setup-page">
        <div className="setup-loading">
          <Spin size="large" />
          <p>检查系统状态...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="setup-page">
      <div className="setup-container">
        <Card className="setup-card">
          <div className="setup-header">
            <Title level={2}>团圆寻亲系统初始化</Title>
            <Text type="secondary">欢迎使用团圆寻亲志愿者系统，请先完成初始化配置</Text>
          </div>

          <Steps current={currentStep} items={steps} className="setup-steps" />

          <div className="setup-content">
            {currentStep === 0 && (
              <Form
                form={dbForm}
                layout="vertical"
                initialValues={{
                  db_type: 'postgres',
                  db_host: 'localhost',
                  db_port: 5432,
                  db_name: 'cntuanyuan',
                }}
              >
                <Alert
                  message="数据库配置"
                  description="请配置系统数据库，支持 PostgreSQL 和 MySQL 8.0"
                  type="info"
                  showIcon
                  style={{ marginBottom: 24 }}
                />

                <Form.Item
                  name="db_type"
                  label="数据库类型"
                  rules={[{ required: true }]}
                >
                  <Radio.Group>
                    <Radio.Button value="postgres">PostgreSQL</Radio.Button>
                    <Radio.Button value="mysql">MySQL 8.0</Radio.Button>
                  </Radio.Group>
                </Form.Item>

                <div className="form-row">
                  <Form.Item
                    name="db_host"
                    label="主机地址"
                    rules={[{ required: true, message: '请输入主机地址' }]}
                    style={{ flex: 1 }}
                  >
                    <Input placeholder="localhost" />
                  </Form.Item>

                  <Form.Item
                    name="db_port"
                    label="端口"
                    rules={[{ required: true, message: '请输入端口' }]}
                    style={{ width: 120 }}
                  >
                    <Input type="number" placeholder={dbType === 'mysql' ? '3306' : '5432'} />
                  </Form.Item>
                </div>

                <div className="form-row">
                  <Form.Item
                    name="db_user"
                    label="用户名"
                    rules={[{ required: true, message: '请输入用户名' }]}
                    style={{ flex: 1 }}
                  >
                    <Input placeholder="postgres" />
                  </Form.Item>

                  <Form.Item
                    name="db_password"
                    label="密码"
                    rules={[{ required: true, message: '请输入密码' }]}
                    style={{ flex: 1 }}
                  >
                    <Input.Password placeholder="数据库密码" />
                  </Form.Item>
                </div>

                <Form.Item
                  name="db_name"
                  label="数据库名称"
                  rules={[{ required: true, message: '请输入数据库名称' }]}
                >
                  <Input placeholder="cntuanyuan" />
                </Form.Item>

                {dbType === 'postgres' && (
                  <Form.Item name="db_ssl_mode" label="SSL 模式">
                    <Radio.Group defaultValue="disable">
                      <Radio value="disable">禁用</Radio>
                      <Radio value="require">需要</Radio>
                      <Radio value="prefer">优先</Radio>
                    </Radio.Group>
                  </Form.Item>
                )}

                {dbType === 'mysql' && (
                  <Form.Item name="db_charset" label="字符集">
                    <Radio.Group defaultValue="utf8mb4">
                      <Radio value="utf8mb4">utf8mb4（推荐）</Radio>
                      <Radio value="utf8">utf8</Radio>
                    </Radio.Group>
                  </Form.Item>
                )}

                <Button
                  type="primary"
                  size="large"
                  block
                  loading={loading}
                  onClick={testDatabase}
                >
                  测试连接并继续
                </Button>
              </Form>
            )}

            {currentStep === 1 && (
              <Form
                form={adminForm}
                layout="vertical"
                initialValues={{
                  admin_nickname: '超级管理员',
                }}
              >
                <Alert
                  message="创建管理员账号"
                  description="请设置系统管理员账号，用于登录和管理系统"
                  type="info"
                  showIcon
                  style={{ marginBottom: 24 }}
                />

                <Form.Item
                  name="admin_phone"
                  label="手机号"
                  rules={[
                    { required: true, message: '请输入手机号' },
                    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确' },
                  ]}
                >
                  <Input placeholder="13800138000" />
                </Form.Item>

                <Form.Item
                  name="admin_nickname"
                  label="昵称"
                  rules={[{ required: true, message: '请输入昵称' }]}
                >
                  <Input placeholder="超级管理员" />
                </Form.Item>

                <Form.Item
                  name="admin_password"
                  label="密码"
                  rules={[
                    { required: true, message: '请输入密码' },
                    { min: 6, message: '密码至少6位' },
                  ]}
                >
                  <Input.Password placeholder="管理员密码" />
                </Form.Item>

                <div className="form-actions">
                  <Button onClick={() => setCurrentStep(0)}>上一步</Button>
                  <Button
                    type="primary"
                    loading={loading}
                    onClick={initialize}
                  >
                    完成初始化
                  </Button>
                </div>
              </Form>
            )}

            {currentStep === 2 && (
              <div className="setup-success">
                <CheckCircleOutlined style={{ fontSize: 64, color: '#52c41a' }} />
                <Title level={3}>初始化成功</Title>
                <Text>系统已初始化完成，即将跳转到登录页面...</Text>
              </div>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
}
