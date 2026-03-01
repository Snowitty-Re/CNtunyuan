import { useState } from 'react'
import {
  Card,
  Tabs,
  Form,
  Input,
  Switch,
  Button,
  Select,
  InputNumber,
  Divider,
  message,
  Space,
  Alert,
  Typography,
} from 'antd'
import {
  SaveOutlined,
  SafetyOutlined,
  SettingOutlined,
  BellOutlined,
  CloudUploadOutlined,
} from '@ant-design/icons'

const { TabPane } = Tabs
const { Title, Text } = Typography

const Settings = () => {
  const [generalForm] = Form.useForm()
  const [securityForm] = Form.useForm()
  const [storageForm] = Form.useForm()
  const [notificationForm] = Form.useForm()
  const [saving, setSaving] = useState(false)

  const handleSave = async (type: string, values: any) => {
    setSaving(true)
    try {
      // TODO: 调用API保存设置
      console.log(`Saving ${type} settings:`, values)
      await new Promise((resolve) => setTimeout(resolve, 500))
      message.success('保存成功')
    } catch (error) {
      message.error('保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <Title level={4}>系统设置</Title>
      <Text type="secondary">管理系统全局配置，仅管理员可访问</Text>

      <Card style={{ marginTop: 16 }}>
        <Tabs defaultActiveKey="general">
          <TabPane
            tab={
              <span>
                <SettingOutlined />
                基础设置
              </span>
            }
            key="general"
          >
            <Form
              form={generalForm}
              layout="vertical"
              initialValues={{
                siteName: '团圆寻亲志愿者系统',
                siteDescription: '帮助寻找走失人员的公益平台',
                maxUploadSize: 50,
                allowRegistration: true,
                requireApproval: true,
              }}
              onFinish={(values) => handleSave('general', values)}
            >
              <Form.Item
                label="站点名称"
                name="siteName"
                rules={[{ required: true, message: '请输入站点名称' }]}
              >
                <Input />
              </Form.Item>

              <Form.Item label="站点描述" name="siteDescription">
                <Input.TextArea rows={2} />
              </Form.Item>

              <Form.Item
                label="最大上传文件大小(MB)"
                name="maxUploadSize"
                rules={[{ required: true }]}
              >
                <InputNumber min={1} max={500} style={{ width: 200 }} />
              </Form.Item>

              <Form.Item
                label="开放注册"
                name="allowRegistration"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label="注册需要审核"
                name="requireApproval"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>
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
            key="security"
          >
            <Form
              form={securityForm}
              layout="vertical"
              initialValues={{
                passwordMinLength: 8,
                passwordRequireUppercase: true,
                passwordRequireNumber: true,
                passwordRequireSpecial: false,
                loginMaxAttempts: 5,
                loginLockoutMinutes: 30,
                sessionTimeout: 120,
                requireTwoFactor: false,
              }}
              onFinish={(values) => handleSave('security', values)}
            >
              <Alert
                message="安全提示"
                description="以下设置影响系统安全性，请谨慎修改"
                type="warning"
                showIcon
                style={{ marginBottom: 16 }}
              />

              <Form.Item
                label="密码最小长度"
                name="passwordMinLength"
                rules={[{ required: true }]}
              >
                <InputNumber min={6} max={32} style={{ width: 200 }} />
              </Form.Item>

              <Form.Item
                label="密码要求大写字母"
                name="passwordRequireUppercase"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label="密码要求数字"
                name="passwordRequireNumber"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label="密码要求特殊字符"
                name="passwordRequireSpecial"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Divider />

              <Form.Item
                label="登录最大尝试次数"
                name="loginMaxAttempts"
                rules={[{ required: true }]}
              >
                <InputNumber min={3} max={10} style={{ width: 200 }} />
              </Form.Item>

              <Form.Item
                label="登录锁定时间(分钟)"
                name="loginLockoutMinutes"
                rules={[{ required: true }]}
              >
                <InputNumber min={5} max={120} style={{ width: 200 }} />
              </Form.Item>

              <Form.Item
                label="会话超时时间(分钟)"
                name="sessionTimeout"
                rules={[{ required: true }]}
              >
                <InputNumber min={15} max={480} style={{ width: 200 }} />
              </Form.Item>

              <Form.Item
                label="强制双因素认证"
                name="requireTwoFactor"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>
                  保存设置
                </Button>
              </Form.Item>
            </Form>
          </TabPane>

          <TabPane
            tab={
              <span>
                <CloudUploadOutlined />
                存储设置
              </span>
            }
            key="storage"
          >
            <Form
              form={storageForm}
              layout="vertical"
              initialValues={{
                storageType: 'local',
                localPath: './uploads',
              }}
              onFinish={(values) => handleSave('storage', values)}
            >
              <Form.Item
                label="存储类型"
                name="storageType"
                rules={[{ required: true }]}
              >
                <Select style={{ width: 200 }}>
                  <Select.Option value="local">本地存储</Select.Option>
                  <Select.Option value="oss">阿里云OSS</Select.Option>
                  <Select.Option value="cos">腾讯云COS</Select.Option>
                </Select>
              </Form.Item>

              <Form.Item
                label="本地存储路径"
                name="localPath"
                rules={[{ required: true }]}
              >
                <Input style={{ width: 400 }} />
              </Form.Item>

              <Alert
                message="对象存储配置"
                description="如需使用OSS或COS，请在后端配置文件中设置相关参数"
                type="info"
                showIcon
                style={{ marginBottom: 16 }}
              />

              <Form.Item>
                <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>
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
            key="notification"
          >
            <Form
              form={notificationForm}
              layout="vertical"
              initialValues={{
                enableEmail: false,
                enableSMS: false,
                enableWechat: true,
                taskAssignNotify: true,
                taskCompleteNotify: true,
                workflowNotify: true,
              }}
              onFinish={(values) => handleSave('notification', values)}
            >
              <Form.Item
                label="启用邮件通知"
                name="enableEmail"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label="启用短信通知"
                name="enableSMS"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label="启用微信通知"
                name="enableWechat"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Divider />

              <Form.Item
                label="任务分配通知"
                name="taskAssignNotify"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label="任务完成通知"
                name="taskCompleteNotify"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label="工作流审批通知"
                name="workflowNotify"
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>

              <Form.Item>
                <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={saving}>
                  保存设置
                </Button>
              </Form.Item>
            </Form>
          </TabPane>
        </Tabs>
      </Card>
    </div>
  )
}

export default Settings
