'use client'

import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  ReloadOutlined,
  SwapOutlined,
} from '@ant-design/icons'
import {
  App,
  Breadcrumb,
  Button,
  Card,
  Col,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Table,
  Tag,
  Tree,
  Typography,
} from 'antd'
import type { DataNode } from 'antd/es/tree'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { listFrom } from '@/lib/data'
import { ORG_STATUS_OPTIONS, ORG_TYPE_OPTIONS, orgStatusLabel, orgTypeLabel } from '@/lib/orgTypes'
import { isAdmin, isSuperAdmin } from '@/lib/rbac'
import { organizationService } from '@/services/organizations'
import type { Organization } from '@/types/api'

type FlatOrg = Organization & { key: string }

function flattenOrgs(nodes: Organization[], acc: FlatOrg[] = []): FlatOrg[] {
  nodes.forEach((n) => {
    acc.push({ ...n, key: n.id })
    if (n.children?.length) flattenOrgs(n.children, acc)
  })
  return acc
}

function toTreeData(nodes: Organization[]): DataNode[] {
  return nodes.map((n) => ({
    key: n.id,
    title: `${n.name}（${orgTypeLabel(n.type)}）`,
    children: n.children?.length ? toTreeData(n.children) : undefined,
  }))
}

/** Exclude self and descendants from parent candidates */
function collectDescendantIds(node: Organization | undefined, set: Set<string>) {
  if (!node) return
  set.add(node.id)
  node.children?.forEach((c) => collectDescendantIds(c, set))
}

function findNode(nodes: Organization[], id: string): Organization | undefined {
  for (const n of nodes) {
    if (n.id === id) return n
    if (n.children?.length) {
      const hit = findNode(n.children, id)
      if (hit) return hit
    }
  }
  return undefined
}

