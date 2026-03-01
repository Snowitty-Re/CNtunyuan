import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../stores/auth'
import { usePermission, RoleLevel } from '../hooks/usePermission.tsx'
import { Result, Button } from 'antd'

interface RouteGuardProps {
  children: React.ReactNode
  minRole?: RoleLevel
  requireAuth?: boolean
}

/**
 * 路由守卫组件
 * 用于控制页面访问权限
 */
export function RouteGuard({ children, minRole, requireAuth = true }: RouteGuardProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const { token, isAuthenticated } = useAuthStore()
  const { roleLevel, isSuperAdmin } = usePermission()

  useEffect(() => {
    // 检查是否需要登录
    if (requireAuth && !isAuthenticated()) {
      navigate('/login', { state: { from: location.pathname } })
      return
    }
  }, [requireAuth, isAuthenticated, navigate, location])

  // 检查角色权限
  if (minRole !== undefined && roleLevel < minRole) {
    return (
      <Result
        status="403"
        title="403"
        subTitle="抱歉，您没有权限访问此页面"
        extra={
          <Button type="primary" onClick={() => navigate('/')}>
            返回首页
          </Button>
        }
      />
    )
  }

  return <>{children}</>
}

/**
 * 管理员路由守卫
 */
export function AdminRoute({ children }: { children: React.ReactNode }) {
  return (
    <RouteGuard minRole={RoleLevel.Admin}>
      {children}
    </RouteGuard>
  )
}

/**
 * 管理者路由守卫
 */
export function ManagerRoute({ children }: { children: React.ReactNode }) {
  return (
    <RouteGuard minRole={RoleLevel.Manager}>
      {children}
    </RouteGuard>
  )
}

/**
 * 超级管理员路由守卫
 */
export function SuperAdminRoute({ children }: { children: React.ReactNode }) {
  return (
    <RouteGuard minRole={RoleLevel.SuperAdmin}>
      {children}
    </RouteGuard>
  )
}

/**
 * 公开路由守卫（不需要登录）
 */
export function PublicRoute({ children }: { children: React.ReactNode }) {
  return (
    <RouteGuard requireAuth={false}>
      {children}
    </RouteGuard>
  )
}

export default RouteGuard
