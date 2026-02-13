import { Link } from 'react-router-dom'
import { useQuery, useQueries } from '@tanstack/react-query'
import { linkApi, taskApi } from '../services/api'
import type { TaskLink, Task, TaskStatus } from '../types'
import {
  GitBranch,
  ArrowRight,
  Link as LinkIcon,
  RefreshCw,
  AlertTriangle,
  BookOpen,
  ChevronDown,
  ChevronUp,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useState } from 'react'

interface DependencyViewProps {
  taskId: string
  currentTask?: Task
}

const linkTypeConfig: Record<string, { icon: typeof LinkIcon; color: string; label: string }> = {
  similar: { icon: GitBranch, color: 'text-blue-500', label: 'Similar' },
  dependent: { icon: ArrowRight, color: 'text-purple-500', label: 'Depends On' },
  retry: { icon: RefreshCw, color: 'text-orange-500', label: 'Retry Of' },
  related: { icon: LinkIcon, color: 'text-gray-500', label: 'Related' },
}

const statusColors: Record<TaskStatus, string> = {
  draft: 'bg-gray-100 text-gray-700 border-gray-200',
  pending_acknowledgment: 'bg-blue-100 text-blue-700 border-blue-200',
  acknowledged: 'bg-indigo-100 text-indigo-700 border-indigo-200',
  active: 'bg-yellow-100 text-yellow-700 border-yellow-200',
  blocked: 'bg-red-100 text-red-700 border-red-200',
  pending_review: 'bg-purple-100 text-purple-700 border-purple-200',
  completed: 'bg-green-100 text-green-700 border-green-200',
  cancelled: 'bg-gray-100 text-gray-400 border-gray-200',
}

interface LinkedTaskWithDetails extends TaskLink {
  targetTask?: Task
}

