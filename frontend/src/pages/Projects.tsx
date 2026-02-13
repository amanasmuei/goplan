import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { projectApi } from '../services/api'
import type { ProjectResponse, CreateProjectRequest, ProjectStatus } from '../types'
import { FolderKanban, Plus, Archive, Trash2, AlertCircle } from 'lucide-react'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

export default function Projects() {
  useDocumentTitle('Projects')
  const [projects, setProjects] = useState<ProjectResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newProject, setNewProject] = useState<CreateProjectRequest>({ name: '', description: '' })
  const [isCreating, setIsCreating] = useState(false)
  const [statusFilter, setStatusFilter] = useState<ProjectStatus | 'all'>('active')

  useEffect(() => {
    fetchProjects()
  }, [statusFilter])

  const fetchProjects = async () => {
    try {
      setIsLoading(true)
      const params = statusFilter !== 'all' ? { status: statusFilter } : undefined
      const response = await projectApi.list(params)
      setProjects(response.projects || [])
      setError(null)
    } catch (err) {
      setError('Failed to load projects')
      console.error('Error fetching projects:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateProject = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newProject.name.trim()) return

    try {
      setIsCreating(true)
      await projectApi.create(newProject)
      setNewProject({ name: '', description: '' })
      setShowCreateModal(false)
      fetchProjects()
    } catch (err) {
      setError('Failed to create project')
      console.error('Error creating project:', err)
    } finally {
      setIsCreating(false)
    }
  }

  const handleArchiveProject = async (projectId: string) => {
    if (!confirm('Are you sure you want to archive this project?')) return

    try {
      await projectApi.archive(projectId)
      fetchProjects()
    } catch (err) {
      setError('Failed to archive project')
      console.error('Error archiving project:', err)
    }
  }

  const handleDeleteProject = async (projectId: string) => {
    if (!confirm('Are you sure you want to delete this project? This cannot be undone.')) return

    try {
      await projectApi.delete(projectId)
      fetchProjects()
    } catch (err) {
      setError('Failed to delete project. Projects with tasks cannot be deleted.')
      console.error('Error deleting project:', err)
    }
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64" aria-live="polite">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        <span className="sr-only">Loading projects...</span>
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Projects</h1>
          <p className="text-gray-600 mt-1">Manage your organization's projects</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          <Plus className="h-5 w-5" />
          New Project
        </button>
      </div>

      {/* Status Filter */}
      <div className="mb-6 flex gap-2" role="group" aria-label="Filter projects by status">
        {(['active', 'archived', 'all'] as const).map((status) => (
          <button
            key={status}
            onClick={() => setStatusFilter(status)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              statusFilter === status
                ? 'bg-primary-100 text-primary-700'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
            aria-pressed={statusFilter === status}
          >
            {status.charAt(0).toUpperCase() + status.slice(1)}
          </button>
        ))}
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700" role="alert">
          <AlertCircle className="h-5 w-5" />
          {error}
        </div>
      )}

      {projects.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <FolderKanban className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No projects found</h3>
          <p className="text-gray-600 mb-4">
            {statusFilter === 'all'
              ? 'Create your first project to get started'
              : `No ${statusFilter} projects found`}
          </p>
          {statusFilter === 'active' && (
            <button
              onClick={() => setShowCreateModal(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
            >
              <Plus className="h-5 w-5" />
              Create Project
            </button>
          )}
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {projects.map((projectResponse) => (
            <div
              key={projectResponse.project.id}
              className={`bg-white rounded-lg border p-6 hover:shadow-md transition-shadow ${
                projectResponse.project.status === 'archived'
                  ? 'border-gray-300 bg-gray-50'
                  : 'border-gray-200'
              }`}
            >
              <div className="flex justify-between items-start mb-4">
                <div className="flex items-center gap-3">
                  <div className={`h-10 w-10 rounded-lg flex items-center justify-center ${
                    projectResponse.project.status === 'archived'
                      ? 'bg-gray-200'
                      : 'bg-primary-100'
                  }`}>
                    <FolderKanban className={`h-5 w-5 ${
                      projectResponse.project.status === 'archived'
                        ? 'text-gray-500'
                        : 'text-primary-600'
                    }`} />
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-900">{projectResponse.project.name}</h3>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-gray-500">{projectResponse.task_count} tasks</span>
                      {projectResponse.project.status === 'archived' && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 bg-gray-200 text-gray-600 text-xs rounded-full">
                          <Archive className="h-3 w-3" />
                          Archived
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                <div className="flex gap-1">
                  {projectResponse.project.status === 'active' && (
                    <button
                      onClick={() => handleArchiveProject(projectResponse.project.id)}
                      className="p-2 text-gray-400 hover:text-yellow-600 transition-colors"
                      aria-label={`Archive ${projectResponse.project.name}`}
                    >
                      <Archive className="h-4 w-4" />
                    </button>
                  )}
                  {projectResponse.task_count === 0 && (
                    <button
                      onClick={() => handleDeleteProject(projectResponse.project.id)}
                      className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                      aria-label={`Delete ${projectResponse.project.name}`}
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  )}
                </div>
              </div>
              {projectResponse.project.description && (
                <p className="text-sm text-gray-600 line-clamp-2">{projectResponse.project.description}</p>
              )}
              {projectResponse.teams && projectResponse.teams.length > 0 && (
                <div className="mt-3 flex flex-wrap gap-1">
                  {projectResponse.teams.map((team) => (
                    <span
                      key={team.id}
                      className="inline-flex items-center px-2 py-0.5 bg-blue-50 text-blue-700 text-xs rounded-full"
                    >
                      {team.name}
                    </span>
                  ))}
                </div>
              )}
              <div className="mt-4 pt-4 border-t border-gray-100">
                <Link
                  to={`/projects/${projectResponse.project.id}`}
                  className="text-sm text-primary-600 hover:text-primary-700 font-medium"
                >
                  View Project &rarr;
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Project Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md" role="dialog" aria-labelledby="create-project-title">
            <h2 id="create-project-title" className="text-xl font-semibold mb-4">Create New Project</h2>
            <form onSubmit={handleCreateProject}>
              <div className="mb-4">
                <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                  Project Name
                </label>
                <input
                  type="text"
                  id="name"
                  value={newProject.name}
                  onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="GoPlan MVP"
                  required
                />
              </div>
              <div className="mb-6">
                <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <textarea
                  id="description"
                  value={newProject.description}
                  onChange={(e) => setNewProject({ ...newProject, description: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="Main product development project"
                  rows={3}
                />
              </div>
              <div className="flex gap-3 justify-end">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating || !newProject.name.trim()}
                  className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 transition-colors"
                >
                  {isCreating ? 'Creating...' : 'Create Project'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
