'use client'

import { PlusOutlined, ReloadOutlined } from '@ant-design/icons'
import {
  App,
  Button,
  Card,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import { useAuthGuard } from '@/hooks/useAuthGuard'
import { listFrom } from '@/lib/data'
import { ACTIONS, assignableRoleOptions, canAssignRole, hasPermission, isAdmin, roleLabel } from '@/lib/rbac'
import { userStatusLabel } from '@/lib/status'
import { isMainlandPhone, phoneRuleMessage } from '@/lib/validators'
import { organizationService } from '@/services/organizations'
import { userService } from '@/services/users'
import type { Organization, User } from '@/types/api'

const ROLE_FILTER_OPTIONS = [
  { value: 'volunteer', label: '志愿者' },
  { value: 'manager', label: '管理者' },
  { value: 'admin', label: '管理员' },
  { value: 'super_admin', label: '超级管理员' },
]

const STATUS_OPTIONS = [
  { value: 'active', label: '正常' },
  { value: 'inactive', label: '待审核' },
  { value: 'banned', label: '禁用' },
]

function flattenOrgs(nodes: Organization[], acc: Organization[] = [], depth = 0): Organization[] {
  nodes.forEach((n) => {
    acc.push({ ...n, name: `${'　'.repeat(depth)}${n.name}` })
    if (n.children?.length) flattenOrgs(n.children, acc, depth + 1)
  })
  return acc
}

function normalizeUser(row: User): User {
  const anyRow = row as Record<string, unknown>
  const orgId = (row.org_id || anyRow.orgId || row.organization?.id || '') as string
  const orgName = (row.org_name || anyRow.orgName || row.organization?.name || '') as string
  return {
    ...row,
    org_id: orgId || null,
    org_name: orgName,
    organization: row.organization || (orgId ? { id: orgId, name: orgName } : null),
  }
}

export default function UsersPage() {
  const { ready, user } = useAuthGuard()
  const { message } = App.useApp()
  const canCreate = hasPermission(user, ACTIONS.USER_CREATE) || isAdmin(user)
  const canModify = isAdmin(user)
  const roleOptions = useMemo(() => assignableRoleOptions(user), [user])

  const [loading, setLoading] = useState(true)
  const [items, setItems] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [roleFilter, setRoleFilter] = useState<string | undefined>()
  const [orgFilter, setOrgFilter] = useState<string | undefined>()
  const [orgs, setOrgs] = useState<Organization[]>([])
  const [orgOptions, setOrgOptions] = useState<{ value: string; label: string }[]>([])

  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editing, setEditing] = useState<User | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [createForm] = Form.useForm()
  const [editForm] = Form.useForm()

  const loadOrgs = useCallback(async () => {
    try {
      const tree = await organizationService.tree()
      const flat = flattenOrgs(tree)
      setOrgs(flat)
      setOrgOptions(flat.map((o) => ({ value: o.id, label: o.name })))
    } catch {
      try {
        const data = await organizationService.list({ page: 1, page_size: 100 })
        const list = listFrom<Organization>(data).list
        setOrgs(list)
        setOrgOptions(list.map((o) => ({ value: o.id, label: o.name })))
      } catch {
        setOrgs([])
        setOrgOptions([])
      }
    }
  }, [])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const data = await userService.list({
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
        status: statusFilter || undefined,
        role: roleFilter || undefined,
        org_id: orgFilter || undefined,
      })
      const normalized = listFrom<User>(data)
      setItems(normalized.list.map(normalizeUser))
      setTotal(normalized.total)
    } catch (e) {
      message.error(e instanceof Error ? e.message : '加载用户失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, keyword, statusFilter, roleFilter, orgFilter, message])

  useEffect(() => {
    if (!ready) return
    loadOrgs()
  }, [ready, loadOrgs])

  useEffect(() => {
    if (!ready || typeof window === 'undefined') return
    const fromQuery = new URLSearchParams(window.location.search).get('org_id')
    if (fromQuery) setOrgFilter(fromQuery)
  }, [ready])

  useEffect(() => {
    if (!ready) return
    load()
  }, [ready, load])

  async function onCreate(values: Record<string, unknown>) {
    const password = String(values.password || '')
    const phone = String(values.phone || '').trim()
    if (password.length < 8) {
      message.error('密码至少 8 位')
      return
    }
    const phoneErr = phoneRuleMessage(phone)
    if (phoneErr) {
      message.error(phoneErr)
      return
    }
    if (!values.org_id) {
      message.error('必须绑定组织')
      return
    }
    const role = String(values.role || 'volunteer')
    if (!canAssignRole(user, role)) {
      message.error('无权分配该角色')
      return
    }
    setSubmitting(true)
    try {
      await userService.create({
        nickname: String(values.nickname || '').trim(),
        phone,
        password,
        role,
        org_id: String(values.org_id),
      })
      message.success('用户创建成功')
      setCreateOpen(false)
      createForm.resetFields()
      setPage(1)
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '创建失败')
    } finally {
      setSubmitting(false)
    }
  }

  function openEdit(row: User) {
    setEditing(row)
    const roles = assignableRoleOptions(user)
    const canChangeRole = canAssignRole(user, row.role || '')
    const role = canChangeRole
      ? row.role
      : roles.some((r) => r.value === row.role)
        ? row.role
        : roles[0]?.value || 'volunteer'
    editForm.setFieldsValue({
      nickname: row.nickname,
      email: row.email,
      role,
      status: row.status || 'active',
      org_id: row.org_id || row.organization?.id || undefined,
      _roleLocked: !canChangeRole && !roles.some((r) => r.value === row.role),
    })
    setEditOpen(true)
  }

  async function onEdit(values: Record<string, unknown>) {
    if (!editing) return
    if (!values.org_id) {
      message.error('必须绑定组织')
      return
    }
    const nextRole = String(values.role || editing.role)
    // 不可分配同级/更高角色时，不提交 role 字段，避免误降级
    const rolePayload = canAssignRole(user, nextRole) ? { role: nextRole } : {}
    if (values.role && !canAssignRole(user, nextRole) && nextRole !== editing.role) {
      message.error('无权将该用户调整为所选角色')
      return
    }
    setSubmitting(true)
    try {
      await userService.update(editing.id, {
        nickname: String(values.nickname || '').trim(),
        email: values.email ? String(values.email) : undefined,
        ...rolePayload,
        status: String(values.status || 'active'),
        org_id: String(values.org_id),
      })
      message.success('用户已更新')
      setEditOpen(false)
      setEditing(null)
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '更新失败')
    } finally {
      setSubmitting(false)
    }
  }

  async function setStatus(row: User, status: string) {
    try {
      await userService.updateStatus(row.id, status)
      message.success('状态已更新')
      await load()
    } catch (e) {
      message.error(e instanceof Error ? e.message : '状态更新失败')
    }
  }

  const statusColor = useMemo(
    () =>
      ({
        active: 'success',
        inactive: 'warning',
        banned: 'error',
      }) as Record<string, string>,
    [],
  )

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
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <div>
            <Typography.Title level={4} style={{ margin: 0 }}>
              人员管理
            </Typography.Title>
            <Typography.Text type="secondary">账号、角色、组织归属与状态</Typography.Text>
          </div>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={() => load()}>
              刷新
            </Button>
            {canCreate ? (
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  createForm.setFieldsValue({ role: 'volunteer', org_id: orgFilter })
                  setCreateOpen(true)
                }}
              >
                新建用户
              </Button>
            ) : null}
          </Space>
        </Space>

        <Card size="small">
          <Space wrap style={{ marginBottom: 12 }}>
            <Input.Search
              allowClear
              placeholder="昵称/手机号"
              style={{ width: 220 }}
              onSearch={(v) => {
                setPage(1)
                setKeyword(v.trim())
              }}
            />
            <Select
              allowClear
              placeholder="状态"
              style={{ width: 120 }}
              options={STATUS_OPTIONS}
              value={statusFilter}
              onChange={(v) => {
                setPage(1)
                setStatusFilter(v)
              }}
            />
            <Select
              allowClear
              placeholder="角色"
              style={{ width: 140 }}
              options={ROLE_FILTER_OPTIONS}
              value={roleFilter}
              onChange={(v) => {
                setPage(1)
                setRoleFilter(v)
              }}
            />
            <Select
              allowClear
              showSearch
              optionFilterProp="label"
              placeholder="组织"
              style={{ width: 220 }}
              options={orgOptions}
              value={orgFilter}
              onChange={(v) => {
                setPage(1)
                setOrgFilter(v)
              }}
            />
          </Space>

          <Table
            rowKey="id"
            size="small"
            loading={loading}
            dataSource={items}
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
            columns={[
              { title: '昵称', dataIndex: 'nickname' },
              { title: '手机号', dataIndex: 'phone', width: 130 },
              {
                title: '角色',
                dataIndex: 'role',
                width: 110,
                render: (r: string) => roleLabel(r),
              },
              {
                title: '状态',
                dataIndex: 'status',
                width: 100,
                render: (s: string) => <Tag color={statusColor[s] || 'default'}>{userStatusLabel(s)}</Tag>,
              },
              {
                title: '组织',
                render: (_, row) => row.organization?.name || row.org_name || '-',
              },
              {
                title: '操作',
                width: 220,
                render: (_, row) =>
                  canModify ? (
                    <Space size="small">
                      <Button type="link" size="small" onClick={() => openEdit(row)}>
                        编辑
                      </Button>
                      {row.status !== 'active' ? (
                        <Button type="link" size="small" onClick={() => setStatus(row, 'active')}>
                          启用
                        </Button>
                      ) : (
                        <Button type="link" size="small" onClick={() => setStatus(row, 'banned')}>
                          禁用
                        </Button>
                      )}
                    </Space>
                  ) : (
                    '-'
                  ),
              },
            ]}
          />
        </Card>
      </Space>

      <Modal
        title="新建用户"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={() => createForm.submit()}
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={createForm} layout="vertical" onFinish={onCreate}>
          <Form.Item name="nickname" label="昵称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item
            name="phone"
            label="手机号"
            rules={[
              { required: true, message: '请输入手机号' },
              {
                validator: async (_, v) => {
                  if (!isMainlandPhone(String(v || ''))) throw new Error('请输入正确的大陆手机号')
                },
              },
            ]}
          >
            <Input maxLength={11} />
          </Form.Item>
          <Form.Item
            name="password"
            label="密码"
            rules={[{ required: true }, { min: 8, message: '至少 8 位' }]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true }]} extra="仅可分配低于当前账号的角色">
            <Select options={roleOptions} />
          </Form.Item>
          <Form.Item name="org_id" label="所属组织" rules={[{ required: true, message: '必须绑定组织' }]}>
            <Select showSearch optionFilterProp="label" options={orgOptions} placeholder="请选择组织" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editing ? `编辑：${editing.nickname || editing.phone}` : '编辑用户'}
        open={editOpen}
        onCancel={() => {
          setEditOpen(false)
          setEditing(null)
        }}
        onOk={() => editForm.submit()}
        confirmLoading={submitting}
        destroyOnClose
      >
        <Form form={editForm} layout="vertical" onFinish={onEdit}>
          <Form.Item name="nickname" label="昵称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input />
          </Form.Item>
          <Form.Item
            name="role"
            label="角色"
            rules={[{ required: true }]}
            extra={
              editing && !canAssignRole(user, editing.role || '')
                ? '当前账号无权调整该用户角色，保存时将保持原角色'
                : '仅可分配低于当前账号的角色'
            }
          >
            <Select
              options={
                editing && editing.role && !roleOptions.some((r) => r.value === editing.role)
                  ? [{ value: editing.role, label: roleLabel(editing.role) }, ...roleOptions]
                  : roleOptions
              }
              disabled={!!(editing && !canAssignRole(user, editing.role || '') && !roleOptions.some((r) => r.value === editing.role))}
            />
          </Form.Item>
          <Form.Item name="status" label="状态" rules={[{ required: true }]}>
            <Select options={STATUS_OPTIONS} />
          </Form.Item>
          <Form.Item name="org_id" label="所属组织" rules={[{ required: true, message: '必须绑定组织' }]}>
            <Select showSearch optionFilterProp="label" options={orgOptions} />
          </Form.Item>
        </Form>
      </Modal>
    </AppShell>
  )
}
