import { useEffect, useState } from 'react'
import { Table, Input, Select, Tag, DatePicker, Card, Row, Col, Button, Modal, InputNumber, message } from 'antd'
import { auditApi } from '@/api/audit'
import { usePermission } from '@/hooks/usePermission'
import type { AuditLog, AuditQuery } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'

const typeMap: Record<string, string> = {
  login: '登录', logout: '登出', create: '创建', update: '更新',
  delete: '删除', query: '查询', upload: '上传', download: '下载',
  export: '导出', other: '其他',
}

const methodColor: Record<string, string> = {
  GET: 'green', POST: 'blue', PUT: 'orange', DELETE: 'red', PATCH: 'cyan',
}

export default function AuditList() {
  const [data, setData] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [query, setQuery] = useState<AuditQuery>({ page: 1, page_size: 20 })
  const [cleanupModal, setCleanupModal] = useState(false)
  const [cleanupDays, setCleanupDays] = useState(90)
  const { hasRole } = usePermission()
  const isSuperAdmin = hasRole('super_admin')

  const fetchData = async () => {
    setLoading(true)
    try {
      const res = await auditApi.list(query)
      setData(res.data.data.list || [])
      setTotal(res.data.data.total)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [query])

  const handleCleanup = async () => {
    await auditApi.cleanup(cleanupDays)
    message.success('清理完成')
    setCleanupModal(false)
    fetchData()
  }

  const columns: ColumnsType<AuditLog> = [
    { title: '时间', dataIndex: 'created_at', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm:ss') },
    { title: '用户', dataIndex: 'username', width: 100 },
    { title: '模块', dataIndex: 'module', width: 100 },
    { title: '操作', dataIndex: 'action', width: 80 },
    { title: '类型', dataIndex: 'type', width: 80, render: (v: string) => <Tag>{typeMap[v] || v}</Tag> },
    { title: '方法', dataIndex: 'request_method', width: 70, render: (v: string) => <Tag color={methodColor[v]}>{v}</Tag> },
    { title: 'IP', dataIndex: 'request_ip', width: 130 },
    { title: '状态', dataIndex: 'status', width: 80, render: (v: string) => <Tag color={v === 'success' ? 'green' : 'red'}>{v === 'success' ? '成功' : '失败'}</Tag> },
    { title: '耗时', dataIndex: 'duration', width: 80, render: (v: number) => `${v}ms` },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>审计日志</h2>
        {isSuperAdmin && <Button danger onClick={() => setCleanupModal(true)}>清理日志</Button>}
      </div>

      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={[12, 12]}>
          <Col flex="auto">
            <Input.Search placeholder="搜索" allowClear style={{ width: 200 }} onSearch={(v) => setQuery((q) => ({ ...q, keyword: v, page: 1 }))} />
          </Col>
          <Col>
            <Select placeholder="模块" allowClear style={{ width: 120 }} onChange={(v) => setQuery((q) => ({ ...q, module: v, page: 1 }))}>
              {['认证', '用户管理', '组织管理', '案件管理', '任务管理', '方言管理', '文件管理'].map((m) => <Select.Option key={m} value={m}>{m}</Select.Option>)}
            </Select>
          </Col>
          <Col>
            <Select placeholder="类型" allowClear style={{ width: 100 }} onChange={(v) => setQuery((q) => ({ ...q, type: v, page: 1 }))}>
              {Object.entries(typeMap).map(([k, v]) => <Select.Option key={k} value={k}>{v}</Select.Option>)}
            </Select>
          </Col>
          <Col>
            <DatePicker.RangePicker
              onChange={(dates) => {
                setQuery((q) => ({
                  ...q, page: 1,
                  start_time: dates?.[0]?.format('YYYY-MM-DD'),
                  end_time: dates?.[1]?.format('YYYY-MM-DD'),
                }))
              }}
            />
          </Col>
        </Row>
      </Card>

      <Table
        rowKey="id"
        columns={columns}
        dataSource={data}
        loading={loading}
        scroll={{ x: 900 }}
        expandable={{
          expandedRowRender: (record) => (
            <div style={{ fontSize: 13, color: '#646a73' }}>
              <p><b>URL:</b> {record.request_url}</p>
              {record.request_body && <p><b>请求体:</b> <code>{record.request_body}</code></p>}
              {record.error_message && <p><b>错误:</b> <span style={{ color: '#f5222d' }}>{record.error_message}</span></p>}
              <p><b>User-Agent:</b> {record.user_agent}</p>
              <p><b>Trace ID:</b> {record.trace_id}</p>
            </div>
          ),
        }}
        pagination={{
          current: query.page, pageSize: query.page_size, total,
          showSizeChanger: true, showTotal: (t) => `共 ${t} 条`,
          onChange: (page, pageSize) => setQuery((q) => ({ ...q, page, page_size: pageSize })),
        }}
      />

      <Modal title="清理日志" open={cleanupModal} onCancel={() => setCleanupModal(false)} onOk={handleCleanup}>
        <p>清理多少天前的日志？</p>
        <InputNumber min={7} max={365} value={cleanupDays} onChange={(v) => setCleanupDays(v || 90)} addonAfter="天" />
      </Modal>
    </div>
  )
}
