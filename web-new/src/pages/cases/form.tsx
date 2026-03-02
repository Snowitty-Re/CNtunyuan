import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Form, Input, Button, Select, DatePicker, Radio, message, Space } from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import { http } from '@/utils/request';
import dayjs from 'dayjs';

const { TextArea } = Input;

export default function CaseFormPage() {
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
      const res: any = await http.get(`/missing-persons/${id}`);
      form.setFieldsValue({
        ...res,
        missing_time: res.missing_time ? dayjs(res.missing_time) : null,
      });
    } catch (error) {
      message.error('获取案件信息失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      const data = {
        ...values,
        missing_time: values.missing_time?.format('YYYY-MM-DD HH:mm:ss'),
      };

      if (isEdit) {
        await http.put(`/missing-persons/${id}`, data);
        message.success('更新成功');
      } else {
        await http.post('/missing-persons', data);
        message.success('创建成功');
      }
      navigate('/cases');
    } catch (error) {
      message.error(isEdit ? '更新失败' : '创建失败');
    }
  };

  return (
    <div>
      <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/cases')}>
            返回
          </Button>
          <span style={{ fontSize: 18, fontWeight: 600 }}>
            {isEdit ? '编辑案件' : '发布寻人'}
          </span>
        </Space>
      </Card>

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          autoComplete="off"
          style={{ maxWidth: 800 }}
        >
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
            <Form.Item
              label="姓名"
              name="name"
              rules={[{ required: true, message: '请输入姓名' }]}
            >
              <Input placeholder="请输入走失人员姓名" />
            </Form.Item>

            <Form.Item
              label="性别"
              name="gender"
              rules={[{ required: true, message: '请选择性别' }]}
            >
              <Radio.Group>
                <Radio value="male">男</Radio>
                <Radio value="female">女</Radio>
              </Radio.Group>
            </Form.Item>

            <Form.Item
              label="年龄"
              name="age"
              rules={[{ required: true, message: '请输入年龄' }]}
            >
              <Input type="number" placeholder="请输入年龄" />
            </Form.Item>

            <Form.Item
              label="案件类型"
              name="case_type"
              rules={[{ required: true, message: '请选择案件类型' }]}
            >
              <Select placeholder="请选择案件类型">
                <Select.Option value="elderly">老人走失</Select.Option>
                <Select.Option value="child">儿童走失</Select.Option>
                <Select.Option value="adult">成人走失</Select.Option>
                <Select.Option value="disability">残障人士走失</Select.Option>
                <Select.Option value="other">其他</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="走失时间"
              name="missing_time"
              rules={[{ required: true, message: '请选择走失时间' }]}
            >
              <DatePicker
                showTime
                format="YYYY-MM-DD HH:mm"
                style={{ width: '100%' }}
                placeholder="请选择走失时间"
              />
            </Form.Item>

            <Form.Item
              label="案件状态"
              name="status"
              initialValue="missing"
              rules={[{ required: true, message: '请选择状态' }]}
            >
              <Select placeholder="请选择状态">
                <Select.Option value="missing">失踪中</Select.Option>
                <Select.Option value="searching">寻找中</Select.Option>
                <Select.Option value="found">已找到</Select.Option>
                <Select.Option value="reunited">已团圆</Select.Option>
              </Select>
            </Form.Item>
          </div>

          <Form.Item
            label="走失地点"
            name="missing_location"
            rules={[{ required: true, message: '请输入走失地点' }]}
          >
            <Input placeholder="请输入详细走失地点" />
          </Form.Item>

          <Form.Item
            label="体貌特征"
            name="appearance"
          >
            <TextArea rows={3} placeholder="请描述走失人员的体貌特征、穿着等" />
          </Form.Item>

          <Form.Item
            label="联系人姓名"
            name="contact_name"
            rules={[{ required: true, message: '请输入联系人姓名' }]}
          >
            <Input placeholder="请输入联系人姓名" />
          </Form.Item>

          <Form.Item
            label="联系人电话"
            name="contact_phone"
            rules={[
              { required: true, message: '请输入联系人电话' },
              { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号' },
            ]}
          >
            <Input placeholder="请输入联系人电话" />
          </Form.Item>

          <Form.Item style={{ marginTop: 24 }}>
            <Space size={16}>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} size="large">
                {isEdit ? '保存修改' : '发布寻人'}
              </Button>
              <Button onClick={() => navigate('/cases')} size="large">
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
