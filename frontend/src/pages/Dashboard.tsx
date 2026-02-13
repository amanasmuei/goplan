import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { taskApi } from '../services/api'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import {
  ListTodo,
  Clock,
  AlertTriangle,
  CheckCircle2,
  Plus,
  TrendingUp,
} from 'lucide-react'

export default function Dashboard() {
  useDocumentTitle('Dashboard')
  const { data: tasksData } = useQuery({
    queryKey: ['tasks'],
    queryFn: () => taskApi.list({ page: 1, page_size: 10 }),
  })

  const tasks = tasksData?.tasks || []

  const stats = {
    total: tasks.length,
    active: tasks.filter((t) => t.status === 'active').length,
    blocked: tasks.filter((t) => t.status === 'blocked').length,
    completed: tasks.filter((t) => t.status === 'completed').length,
  }

  const recentTasks = tasks.slice(0, 5)

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-gray-500 mt-1">
            Overview of your task intelligence metrics
          </p>
        </div>
        <Link
          to="/tasks/new"
          className="bg-primary-600 text-white px-4 py-2 rounded-lg font-medium
                   hover:bg-primary-700 transition-colors flex items-center gap-2"
        >
          <Plus className="h-5 w-5" />
          New Task
        </Link>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-lg bg-blue-50 flex items-center justify-center">
              <ListTodo className="h-6 w-6 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Total Tasks</p>
              <p className="text-2xl font-bold text-gray-900">{stats.total}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-lg bg-yellow-50 flex items-center justify-center">
              <Clock className="h-6 w-6 text-yellow-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Active</p>
              <p className="text-2xl font-bold text-gray-900">{stats.active}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-lg bg-red-50 flex items-center justify-center">
              <AlertTriangle className="h-6 w-6 text-red-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Blocked</p>
              <p className="text-2xl font-bold text-gray-900">{stats.blocked}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-200">
          <div className="flex items-center gap-4">
            <div className="h-12 w-12 rounded-lg bg-green-50 flex items-center justify-center">
              <CheckCircle2 className="h-6 w-6 text-green-600" />
            </div>
            <div>
              <p className="text-sm text-gray-500">Completed</p>
              <p className="text-2xl font-bold text-gray-900">{stats.completed}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Tasks */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200">
          <div className="p-6 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">Recent Tasks</h2>
          </div>
          <div className="divide-y divide-gray-100">
            {recentTasks.length > 0 ? (
              recentTasks.map((task) => (
                <Link
                  key={task.id}
                  to={`/tasks/${task.id}`}
                  className="flex items-center justify-between p-4 hover:bg-gray-50 transition-colors"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {task.title}
                    </p>
                    <p className="text-xs text-gray-500 mt-1">
                      {new Date(task.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <span
                    className={`px-2 py-1 text-xs font-medium rounded-full ${
                      task.status === 'completed'
                        ? 'bg-green-100 text-green-700'
                        : task.status === 'blocked'
                        ? 'bg-red-100 text-red-700'
                        : task.status === 'active'
                        ? 'bg-yellow-100 text-yellow-700'
                        : 'bg-gray-100 text-gray-700'
                    }`}
                  >
                    {task.status.replace('_', ' ')}
                  </span>
                </Link>
              ))
            ) : (
              <div className="p-8 text-center text-gray-500">
                <ListTodo className="h-12 w-12 mx-auto mb-2 text-gray-300" />
                <p>No tasks yet. Create your first task to get started.</p>
              </div>
            )}
          </div>
          {tasks.length > 5 && (
            <div className="p-4 border-t border-gray-100">
              <Link
                to="/tasks"
                className="text-sm text-primary-600 hover:text-primary-700 font-medium"
              >
                View all tasks
              </Link>
            </div>
          )}
        </div>

        {/* Insights */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200">
          <div className="p-6 border-b border-gray-200">
            <h2 className="text-lg font-semibold text-gray-900">Insights</h2>
          </div>
          <div className="p-6 space-y-4">
            <div className="flex items-start gap-4">
              <div className="h-10 w-10 rounded-lg bg-primary-50 flex items-center justify-center flex-shrink-0">
                <TrendingUp className="h-5 w-5 text-primary-600" />
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900">
                  Prediction Accuracy
                </p>
                <p className="text-sm text-gray-500 mt-1">
                  Start completing tasks with reviews to see prediction accuracy
                  metrics.
                </p>
              </div>
            </div>

            <div className="flex items-start gap-4">
              <div className="h-10 w-10 rounded-lg bg-orange-50 flex items-center justify-center flex-shrink-0">
                <AlertTriangle className="h-5 w-5 text-orange-600" />
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900">
                  Common Blockers
                </p>
                <p className="text-sm text-gray-500 mt-1">
                  Log blockers to identify patterns and prevent repeated delays.
                </p>
              </div>
            </div>

            <div className="flex items-start gap-4">
              <div className="h-10 w-10 rounded-lg bg-green-50 flex items-center justify-center flex-shrink-0">
                <CheckCircle2 className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm font-medium text-gray-900">
                  Context Linking Rate
                </p>
                <p className="text-sm text-gray-500 mt-1">
                  Link tasks to historical context to improve future predictions.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
