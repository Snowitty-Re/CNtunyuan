import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Card, Form, Input, Button, Select, DatePicker, message, Space, Slider } from 'antd';
import { ArrowLeftOutlined, SaveOutlined } from '@ant-design/icons';
import { http } from '@/utils/request';
import dayjs from 'dayjs';

const { TextArea } = Input;

export default function TaskFormPage() {
  const navigate = useNavigate();
  const { id } = useParams();
  const [form] = Form.useForm();
  const isEdit = !!id;
  const [cases, setCases] = useState<any[]>([]);
  const [users, setUsers] = useState<any[]>([]);

  useEffect(() => {
    fetchOptions();
    if (isEdit) {
      fetchDetail();
    }
  }, [id]);

  const fetchOptions = async () => {
    try {
      const [casesRes, usersRes] = await Promise.all([
        http.get('/missing-persons?limit=100'),
        http.get('/users?limit=100'),
      ]);
      setCases((casesRes as any).list || []);
      setUsers((usersRes as any).list || []);
    } catch (error) {
      console.error('获取选项失败', error);
    }
  };

  const fetchDetail = async () => {
    try {
      const res: any = await http.get(`/tasks/${id}`);
      form.setFieldsValue({
        ...res,
        deadline: res.deadline ? dayjs(res.deadline) : null,
        missing_person_id: res.missing_person_id,
        assignee_id: res.assignee_id,
      });
    } catch (error) {
      message.error('获取任务信息失败');
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      const data = {
        ...values,
        deadline: values.deadline?.format('YYYY-MM-DD HH:mm:ss'),
      };

      if (isEdit) {
        await http.put(`/tasks/${id}`, data);
        message.success('更新成功');
      } else {
        await http.post('/tasks', data);
        message.success('创建成功');
      }
      navigate('/tasks');
    } catch (error) {
      message.error(isEdit ? '更新失败' : '创建失败');
    }
  };

  return (
    <div>
      <Card style={{ marginBottom: 24 }} bodyStyle={{ padding: 16 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/tasks')}>
            返回
          </Button>
          <span style={{ fontSize: 18, fontWeight: 600 }}>
            {isEdit ? '编辑任务' : '创建任务'}
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
          <Form.Item
            label="任务标题"
            name="title"
            rules={[{ required: true, message: '请输入任务标题' }]}
          >
            <Input placeholder="请输入任务标题" />
          </Form.Item>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
            <Form.Item
              label="任务类型"
              name="type"
              rules={[{ required: true, message: '请选择任务类型' }]}
            >
              <Select placeholder="请选择任务类型">
                <Select.Option value="search">实地寻访</Select.Option>
                <Select.Option value="phone">电话核实</Select.Option>
                <Select.Option value="coordination">协调沟通</Select.Option>
                <Select.Option value="documentation">资料整理</Select.Option>
                <Select.Option value="other">其他</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="优先级"
              name="priority"
              initialValue="normal"
              rules={[{ required: true, message: '请选择优先级' }]}
            >
              <Select placeholder="请选择优先级">
                <Select.Option value="urgent">紧急</Select.Option>
                <Select.Option value="high">高</Select.Option>
                <Select.Option value="normal">普通</Select.Option>
                <Select.Option value="low">低</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="关联案件"
              name="missing_person_id"
            >
              <Select placeholder="请选择关联案件（可选）" allowClear>
                {cases.map((c) => (
                  <Select.Option key={c.id} value={c.id}>
                    {c.case_no} - {c.name}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>

            <Form.Item
              label="执行人"
              name="assignee_id"
            >
              <Select placeholder="请选择执行人（可选）" allowClear>
                {users.map((u) => (
                  <Select.Option key={u.id} value={u.id}>
                    {u.nickname}
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>

            <Form.Item
              label="截止时间"
              name="deadline"
            >
              <DatePicker
                showTime
                format="YYYY-MM-DD HH:mm"
                style={{ width: '100%' }}
                placeholder="请选择截止时间"
              />
            </Form.Item>

            <Form.Item
              label="当前进度"
              name="progress"
              initialValue={0}
            >
              <Slider min={0} max={100} marks={{ 0: '0%', 50: '50%', 100: '100%' }} />
            </Form.Item>
          </div>

          <Form.Item
            label="任务描述"
            name="description"
          >
            <TextArea rows={4} placeholder="请详细描述任务内容" />
          </Form.Item>

          <Form.Item style={{ marginTop: 24 }}>
            <Space size={16}>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} size="large">
                {isEdit ? '保存修改' : '创建任务'}
              </Button>
              <Button onClick={() => navigate('/tasks')} size="large">
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
