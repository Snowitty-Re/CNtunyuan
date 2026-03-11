import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Form, Input, Button, Row, Col, message, Spin, Space } from 'antd';
import PageContainer from '@/components/PageContainer';
import FileUpload from '@/components/FileUpload';
import { createDialect, getDialect, updateDialect } from '@/api/dialect';

export default function DialectFormPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [initialLoading, setInitialLoading] = useState(!!id);
  const isEdit = !!id;

  useEffect(() => {
    if (!id) return;
    getDialect(id)
      .then((data) => form.setFieldsValue(data))
      .finally(() => setInitialLoading(false));
  }, [id, form]);

  const onFinish = async (values: any) => {
    setLoading(true);
    try {
      if (isEdit) {
        await updateDialect(id!, values);
        message.success('更新成功');
      } else {
        await createDialect(values);
        message.success('创建成功');
      }
      navigate('/dialects');
    } catch { /* interceptor */ } finally {
      setLoading(false);
    }
  };

  if (initialLoading) return <div style={{ textAlign: 'center', paddingTop: 120 }}><Spin size="large" /></div>;

  return (
    <PageContainer
      title={isEdit ? '编辑方言' : '上传方言'}
      breadcrumb={[{ title: '首页' }, { title: '方言库' }, { title: isEdit ? '编辑' : '上传' }]}
    >
      <Card>
        <Form form={form} layout="vertical" onFinish={onFinish} style={{ maxWidth: 800 }}>
          <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
          <Row gutter={16}>
            <Col xs={24} sm={12}><Form.Item name="region" label="地区" rules={[{ required: true }]}><Input /></Form.Item></Col>
            <Col xs={24} sm={12}><Form.Item name="dialect_type" label="方言类型"><Input placeholder="如: 粤语、闽南语" /></Form.Item></Col>
          </Row>
          <Row gutter={16}>
            <Col xs={24} sm={12}><Form.Item name="province" label="省份"><Input /></Form.Item></Col>
            <Col xs={24} sm={12}><Form.Item name="city" label="城市"><Input /></Form.Item></Col>
          </Row>
          {!isEdit && (
            <Form.Item name="audio_url" label="音频文件URL" rules={[{ required: true, message: '请上传音频或填写URL' }]}>
              <Input placeholder="音频文件URL" />
            </Form.Item>
          )}
          {!isEdit && (
            <Card size="small" style={{ marginBottom: 24 }}>
              <FileUpload
                accept="audio/*"
                maxSize={50}
                onSuccess={(file) => form.setFieldsValue({ audio_url: file.url, duration: 0, file_size: file.size, format: file.mime_type })}
              />
            </Card>
          )}
          <Form.Item name="content" label="内容"><Input.TextArea rows={2} /></Form.Item>
          <Form.Item name="tags" label="标签"><Input placeholder="多个标签用逗号分隔" /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>{isEdit ? '保存' : '上传'}</Button>
              <Button onClick={() => navigate(-1)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
}
