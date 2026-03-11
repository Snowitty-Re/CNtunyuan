import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Form, Input, Select, Button, DatePicker, Row, Col, message, Spin, Space } from 'antd';
import PageContainer from '@/components/PageContainer';
import { createTask, getTask, updateTask } from '@/api/task';
import { TASK_PRIORITY_MAP } from '@/utils/constants';
import dayjs from 'dayjs';

const TASK_TYPES = [
  { label: '搜寻', value: 'search' },
  { label: '走访', value: 'visit' },
  { label: '核实', value: 'verify' },
  { label: '联络', value: 'contact' },
  { label: '其他', value: 'other' },
];

export default function TaskFormPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(!!id);
  const isEdit = !!id;

  useEffect(() => {
    if (!id) return;
    getTask(id)
      .then((data) => {
        form.setFieldsValue({
          ...data,
          deadline: data.deadline ? dayjs(data.deadline) : undefined,
        });
      })
      .finally(() => setInitialLoading(false));
  }, [id, form]);

  const onFinish = async (values: any) => {
    setLoading(true);
    const payload = { ...values, deadline: values.deadline?.toISOString() };
    try {
      if (isEdit) {
        await updateTask(id!, payload);
        message.success('更新成功');
      } else {
        await createTask(payload);
        message.success('创建成功');
      }
      navigate('/tasks');
    } catch { /* interceptor */ } finally {
      setLoading(false);
    }
  };

  if (initialLoading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;

  return (
    <PageContainer
      title={isEdit ? '编辑任务' : '新建任务'}
      breadcrumb={[{ title: '首页' }, { title: '任务管理' }, { title: isEdit ? '编辑' : '新建' }]}
    >
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 800 }}>
          <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item name="type" label="类型" rules={[{ required: true }]}>
                <Select options={TASK_TYPES} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="priority" label="优先级">
                <Select options={Object.entries(TASK_PRIORITY_MAP).map(([k, v]) => ({ label: v, value: k }))} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="deadline" label="截止日期">
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Row gutter={16}>
            <Col xs={24} sm={8}><Form.Item name="province" label="省份"><Input /></Form.Item></Col>
            <Col xs={24} sm={8}><Form.Item name="city" label="城市"><Input /></Form.Item></Col>
            <Col xs={24} sm={8}><Form.Item name="district" label="区县"><Input /></Form.Item></Col>
          </Row>
          <Form.Item name="address" label="详细地址"><Input /></Form.Item>
          <Form.Item name="missing_person_id" label="关联走失人员ID"><Input placeholder="可选" /></Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>{isEdit ? '保存' : '创建'}</Button>
              <Button onClick={() => navigate(-1)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
}
