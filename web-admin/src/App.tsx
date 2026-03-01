import { Routes, Route, Navigate } from 'react-router-dom'
import { App as AntApp } from 'antd'
import Login from './pages/Login'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Users from './pages/Users'
import Organizations from './pages/Organizations'
import MissingPersons from './pages/MissingPersons'
import Dialects from './pages/Dialects'
import Tasks from './pages/Tasks'
import Workflows from './pages/Workflows'
import BigScreen from './pages/BigScreen'
import OperationLogs from './pages/OperationLogs'
import Settings from './pages/Settings'
import { useAuthStore } from './stores/auth'
import { RouteGuard, AdminRoute, SuperAdminRoute } from './components/RouteGuard'
import { RoleLevel } from './hooks/usePermission.tsx'

// 受保护路由组件
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((state) => state.token)
  
  if (!token) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

// 权限路由组件
function PermissionRoute({ 
  children, 
  minRole 
}: { 
  children: React.ReactNode
  minRole: RoleLevel 
}) {
  return (
    <ProtectedRoute>
      <RouteGuard minRole={minRole}>
        {children}
      </RouteGuard>
    </ProtectedRoute>
  )
}

function App() {
  return (
    <AntApp>
      <Routes>
        {/* 登录页 */}
        <Route path="/login" element={<Login />} />
        
        {/* 受保护的路由 */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          {/* 所有登录用户可访问 */}
          <Route index element={<Dashboard />} />
          <Route path="organizations" element={<Organizations />} />
          <Route path="missing-persons" element={<MissingPersons />} />
          <Route path="dialects" element={<Dialects />} />
          <Route path="tasks" element={<Tasks />} />
          <Route path="workflows" element={<Workflows />} />
          
          {/* 管理员可访问 */}
          <Route 
            path="users" 
            element={
              <PermissionRoute minRole={RoleLevel.Admin}>
                <Users />
              </PermissionRoute>
            } 
          />
          <Route 
            path="settings" 
            element={
              <PermissionRoute minRole={RoleLevel.Admin}>
                <Settings />
              </PermissionRoute>
            } 
          />
          
          {/* 超级管理员可访问 */}
          <Route 
            path="logs" 
            element={
              <PermissionRoute minRole={RoleLevel.SuperAdmin}>
                <OperationLogs />
              </PermissionRoute>
            } 
          />
        </Route>
        
        {/* 数据大屏 */}
        <Route path="/big-screen" element={<BigScreen />} />
        
        {/* 404 */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AntApp>
  )
}

export default App
