import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { taskApi } from '../services/api'
import { useDebounce } from '../hooks/useDebounce'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import type { TaskStatus } from '../types'
import {
  Plus,
  Search,
  Filter,
  ChevronLeft,
  ChevronRight,
  ListTodo,
} from 'lucide-react'
import { clsx } from 'clsx'

const statusOptions: { value: TaskStatus | ''; label: string }[] = [
  { value: '', label: 'All Status' },
  { value: 'draft', label: 'Draft' },
  { value: 'pending_acknowledgment', label: 'Pending Acknowledgment' },
  { value: 'acknowledged', label: 'Acknowledged' },
  { value: 'active', label: 'Active' },
  { value: 'blocked', label: 'Blocked' },
  { value: 'pending_review', label: 'Pending Review' },
  { value: 'completed', label: 'Completed' },
]

export default function TaskList() {
  useDocumentTitle('Tasks')
  const [search, setSearch] = useState('')
  const debouncedSearch = useDebounce(search, 300)
  const [status, setStatus] = useState<TaskStatus | ''>('')
  const [page, setPage] = useState(1)
  const pageSize = 10

  const { data, isLoading } = useQuery({
    queryKey: ['tasks', { search: debouncedSearch, status, page, pageSize }],
    queryFn: () =>
      taskApi.list({
        search: debouncedSearch || undefined,
        status: status || undefined,
        page,
        page_size: pageSize,
      }),
  })

  const tasks = data?.tasks || []
  const totalPages = data?.total_pages || 1

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Tasks</h1>
          <p className="text-gray-500 mt-1">
            {data?.total || 0} total tasks
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

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4">
        <div className="flex flex-wrap gap-4">
          <div className="flex-1 min-w-[200px]">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-400" aria-hidden="true" />
              <input
                type="text"
                placeholder="Search tasks..."
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value)
                  setPage(1)
                }}
                aria-label="Search tasks"
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg
                         focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              />
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Filter className="h-5 w-5 text-gray-400" aria-hidden="true" />
            <select
              value={status}
              onChange={(e) => {
                setStatus(e.target.value as TaskStatus | '')
                setPage(1)
              }}
              aria-label="Filter by status"
              className="border border-gray-300 rounded-lg px-3 py-2
                       focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            >
              {statusOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Task List */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200">
        {isLoading ? (
          <div className="p-8 text-center text-gray-500" aria-live="polite">Loading...</div>
        ) : tasks.length > 0 ? (
          <div className="divide-y divide-gray-100">
            {tasks.map((task) => (
              <Link
                key={task.id}
                to={`/tasks/${task.id}`}
                className="flex items-center p-4 hover:bg-gray-50 transition-colors"
              >
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3">
                    <h3 className="text-sm font-medium text-gray-900 truncate">
                      {task.title}
                    </h3>
                    <span
                      className={clsx(
                        'px-2 py-0.5 text-xs font-medium rounded-full',
                        task.status === 'completed' && 'bg-green-100 text-green-700',
                        task.status === 'blocked' && 'bg-red-100 text-red-700',
                        task.status === 'active' && 'bg-yellow-100 text-yellow-700',
                        task.status === 'pending_acknowledgment' &&
                          'bg-blue-100 text-blue-700',
                        task.status === 'draft' && 'bg-gray-100 text-gray-700'
                      )}
                    >
                      {task.status.replace(/_/g, ' ')}
                    </span>
                  </div>
                  <p className="text-sm text-gray-500 mt-1 truncate">
                    {task.description.slice(0, 100)}
                    {task.description.length > 100 && '...'}
                  </p>
                  <div className="flex items-center gap-4 mt-2 text-xs text-gray-400">
                    <span>Created {new Date(task.created_at).toLocaleDateString()}</span>
                    {task.estimated_days && (
                      <span>Est: {task.estimated_days} days</span>
                    )}
                    {task.prediction_confidence && (
                      <span>
                        Predicted: {task.predicted_days_low}-{task.predicted_days_high} days
                      </span>
                    )}
                  </div>
                </div>
              </Link>
            ))}
          </div>
        ) : (
          <div className="p-8 text-center text-gray-500">
            <ListTodo className="h-12 w-12 mx-auto mb-2 text-gray-300" />
            <p>No tasks found</p>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between px-4 py-3 border-t border-gray-100">
            <div className="text-sm text-gray-500">
              Page {page} of {totalPages}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                aria-label="Previous page"
                className="p-2 rounded-lg border border-gray-300 hover:bg-gray-50
                         disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronLeft className="h-5 w-5" />
              </button>
              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                aria-label="Next page"
                className="p-2 rounded-lg border border-gray-300 hover:bg-gray-50
                         disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronRight className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
