import { useState, useEffect, useCallback } from 'react'
import {
  Card,
  Table,
  DatePicker,
  Select,
  Input,
  Button,
  Space,
  Tag,
  Tooltip,
  message,
  Statistic,
  Row,
  Col,
  Descriptions,
  Modal,
  Popconfirm,
} from 'antd'
import { SearchOutlined, ReloadOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons'
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table'
import type { Dayjs } from 'dayjs'
import dayjs from 'dayjs'

interface OperationLog {
  id: string
  user_id: string
  role: string
  org_id: string
  module: string
  action: string
  method: string
  path: string
  query: string
  body: string
  ip: string
  user_agent: string
  status: number
  duration: number
  error: string
  created_at: string
}

interface LogSummary {
  total_requests: number
  success_count: number
  error_count: number
  unique_users: number
  avg_duration: number
}

const { RangePicker } = DatePicker
const { Option } = Select

const OperationLogs = () => {
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: 1,
    pageSize: 20,
    total: 0,
  })
  const [filters, setFilters] = useState({
    user_id: '',
    role: undefined as string | undefined,
    module: undefined as string | undefined,
    action: undefined as string | undefined,
    status: undefined as number | undefined,
    dateRange: null as [Dayjs, Dayjs] | null,
    keyword: '',
  })
  const [selectedLog, setSelectedLog] = useState<OperationLog | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)
  const [logsData, setLogsData] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  // 获取日志列表
  const fetchLogs = useCallback(async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams({
        page: String(pagination.current || 1),
        page_size: String(pagination.pageSize || 20),
      })

      if (filters.user_id) params.append('user_id', filters.user_id)
      if (filters.role) params.append('role', filters.role)
      if (filters.module) params.append('module', filters.module)
      if (filters.action) params.append('action', filters.action)
      if (filters.status !== undefined) params.append('status', String(filters.status))
      if (filters.keyword) params.append('keyword', filters.keyword)
      if (filters.dateRange) {
        params.append('start_time', filters.dateRange[0].toISOString())
        params.append('end_time', filters.dateRange[1].toISOString())
      }

      const res = await fetch(`/api/v1/operation-logs?${params.toString()}`, {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
        },
      })

      if (!res.ok) {
        throw new Error('获取日志失败')
      }

      const data = await res.json()
      setLogsData(data)
      if (data.data) {
        setPagination((prev) => ({
          ...prev,
          total: data.data.total,
        }))
      }
    } catch (error) {
      message.error('获取日志失败')
    } finally {
      setLoading(false)
    }
  }, [pagination.current, pagination.pageSize, filters])

  useEffect(() => {
    fetchLogs()
  }, [fetchLogs])

  // 获取统计摘要
  const [summaryData, setSummaryData] = useState<any>(null)
  
  useEffect(() => {
    const fetchSummary = async () => {
      const res = await fetch('/api/v1/operation-logs/stats/summary?days=7', {
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
        },
      })
      if (res.ok) {
        const data = await res.json()
        setSummaryData(data)
      }
    }
    fetchSummary()
    const interval = setInterval(fetchSummary, 30000)
    return () => clearInterval(interval)
  }, [])

  const summary: LogSummary | null = summaryData?.data || null

  // 清理旧日志
  const handleCleanup = async (days: number) => {
    try {
      const res = await fetch(`/api/v1/operation-logs/cleanup?days=${days}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
        },
      })

      if (!res.ok) {
        throw new Error('清理失败')
      }

      message.success(`已清理 ${days} 天前的日志`)
      fetchLogs()
    } catch (error) {
      message.error('清理失败')
    }
  }

  const columns: ColumnsType<OperationLog> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (text) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '用户',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 100,
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <span>{record.user_id?.slice(0, 8) || '系统'}</span>
          <Tag size="small" color={getRoleColor(record.role)}>
            {getRoleLabel(record.role)}
          </Tag>
        </Space>
      ),
    },
    {
      title: '模块',
      dataIndex: 'module',
      key: 'module',
      width: 100,
      render: (text) => <Tag>{text}</Tag>,
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 80,
      render: (text) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: '请求',
      key: 'request',
      render: (_, record) => (
        <Space direction="vertical" size={0}>
          <Tag color={getMethodColor(record.method)}>{record.method}</Tag>
          <Tooltip title={record.path}>
            <span style={{ fontSize: 12, color: '#666' }}>
              {record.path.length > 40 ? record.path.slice(0, 40) + '...' : record.path}
            </span>
          </Tooltip>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status) => (
        <Tag color={status >= 400 ? 'error' : status >= 300 ? 'warning' : 'success'}>
          {status}
        </Tag>
      ),
    },
    {
      title: '耗时',
      dataIndex: 'duration',
      key: 'duration',
      width: 90,
      render: (duration) => (
        <span style={{ color: duration > 1000 ? '#f5222d' : '#52c41a' }}>
          {duration}ms
        </span>
      ),
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      key: 'ip',
      width: 120,
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      fixed: 'right',
      render: (_, record) => (
        <Button
          type="text"
          icon={<EyeOutlined />}
          onClick={() => {
            setSelectedLog(record)
            setDetailVisible(true)
          }}
        />
      ),
    },
  ]

  const getRoleColor = (role: string) => {
    const colors: Record<string, string> = {
      super_admin: 'red',
      admin: 'orange',
      manager: 'blue',
      volunteer: 'green',
    }
    return colors[role] || 'default'
  }

  const getRoleLabel = (role: string) => {
    const labels: Record<string, string> = {
      super_admin: '超级管理员',
      admin: '管理员',
      manager: '管理者',
      volunteer: '志愿者',
    }
    return labels[role] || role
  }

  const getMethodColor = (method: string) => {
    const colors: Record<string, string> = {
      GET: 'green',
      POST: 'blue',
      PUT: 'orange',
      PATCH: 'orange',
      DELETE: 'red',
    }
    return colors[method] || 'default'
  }

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination(newPagination)
  }

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={4}>
            <Statistic
              title="总请求数"
              value={summary?.total_requests || 0}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="成功请求"
              value={summary?.success_count || 0}
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="失败请求"
              value={summary?.error_count || 0}
              valueStyle={{ color: '#f5222d' }}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="独立用户"
              value={summary?.unique_users || 0}
            />
          </Col>
          <Col span={4}>
            <Statistic
              title="平均耗时"
              value={summary?.avg_duration || 0}
              suffix="ms"
            />
          </Col>
          <Col span={4} style={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
            <Popconfirm
              title="清理日志"
              description="确定要清理30天前的日志吗？"
              onConfirm={() => handleCleanup(30)}
              okText="确定"
              cancelText="取消"
            >
              <Button icon={<DeleteOutlined />} danger>
                清理旧日志
              </Button>
            </Popconfirm>
          </Col>
        </Row>
      </Card>

      <Card
        title="操作日志"
        extra={
          <Button icon={<ReloadOutlined />} onClick={fetchLogs}>
            刷新
          </Button>
        }
      >
        <Space style={{ marginBottom: 16, flexWrap: 'wrap' }}>
          <RangePicker
            value={filters.dateRange}
            onChange={(dates) => setFilters({ ...filters, dateRange: dates as [Dayjs, Dayjs] })}
            showTime
          />
          <Select
            placeholder="角色"
            value={filters.role}
            onChange={(value) => setFilters({ ...filters, role: value })}
            style={{ width: 120 }}
            allowClear
          >
            <Option value="super_admin">超级管理员</Option>
            <Option value="admin">管理员</Option>
            <Option value="manager">管理者</Option>
            <Option value="volunteer">志愿者</Option>
          </Select>
          <Select
            placeholder="模块"
            value={filters.module}
            onChange={(value) => setFilters({ ...filters, module: value })}
            style={{ width: 120 }}
            allowClear
          >
            <Option value="auth">认证</Option>
            <Option value="users">用户</Option>
            <Option value="organizations">组织</Option>
            <Option value="missing-persons">走失人员</Option>
            <Option value="dialects">方言</Option>
            <Option value="tasks">任务</Option>
            <Option value="workflows">工作流</Option>
          </Select>
          <Select
            placeholder="状态"
            value={filters.status}
            onChange={(value) => setFilters({ ...filters, status: value })}
            style={{ width: 100 }}
            allowClear
          >
            <Option value={200}>200</Option>
            <Option value={400}>400</Option>
            <Option value={401}>401</Option>
            <Option value={403}>403</Option>
            <Option value={404}>404</Option>
            <Option value={500}>500</Option>
          </Select>
          <Input.Search
            placeholder="搜索关键词"
            value={filters.keyword}
            onChange={(e) => setFilters({ ...filters, keyword: e.target.value })}
            onSearch={() => refresh()}
            style={{ width: 200 }}
            allowClear
          />
        </Space>

        <Table
          columns={columns}
          dataSource={logsData?.data?.list || []}
          loading={loading}
          pagination={pagination}
          onChange={handleTableChange}
          rowKey="id"
          scroll={{ x: 1200 }}
        />
      </Card>

      <Modal
        title="日志详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={800}
      >
        {selectedLog && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="ID">{selectedLog.id}</Descriptions.Item>
            <Descriptions.Item label="时间">
              {dayjs(selectedLog.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
            <Descriptions.Item label="用户ID">{selectedLog.user_id}</Descriptions.Item>
            <Descriptions.Item label="角色">
              <Tag color={getRoleColor(selectedLog.role)}>
                {getRoleLabel(selectedLog.role)}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="模块">{selectedLog.module}</Descriptions.Item>
            <Descriptions.Item label="操作">{selectedLog.action}</Descriptions.Item>
            <Descriptions.Item label="方法">
              <Tag color={getMethodColor(selectedLog.method)}>{selectedLog.method}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={selectedLog.status >= 400 ? 'error' : 'success'}>
                {selectedLog.status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="路径" span={2}>
              {selectedLog.path}
            </Descriptions.Item>
            <Descriptions.Item label="查询参数" span={2}>
              <pre style={{ margin: 0, maxHeight: 100, overflow: 'auto' }}>
                {selectedLog.query || '无'}
              </pre>
            </Descriptions.Item>
            <Descriptions.Item label="请求Body" span={2}>
              <pre style={{ margin: 0, maxHeight: 200, overflow: 'auto' }}>
                {selectedLog.body || '无'}
              </pre>
            </Descriptions.Item>
            <Descriptions.Item label="IP">{selectedLog.ip}</Descriptions.Item>
            <Descriptions.Item label="耗时">{selectedLog.duration}ms</Descriptions.Item>
            <Descriptions.Item label="User-Agent" span={2}>
              {selectedLog.user_agent}
            </Descriptions.Item>
            {selectedLog.error && (
              <Descriptions.Item label="错误信息" span={2}>
                <pre style={{ margin: 0, color: '#f5222d', maxHeight: 200, overflow: 'auto' }}>
                  {selectedLog.error}
                </pre>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>
    </div>
  )
}

export default OperationLogs
