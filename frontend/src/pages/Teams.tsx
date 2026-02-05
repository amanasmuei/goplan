import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { teamApi } from '../services/api'
import type { TeamResponse, CreateTeamRequest } from '../types'
import { Users, Plus, Settings, Trash2, AlertCircle } from 'lucide-react'

export default function Teams() {
  const [teams, setTeams] = useState<TeamResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [newTeam, setNewTeam] = useState<CreateTeamRequest>({ name: '', description: '' })
  const [isCreating, setIsCreating] = useState(false)

  useEffect(() => {
    fetchTeams()
  }, [])

  const fetchTeams = async () => {
    try {
      setIsLoading(true)
      const response = await teamApi.list()
      setTeams(response.teams || [])
      setError(null)
    } catch (err) {
      setError('Failed to load teams')
      console.error('Error fetching teams:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreateTeam = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTeam.name.trim()) return

    try {
      setIsCreating(true)
      await teamApi.create(newTeam)
      setNewTeam({ name: '', description: '' })
      setShowCreateModal(false)
      fetchTeams()
    } catch (err) {
      setError('Failed to create team')
      console.error('Error creating team:', err)
    } finally {
      setIsCreating(false)
    }
  }

  const handleDeleteTeam = async (teamId: string) => {
    if (!confirm('Are you sure you want to delete this team?')) return

    try {
      await teamApi.delete(teamId)
      fetchTeams()
    } catch (err) {
      setError('Failed to delete team')
      console.error('Error deleting team:', err)
    }
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Teams</h1>
          <p className="text-gray-600 mt-1">Manage your organization's teams</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          <Plus className="h-5 w-5" />
          New Team
        </button>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2 text-red-700">
          <AlertCircle className="h-5 w-5" />
          {error}
        </div>
      )}

      {teams.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No teams yet</h3>
          <p className="text-gray-600 mb-4">Create your first team to start collaborating</p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
          >
            <Plus className="h-5 w-5" />
            Create Team
          </button>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {teams.map((teamResponse) => (
            <div
              key={teamResponse.team.id}
              className="bg-white rounded-lg border border-gray-200 p-6 hover:shadow-md transition-shadow"
            >
              <div className="flex justify-between items-start mb-4">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-lg bg-primary-100 flex items-center justify-center">
                    <Users className="h-5 w-5 text-primary-600" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-900">{teamResponse.team.name}</h3>
                    <p className="text-sm text-gray-500">{teamResponse.member_count} members</p>
                  </div>
                </div>
                <div className="flex gap-1">
                  <Link
                    to={`/teams/${teamResponse.team.id}`}
                    className="p-2 text-gray-400 hover:text-primary-600 transition-colors"
                    title="Team settings"
                  >
                    <Settings className="h-4 w-4" />
                  </Link>
                  <button
                    onClick={() => handleDeleteTeam(teamResponse.team.id)}
                    className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                    title="Delete team"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
              {teamResponse.team.description && (
                <p className="text-sm text-gray-600 line-clamp-2">{teamResponse.team.description}</p>
              )}
              <div className="mt-4 pt-4 border-t border-gray-100">
                <Link
                  to={`/teams/${teamResponse.team.id}`}
                  className="text-sm text-primary-600 hover:text-primary-700 font-medium"
                >
                  View Team &rarr;
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Team Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold mb-4">Create New Team</h2>
            <form onSubmit={handleCreateTeam}>
              <div className="mb-4">
                <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                  Team Name
                </label>
                <input
                  type="text"
                  id="name"
                  value={newTeam.name}
                  onChange={(e) => setNewTeam({ ...newTeam, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="Engineering"
                  required
                />
              </div>
              <div className="mb-6">
                <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <textarea
                  id="description"
                  value={newTeam.description}
                  onChange={(e) => setNewTeam({ ...newTeam, description: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="Core engineering team"
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
                  disabled={isCreating || !newTeam.name.trim()}
                  className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 transition-colors"
                >
                  {isCreating ? 'Creating...' : 'Create Team'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
