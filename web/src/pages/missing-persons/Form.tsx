import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Form, Input, Select, Button, DatePicker, InputNumber, Row, Col, message, Spin, Space } from 'antd';
import PageContainer from '@/components/PageContainer';
import { createMissingPerson, getMissingPerson, updateMissingPerson } from '@/api/missingPerson';
import { GENDER_MAP, URGENCY_MAP } from '@/utils/constants';
import dayjs from 'dayjs';

export default function MissingPersonFormPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(!!id);
  const isEdit = !!id;

  useEffect(() => {
    if (!id) return;
    getMissingPerson(id)
      .then((data) => {
        form.setFieldsValue({
          ...data,
          missing_time: data.missing_time ? dayjs(data.missing_time) : undefined,
          birth_date: data.birth_date ? dayjs(data.birth_date) : undefined,
          urgency_level: data.urgency,
        });
      })
      .finally(() => setInitialLoading(false));
  }, [id, form]);

  const onFinish = async (values: any) => {
    setLoading(true);
    const payload = {
      ...values,
      missing_time: values.missing_time?.toISOString(),
      birth_date: values.birth_date?.toISOString(),
    };
    try {
      if (isEdit) {
        await updateMissingPerson(id!, payload);
        message.success('更新成功');
      } else {
        await createMissingPerson(payload);
        message.success('创建成功');
      }
      navigate('/missing-persons');
    } catch { /* interceptor */ } finally {
      setLoading(false);
    }
  };

  if (initialLoading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;

  return (
    <PageContainer
      title={isEdit ? '编辑案件' : '登记走失人员'}
      breadcrumb={[{ title: '首页' }, { title: '走失人员' }, { title: isEdit ? '编辑' : '登记' }]}
    >
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 800 }}>
          <Row gutter={16}>
            <Col xs={24} sm={12}>
              <Form.Item name="name" label="姓名" rules={[{ required: true }]}><Input /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="gender" label="性别" rules={[{ required: true }]}>
                <Select options={Object.entries(GENDER_MAP).map(([k, v]) => ({ label: v, value: k }))} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="age" label="年龄"><InputNumber min={0} max={150} style={{ width: '100%' }} /></Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="height" label="身高(cm)"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="weight" label="体重(kg)"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="birth_date" label="出生日期"><DatePicker style={{ width: '100%' }} /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="missing_time" label="走失时间" rules={[{ required: true }]}>
                <DatePicker showTime style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="province" label="省份"><Input /></Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="city" label="城市"><Input /></Form.Item>
            </Col>
            <Col xs={24} sm={8}>
              <Form.Item name="district" label="区县"><Input /></Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="address" label="详细地址"><Input /></Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="clothes" label="衣着描述"><Input.TextArea rows={2} /></Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="features" label="体貌特征"><Input.TextArea rows={2} /></Form.Item>
            </Col>
            <Col xs={24}>
              <Form.Item name="description" label="详细描述"><Input.TextArea rows={3} /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="urgency_level" label="紧急程度">
                <Select options={Object.entries(URGENCY_MAP).map(([k, v]) => ({ label: v, value: k }))} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="photo_url" label="照片URL"><Input /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="contact_name" label="联系人" rules={[{ required: true }]}><Input /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="contact_phone" label="联系电话" rules={[{ required: true }]}><Input /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="contact_rel" label="与走失者关系"><Input /></Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="alt_contact" label="备用联系方式"><Input /></Form.Item>
            </Col>
          </Row>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>{isEdit ? '保存' : '提交'}</Button>
              <Button onClick={() => navigate(-1)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
}
