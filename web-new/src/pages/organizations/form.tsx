import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Form, Input, Button, Select, Radio, message, Space } from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import { http } from '@/utils/request';

const { TextArea } = Input;

export default function OrganizationFormPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const [form] = Form.useForm();
  const isEdit = !!id;

  useEffect(() => {
    if (isEdit) {
      fetchDetail();
    }
  }, [id]);

  const fetchDetail = async () => {
    try {
      const res: any = await http.get(`/organizations/${id}`);
      form.setFieldsValue({
        name: res.name,
        code: res.code,
        type: res.type,
        province: res.province,
        city: res.city,
        district: res.district,
        address: res.address,
        status: res.status,
        description: res.description,
      });
    } catch (error) {
      message.error('获取组织信息失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      if (isEdit) {
        await http.put(`/organizations/${id}`, values);
        message.success('更新成功');
      } else {
        await http.post('/organizations', values);
        message.success('创建成功');
      }
      navigate('/organizations');
    } catch (error) {
      message.error(isEdit ? '更新失败' : '创建失败');
    }
  };

  return (
    <div>
      <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/organizations')}>
            返回
          </Button>
          <span style={{ fontSize: 18, fontWeight: 600 }}>
            {isEdit ? '编辑组织' : '添加组织'}
          </span>
        </Space>
      </Card>

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
          style={{ maxWidth: 600 }}
        >
          <Form.Item
            label="组织名称"
            name="name"
            rules={[{ required: true, message: '请输入组织名称' }]}
          >
            <Input placeholder="请输入组织名称" />
          </Form.Item>

          <Form.Item
            label="组织代码"
            name="code"
            rules={[{ required: true, message: '请输入组织代码' }]}
          >
            <Input placeholder="请输入组织代码，如：beijing" disabled={isEdit} />
          </Form.Item>

          <Form.Item
            label="组织类型"
            name="type"
            rules={[{ required: true, message: '请选择组织类型' }]}
          >
            <Select placeholder="请选择组织类型">
              <Select.Option value="root">总部</Select.Option>
              <Select.Option value="province">省级</Select.Option>
              <Select.Option value="city">市级</Select.Option>
              <Select.Option value="district">区级</Select.Option>
              <Select.Option value="street">街道</Select.Option>
            </Select>
          </Form.Item>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 16 }}>
            <Form.Item label="省" name="province">
              <Input placeholder="省" />
            </Form.Item>
            <Form.Item label="市" name="city">
              <Input placeholder="市" />
            </Form.Item>
            <Form.Item label="区/县" name="district">
              <Input placeholder="区/县" />
            </Form.Item>
          </div>

          <Form.Item label="详细地址" name="address">
            <Input placeholder="请输入详细地址" />
          </Form.Item>

          <Form.Item label="组织描述" name="description">
            <TextArea rows={3} placeholder="请输入组织描述" />
          </Form.Item>

          <Form.Item
            label="状态"
            name="status"
            initialValue="active"
            rules={[{ required: true, message: '请选择状态' }]}
          >
            <Radio.Group>
              <Radio value="active">正常</Radio>
              <Radio value="inactive">禁用</Radio>
            </Radio.Group>
          </Form.Item>

          <Form.Item style={{ marginTop: 24 }}>
            <Space size={16}>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} size="large">
                {isEdit ? '保存修改' : '添加组织'}
              </Button>
              <Button onClick={() => navigate('/organizations')} size="large">
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
