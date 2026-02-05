import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { projectApi, teamApi, taskApi } from '../services/api'
import type { ProjectResponse, TeamResponse, Task } from '../types'
import {
  FolderKanban,
  ArrowLeft,
  Plus,
  Trash2,
  AlertCircle,
  Users,
  ListTodo,
  Archive,
  Calendar
} from 'lucide-react'

export default function ProjectDetail() {
  const { id } = useParams<{ id: string }>()
  const [project, setProject] = useState<ProjectResponse | null>(null)
  const [tasks, setTasks] = useState<Task[]>([])
  const [availableTeams, setAvailableTeams] = useState<TeamResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showAssignTeamModal, setShowAssignTeamModal] = useState(false)
  const [selectedTeamId, setSelectedTeamId] = useState('')
  const [isAssigning, setIsAssigning] = useState(false)

  useEffect(() => {
    if (id) {
      fetchProject()
      fetchTasks()
      fetchAvailableTeams()
    }
  }, [id])

  const fetchProject = async () => {
    if (!id) return
    try {
      setIsLoading(true)
      const response = await projectApi.get(id)
      setProject(response)
      setError(null)
    } catch (err) {
      setError('Failed to load project')
      console.error('Error fetching project:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const fetchTasks = async () => {
    if (!id) return
    try {
      const response = await taskApi.list({ project_id: id, page_size: 10 })
      setTasks(response.tasks || [])
    } catch (err) {
      console.error('Error fetching tasks:', err)
    }
  }

  const fetchAvailableTeams = async () => {
    try {
      const response = await teamApi.list()
      setAvailableTeams(response.teams || [])
    } catch (err) {
      console.error('Error fetching teams:', err)
    }
  }

  const handleAssignTeam = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id || !selectedTeamId) return

    try {
      setIsAssigning(true)
      await projectApi.assignTeams(id, { team_ids: [selectedTeamId] })
      setSelectedTeamId('')
      setShowAssignTeamModal(false)
      fetchProject()
    } catch (err) {
      setError('Failed to assign team')
      console.error('Error assigning team:', err)
    } finally {
      setIsAssigning(false)
    }
  }

  const handleRemoveTeam = async (teamId: string, teamName: string) => {
    if (!id) return
    if (!confirm(`Are you sure you want to remove ${teamName} from this project?`)) return

    try {
      await projectApi.removeTeam(id, teamId)
      fetchProject()
    } catch (err) {
      setError('Failed to remove team')
      console.error('Error removing team:', err)
    }
  }

  const handleArchive = async () => {
    if (!id) return
    if (!confirm('Are you sure you want to archive this project?')) return

    try {
      await projectApi.archive(id)
      fetchProject()
    } catch (err) {
      setError('Failed to archive project')
      console.error('Error archiving project:', err)
    }
  }

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      draft: 'bg-gray-100 text-gray-700',
      pending_acknowledgment: 'bg-yellow-100 text-yellow-700',
      acknowledged: 'bg-blue-100 text-blue-700',
      active: 'bg-green-100 text-green-700',
      blocked: 'bg-red-100 text-red-700',
      pending_review: 'bg-purple-100 text-purple-700',
      completed: 'bg-emerald-100 text-emerald-700',
      cancelled: 'bg-gray-100 text-gray-500',
    }
    return colors[status] || 'bg-gray-100 text-gray-700'
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (!project) {
    return (
      <div className="text-center py-12">
        <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">Project not found</h3>
        <Link to="/projects" className="text-primary-600 hover:text-primary-700">
          &larr; Back to Projects
        </Link>
      </div>
    )
  }

  const assignedTeamIds = new Set(project.teams?.map((t) => t.id) || [])
  const unassignedTeams = availableTeams.filter((t) => !assignedTeamIds.has(t.team.id))

  return (
    <div>
      {/* Header */}
      <div className="mb-6">
        <Link
          to="/projects"
          className="inline-flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Projects
        </Link>
        <div className="flex justify-between items-start">
          <div className="flex items-center gap-4">
            <div className={`h-14 w-14 rounded-xl flex items-center justify-center ${
              project.project.status === 'archived' ? 'bg-gray-200' : 'bg-primary-100'
            }`}>
              <FolderKanban className={`h-7 w-7 ${
                project.project.status === 'archived' ? 'text-gray-500' : 'text-primary-600'
              }`} />
            </div>
            <div>
              <div className="flex items-center gap-3">
                <h1 className="text-2xl font-bold text-gray-900">{project.project.name}</h1>
                {project.project.status === 'archived' && (
                  <span className="inline-flex items-center gap-1 px-2 py-1 bg-gray-200 text-gray-600 text-sm rounded-full">
                    <Archive className="h-3 w-3" />
                    Archived
                  </span>
                )}
              </div>
              {project.project.description && (
                <p className="text-gray-600 mt-1">{project.project.description}</p>
              )}
            </div>
          </div>
          <div className="flex gap-2">
            {project.project.status === 'active' && (
              <>
                <button
                  onClick={() => setShowAssignTeamModal(true)}
                  className="flex items-center gap-2 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <Users className="h-5 w-5" />
                  Assign Team
                </button>
                <button
                  onClick={handleArchive}
                  className="flex items-center gap-2 px-4 py-2 border border-yellow-300 text-yellow-700 rounded-lg hover:bg-yellow-50 transition-colors"
                >
                  <Archive className="h-5 w-5" />
                  Archive
                </button>
              </>
            )}
            <Link
              to={`/tasks/new?project_id=${project.project.id}`}
              className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
            >
              <Plus className="h-5 w-5" />
              New Task
            </Link>
          </div>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
          <AlertCircle className="h-5 w-5" />
          {error}
          <button onClick={() => setError(null)} className="ml-auto text-red-500 hover:text-red-700">
            &times;
          </button>
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Tasks */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-semibold text-gray-900">
                Tasks ({project.task_count})
              </h2>
              <Link
                to={`/tasks?project_id=${project.project.id}`}
                className="text-sm text-primary-600 hover:text-primary-700"
              >
                View All &rarr;
              </Link>
            </div>
            {tasks.length > 0 ? (
              <div className="space-y-3">
                {tasks.map((task) => (
                  <Link
                    key={task.id}
                    to={`/tasks/${task.id}`}
                    className="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <ListTodo className="h-5 w-5 text-gray-400" />
                      <div>
                        <p className="font-medium text-gray-900">{task.title}</p>
                        <p className="text-sm text-gray-500 line-clamp-1">{task.description}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      {task.estimated_days && (
                        <span className="text-sm text-gray-500">
                          {task.estimated_days}d est.
                        </span>
                      )}
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(task.status)}`}>
                        {task.status.replace(/_/g, ' ')}
                      </span>
                    </div>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <ListTodo className="h-10 w-10 text-gray-300 mx-auto mb-3" />
                <p className="text-gray-500">No tasks yet</p>
                <Link
                  to={`/tasks/new?project_id=${project.project.id}`}
                  className="inline-flex items-center gap-2 mt-3 text-primary-600 hover:text-primary-700"
                >
                  <Plus className="h-4 w-4" />
                  Create first task
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Teams */}
          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Teams ({project.teams?.length || 0})
            </h2>
            {project.teams && project.teams.length > 0 ? (
              <div className="space-y-3">
                {project.teams.map((team) => (
                  <div
                    key={team.id}
                    className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                  >
                    <Link
                      to={`/teams/${team.id}`}
                      className="flex items-center gap-3 hover:text-primary-600"
                    >
                      <Users className="h-5 w-5 text-primary-600" />
                      <span className="font-medium">{team.name}</span>
                    </Link>
                    <button
                      onClick={() => handleRemoveTeam(team.id, team.name)}
                      className="p-1 text-gray-400 hover:text-red-600 transition-colors"
                      title="Remove team"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-gray-500 text-center py-4">No teams assigned</p>
            )}
            {unassignedTeams.length > 0 && project.project.status === 'active' && (
              <button
                onClick={() => setShowAssignTeamModal(true)}
                className="w-full mt-4 flex items-center justify-center gap-2 px-4 py-2 border border-dashed border-gray-300 text-gray-600 rounded-lg hover:border-primary-500 hover:text-primary-600 transition-colors"
              >
                <Plus className="h-4 w-4" />
                Assign Team
              </button>
            )}
          </div>

          {/* Project Info */}
          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Details</h2>
            <div className="space-y-3 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-gray-500">Status</span>
                <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                  project.project.status === 'active'
                    ? 'bg-green-100 text-green-700'
                    : 'bg-gray-100 text-gray-600'
                }`}>
                  {project.project.status}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-500">Total Tasks</span>
                <span className="font-medium">{project.task_count}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-500">Created</span>
                <span className="font-medium flex items-center gap-1">
                  <Calendar className="h-4 w-4 text-gray-400" />
                  {formatDate(project.project.created_at)}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Assign Team Modal */}
      {showAssignTeamModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold mb-4">Assign Team to Project</h2>
            <form onSubmit={handleAssignTeam}>
              <div className="mb-6">
                <label htmlFor="team" className="block text-sm font-medium text-gray-700 mb-1">
                  Select Team
                </label>
                {unassignedTeams.length > 0 ? (
                  <select
                    id="team"
                    value={selectedTeamId}
                    onChange={(e) => setSelectedTeamId(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                    required
                  >
                    <option value="">Choose a team...</option>
                    {unassignedTeams.map((teamResponse) => (
                      <option key={teamResponse.team.id} value={teamResponse.team.id}>
                        {teamResponse.team.name} ({teamResponse.member_count} members)
                      </option>
                    ))}
                  </select>
                ) : (
                  <p className="text-gray-500 py-2">All teams are already assigned to this project.</p>
                )}
              </div>
              <div className="flex gap-3 justify-end">
                <button
                  type="button"
                  onClick={() => setShowAssignTeamModal(false)}
                  className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  Cancel
                </button>
                {unassignedTeams.length > 0 && (
                  <button
                    type="submit"
                    disabled={isAssigning || !selectedTeamId}
                    className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 transition-colors"
                  >
                    {isAssigning ? 'Assigning...' : 'Assign Team'}
                  </button>
                )}
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
