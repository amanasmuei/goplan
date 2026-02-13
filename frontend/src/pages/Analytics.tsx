import { useQuery } from '@tanstack/react-query'
import { taskApi } from '../services/api'
import {
  BarChart3,
  TrendingUp,
  TrendingDown,
  Clock,
  AlertTriangle,
  CheckCircle2,
  Target,
} from 'lucide-react'
import { clsx } from 'clsx'
import type { Task } from '../types'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

interface AnalyticsData {
  totalTasks: number
  completedTasks: number
  avgAccuracy: number
  avgCycleTime: number
  blockersByType: Record<string, number>
  predictionAccuracy: {
    underEstimated: number
    accurate: number
    overEstimated: number
  }
  tasksByStatus: Record<string, number>
  recentCompletions: Task[]
}

function calculateAnalytics(tasks: Task[]): AnalyticsData {
  const completedTasks = tasks.filter((t) => t.status === 'completed')

  // Calculate prediction accuracy
  let totalVariance = 0
  let varianceCount = 0
  let underEstimated = 0
  let accurate = 0
  let overEstimated = 0

  completedTasks.forEach((task) => {
    if (task.actual_days && task.predicted_days_low && task.predicted_days_high) {
      const midPrediction = (task.predicted_days_low + task.predicted_days_high) / 2
      const variance = Math.abs(task.actual_days - midPrediction) / midPrediction
      totalVariance += variance
      varianceCount++

      if (task.actual_days < task.predicted_days_low) {
        overEstimated++
      } else if (task.actual_days > task.predicted_days_high) {
        underEstimated++
      } else {
        accurate++
      }
    }
  })

  // Calculate average cycle time
  let totalCycleTime = 0
  let cycleCount = 0
  completedTasks.forEach((task) => {
    if (task.actual_days) {
      totalCycleTime += task.actual_days
      cycleCount++
    }
  })

  // Count tasks by status
  const tasksByStatus: Record<string, number> = {}
  tasks.forEach((task) => {
    tasksByStatus[task.status] = (tasksByStatus[task.status] || 0) + 1
  })

  return {
    totalTasks: tasks.length,
    completedTasks: completedTasks.length,
    avgAccuracy: varianceCount > 0 ? (1 - totalVariance / varianceCount) * 100 : 0,
    avgCycleTime: cycleCount > 0 ? totalCycleTime / cycleCount : 0,
    blockersByType: {}, // Would need blocker data
    predictionAccuracy: { underEstimated, accurate, overEstimated },
    tasksByStatus,
    recentCompletions: completedTasks.slice(0, 5),
  }
}

function StatCard({
  title,
  value,
  subtitle,
  icon: Icon,
  trend,
  color = 'blue',
}: {
  title: string
  value: string | number
  subtitle?: string
  icon: React.ComponentType<{ className?: string }>
  trend?: 'up' | 'down'
  color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple'
}) {
  const colorClasses = {
    blue: 'bg-blue-50 text-blue-600',
    green: 'bg-green-50 text-green-600',
    yellow: 'bg-yellow-50 text-yellow-600',
    red: 'bg-red-50 text-red-600',
    purple: 'bg-purple-50 text-purple-600',
  }

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500">{title}</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{value}</p>
          {subtitle && <p className="text-sm text-gray-500 mt-1">{subtitle}</p>}
        </div>
        <div className={clsx('p-3 rounded-lg', colorClasses[color])}>
          <Icon className="h-6 w-6" />
        </div>
      </div>
      {trend && (
        <div className="mt-4 flex items-center gap-1 text-sm">
          {trend === 'up' ? (
            <>
              <TrendingUp className="h-4 w-4 text-green-500" />
              <span className="text-green-600">Improving</span>
            </>
          ) : (
            <>
              <TrendingDown className="h-4 w-4 text-red-500" />
              <span className="text-red-600">Needs attention</span>
            </>
          )}
        </div>
      )}
    </div>
  )
}

