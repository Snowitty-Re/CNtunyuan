import { createBrowserRouter, Navigate } from 'react-router-dom'
import { lazy, Suspense } from 'react'
import { Spin } from 'antd'
import MainLayout from '@/layouts/MainLayout'
import RouteGuard from '@/components/RouteGuard'

const Loading = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: 300 }}>
    <Spin size="large" />
  </div>
)

function L(factory: () => Promise<{ default: React.ComponentType }>) {
  const Comp = lazy(factory)
  return (
    <Suspense fallback={<Loading />}>
      <Comp />
    </Suspense>
  )
}

const router = createBrowserRouter([
  { path: '/login', element: L(() => import('@/pages/login/Login')) },
  {
    path: '/',
    element: <RouteGuard><MainLayout /></RouteGuard>,
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: L(() => import('@/pages/dashboard/Dashboard')) },
      // Cases
      { path: 'cases', element: L(() => import('@/pages/cases/List')) },
      { path: 'cases/new', element: L(() => import('@/pages/cases/Form')) },
      { path: 'cases/:id', element: L(() => import('@/pages/cases/Detail')) },
      { path: 'cases/:id/edit', element: L(() => import('@/pages/cases/Form')) },
      // Tasks
      { path: 'tasks', element: L(() => import('@/pages/tasks/List')) },
      { path: 'tasks/new', element: L(() => import('@/pages/tasks/Form')) },
      { path: 'tasks/:id', element: L(() => import('@/pages/tasks/Detail')) },
      { path: 'tasks/:id/edit', element: L(() => import('@/pages/tasks/Form')) },
      // Users
      { path: 'users', element: L(() => import('@/pages/users/List')) },
      { path: 'users/new', element: L(() => import('@/pages/users/Form')) },
      { path: 'users/:id/edit', element: L(() => import('@/pages/users/Form')) },
      // Organizations
      { path: 'organizations', element: L(() => import('@/pages/organizations/List')) },
      { path: 'organizations/new', element: L(() => import('@/pages/organizations/Form')) },
      { path: 'organizations/:id/edit', element: L(() => import('@/pages/organizations/Form')) },
      // Dialects
      { path: 'dialects', element: L(() => import('@/pages/dialects/List')) },
      { path: 'dialects/new', element: L(() => import('@/pages/dialects/Form')) },
      { path: 'dialects/:id/edit', element: L(() => import('@/pages/dialects/Form')) },
      // Audit
      { path: 'audit', element: L(() => import('@/pages/audit/List')) },
      // Profile
      { path: 'profile', element: L(() => import('@/pages/profile/Profile')) },
    ],
  },
])

export default router
