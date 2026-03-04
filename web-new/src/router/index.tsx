import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '@/components/layout/MainLayout';
import RouteGuard from '@/components/layout/RouteGuard';
import LoginPage from '@/pages/login';
import DashboardPage from '@/pages/dashboard';

// 案件管理
import CasesPage from '@/pages/cases';
import CaseDetailPage from '@/pages/cases/detail';
import CaseFormPage from '@/pages/cases/form';

// 任务管理
import TasksPage from '@/pages/tasks';
import TaskDetailPage from '@/pages/tasks/detail';
import TaskFormPage from '@/pages/tasks/form';

// 志愿者管理
import VolunteersPage from '@/pages/volunteers';
import VolunteerDetailPage from '@/pages/volunteers/detail';
import VolunteerFormPage from '@/pages/volunteers/form';

// 组织架构
import OrganizationsPage from '@/pages/organizations';
import OrganizationFormPage from '@/pages/organizations/form';

// 方言管理
import DialectsPage from '@/pages/dialects';
import DialectFormPage from '@/pages/dialects/form';

// 个人中心
import ProfilePage from '@/pages/profile';

// 系统设置
import SettingsPage from '@/pages/settings';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: (
      <RouteGuard>
        <MainLayout />
      </RouteGuard>
    ),
    children: [
      {
        path: '/',
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: '/dashboard',
        element: <DashboardPage />,
      },
      // 案件管理
      {
        path: '/cases',
        element: <CasesPage />,
      },
      {
        path: '/cases/create',
        element: <CaseFormPage />,
      },
      {
        path: '/cases/:id',
        element: <CaseDetailPage />,
      },
      {
        path: '/cases/:id/edit',
        element: <CaseFormPage />,
      },
      // 任务管理
      {
        path: '/tasks',
        element: <TasksPage />,
      },
      {
        path: '/tasks/create',
        element: <TaskFormPage />,
      },
      {
        path: '/tasks/:id',
        element: <TaskDetailPage />,
      },
      {
        path: '/tasks/:id/edit',
        element: <TaskFormPage />,
      },
      // 志愿者管理
      {
        path: '/volunteers',
        element: <VolunteersPage />,
      },
      {
        path: '/volunteers/create',
        element: <VolunteerFormPage />,
      },
      {
        path: '/volunteers/:id',
        element: <VolunteerDetailPage />,
      },
      {
        path: '/volunteers/:id/edit',
        element: <VolunteerFormPage />,
      },
      // 组织架构
      {
        path: '/organizations',
        element: <OrganizationsPage />,
      },
      {
        path: '/organizations/create',
        element: <OrganizationFormPage />,
      },
      {
        path: '/organizations/:id/edit',
        element: <OrganizationFormPage />,
      },
      // 方言管理
      {
        path: '/dialects',
        element: <DialectsPage />,
      },
      {
        path: '/dialects/create',
        element: <DialectFormPage />,
      },
      {
        path: '/dialects/:id/edit',
        element: <DialectFormPage />,
      },
      // 个人中心
      {
        path: '/profile',
        element: <ProfilePage />,
      },
      // 系统设置
      {
        path: '/settings',
        element: <SettingsPage />,
      },
    ],
  },
]);
