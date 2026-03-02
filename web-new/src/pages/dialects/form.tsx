import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Form, Input, Button, Upload, message, Space } from 'antd';
import { ArrowLeftOutlined, SaveOutlined, UploadOutlined } from '@ant-design/icons';
import { http } from '@/utils/request';
import type { UploadFile } from 'antd/es/upload/interface';

const { TextArea } = Input;

export default function DialectFormPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const [form] = Form.useForm();
  const isEdit = !!id;
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [uploading, setUploading] = useState(false);
  const [audioUrl, setAudioUrl] = useState('');

  useEffect(() => {
    if (isEdit) {
      fetchDetail();
    }
  }, [id]);

  const fetchDetail = async () => {
    try {
      const res: any = await http.get(`/dialects/${id}`);
      form.setFieldsValue({
        title: res.title,
        description: res.description,
        province: res.province,
        city: res.city,
        district: res.district,
      });
      setAudioUrl(res.audio_url);
    } catch (error) {
      message.error('获取方言信息失败');
    }
  };

  const handleUpload = async (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('type', 'audio');

    try {
      setUploading(true);
      const res: any = await http.post('/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      setAudioUrl(res.url);
      message.success('上传成功');
      return res.url;
    } catch (error) {
      message.error('上传失败');
      return null;
    } finally {
      setUploading(false);
    }
  };

  const handleSubmit = async (values: any) => {
    if (!audioUrl && !isEdit) {
      message.error('请上传音频文件');
      return;
    }

    try {
      const data = {
        ...values,
        audio_url: audioUrl,
        duration: 15, // 默认时长，实际应该从音频文件获取
      };

      if (isEdit) {
        await http.put(`/dialects/${id}`, data);
        message.success('更新成功');
      } else {
        await http.post('/dialects', data);
        message.success('创建成功');
      }
      navigate('/dialects');
    } catch (error) {
      message.error(isEdit ? '更新失败' : '创建失败');
    }
  };

  return (
    <div>
      <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/dialects')}>
            返回
          </Button>
          <span style={{ fontSize: 18, fontWeight: 600 }}>
            {isEdit ? '编辑方言' : '录制方言'}
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
            label="标题"
            name="title"
            rules={[{ required: true, message: '请输入标题' }]}
          >
            <Input placeholder="请输入方言样本标题" />
          </Form.Item>

          <Form.Item label="音频文件">
            <Upload
              accept="audio/*"
              beforeUpload={async (file) => {
                await handleUpload(file);
                return false;
              }}
              fileList={fileList}
              onChange={({ fileList }) => setFileList(fileList)}
              maxCount={1}
            >
              <Button icon={<UploadOutlined />} loading={uploading}>
                {audioUrl ? '重新上传' : '上传音频'}
              </Button>
            </Upload>
            {audioUrl && (
              <div style={{ marginTop: 8 }}>
                <audio src={audioUrl} controls style={{ width: '100%' }} />
              </div>
            )}
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

          <Form.Item label="描述" name="description">
            <TextArea rows={4} placeholder="请输入方言描述信息" />
          </Form.Item>

          <Form.Item style={{ marginTop: 24 }}>
            <Space size={16}>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} size="large">
                {isEdit ? '保存修改' : '保存方言'}
              </Button>
              <Button onClick={() => navigate('/dialects')} size="large">
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