export default function DependencyView({ taskId, currentTask }: DependencyViewProps) {
  const [expandedTasks, setExpandedTasks] = useState<Set<string>>(new Set())

  const { data: linksData, isLoading } = useQuery({
    queryKey: ['task-links', taskId],
    queryFn: () => linkApi.list(taskId),
    enabled: !!taskId,
  })

  const links = linksData?.links || []

  // Fetch details for each linked task
  const linkedTaskQueries = useQueries({
    queries: links.map((link) => ({
      queryKey: ['task', link.target_task_id],
      queryFn: () => taskApi.get(link.target_task_id),
      enabled: !!link.target_task_id,
    })),
  })

  const linkedTasks: LinkedTaskWithDetails[] = links.map((link, index) => ({
    ...link,
    targetTask: linkedTaskQueries[index]?.data,
  }))

  const toggleExpand = (taskId: string) => {
    setExpandedTasks((prev) => {
      const next = new Set(prev)
      if (next.has(taskId)) {
        next.delete(taskId)
      } else {
        next.add(taskId)
      }
      return next
    })
  }

  if (isLoading) {
    return (
      <div className="p-4 text-center text-gray-500">
        Loading task chain...
      </div>
    )
  }

  if (links.length === 0) {
    return (
      <div className="p-6 text-center">
        <GitBranch className="h-12 w-12 mx-auto mb-3 text-gray-300" />
        <p className="text-gray-500">No linked tasks in this chain</p>
        <p className="text-sm text-gray-400 mt-1">
          Link similar or related tasks to build context
        </p>
      </div>
    )
  }

  // Group links by type
  const groupedLinks = linkedTasks.reduce((acc, link) => {
    const type = link.link_type
    if (!acc[type]) acc[type] = []
    acc[type].push(link)
    return acc
  }, {} as Record<string, LinkedTaskWithDetails[]>)

  return (
    <div className="space-y-6">
      {/* Current Task Node */}
      {currentTask && (
        <div className="relative">
          <div className="bg-primary-50 border-2 border-primary-500 rounded-lg p-4">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs font-medium text-primary-600 uppercase">
                Current Task
              </span>
            </div>
            <h3 className="font-semibold text-gray-900">{currentTask.title}</h3>
            <div className="flex items-center gap-2 mt-2">
              <span
                className={clsx(
                  'px-2 py-0.5 text-xs font-medium rounded-full border',
                  statusColors[currentTask.status]
                )}
              >
                {currentTask.status.replace(/_/g, ' ')}
              </span>
              {currentTask.actual_days && (
                <span className="text-xs text-gray-500">
                  {currentTask.actual_days.toFixed(1)} days
                </span>
              )}
            </div>
          </div>
          {/* Connector line */}
          {links.length > 0 && (
            <div className="absolute left-1/2 -bottom-4 w-0.5 h-4 bg-gray-300" />
          )}
        </div>
      )}

      {/* Linked Tasks by Type */}
      {Object.entries(groupedLinks).map(([type, typeLinks]) => {
        const config = linkTypeConfig[type] || linkTypeConfig.related
        const Icon = config.icon

        return (
          <div key={type} className="relative">
            {/* Type Header */}
            <div className="flex items-center gap-2 mb-3">
              <Icon className={clsx('h-5 w-5', config.color)} />
              <span className="text-sm font-medium text-gray-700">
                {config.label} ({typeLinks.length})
              </span>
            </div>

            {/* Task Cards */}
            <div className="space-y-3 ml-7">
              {typeLinks.map((link) => {
                const task = link.targetTask
                const isExpanded = expandedTasks.has(link.target_task_id)
                const hasBlockers = task?.status === 'blocked'
                const hasLessons = false // Would need to fetch review data

                return (
                  <div
                    key={link.id}
                    className={clsx(
                      'border rounded-lg overflow-hidden transition-all',
                      hasBlockers ? 'border-red-200' : 'border-gray-200'
                    )}
                  >
                    {/* Task Header */}
                    <div
                      className={clsx(
                        'p-4 cursor-pointer hover:bg-gray-50 transition-colors',
                        hasBlockers && 'bg-red-50'
                      )}
                      onClick={() => toggleExpand(link.target_task_id)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1 min-w-0">
                          <Link
                            to={`/tasks/${link.target_task_id}`}
                            className="text-sm font-medium text-gray-900 hover:text-primary-600 truncate block"
                            onClick={(e) => e.stopPropagation()}
                          >
                            {task?.title || 'Loading...'}
                          </Link>
                          {task && (
                            <div className="flex items-center gap-3 mt-2">
                              <span
                                className={clsx(
                                  'px-2 py-0.5 text-xs font-medium rounded-full border',
                                  statusColors[task.status]
                                )}
                              >
                                {task.status.replace(/_/g, ' ')}
                              </span>
                              {task.actual_days && (
                                <span className="text-xs text-gray-500">
                                  Took {task.actual_days.toFixed(1)} days
                                </span>
                              )}
                              {task.estimated_days && (
                                <span className="text-xs text-gray-500">
                                  Est: {task.estimated_days} days
                                </span>
                              )}
                            </div>
                          )}
                        </div>
                        <div className="flex items-center gap-2 ml-4">
                          {hasBlockers && (
                            <AlertTriangle className="h-4 w-4 text-red-500" />
                          )}
                          {hasLessons && (
                            <BookOpen className="h-4 w-4 text-blue-500" />
                          )}
                          {isExpanded ? (
                            <ChevronUp className="h-4 w-4 text-gray-400" />
                          ) : (
                            <ChevronDown className="h-4 w-4 text-gray-400" />
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Expanded Details */}
                    {isExpanded && task && (
                      <div className="px-4 pb-4 border-t border-gray-100">
                        <div className="pt-3 space-y-3">
                          {/* Description Preview */}
                          <div>
                            <p className="text-xs font-medium text-gray-500 mb-1">
                              Description
                            </p>
                            <p className="text-sm text-gray-700">
                              {task.description.slice(0, 200)}
                              {task.description.length > 200 && '...'}
                            </p>
                          </div>

                          {/* Timing Comparison */}
                          {(task.estimated_days || task.actual_days) && (
                            <div className="flex gap-4">
                              {task.estimated_days && (
                                <div>
                                  <p className="text-xs font-medium text-gray-500">
                                    Estimated
                                  </p>
                                  <p className="text-sm font-medium text-gray-900">
                                    {task.estimated_days} days
                                  </p>
                                </div>
                              )}
                              {task.actual_days && (
                                <div>
                                  <p className="text-xs font-medium text-gray-500">
                                    Actual
                                  </p>
                                  <p
                                    className={clsx(
                                      'text-sm font-medium',
                                      task.estimated_days &&
                                        task.actual_days > task.estimated_days * 1.2
                                        ? 'text-red-600'
                                        : 'text-green-600'
                                    )}
                                  >
                                    {task.actual_days.toFixed(1)} days
                                    {task.estimated_days && (
                                      <span className="text-xs text-gray-500 ml-1">
                                        ({task.actual_days > task.estimated_days ? '+' : ''}
                                        {((task.actual_days / task.estimated_days - 1) * 100).toFixed(0)}%)
                                      </span>
                                    )}
                                  </p>
                                </div>
                              )}
                            </div>
                          )}

                          {/* Link Notes */}
                          {link.notes && (
                            <div>
                              <p className="text-xs font-medium text-gray-500 mb-1">
                                Link Notes
                              </p>
                              <p className="text-sm text-gray-700 italic">
                                "{link.notes}"
                              </p>
                            </div>
                          )}

                          {/* View Full Task Link */}
                          <Link
                            to={`/tasks/${link.target_task_id}`}
                            className="inline-flex items-center gap-1 text-sm text-primary-600 hover:text-primary-700"
                          >
                            View full task
                            <ArrowRight className="h-4 w-4" />
                          </Link>
                        </div>
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        )
      })}

      {/* Chain Summary */}
      <div className="bg-gray-50 rounded-lg p-4 text-sm">
        <h4 className="font-medium text-gray-700 mb-2">Chain Summary</h4>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-gray-500">Total Linked</p>
            <p className="text-lg font-semibold text-gray-900">{links.length}</p>
          </div>
          <div>
            <p className="text-gray-500">Completed</p>
            <p className="text-lg font-semibold text-green-600">
              {linkedTasks.filter((l) => l.targetTask?.status === 'completed').length}
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