function AccuracyChart({ data }: { data: AnalyticsData['predictionAccuracy'] }) {
  const total = data.underEstimated + data.accurate + data.overEstimated
  if (total === 0) return <p className="text-gray-500 text-center py-4">No data available</p>

  const segments = [
    { label: 'Under-estimated', value: data.underEstimated, color: 'bg-red-500' },
    { label: 'Accurate', value: data.accurate, color: 'bg-green-500' },
    { label: 'Over-estimated', value: data.overEstimated, color: 'bg-blue-500' },
  ]

  return (
    <div className="space-y-4">
      {/* Stacked bar */}
      <div className="h-8 flex rounded-lg overflow-hidden" role="img" aria-label="Prediction accuracy breakdown chart">
        {segments.map((seg) => (
          <div
            key={seg.label}
            className={clsx(seg.color, 'transition-all')}
            style={{ width: `${(seg.value / total) * 100}%` }}
            title={`${seg.label}: ${seg.value}`}
          />
        ))}
      </div>

      {/* Legend */}
      <div className="flex flex-wrap gap-4 justify-center">
        {segments.map((seg) => (
          <div key={seg.label} className="flex items-center gap-2">
            <div className={clsx('w-3 h-3 rounded', seg.color)} />
            <span className="text-sm text-gray-600">
              {seg.label}: {seg.value} ({((seg.value / total) * 100).toFixed(0)}%)
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

function StatusChart({ data }: { data: Record<string, number> }) {
  const statusColors: Record<string, string> = {
    draft: 'bg-gray-400',
    pending_acknowledgment: 'bg-blue-400',
    acknowledged: 'bg-indigo-400',
    active: 'bg-yellow-400',
    blocked: 'bg-red-400',
    pending_review: 'bg-purple-400',
    completed: 'bg-green-400',
    cancelled: 'bg-gray-300',
  }

  const total = Object.values(data).reduce((a, b) => a + b, 0)
  if (total === 0) return <p className="text-gray-500 text-center py-4">No tasks yet</p>

  return (
    <div className="space-y-3">
      {Object.entries(data).map(([status, count]) => (
        <div key={status} className="flex items-center gap-3">
          <div className="w-32 text-sm text-gray-600 capitalize">
            {status.replace(/_/g, ' ')}
          </div>
          <div className="flex-1 bg-gray-100 rounded-full h-4 overflow-hidden">
            <div
              className={clsx(statusColors[status] || 'bg-gray-400', 'h-full rounded-full')}
              style={{ width: `${(count / total) * 100}%` }}
            />
          </div>
          <div className="w-12 text-right text-sm font-medium text-gray-700">{count}</div>
        </div>
      ))}
    </div>
  )
}

export default function Analytics() {
  useDocumentTitle('Analytics')
  const { data: tasksData, isLoading } = useQuery({
    queryKey: ['tasks', { page: 1, page_size: 250 }],
    queryFn: () => taskApi.list({ page: 1, page_size: 250 }),
  })

  if (isLoading) {
    return (
      <div className="p-8 text-center" aria-live="polite">
        <div className="animate-spin h-8 w-8 border-4 border-primary-500 border-t-transparent rounded-full mx-auto" />
        <p className="mt-4 text-gray-500">Loading analytics...</p>
      </div>
    )
  }

  const tasks = tasksData?.tasks || []
  const analytics = calculateAnalytics(tasks)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Analytics Dashboard</h1>
        <p className="text-gray-500 mt-1">
          Insights from {analytics.totalTasks} tasks in your organization
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Tasks"
          value={analytics.totalTasks}
          icon={BarChart3}
          color="blue"
        />
        <StatCard
          title="Completed"
          value={analytics.completedTasks}
          subtitle={`${((analytics.completedTasks / analytics.totalTasks) * 100 || 0).toFixed(0)}% completion rate`}
          icon={CheckCircle2}
          color="green"
        />
        <StatCard
          title="Prediction Accuracy"
          value={`${analytics.avgAccuracy.toFixed(0)}%`}
          subtitle="Average across completed tasks"
          icon={Target}
          color="purple"
          trend={analytics.avgAccuracy >= 70 ? 'up' : 'down'}
        />
        <StatCard
          title="Avg Cycle Time"
          value={`${analytics.avgCycleTime.toFixed(1)} days`}
          subtitle="From start to completion"
          icon={Clock}
          color="yellow"
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Prediction Accuracy Breakdown */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Target className="h-5 w-5 text-purple-500" />
            Prediction Accuracy Breakdown
          </h2>
          <p className="text-sm text-gray-500 mb-4">
            How well do system predictions match actual completion times?
          </p>
          <AccuracyChart data={analytics.predictionAccuracy} />
        </div>

        {/* Tasks by Status */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <BarChart3 className="h-5 w-5 text-blue-500" />
            Tasks by Status
          </h2>
          <p className="text-sm text-gray-500 mb-4">
            Current distribution of tasks across workflow stages
          </p>
          <StatusChart data={analytics.tasksByStatus} />
        </div>
      </div>

      {/* Recent Completions */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <CheckCircle2 className="h-5 w-5 text-green-500" />
          Recent Completions
        </h2>
        {analytics.recentCompletions.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="text-left py-3 px-4 text-sm font-medium text-gray-500">Task</th>
                  <th className="text-right py-3 px-4 text-sm font-medium text-gray-500">Estimated</th>
                  <th className="text-right py-3 px-4 text-sm font-medium text-gray-500">Predicted</th>
                  <th className="text-right py-3 px-4 text-sm font-medium text-gray-500">Actual</th>
                  <th className="text-right py-3 px-4 text-sm font-medium text-gray-500">Variance</th>
                </tr>
              </thead>
              <tbody>
                {analytics.recentCompletions.map((task) => {
                  const predicted = task.predicted_days_low && task.predicted_days_high
                    ? (task.predicted_days_low + task.predicted_days_high) / 2
                    : null
                  const variance = predicted && task.actual_days
                    ? ((task.actual_days - predicted) / predicted) * 100
                    : null

                  return (
                    <tr key={task.id} className="border-b border-gray-100 hover:bg-gray-50">
                      <td className="py-3 px-4">
                        <a
                          href={`/tasks/${task.id}`}
                          className="text-primary-600 hover:underline font-medium"
                        >
                          {task.title}
                        </a>
                      </td>
                      <td className="py-3 px-4 text-right text-sm text-gray-600">
                        {task.estimated_days ? `${task.estimated_days}d` : '-'}
                      </td>
                      <td className="py-3 px-4 text-right text-sm text-gray-600">
                        {task.predicted_days_low && task.predicted_days_high
                          ? `${task.predicted_days_low.toFixed(1)}-${task.predicted_days_high.toFixed(1)}d`
                          : '-'}
                      </td>
                      <td className="py-3 px-4 text-right text-sm font-medium">
                        {task.actual_days ? `${task.actual_days.toFixed(1)}d` : '-'}
                      </td>
                      <td className="py-3 px-4 text-right">
                        {variance !== null ? (
                          <span
                            className={clsx(
                              'text-sm font-medium',
                              Math.abs(variance) <= 10
                                ? 'text-green-600'
                                : variance > 0
                                ? 'text-red-600'
                                : 'text-blue-600'
                            )}
                          >
                            {variance > 0 ? '+' : ''}{variance.toFixed(0)}%
                          </span>
                        ) : (
                          '-'
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-gray-500 text-center py-8">
            No completed tasks yet. Complete some tasks to see analytics.
          </p>
        )}
      </div>

      {/* Insights */}
      <div className="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-xl p-6 text-white">
        <h2 className="text-lg font-semibold mb-2 flex items-center gap-2">
          <AlertTriangle className="h-5 w-5" />
          Key Insights
        </h2>
        <ul className="space-y-2 text-sm text-white/90">
          {analytics.avgAccuracy < 70 && (
            <li>
              Prediction accuracy is below 70%. Consider reviewing task descriptions for more detail.
            </li>
          )}
          {analytics.predictionAccuracy.underEstimated > analytics.predictionAccuracy.overEstimated && (
            <li>
              Tasks are frequently taking longer than predicted. Common blockers may need addressing.
            </li>
          )}
          {analytics.tasksByStatus['blocked'] > 0 && (
            <li>
              {analytics.tasksByStatus['blocked']} task(s) are currently blocked. Review and resolve blockers.
            </li>
          )}
          {analytics.completedTasks === 0 && (
            <li>
              No tasks completed yet. As you complete tasks, insights will improve.
            </li>
          )}
          {analytics.avgAccuracy >= 80 && (
            <li>
              Excellent prediction accuracy! The system is learning well from your task history.
            </li>
          )}
        </ul>
      </div>
    </div>
  )
}
