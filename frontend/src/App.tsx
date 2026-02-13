import { lazy, Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './store/authStore'
import ErrorBoundary from './components/ErrorBoundary'
import Layout from './components/Layout'

const Login = lazy(() => import('./pages/Login'))
const Register = lazy(() => import('./pages/Register'))
const Dashboard = lazy(() => import('./pages/Dashboard'))
const TaskList = lazy(() => import('./pages/TaskList'))
const CreateTask = lazy(() => import('./pages/CreateTask'))
const TaskDetail = lazy(() => import('./pages/TaskDetail'))
const Teams = lazy(() => import('./pages/Teams'))
const TeamDetail = lazy(() => import('./pages/TeamDetail'))
const Projects = lazy(() => import('./pages/Projects'))
const ProjectDetail = lazy(() => import('./pages/ProjectDetail'))
const Analytics = lazy(() => import('./pages/Analytics'))

function LoadingFallback() {
  return (
    <div className="flex justify-center items-center h-64" aria-live="polite">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      <span className="sr-only">Loading page...</span>
    </div>
  )
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <Routes>
        <Route path="/login" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load login page. Please refresh.</div>}><Login /></ErrorBoundary>} />
        <Route path="/register" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load registration page. Please refresh.</div>}><Register /></ErrorBoundary>} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="/dashboard" replace />} />
          <Route path="dashboard" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load dashboard. Please refresh.</div>}><Dashboard /></ErrorBoundary>} />
          <Route path="tasks" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load tasks. Please refresh.</div>}><TaskList /></ErrorBoundary>} />
          <Route path="tasks/new" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load task creation. Please refresh.</div>}><CreateTask /></ErrorBoundary>} />
          <Route path="tasks/:id" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load task details. Please refresh.</div>}><TaskDetail /></ErrorBoundary>} />
          <Route path="teams" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load teams. Please refresh.</div>}><Teams /></ErrorBoundary>} />
          <Route path="teams/:id" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load team details. Please refresh.</div>}><TeamDetail /></ErrorBoundary>} />
          <Route path="projects" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load projects. Please refresh.</div>}><Projects /></ErrorBoundary>} />
          <Route path="projects/:id" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load project details. Please refresh.</div>}><ProjectDetail /></ErrorBoundary>} />
          <Route path="analytics" element={<ErrorBoundary fallback={<div className="p-8 text-center text-red-600" role="alert">Failed to load analytics. Please refresh.</div>}><Analytics /></ErrorBoundary>} />
        </Route>
      </Routes>
    </Suspense>
  )
}

export default App
