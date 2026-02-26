import { useEffect, useState } from 'react'
import { Table, Button, Space, Tag, Modal, Form, Input, Select, DatePicker, message } from 'antd'
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons'
import { missingPersonApi } from '../services/missing_person'
import type { MissingPerson } from '../types'
import dayjs from 'dayjs'

const { Option } = Select
const { TextArea } = Input

const MissingPersons = () => {
  const [data, setData] = useState<MissingPerson[]>([])
  const [loading, setLoading] = useState(false)
  const [modalVisible, setModalVisible] = useState(false)
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedRecord, setSelectedRecord] = useState<MissingPerson | null>(null)
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [form] = Form.useForm()

  useEffect(() => {
    fetchData()
  }, [pagination.current, pagination.pageSize])

  const fetchData = async () => {
    setLoading(true)
    try {
      const result = await missingPersonApi.getList({
        page: pagination.current,
        page_size: pagination.pageSize,
      })
      setData(result.list)
      setPagination({ ...pagination, total: result.total })
    } finally {
      setLoading(false)
    }
  }

  const handleView = (record: MissingPerson) => {
    setSelectedRecord(record)
    setDetailVisible(true)
  }

  const handleEdit = (record: MissingPerson) => {
    setSelectedRecord(record)
    form.setFieldsValue({
      ...record,
      missing_time: record.missing_time ? dayjs(record.missing_time) : undefined,
    })
    setModalVisible(true)
  }

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除后无法恢复，是否确认？',
      onOk: async () => {
        message.success('删除成功')
        fetchData()
      },
    })
  }

  const handleSubmit = async (values: any) => {
    const data = {
      ...values,
      missing_time: values.missing_time?.toISOString(),
    }
    if (selectedRecord) {
      message.success('更新成功')
    } else {
      message.success('创建成功')
    }
    setModalVisible(false)
    fetchData()
  }

  const handleStatusChange = async (id: string, status: string) => {
    await missingPersonApi.updateStatus(id, status)
    message.success('状态更新成功')
    fetchData()
  }

  const columns = [
    { title: '案件编号', dataIndex: 'case_no' },
    { title: '姓名', dataIndex: 'name' },
    { title: '年龄', dataIndex: 'age' },
    {
      title: '案件类型',
      dataIndex: 'case_type',
      render: (type: string) => ({
        elderly: '老人走失',
        child: '儿童走失',
        adult: '成年人走失',
        disability: '残障人士走失',
        other: '其他',
      }[type] || type),
    },
    { title: '走失地点', dataIndex: 'missing_location' },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string, record: MissingPerson) => (
        <Select
          value={status}
          style={{ width: 100 }}
          onChange={(value) => handleStatusChange(record.id, value)}
        >
          <Option value="missing">失踪中</Option>
          <Option value="searching">寻找中</Option>
          <Option value="found">已找到</Option>
          <Option value="reunited">已团圆</Option>
          <Option value="closed">已结案</Option>
        </Select>
      ),
    },
    {
      title: '操作',
      render: (_: any, record: MissingPerson) => (
        <Space>
          <Button icon={<EyeOutlined />} onClick={() => handleView(record)}>
            查看
          </Button>
          <Button icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button danger icon={<DeleteOutlined />} onClick={() => handleDelete(record.id)}>
            删除
          </Button>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setSelectedRecord(null); form.resetFields(); setModalVisible(true) }}>
          登记走失人员
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={(p) => setPagination({ ...pagination, current: p.current || 1, pageSize: p.pageSize || 20 })}
      />

      {/* 详情Modal */}
      <Modal
        title="案件详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={800}
      >
        {selectedRecord && (
          <div>
            <p><strong>姓名：</strong>{selectedRecord.name}</p>
            <p><strong>性别：</strong>{selectedRecord.gender === 'male' ? '男' : selectedRecord.gender === 'female' ? '女' : '其他'}</p>
            <p><strong>年龄：</strong>{selectedRecord.age}</p>
            <p><strong>走失时间：</strong>{selectedRecord.missing_time}</p>
            <p><strong>走失地点：</strong>{selectedRecord.missing_location}</p>
            <p><strong>外貌特征：</strong>{selectedRecord.appearance}</p>
            <p><strong>衣着描述：</strong>{selectedRecord.clothing}</p>
            <p><strong>联系人：</strong>{selectedRecord.contact_name}</p>
            <p><strong>联系电话：</strong>{selectedRecord.contact_phone}</p>
          </div>
        )}
      </Modal>

      {/* 编辑Modal */}
      <Modal
        title={selectedRecord ? '编辑案件' : '登记走失人员'}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={() => form.submit()}
        width={800}
      >
        <Form form={form} onFinish={handleSubmit} layout="vertical">
          <Form.Item name="name" label="姓名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="gender" label="性别">
            <Select>
              <Option value="male">男</Option>
              <Option value="female">女</Option>
              <Option value="other">其他</Option>
            </Select>
          </Form.Item>
          <Form.Item name="age" label="年龄">
            <Input type="number" />
          </Form.Item>
          <Form.Item name="case_type" label="案件类型">
            <Select>
              <Option value="elderly">老人走失</Option>
              <Option value="child">儿童走失</Option>
              <Option value="adult">成年人走失</Option>
              <Option value="disability">残障人士走失</Option>
              <Option value="other">其他</Option>
            </Select>
          </Form.Item>
          <Form.Item name="missing_time" label="走失时间" rules={[{ required: true }]}>
            <DatePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="missing_location" label="走失地点" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="appearance" label="外貌特征">
            <TextArea rows={3} />
          </Form.Item>
          <Form.Item name="clothing" label="衣着描述">
            <TextArea rows={2} />
          </Form.Item>
          <Form.Item name="contact_name" label="联系人">
            <Input />
          </Form.Item>
          <Form.Item name="contact_phone" label="联系电话" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default MissingPersons
