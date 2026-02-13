import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { taskApi, linkApi, blockerApi, reviewApi } from '../services/api'
import {
  ArrowLeft,
  Clock,
  AlertTriangle,
  CheckCircle2,
  Link as LinkIcon,
  Play,
  Flag,
  Star,
  GitBranch,
} from 'lucide-react'
import { clsx } from 'clsx'
import { useState } from 'react'
import DependencyView from '../components/DependencyView'
import AcknowledgmentDialog from '../components/AcknowledgmentDialog'
import CompletionReviewForm from '../components/CompletionReviewForm'
import type { AcknowledgmentRequest, CreateReviewRequest } from '../types'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

export default function TaskDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const [showReviewForm, setShowReviewForm] = useState(false)
  const [showAcknowledgmentDialog, setShowAcknowledgmentDialog] = useState(false)

  const { data: task, isLoading } = useQuery({
    queryKey: ['task', id],
    queryFn: () => taskApi.get(id!),
    enabled: !!id,
  })

  useDocumentTitle(task?.title ? `${task.title}` : 'Task Detail')

  const { data: linksData } = useQuery({
    queryKey: ['task-links', id],
    queryFn: () => linkApi.list(id!),
    enabled: !!id,
  })

  const { data: blockersData } = useQuery({
    queryKey: ['task-blockers', id],
    queryFn: () => blockerApi.list(id!),
    enabled: !!id,
  })

  const { data: similarData } = useQuery({
    queryKey: ['task-similar', id],
    queryFn: () => taskApi.getSimilar(id!),
    enabled: !!id,
  })

  const acknowledgeMutation = useMutation({
    mutationFn: (data: AcknowledgmentRequest) => taskApi.acknowledge(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', id] })
      setShowAcknowledgmentDialog(false)
    },
  })

  const startMutation = useMutation({
    mutationFn: () => taskApi.start(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['task', id] }),
  })

  const completeMutation = useMutation({
    mutationFn: () => taskApi.complete(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['task', id] }),
  })

  const reviewMutation = useMutation({
    mutationFn: (data: CreateReviewRequest) => reviewApi.create(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['task', id] })
      setShowReviewForm(false)
    },
  })

  if (isLoading) {
    return (
      <div className="p-8 text-center" aria-live="polite">
        Loading...
        <span className="sr-only">Loading task details</span>
      </div>
    )
  }

  if (!task) {
    return (
      <div className="p-8 text-center" role="alert">
        Task not found
      </div>
    )
  }

  const links = linksData?.links || []
  const blockers = blockersData?.blockers || []
  const similar = similarData?.similar_tasks || []

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate(-1)}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          aria-label="Go back"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">{task.title}</h1>
          <div className="flex items-center gap-4 mt-2">
            <span
              className={clsx(
                'px-3 py-1 text-sm font-medium rounded-full',
                task.status === 'completed' && 'bg-green-100 text-green-700',
                task.status === 'blocked' && 'bg-red-100 text-red-700',
                task.status === 'active' && 'bg-yellow-100 text-yellow-700',
                task.status === 'pending_acknowledgment' && 'bg-blue-100 text-blue-700',
                task.status === 'draft' && 'bg-gray-100 text-gray-700',
                task.status === 'pending_review' && 'bg-purple-100 text-purple-700'
              )}
            >
              {task.status.replace(/_/g, ' ')}
            </span>
            <span className="text-sm text-gray-500">
              Created {new Date(task.created_at).toLocaleDateString()}
            </span>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          {task.status === 'pending_acknowledgment' && (
            <button
              onClick={() => setShowAcknowledgmentDialog(true)}
              className="bg-blue-600 text-white px-4 py-2 rounded-lg font-medium
                       hover:bg-blue-700 transition-colors flex items-center gap-2"
            >
              <CheckCircle2 className="h-5 w-5" />
              Acknowledge
            </button>
          )}
          {task.status === 'acknowledged' && (
            <button
              onClick={() => startMutation.mutate()}
              disabled={startMutation.isPending}
              className="bg-green-600 text-white px-4 py-2 rounded-lg font-medium
                       hover:bg-green-700 transition-colors flex items-center gap-2"
            >
              <Play className="h-5 w-5" />
              Start Task
            </button>
          )}
          {task.status === 'active' && (
            <button
              onClick={() => completeMutation.mutate()}
              disabled={completeMutation.isPending}
              className="bg-primary-600 text-white px-4 py-2 rounded-lg font-medium
                       hover:bg-primary-700 transition-colors flex items-center gap-2"
            >
              <Flag className="h-5 w-5" />
              Complete
            </button>
          )}
          {task.status === 'pending_review' && !showReviewForm && (
            <button
              onClick={() => setShowReviewForm(true)}
              className="bg-purple-600 text-white px-4 py-2 rounded-lg font-medium
                       hover:bg-purple-700 transition-colors flex items-center gap-2"
            >
              <Star className="h-5 w-5" />
              Submit Review
            </button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Description */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Description</h2>
            <p className="text-gray-700 whitespace-pre-wrap">{task.description}</p>
          </div>

          {/* Predictions */}
          {(task.predicted_days_low || task.planning_quality_score) && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Predictions</h2>
              <div className="grid grid-cols-2 gap-4">
                {task.predicted_days_low && (
                  <div className="p-4 bg-blue-50 rounded-lg">
                    <div className="flex items-center gap-2 text-blue-700 mb-1">
                      <Clock className="h-5 w-5" />
                      <span className="font-medium">Timeline</span>
                    </div>
                    <p className="text-2xl font-bold text-blue-900">
                      {task.predicted_days_low}-{task.predicted_days_high} days
                    </p>
                    <p className="text-sm text-blue-600 mt-1">
                      {((task.prediction_confidence || 0) * 100).toFixed(0)}% confidence
                    </p>
                  </div>
                )}
                {task.planning_quality_score && (
                  <div className="p-4 bg-green-50 rounded-lg">
                    <div className="flex items-center gap-2 text-green-700 mb-1">
                      <CheckCircle2 className="h-5 w-5" />
                      <span className="font-medium">Planning Quality</span>
                    </div>
                    <p className="text-2xl font-bold text-green-900">
                      {task.planning_quality_score}/100
                    </p>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Review Form */}
          {showReviewForm && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Completion Review
              </h2>
              <CompletionReviewForm
                task={task}
                onSubmit={(data) => reviewMutation.mutate(data)}
                onCancel={() => setShowReviewForm(false)}
                isPending={reviewMutation.isPending}
              />
            </div>
          )}

          {/* Blockers */}
          {blockers.length > 0 && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Blockers</h2>
              <div className="space-y-3">
                {blockers.map((blocker) => (
                  <div
                    key={blocker.id}
                    className={clsx(
                      'p-4 rounded-lg border',
                      blocker.resolved_at
                        ? 'bg-gray-50 border-gray-200'
                        : 'bg-red-50 border-red-200'
                    )}
                  >
                    <div className="flex items-center gap-2 mb-2">
                      <AlertTriangle
                        className={clsx(
                          'h-5 w-5',
                          blocker.resolved_at ? 'text-gray-400' : 'text-red-600'
                        )}
                      />
                      <span className="font-medium capitalize">
                        {blocker.blocker_type.replace(/_/g, ' ')}
                      </span>
                      {blocker.resolved_at && (
                        <span className="text-xs text-gray-500">
                          Resolved ({blocker.days_blocked} days)
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-gray-700">{blocker.description}</p>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Task Info */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h3 className="font-semibold text-gray-900 mb-4">Details</h3>
            <dl className="space-y-3 text-sm">
              {task.estimated_days && (
                <div>
                  <dt className="text-gray-500">Estimate</dt>
                  <dd className="font-medium">{task.estimated_days} days</dd>
                </div>
              )}
              {task.actual_days && (
                <div>
                  <dt className="text-gray-500">Actual</dt>
                  <dd className="font-medium">{task.actual_days.toFixed(1)} days</dd>
                </div>
              )}
              {task.started_at && (
                <div>
                  <dt className="text-gray-500">Started</dt>
                  <dd className="font-medium">
                    {new Date(task.started_at).toLocaleDateString()}
                  </dd>
                </div>
              )}
              {task.completed_at && (
                <div>
                  <dt className="text-gray-500">Completed</dt>
                  <dd className="font-medium">
                    {new Date(task.completed_at).toLocaleDateString()}
                  </dd>
                </div>
              )}
            </dl>
          </div>

          {/* Linked Tasks */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h3 className="font-semibold text-gray-900 mb-4 flex items-center gap-2">
              <LinkIcon className="h-5 w-5" />
              Linked Tasks ({links.length})
            </h3>
            {links.length > 0 ? (
              <div className="space-y-2">
                {links.map((link) => (
                  <Link
                    key={link.id}
                    to={`/tasks/${link.target_task_id}`}
                    className="block p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                  >
                    <span className="text-xs text-gray-500 uppercase">
                      {link.link_type}
                    </span>
                  </Link>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500">No linked tasks</p>
            )}
          </div>

          {/* Similar Tasks */}
          {similar.length > 0 && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h3 className="font-semibold text-gray-900 mb-4">Similar Tasks</h3>
              <div className="space-y-3">
                {similar.slice(0, 5).map((s) => (
                  <Link
                    key={s.id}
                    to={`/tasks/${s.id}`}
                    className="block p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                  >
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {s.title}
                    </p>
                    <div className="flex items-center gap-2 mt-1 text-xs text-gray-500">
                      <span>{(s.similarity_score * 100).toFixed(0)}% match</span>
                      {s.actual_days && <span>{s.actual_days} days</span>}
                    </div>
                  </Link>
                ))}
              </div>
            </div>
          )}

          {/* Context Chain View */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <h3 className="font-semibold text-gray-900 mb-4 flex items-center gap-2">
              <GitBranch className="h-5 w-5" />
              Context Chain
            </h3>
            <DependencyView taskId={id!} currentTask={task} />
          </div>
        </div>
      </div>

      {/* Acknowledgment Dialog */}
      <AcknowledgmentDialog
        task={task}
        isOpen={showAcknowledgmentDialog}
        onClose={() => setShowAcknowledgmentDialog(false)}
        onAcknowledge={(data) => acknowledgeMutation.mutate(data)}
        isPending={acknowledgeMutation.isPending}
      />
    </div>
  )
}
