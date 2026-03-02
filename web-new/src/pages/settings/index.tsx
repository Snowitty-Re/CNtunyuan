import { useState, useEffect } from 'react';
import { Card, Form, Input, Button, Switch, Select, InputNumber, Tabs, message, Divider, Alert } from 'antd';
import { SaveOutlined, SettingOutlined, SafetyOutlined, BellOutlined, DatabaseOutlined } from '@ant-design/icons';
import { http } from '@/utils/request';

const { TabPane } = Tabs;

export default function SettingsPage() {
  const [generalForm] = Form.useForm();
  const [securityForm] = Form.useForm();
  const [storageForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('1');

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const res: any = await http.get('/configs');
      generalForm.setFieldsValue(res.general || {});
      securityForm.setFieldsValue(res.security || {});
      storageForm.setFieldsValue(res.storage || {});
    } catch (error) {
      // 使用默认配置
    }
  };

  const handleSaveGeneral = async (values: any) => {
    setLoading(true);
    try {
      await http.post('/configs', { type: 'general', data: values });
      message.success('保存成功');
    } catch (error) {
      message.error('保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveSecurity = async (values: any) => {
    setLoading(true);
    try {
      await http.post('/configs', { type: 'security', data: values });
      message.success('保存成功');
    } catch (error) {
      message.error('保存失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSaveStorage = async (values: any) => {
    setLoading(true);
    try {
      await http.post('/configs', { type: 'storage', data: values });
      message.success('保存成功');
    } catch (error) {
      message.error('保存失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Card style={{ marginBottom: 24 }}>
        <div style={{ fontSize: 18, fontWeight: 600 }}>
          <SettingOutlined style={{ marginRight: 8 }} />
          系统设置
        </div>
        <div style={{ color: '#8f959e', marginTop: 8 }}>
          配置系统参数，管理全局设置
        </div>
      </Card>

      <Card>
        <Tabs activeKey={activeTab} onChange={setActiveTab}>
          <TabPane
            tab={
              <span>
                <SettingOutlined />
                基础设置
              </span>
            }
            key="1"
          >
            <Form
              form={generalForm}
              layout="vertical"
              onFinish={handleSaveGeneral}
              autoComplete="off"
              style={{ maxWidth: 600 }}
              initialValues={{
                site_name: '团圆寻亲志愿者系统',
                site_description: '帮助寻找走失人员的公益平台',
                enable_registration: true,
                enable_notification: true,
              }}
            >
              <Form.Item
                label="系统名称"
                name="site_name"
                rules={[{ required: true, message: '请输入系统名称' }]}
              >
                <Input placeholder="请输入系统名称" />
              </Form.Item>

              <Form.Item
                label="系统描述"
                name="site_description"
              >
                <Input.TextArea
                  rows={3}
                  placeholder="请输入系统描述"
                />
              </Form.Item>

              <Form.Item
                label="系统Logo URL"
                name="site_logo"
              >
                <Input placeholder="请输入Logo URL" />
              </Form.Item>

              <Form.Item
                label="客服电话"
                name="service_phone"
                rules={[{ pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' }]}
              >
                <Input placeholder="请输入客服电话" />
              </Form.Item>

              <Form.Item
                label="客服邮箱"
                name="service_email"
                rules={[{ type: 'email', message: '请输入正确的邮箱' }]}
              >
                <Input placeholder="请输入客服邮箱" />
              </Form.Item>

              <Divider />

              <Form.Item
                label="功能开关"
                name="enable_registration"
                valuePropName="checked"
              >
                <Switch checkedChildren="开启" unCheckedChildren="关闭" />
              </Form.Item>

              <Form.Item
                label="消息通知"
                name="enable_notification"
                valuePropName="checked"
              >
                <Switch checkedChildren="开启" unCheckedChildren="关闭" />
              </Form.Item>

              <Form.Item style={{ marginTop: 24 }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  icon={<SaveOutlined />}
                  loading={loading}
                  style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
                >
                  保存设置
                </Button>
              </Form.Item>
            </Form>
          </TabPane>

          <TabPane
            tab={
              <span>
                <SafetyOutlined />
                安全设置
              </span>
            }
            key="2"
          >
            <Form
              form={securityForm}
              layout="vertical"
              onFinish={handleSaveSecurity}
              autoComplete="off"
              style={{ maxWidth: 600 }}
              initialValues={{
                login_max_attempts: 5,
                password_min_length: 6,
                session_timeout: 7,
                enable_captcha: true,
                enable_2fa: false,
              }}
            >
              <Alert
                message="安全提示"
                description="合理配置安全参数可以有效保护系统安全"
                type="info"
                showIcon
                style={{ marginBottom: 24 }}
              />

              <Form.Item
                label="登录最大尝试次数"
                name="login_max_attempts"
                rules={[{ required: true }]}
              >
                <InputNumber
                  min={3}
                  max={10}
                  style={{ width: '100%' }}
                  placeholder="登录失败多少次后锁定账号"
                />
              </Form.Item>

              <Form.Item
                label="密码最小长度"
                name="password_min_length"
                rules={[{ required: true }]}
              >
                <InputNumber
                  min={6}
                  max={20}
                  style={{ width: '100%' }}
                  placeholder="密码最小长度"
                />
              </Form.Item>

              <Form.Item
                label="会话超时（天）"
                name="session_timeout"
                rules={[{ required: true }]}
              >
                <InputNumber
                  min={1}
                  max={30}
                  style={{ width: '100%' }}
                  placeholder="用户登录会话保持天数"
                />
              </Form.Item>

              <Divider />

              <Form.Item
                label="启用验证码"
                name="enable_captcha"
                valuePropName="checked"
              >
                <Switch checkedChildren="开启" unCheckedChildren="关闭" />
              </Form.Item>

              <Form.Item
                label="启用双因素认证"
                name="enable_2fa"
                valuePropName="checked"
              >
                <Switch checkedChildren="开启" unCheckedChildren="关闭" />
              </Form.Item>

              <Form.Item style={{ marginTop: 24 }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  icon={<SaveOutlined />}
                  loading={loading}
                  style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
                >
                  保存设置
                </Button>
              </Form.Item>
            </Form>
          </TabPane>

          <TabPane
            tab={
              <span>
                <DatabaseOutlined />
                存储设置
              </span>
            }
            key="3"
          >
            <Form
              form={storageForm}
              layout="vertical"
              onFinish={handleSaveStorage}
              autoComplete="off"
              style={{ maxWidth: 600 }}
              initialValues={{
                storage_type: 'local',
                max_file_size: 50,
                allowed_image_types: 'jpg,png,gif',
                allowed_audio_types: 'mp3,wav',
                allowed_video_types: 'mp4',
              }}
            >
              <Form.Item
                label="存储类型"
                name="storage_type"
                rules={[{ required: true }]}
              >
                <Select placeholder="请选择存储类型">
                  <Select.Option value="local">本地存储</Select.Option>
                  <Select.Option value="oss">阿里云OSS</Select.Option>
                  <Select.Option value="cos">腾讯云COS</Select.Option>
                </Select>
              </Form.Item>

              <Form.Item
                label="最大文件大小（MB）"
                name="max_file_size"
                rules={[{ required: true }]}
              >
                <InputNumber
                  min={1}
                  max={500}
                  style={{ width: '100%' }}
                  placeholder="单个文件最大上传大小"
                />
              </Form.Item>

              <Divider />

              <Form.Item
                label="允许的图片格式"
                name="allowed_image_types"
              >
                <Input placeholder="如：jpg,png,gif" />
              </Form.Item>

              <Form.Item
                label="允许的音频格式"
                name="allowed_audio_types"
              >
                <Input placeholder="如：mp3,wav" />
              </Form.Item>

              <Form.Item
                label="允许的视频格式"
                name="allowed_video_types"
              >
                <Input placeholder="如：mp4" />
              </Form.Item>

              <Divider />

              <Form.Item style={{ marginTop: 24 }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  icon={<SaveOutlined />}
                  loading={loading}
                  style={{ backgroundColor: '#e67e22', borderColor: '#e67e22' }}
                >
                  保存设置
                </Button>
              </Form.Item>
            </Form>
          </TabPane>

          <TabPane
            tab={
              <span>
                <BellOutlined />
                通知设置
              </span>
            }
            key="4"
          >
            <Form
              layout="vertical"
              autoComplete="off"
              style={{ maxWidth: 600 }}
            >
              <Alert
                message="功能开发中"
                description="消息通知设置功能正在开发中，敬请期待"
                type="warning"
                showIcon
              />
            </Form>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  );
}
