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
import BigScreen from './pages/BigScreen'
import { useAuthStore } from './stores/auth'

// 受保护路由组件
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((state) => state.token)
  
  if (!token) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
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
          <Route index element={<Dashboard />} />
          <Route path="users" element={<Users />} />
          <Route path="organizations" element={<Organizations />} />
          <Route path="missing-persons" element={<MissingPersons />} />
          <Route path="dialects" element={<Dialects />} />
          <Route path="tasks" element={<Tasks />} />
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
