import { lazy } from 'react';
import type { RouteObject } from 'react-router-dom';
import AdminLayout from '@/layouts/AdminLayout';
import BlankLayout from '@/layouts/BlankLayout';
import AuthGuard from './AuthGuard';
import RoleGuard from './RoleGuard';

const Login = lazy(() => import('@/pages/login'));
const ResetPassword = lazy(() => import('@/pages/login/ResetPassword'));
const Dashboard = lazy(() => import('@/pages/dashboard'));
const Users = lazy(() => import('@/pages/users'));
const UserDetail = lazy(() => import('@/pages/users/Detail'));
const Organizations = lazy(() => import('@/pages/organizations'));
const OrgTree = lazy(() => import('@/pages/organizations/OrgTree'));
const MissingPersons = lazy(() => import('@/pages/missing-persons'));
const MissingPersonDetail = lazy(() => import('@/pages/missing-persons/Detail'));
const MissingPersonForm = lazy(() => import('@/pages/missing-persons/Form'));
const Tasks = lazy(() => import('@/pages/tasks'));
const TaskDetail = lazy(() => import('@/pages/tasks/Detail'));
const TaskForm = lazy(() => import('@/pages/tasks/Form'));
const Dialects = lazy(() => import('@/pages/dialects'));
const DialectDetail = lazy(() => import('@/pages/dialects/Detail'));
const DialectForm = lazy(() => import('@/pages/dialects/Form'));
const Files = lazy(() => import('@/pages/files'));
const FileStats = lazy(() => import('@/pages/files/Stats'));
const AuditLogs = lazy(() => import('@/pages/audit'));
const AuditDetail = lazy(() => import('@/pages/audit/Detail'));
const AuditStats = lazy(() => import('@/pages/audit/Stats'));
const Profile = lazy(() => import('@/pages/profile'));
const ChangePassword = lazy(() => import('@/pages/profile/ChangePassword'));
const SystemHealth = lazy(() => import('@/pages/system/Health'));
const Error403 = lazy(() => import('@/pages/error/403'));
const Error404 = lazy(() => import('@/pages/error/404'));
const Error500 = lazy(() => import('@/pages/error/500'));

export const routes: RouteObject[] = [
  {
    element: <BlankLayout />,
    children: [
      { path: '/login', element: <Login /> },
      { path: '/reset-password', element: <ResetPassword /> },
      { path: '/403', element: <Error403 /> },
      { path: '/500', element: <Error500 /> },
    ],
  },
  {
    element: (
      <AuthGuard>
        <AdminLayout />
      </AuthGuard>
    ),
    children: [
      { path: '/', element: <Dashboard /> },
      { path: '/dashboard', element: <Dashboard /> },
      { path: '/missing-persons', element: <MissingPersons /> },
      { path: '/missing-persons/new', element: <MissingPersonForm /> },
      { path: '/missing-persons/:id', element: <MissingPersonDetail /> },
      { path: '/missing-persons/:id/edit', element: <MissingPersonForm /> },
      { path: '/tasks', element: <Tasks /> },
      { path: '/tasks/new', element: <TaskForm /> },
      { path: '/tasks/:id', element: <TaskDetail /> },
      { path: '/tasks/:id/edit', element: <TaskForm /> },
      { path: '/dialects', element: <Dialects /> },
      { path: '/dialects/new', element: <DialectForm /> },
      { path: '/dialects/:id', element: <DialectDetail /> },
      { path: '/dialects/:id/edit', element: <DialectForm /> },
      {
        path: '/users',
        element: <RoleGuard minRole="admin"><Users /></RoleGuard>,
      },
      {
        path: '/users/:id',
        element: <RoleGuard minRole="admin"><UserDetail /></RoleGuard>,
      },
      {
        path: '/organizations',
        element: <RoleGuard minRole="admin"><Organizations /></RoleGuard>,
      },
      {
        path: '/organizations/tree',
        element: <RoleGuard minRole="admin"><OrgTree /></RoleGuard>,
      },
      {
        path: '/files',
        element: <RoleGuard minRole="manager"><Files /></RoleGuard>,
      },
      {
        path: '/files/stats',
        element: <RoleGuard minRole="manager"><FileStats /></RoleGuard>,
      },
      {
        path: '/audit',
        element: <RoleGuard minRole="manager"><AuditLogs /></RoleGuard>,
      },
      {
        path: '/audit/:id',
        element: <RoleGuard minRole="manager"><AuditDetail /></RoleGuard>,
      },
      {
        path: '/audit/stats',
        element: <RoleGuard minRole="manager"><AuditStats /></RoleGuard>,
      },
      {
        path: '/system/health',
        element: <RoleGuard minRole="admin"><SystemHealth /></RoleGuard>,
      },
      { path: '/profile', element: <Profile /> },
      { path: '/profile/change-password', element: <ChangePassword /> },
    ],
  },
  { path: '*', element: <Error404 /> },
];