export default function OrganizationsPage() {
  const { ready, user } = useAuthGuard({ requireAdmin: true })
  const { message } = App.useApp()
  const canWrite = isAdmin(user)
  const allowRootOrg = isSuperAdmin(user)

  const [loading, setLoading] = useState(true)
  const [tree, setTree] = useState<Organization[]>([])
  const [list, setList] = useState<Organization[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [typeFilter, setTypeFilter] = useState<string | undefined>()
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [path, setPath] = useState<Organization[]>([])

  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [moveOpen, setMoveOpen] = useState(false)
  const [editing, setEditing] = useState<Organization | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()
  const [moveForm] = Form.useForm()

  const flat = useMemo(() => flattenOrgs(tree), [tree])
  const selected = useMemo(
    () => (selectedId ? flat.find((o) => o.id === selectedId) || list.find((o) => o.id === selectedId) || null : null),
    [selectedId, flat, list],
  )

  const blockedParentIds = useMemo(() => {
    const set = new Set<string>()
    if (editing) collectDescendantIds(findNode(tree, editing.id), set)
    return set
  }, [editing, tree])

  const parentOptions = useMemo(
    () =>
      flat
        .filter((o) => !blockedParentIds.has(o.id))
        .map((o) => ({ value: o.id, label: `${o.name}（${orgTypeLabel(o.type)}）` })),
    [flat, blockedParentIds],
  )

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [treeData, listData] = await Promise.all([
        organizationService.tree(),
        organizationService.list({
          page,
          page_size: pageSize,
          keyword: keyword || undefined,
          type: typeFilter || undefined,
        }),
      ])
      setTree(treeData)
      const pageResult = listFrom<Organization>(listData)
      setList(pageResult.list)
      setTotal(pageResult.total)
      if (!selectedId && treeData[0]?.id) setSelectedId(treeData[0].id)
    } catch (e) {
      message.error(e instanceof Error ? e.message : '加载组织失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, keyword, typeFilter, message, selectedId])

  useEffect(() => {
    if (!ready) return
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, page, pageSize, keyword, typeFilter])

  useEffect(() => {
    if (!selectedId || !ready) {
      setPath([])
      return
    }
    organizationService
      .path(selectedId)
      .then((p) => setPath(Array.isArray(p) ? p : []))
      .catch(() => setPath([]))
  }, [selectedId, ready])

  async function onCreate(values: Record<string, unknown>) {
    const parentId = values.parent_id ? String(values.parent_id) : ''
    if (!parentId && !allowRootOrg) {
      message.error('请选择父组织（仅超级管理员可创建顶级组织）')
      return
    }
    setSubmitting(true)
    try {
      await organizationService.create({
        name: String(values.name || '').trim(),
        code: String(values.code || '').trim(),
        type: String(values.type || 'team'),
        parent_id: parentId || undefined,
        description: values.description ? String(values.description) : undefined,
        address: values.address ? String(values.address) : undefined,
        contact_name: values.contact_name ? String(values.contact_name) : undefined,
        contact_phone: values.contact_phone ? String(values.contact_phone) : undefined,
        sort_order: typeof values.sort_order === 'number' ? values.sort_order : undefined,
      })
      message.success('组织创建成功')
      setCreateOpen(false)
      createForm.resetFields()
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '创建失败')
    } finally {
      setSubmitting(false)
    }
  }

  function openEdit(org: Organization) {
    setEditing(org)
    editForm.setFieldsValue({
      name: org.name,
      code: org.code,
      description: org.description,
      address: org.address,
      contact_name: org.contact_name,
      contact_phone: org.contact_phone,
      status: org.status || 'active',
      sort_order: org.sort_order ?? 0,
    })
    setEditOpen(true)
  }

  async function onEdit(values: Record<string, unknown>) {
    if (!editing) return
    setSubmitting(true)
    try {
      await organizationService.update(editing.id, {
        name: String(values.name || '').trim(),
        code: String(values.code || '').trim(),
        description: values.description != null ? String(values.description) : undefined,
        address: values.address != null ? String(values.address) : undefined,
        contact_name: values.contact_name != null ? String(values.contact_name) : undefined,
        contact_phone: values.contact_phone != null ? String(values.contact_phone) : undefined,
        status: values.status ? String(values.status) : undefined,
        sort_order: typeof values.sort_order === 'number' ? values.sort_order : undefined,
      })
      message.success('组织已更新')
      setEditOpen(false)
      setEditing(null)
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '更新失败')
    } finally {
      setSubmitting(false)
    }
  }

  function openMove(org: Organization) {
    setEditing(org)
    moveForm.setFieldsValue({ new_parent_id: org.parent_id || undefined })
    setMoveOpen(true)
  }

  async function onMove(values: Record<string, unknown>) {
    if (!editing) return
    const parentId = String(values.new_parent_id || '').trim()
    if (!parentId && !allowRootOrg) {
      message.warning('请选择目标父组织')
      return
    }
    setSubmitting(true)
    try {
      // 超管允许空 parent 表示移到顶级（后端支持）
      await organizationService.move(editing.id, parentId || '')
      message.success('组织已移动')
      setMoveOpen(false)
      setEditing(null)
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '移动失败')
    } finally {
      setSubmitting(false)
    }
  }

  async function onDelete(org: Organization) {
    try {
      await organizationService.remove(org.id)
      message.success('已删除')
      if (selectedId === org.id) setSelectedId(null)
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '删除失败')
    }
  }

  if (!ready) {
    return (
      <AppShell>
        <Card loading />
      </AppShell>
    )
  }

  return (
    <AppShell>
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Typography.Title level={4} style={{ margin: 0 }}>
              组织管理
            </Typography.Title>
            <Typography.Text type="secondary">层级结构、编码治理与组织信息（管理员）</Typography.Text>
          </Col>
          <Col>
            <Space>
              <Button icon={<ReloadOutlined />} onClick={() => load()}>
                刷新
              </Button>
              {canWrite ? (
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={() => {
                    createForm.setFieldsValue({
                      type: 'team',
                      parent_id: selectedId || undefined,
                      sort_order: 0,
                    })
                    setCreateOpen(true)
                  }}
                >
                  新建组织
                </Button>
              ) : null}
            </Space>
          </Col>
        </Row>

        <Row gutter={16}>
          <Col xs={24} lg={8}>
            <Card title="组织树" size="small" loading={loading}>
              {tree.length ? (
                <Tree
                  showLine
                  defaultExpandAll
                  selectedKeys={selectedId ? [selectedId] : []}
                  treeData={toTreeData(tree)}
                  onSelect={(keys) => {
                    if (keys[0]) setSelectedId(String(keys[0]))
                  }}
                />
              ) : (
                <Empty description="暂无组织树" />
              )}
            </Card>
          </Col>
          <Col xs={24} lg={16}>
            <Card size="small" style={{ marginBottom: 16 }}>
              {path.length ? (
                <Breadcrumb
                  items={path.map((p) => ({
                    title: (
                      <a
                        onClick={(e) => {
                          e.preventDefault()
                          setSelectedId(p.id)
                        }}
                      >
                        {p.name}
                      </a>
                    ),
                  }))}
                />
              ) : (
                <Typography.Text type="secondary">选择左侧节点查看路径</Typography.Text>
              )}
              {selected ? (
                <div style={{ marginTop: 12 }}>
                  <Space wrap>
                    <Typography.Text strong>{selected.name}</Typography.Text>
                    <Tag>{orgTypeLabel(selected.type)}</Tag>
                    <Tag color={selected.status === 'inactive' ? 'default' : 'success'}>
                      {orgStatusLabel(selected.status)}
                    </Tag>
                    <Typography.Text type="secondary">编码 {selected.code}</Typography.Text>
                  </Space>
                  {canWrite ? (
                    <div style={{ marginTop: 12 }}>
                      <Space>
                        <Button size="small" icon={<EditOutlined />} onClick={() => openEdit(selected)}>
                          编辑
                        </Button>
                        <Button size="small" icon={<SwapOutlined />} onClick={() => openMove(selected)}>
                          移动
                        </Button>
                        <Popconfirm title={`确认删除「${selected.name}」？`} onConfirm={() => onDelete(selected)}>
                          <Button size="small" danger icon={<DeleteOutlined />}>
                            删除
                          </Button>
                        </Popconfirm>
                        <Button
                          size="small"
                          type="link"
                          href={`/users?org_id=${selected.id}`}
                          onClick={(e) => {
                            e.preventDefault()
                            window.location.href = `/users?org_id=${selected.id}`
                          }}
                        >
                          查看成员
                        </Button>
                      </Space>
                    </div>
                  ) : null}
                  {selected.description ? (
                    <Typography.Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0 }}>
                      {selected.description}
                    </Typography.Paragraph>
                  ) : null}
                </div>
              ) : null}
            </Card>

            <Card title="组织列表" size="small">
              <Space wrap style={{ marginBottom: 12 }}>
                <Input.Search
                  allowClear
                  placeholder="名称/编码"
                  style={{ width: 220 }}
                  onSearch={(v) => {
                    setPage(1)
                    setKeyword(v.trim())
                  }}
                />
                <Select
                  allowClear
                  placeholder="类型"
                  style={{ width: 140 }}
                  options={ORG_TYPE_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
                  onChange={(v) => {
                    setPage(1)
                    setTypeFilter(v)
                  }}
                />
              </Space>
              <Table
                rowKey="id"
                size="small"
                loading={loading}
                dataSource={list}
                pagination={{
                  current: page,
                  pageSize,
                  total,
                  showSizeChanger: true,
                  onChange: (p, ps) => {
                    setPage(p)
                    setPageSize(ps)
                  },
                }}
                onRow={(row) => ({
                  onClick: () => setSelectedId(row.id),
                  style: { cursor: 'pointer' },
                })}
                columns={[
                  { title: '名称', dataIndex: 'name' },
                  { title: '编码', dataIndex: 'code', width: 120 },
                  {
                    title: '类型',
                    dataIndex: 'type',
                    width: 100,
                    render: (t: string) => orgTypeLabel(t),
                  },
                  {
                    title: '状态',
                    dataIndex: 'status',
                    width: 90,
                    render: (s: string) => (
                      <Tag color={s === 'inactive' ? 'default' : 'success'}>{orgStatusLabel(s)}</Tag>
                    ),
                  },
                  {
                    title: '操作',
                    width: 200,
                    render: (_, row) =>
                      canWrite ? (
                        <Space size="small" onClick={(e) => e.stopPropagation()}>
                          <Button type="link" size="small" onClick={() => openEdit(row)}>
                            编辑
                          </Button>
                          <Button type="link" size="small" onClick={() => openMove(row)}>
                            移动
                          </Button>
                          <Popconfirm title="确认删除？" onConfirm={() => onDelete(row)}>
                            <Button type="link" size="small" danger>
                              删除
                            </Button>
                          </Popconfirm>
                        </Space>
                      ) : (
                        '-'
                      ),
                  },
                ]}
              />
            </Card>
          </Col>
        </Row>
      </Space>

      <Modal
        title="新建组织"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={() => createForm.submit()}
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={createForm} layout="vertical" onFinish={onCreate}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true, message: '请输入唯一编码' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={ORG_TYPE_OPTIONS.map((o) => ({ value: o.value, label: o.label }))} />
          </Form.Item>
          <Form.Item
            name="parent_id"
            label="父组织"
            rules={allowRootOrg ? [] : [{ required: true, message: '请选择父组织' }]}
            extra={allowRootOrg ? '不选则创建为顶级组织' : '必须挂在可管理的父组织下'}
          >
            <Select allowClear={allowRootOrg} showSearch optionFilterProp="label" options={parentOptions} />
          </Form.Item>
          <Form.Item name="contact_name" label="联系人">
            <Input />
          </Form.Item>
          <Form.Item name="contact_phone" label="联系电话">
            <Input />
          </Form.Item>
          <Form.Item name="address" label="地址">
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editing ? `编辑：${editing.name}` : '编辑组织'}
        open={editOpen}
        onCancel={() => {
          setEditOpen(false)
          setEditing(null)
        }}
        onOk={() => editForm.submit()}
        confirmLoading={submitting}
        destroyOnClose
      >
        <Typography.Paragraph type="secondary" style={{ marginBottom: 12 }}>
          类型创建后不可修改；调整层级请使用「移动」。
        </Typography.Paragraph>
        {editing ? (
          <Space style={{ marginBottom: 12 }}>
            <Tag>{orgTypeLabel(editing.type)}</Tag>
            <Typography.Text type="secondary">ID {editing.id}</Typography.Text>
          </Space>
        ) : null}
        <Form form={editForm} layout="vertical" onFinish={onEdit}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="code" label="编码" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select options={ORG_STATUS_OPTIONS.map((o) => ({ value: o.value, label: o.label }))} />
          </Form.Item>
          <Form.Item name="contact_name" label="联系人">
            <Input />
          </Form.Item>
          <Form.Item name="contact_phone" label="联系电话">
            <Input />
          </Form.Item>
          <Form.Item name="address" label="地址">
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="sort_order" label="排序">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editing ? `移动：${editing.name}` : '移动组织'}
        open={moveOpen}
        onCancel={() => {
          setMoveOpen(false)
          setEditing(null)
        }}
        onOk={() => moveForm.submit()}
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={moveForm} layout="vertical" onFinish={onMove}>
          <Form.Item
            name="new_parent_id"
            label="新父组织"
            rules={allowRootOrg ? [] : [{ required: true, message: '请选择父组织' }]}
            extra={allowRootOrg ? '不选表示移到顶级（仅超级管理员）' : '仅可移到可管理范围内的父组织'}
          >
            <Select allowClear={allowRootOrg} showSearch optionFilterProp="label" options={parentOptions} placeholder={allowRootOrg ? '顶级组织' : '选择父组织'} />
          </Form.Item>
        </Form>
      </Modal>
    </AppShell>
  )
}
