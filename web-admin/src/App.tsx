import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/auth'
import Login from './pages/Login'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Users from './pages/Users'
import Organizations from './pages/Organizations'
import MissingPersons from './pages/MissingPersons'
import Dialects from './pages/Dialects'
import Tasks from './pages/Tasks'
import BigScreen from './pages/BigScreen'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((state) => state.token)
  return token ? <>{children}</> : <Navigate to="/login" />
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/big-screen" element={<BigScreen />} />
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
      </Routes>
    </BrowserRouter>
  )
}

export default App
