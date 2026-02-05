import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { taskApi, linkApi, justificationApi, projectApi } from '../services/api'
import type { SimilarTask, LinkType, CreateTaskRequest, Predictions, Assessment, ProjectResponse } from '../types'
import {
  ArrowLeft,
  CheckCircle2,
  Link as LinkIcon,
} from 'lucide-react'
import { clsx } from 'clsx'
import PredictionDisplay from '../components/PredictionDisplay'

export default function CreateTask() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const initialProjectId = searchParams.get('project_id') || ''

  const [step, setStep] = useState<'form' | 'similar' | 'done'>('form')
  const [projects, setProjects] = useState<ProjectResponse[]>([])
  const [isLoadingProjects, setIsLoadingProjects] = useState(true)
  const [taskData, setTaskData] = useState<CreateTaskRequest>({
    title: '',
    description: '',
    project_id: initialProjectId,
    estimated_days: undefined,
    tags: [],
  })

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        const response = await projectApi.list({ status: 'active' })
        setProjects(response.projects || [])
        // Set first project if none selected and projects exist
        if (!initialProjectId && response.projects?.length > 0) {
          setTaskData(d => ({ ...d, project_id: response.projects[0].project.id }))
        }
      } catch (err) {
        console.error('Error fetching projects:', err)
      } finally {
        setIsLoadingProjects(false)
      }
    }
    fetchProjects()
  }, [initialProjectId])
  const [createdTask, setCreatedTask] = useState<{
    id: string
    similar: SimilarTask[]
    predictions?: Predictions
    assessment?: Assessment
  } | null>(null)
  const [selectedLinks, setSelectedLinks] = useState<{
    taskId: string
    type: LinkType
  }[]>([])
  const [showJustification, setShowJustification] = useState(false)
  const [justification, setJustification] = useState({
    checked_same_project: false,
    checked_same_stakeholders: false,
    checked_same_dependencies: false,
    justification_text: '',
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateTaskRequest) => taskApi.create(data),
    onSuccess: (response) => {
      setCreatedTask({
        id: response.task.id,
        similar: response.similar_tasks || [],
        predictions: response.predictions,
        assessment: response.planning_assessment,
      })
      if ((response.similar_tasks?.length || 0) > 0) {
        setStep('similar')
      } else {
        setShowJustification(true)
        setStep('similar')
      }
    },
  })

  const linkMutation = useMutation({
    mutationFn: async () => {
      if (!createdTask) return
      for (const link of selectedLinks) {
        await linkApi.create(createdTask.id, {
          target_task_id: link.taskId,
          link_type: link.type,
        })
      }
    },
  })

  const justificationMutation = useMutation({
    mutationFn: async () => {
      if (!createdTask) return
      await justificationApi.create(createdTask.id, justification)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createMutation.mutate(taskData)
  }

  const handleLinkToggle = (taskId: string) => {
    setSelectedLinks((prev) => {
      const exists = prev.find((l) => l.taskId === taskId)
      if (exists) {
        return prev.filter((l) => l.taskId !== taskId)
      }
      return [...prev, { taskId, type: 'similar' as LinkType }]
    })
  }

  const handleProceed = async () => {
    if (selectedLinks.length > 0) {
      await linkMutation.mutateAsync()
    } else if (
      justification.checked_same_project &&
      justification.checked_same_stakeholders &&
      justification.checked_same_dependencies &&
      justification.justification_text.length >= 50
    ) {
      await justificationMutation.mutateAsync()
    }
    navigate(`/tasks/${createdTask?.id}`)
  }

  const canProceed =
    selectedLinks.length > 0 ||
    (justification.checked_same_project &&
      justification.checked_same_stakeholders &&
      justification.checked_same_dependencies &&
      justification.justification_text.length >= 50)

  return (
    <div className="max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-4 mb-8">
        <button
          onClick={() => navigate(-1)}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h1 className="text-2xl font-bold text-gray-900">Create New Task</h1>
      </div>

      {step === 'form' && (
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="space-y-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Title *
                </label>
                <input
                  type="text"
                  value={taskData.title}
                  onChange={(e) =>
                    setTaskData((d) => ({ ...d, title: e.target.value }))
                  }
                  className="w-full border border-gray-300 rounded-lg p-3
                           focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="Brief summary of the task"
                  required
                  minLength={5}
                  maxLength={500}
                />
                <p className="text-xs text-gray-500 mt-1">
                  5-500 characters
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Description *
                </label>
                <textarea
                  value={taskData.description}
                  onChange={(e) =>
                    setTaskData((d) => ({ ...d, description: e.target.value }))
                  }
                  className="w-full border border-gray-300 rounded-lg p-3
                           focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  rows={6}
                  placeholder="Detailed description including objectives, dependencies, risks, and acceptance criteria..."
                  required
                  minLength={50}
                />
                <p className="text-xs text-gray-500 mt-1">
                  {taskData.description.length}/50 minimum characters
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Project *
                </label>
                {isLoadingProjects ? (
                  <div className="w-full border border-gray-300 rounded-lg p-3 bg-gray-50 text-gray-500">
                    Loading projects...
                  </div>
                ) : projects.length === 0 ? (
                  <div className="text-sm text-red-600">
                    No active projects found. Please <a href="/projects" className="underline">create a project</a> first.
                  </div>
                ) : (
                  <select
                    value={taskData.project_id}
                    onChange={(e) =>
                      setTaskData((d) => ({ ...d, project_id: e.target.value }))
                    }
                    className="w-full border border-gray-300 rounded-lg p-3
                             focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                    required
                  >
                    <option value="">Select a project...</option>
                    {projects.map((p) => (
                      <option key={p.project.id} value={p.project.id}>
                        {p.project.name}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Estimated Days
                </label>
                <input
                  type="number"
                  value={taskData.estimated_days || ''}
                  onChange={(e) =>
                    setTaskData((d) => ({
                      ...d,
                      estimated_days: e.target.value
                        ? parseFloat(e.target.value)
                        : undefined,
                    }))
                  }
                  className="w-48 border border-gray-300 rounded-lg p-3
                           focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="e.g., 5"
                  min={0.5}
                  step={0.5}
                />
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-4">
            <button
              type="button"
              onClick={() => navigate(-1)}
              className="px-6 py-2 border border-gray-300 rounded-lg font-medium
                       hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createMutation.isPending || taskData.description.length < 50 || !taskData.project_id}
              className="px-6 py-2 bg-primary-600 text-white rounded-lg font-medium
                       hover:bg-primary-700 transition-colors disabled:opacity-50"
            >
              {createMutation.isPending ? 'Creating...' : 'Create Task'}
            </button>
          </div>
        </form>
      )}

      {step === 'similar' && createdTask && (
        <div className="space-y-6">
          {/* Predictions Display */}
          <PredictionDisplay
            predictions={createdTask.predictions}
            assessment={createdTask.assessment}
            userEstimate={taskData.estimated_days}
          />

          {/* Similar Tasks */}
          {createdTask.similar.length > 0 && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
              <h2 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                <LinkIcon className="h-5 w-5" />
                Similar Historical Tasks
              </h2>
              <p className="text-sm text-gray-500 mb-4">
                Select tasks to link (learn from their outcomes)
              </p>
              <div className="space-y-3">
                {createdTask.similar.map((task) => (
                  <button
                    key={task.id}
                    onClick={() => handleLinkToggle(task.id)}
                    className={clsx(
                      'w-full text-left p-4 rounded-lg border-2 transition-colors',
                      selectedLinks.find((l) => l.taskId === task.id)
                        ? 'border-primary-500 bg-primary-50'
                        : 'border-gray-200 hover:border-gray-300'
                    )}
                  >
                    <div className="flex items-center justify-between mb-2">
                      <span className="font-medium text-gray-900">{task.title}</span>
                      <span className="text-sm text-gray-500">
                        {(task.similarity_score * 100).toFixed(0)}% match
                      </span>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-gray-600">
                      <span
                        className={clsx(
                          'px-2 py-0.5 rounded-full text-xs font-medium',
                          task.status === 'completed' && 'bg-green-100 text-green-700',
                          task.status === 'blocked' && 'bg-red-100 text-red-700'
                        )}
                      >
                        {task.status}
                      </span>
                      {task.actual_days && <span>Took {task.actual_days} days</span>}
                      {task.estimated_days && (
                        <span>Estimated {task.estimated_days} days</span>
                      )}
                    </div>
                    {task.lessons_learned_excerpt && (
                      <p className="text-sm text-gray-500 mt-2 italic">
                        "{task.lessons_learned_excerpt}"
                      </p>
                    )}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Justification */}
          {(createdTask.similar.length === 0 || showJustification) && (
            <div className="bg-yellow-50 rounded-xl p-6 border border-yellow-200">
              <h2 className="text-lg font-semibold text-yellow-900 mb-4">
                {createdTask.similar.length === 0
                  ? 'No similar tasks found'
                  : 'Proceeding without links'}
              </h2>
              <p className="text-sm text-yellow-700 mb-4">
                Please confirm you've checked for similar tasks:
              </p>
              <div className="space-y-3 mb-4">
                {[
                  { key: 'checked_same_project', label: 'Checked tasks in the same project' },
                  { key: 'checked_same_stakeholders', label: 'Checked tasks with same stakeholders' },
                  { key: 'checked_same_dependencies', label: 'Checked tasks with similar dependencies' },
                ].map((item) => (
                  <label key={item.key} className="flex items-center gap-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={justification[item.key as keyof typeof justification] as boolean}
                      onChange={(e) =>
                        setJustification((j) => ({
                          ...j,
                          [item.key]: e.target.checked,
                        }))
                      }
                      className="h-5 w-5 rounded border-yellow-300 text-yellow-600
                               focus:ring-yellow-500"
                    />
                    <span className="text-sm text-yellow-800">{item.label}</span>
                  </label>
                ))}
              </div>
              <div>
                <label className="block text-sm font-medium text-yellow-800 mb-2">
                  Why is this genuinely new work? (min 50 characters)
                </label>
                <textarea
                  value={justification.justification_text}
                  onChange={(e) =>
                    setJustification((j) => ({ ...j, justification_text: e.target.value }))
                  }
                  className="w-full border border-yellow-300 rounded-lg p-3
                           focus:ring-2 focus:ring-yellow-500 focus:border-yellow-500
                           bg-white"
                  rows={3}
                  placeholder="Explain why no historical context applies..."
                />
                <p className="text-xs text-yellow-600 mt-1">
                  {justification.justification_text.length}/50 characters
                </p>
              </div>
            </div>
          )}

          {selectedLinks.length === 0 && createdTask.similar.length > 0 && !showJustification && (
            <button
              onClick={() => setShowJustification(true)}
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              No links apply? Provide justification instead
            </button>
          )}

          <div className="flex justify-end gap-4">
            <button
              onClick={handleProceed}
              disabled={!canProceed || linkMutation.isPending || justificationMutation.isPending}
              className="px-6 py-2 bg-primary-600 text-white rounded-lg font-medium
                       hover:bg-primary-700 transition-colors disabled:opacity-50
                       flex items-center gap-2"
            >
              <CheckCircle2 className="h-5 w-5" />
              {selectedLinks.length > 0
                ? `Link ${selectedLinks.length} Task(s) & Continue`
                : 'Submit Justification & Continue'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
