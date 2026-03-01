import { createBrowserRouter, Navigate } from 'react-router-dom';
import MainLayout from '@/components/layout/MainLayout';
import LoginPage from '@/pages/login';
import DashboardPage from '@/pages/dashboard';
import CasesPage from '@/pages/cases';
import TasksPage from '@/pages/tasks';
import VolunteersPage from '@/pages/volunteers';
import OrganizationsPage from '@/pages/organizations';
import DialectsPage from '@/pages/dialects';

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: <MainLayout />,
    children: [
      {
        path: '/',
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: '/dashboard',
        element: <DashboardPage />,
      },
      {
        path: '/cases',
        element: <CasesPage />,
      },
      {
        path: '/tasks',
        element: <TasksPage />,
      },
      {
        path: '/volunteers',
        element: <VolunteersPage />,
      },
      {
        path: '/organizations',
        element: <OrganizationsPage />,
      },
      {
        path: '/dialects',
        element: <DialectsPage />,
      },
    ],
  },
]);
