'use client'

import {
  AuditOutlined,
  ClusterOutlined,
  DashboardOutlined,
  FileOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SettingOutlined,
  SoundOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons'
import { Alert, Avatar, Button, Card, Layout, Menu, Space, Tag, Typography, theme } from 'antd'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { PropsWithChildren, useEffect, useMemo, useState } from 'react'
import { useAuth } from '@/components/providers/AuthProvider'
import { SafeImage } from '@/components/shared/SafeImage'
import { useSiteBrand } from '@/hooks/useSiteBrand'
import { isNavActive, navItemsForUser, resolveWorkbench, workbenchLabel } from '@/lib/nav'
import { roleLabel } from '@/lib/rbac'

const { Header, Sider, Content } = Layout

const iconFor = (href: string) => {
  if (href.startsWith('/dashboard')) return <DashboardOutlined />
  if (href.startsWith('/organizations')) return <ClusterOutlined />
  if (href.startsWith('/users')) return <TeamOutlined />
  if (href.startsWith('/cases')) return <FileOutlined />
  if (href.startsWith('/tasks')) return <FileOutlined />
  if (href.startsWith('/dialects')) return <SoundOutlined />
  if (href.startsWith('/attachments')) return <FileOutlined />
  if (href.startsWith('/audit')) return <AuditOutlined />
  if (href.startsWith('/settings') || href.startsWith('/site') || href.startsWith('/feature') || href.startsWith('/monitor')) {
    return <SettingOutlined />
  }
  return <UserOutlined />
}

export function AppShell({ children }: PropsWithChildren) {
  const pathname = usePathname()
  const router = useRouter()
  const { user, logout, blockReason, blockMessage } = useAuth()
  const brand = useSiteBrand()
  const [collapsed, setCollapsed] = useState(false)
  const { token } = theme.useToken()

  const nav = useMemo(() => navItemsForUser(user), [user])
  const hrefs = useMemo(() => nav.map((n) => n.href), [nav])
  const workbench = resolveWorkbench(user)

  useEffect(() => {
    document.title = brand.title
  }, [brand.title])

  const selectedKeys = useMemo(() => {
    const active = nav.find((item) => isNavActive(pathname, item.href, hrefs))
    return active ? [active.href] : []
  }, [nav, pathname, hrefs])

  if (blockReason) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24 }}>
        <Card style={{ maxWidth: 480, width: '100%' }}>
          <Typography.Title level={4}>无法进入管理端</Typography.Title>
          <Alert type="warning" showIcon message={blockMessage} style={{ marginBottom: 16 }} />
          <Typography.Paragraph type="secondary">
            {blockReason === 'no_phone'
              ? '请使用已绑定真实手机号的账号登录，或联系管理员完善资料。'
              : '请联系管理员审批通过后再登录。'}
          </Typography.Paragraph>
          <Button
            type="primary"
            danger
            onClick={async () => {
              await logout()
              router.replace('/login')
            }}
          >
            退出登录
          </Button>
        </Card>
      </div>
    )
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        trigger={null}
        width={240}
        style={{
          background: 'linear-gradient(165deg, #8f4f1f 0%, #a95f24 42%, #c56a2c 100%)',
        }}
      >
        <div style={{ padding: collapsed ? 12 : '16px 14px', color: '#fff9f2' }}>
          <Space align="center" size={12}>
            {brand.logoUrl ? (
              <SafeImage src={brand.logoUrl} alt={brand.orgName} width={40} height={40} style={{ borderRadius: 12 }} />
            ) : (
              <Avatar shape="square" size={40} style={{ background: 'rgba(255,255,255,0.2)', borderRadius: 12 }}>
                {brand.orgName.slice(0, 1)}
              </Avatar>
            )}
            {!collapsed ? (
              <div>
                <div style={{ fontWeight: 700, lineHeight: 1.2 }}>{brand.orgName}</div>
                <Typography.Text style={{ color: 'rgba(255,249,242,0.78)', fontSize: 12 }}>
                  {workbenchLabel(workbench)}
                </Typography.Text>
              </div>
            ) : null}
          </Space>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={selectedKeys}
          style={{ background: 'transparent', border: 'none' }}
          items={nav.map((item) => ({
            key: item.href,
            icon: iconFor(item.href),
            label: <Link href={item.href}>{item.label}</Link>,
          }))}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: token.colorBgContainer,
            padding: '0 20px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
          }}
        >
          <Space>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed((v) => !v)}
            />
            <div>
              <Typography.Title level={5} style={{ margin: 0 }}>
                {brand.title}
              </Typography.Title>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                {brand.subtitle}
              </Typography.Text>
            </div>
          </Space>
          <Space size="middle">
            <Avatar src={user?.avatar || '/default-avatar.svg'} icon={<UserOutlined />} />
            <span>{user?.nickname || user?.phone || '未登录'}</span>
            <Tag color="orange">{roleLabel(user?.role)}</Tag>
            <Button
              type="text"
              icon={<LogoutOutlined />}
              onClick={async () => {
                await logout()
                router.replace('/login')
              }}
            >
              退出
            </Button>
          </Space>
        </Header>
        <Content style={{ margin: 20 }}>{children}</Content>
      </Layout>
    </Layout>
  )
}
