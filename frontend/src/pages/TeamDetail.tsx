import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { teamApi } from '../services/api'
import type { TeamResponse, ProjectResponse, AddTeamMemberRequest, TeamRole } from '../types'
import {
  Users,
  ArrowLeft,
  Plus,
  Trash2,
  AlertCircle,
  Shield,
  UserCog,
  User,
  Eye,
  FolderKanban
} from 'lucide-react'

const roleColors: Record<TeamRole, string> = {
  owner: 'bg-purple-100 text-purple-700',
  manager: 'bg-blue-100 text-blue-700',
  member: 'bg-green-100 text-green-700',
  viewer: 'bg-gray-100 text-gray-600',
}

export default function TeamDetail() {
  const { id } = useParams<{ id: string }>()
  const [team, setTeam] = useState<TeamResponse | null>(null)
  const [projects, setProjects] = useState<ProjectResponse[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showAddMemberModal, setShowAddMemberModal] = useState(false)
  const [newMember, setNewMember] = useState<AddTeamMemberRequest>({ user_id: '', role: 'member' })
  const [isAdding, setIsAdding] = useState(false)

  useEffect(() => {
    if (id) {
      fetchTeam()
      fetchProjects()
    }
  }, [id])

  const fetchTeam = async () => {
    if (!id) return
    try {
      setIsLoading(true)
      const response = await teamApi.get(id)
      setTeam(response)
      setError(null)
    } catch (err) {
      setError('Failed to load team')
      console.error('Error fetching team:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const fetchProjects = async () => {
    if (!id) return
    try {
      const response = await teamApi.listProjects(id)
      setProjects(response.projects || [])
    } catch (err) {
      console.error('Error fetching team projects:', err)
    }
  }

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id || !newMember.user_id.trim()) return

    try {
      setIsAdding(true)
      await teamApi.addMember(id, newMember)
      setNewMember({ user_id: '', role: 'member' })
      setShowAddMemberModal(false)
      fetchTeam()
    } catch (err) {
      setError('Failed to add member')
      console.error('Error adding member:', err)
    } finally {
      setIsAdding(false)
    }
  }

  const handleUpdateRole = async (userId: string, role: TeamRole) => {
    if (!id) return
    try {
      await teamApi.updateMemberRole(id, userId, { role })
      fetchTeam()
    } catch (err) {
      setError('Failed to update member role')
      console.error('Error updating role:', err)
    }
  }

  const handleRemoveMember = async (userId: string, memberName: string) => {
    if (!id) return
    if (!confirm(`Are you sure you want to remove ${memberName} from this team?`)) return

    try {
      await teamApi.removeMember(id, userId)
      fetchTeam()
    } catch (err) {
      setError('Failed to remove member')
      console.error('Error removing member:', err)
    }
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    )
  }

  if (!team) {
    return (
      <div className="text-center py-12">
        <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
        <h3 className="text-lg font-medium text-gray-900 mb-2">Team not found</h3>
        <Link to="/teams" className="text-primary-600 hover:text-primary-700">
          &larr; Back to Teams
        </Link>
      </div>
    )
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-6">
        <Link
          to="/teams"
          className="inline-flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Teams
        </Link>
        <div className="flex justify-between items-start">
          <div className="flex items-center gap-4">
            <div className="h-14 w-14 rounded-xl bg-primary-100 flex items-center justify-center">
              <Users className="h-7 w-7 text-primary-600" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{team.team.name}</h1>
              {team.team.description && (
                <p className="text-gray-600 mt-1">{team.team.description}</p>
              )}
            </div>
          </div>
          <button
            onClick={() => setShowAddMemberModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
          >
            <Plus className="h-5 w-5" />
            Add Member
          </button>
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
        {/* Members */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Members ({team.member_count})
            </h2>
            {team.members && team.members.length > 0 ? (
              <div className="space-y-3">
                {team.members.map((member) => (
                    <div
                      key={member.id}
                      className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                    >
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-primary-100 flex items-center justify-center">
                          <span className="text-sm font-medium text-primary-700">
                            {member.user?.name?.charAt(0) || 'U'}
                          </span>
                        </div>
                        <div>
                          <p className="font-medium text-gray-900">
                            {member.user?.name || 'Unknown User'}
                          </p>
                          <p className="text-sm text-gray-500">{member.user?.email}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <select
                          value={member.role}
                          onChange={(e) => handleUpdateRole(member.user_id, e.target.value as TeamRole)}
                          className={`px-3 py-1.5 rounded-lg text-sm font-medium border-0 cursor-pointer ${roleColors[member.role]}`}
                        >
                          <option value="owner">Owner</option>
                          <option value="manager">Manager</option>
                          <option value="member">Member</option>
                          <option value="viewer">Viewer</option>
                        </select>
                        <button
                          onClick={() => handleRemoveMember(member.user_id, member.user?.name || 'this user')}
                          className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                          title="Remove member"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                  ))}
              </div>
            ) : (
              <p className="text-gray-500 text-center py-8">No members yet</p>
            )}
          </div>
        </div>

        {/* Projects */}
        <div>
          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Projects ({projects.length})
            </h2>
            {projects.length > 0 ? (
              <div className="space-y-3">
                {projects.map((projectResponse) => (
                  <Link
                    key={projectResponse.project.id}
                    to={`/projects/${projectResponse.project.id}`}
                    className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                  >
                    <FolderKanban className="h-5 w-5 text-primary-600" />
                    <div>
                      <p className="font-medium text-gray-900">{projectResponse.project.name}</p>
                      <p className="text-sm text-gray-500">{projectResponse.task_count} tasks</p>
                    </div>
                  </Link>
                ))}
              </div>
            ) : (
              <p className="text-gray-500 text-center py-8">No projects assigned</p>
            )}
          </div>

          {/* Role Legend */}
          <div className="mt-6 bg-white rounded-lg border border-gray-200 p-6">
            <h3 className="text-sm font-semibold text-gray-900 mb-3">Role Permissions</h3>
            <div className="space-y-2 text-sm">
              <div className="flex items-center gap-2">
                <Shield className="h-4 w-4 text-purple-600" />
                <span className="text-gray-600"><strong>Owner</strong> - Full control, can delete team</span>
              </div>
              <div className="flex items-center gap-2">
                <UserCog className="h-4 w-4 text-blue-600" />
                <span className="text-gray-600"><strong>Manager</strong> - Manage members, edit team</span>
              </div>
              <div className="flex items-center gap-2">
                <User className="h-4 w-4 text-green-600" />
                <span className="text-gray-600"><strong>Member</strong> - Work on projects</span>
              </div>
              <div className="flex items-center gap-2">
                <Eye className="h-4 w-4 text-gray-500" />
                <span className="text-gray-600"><strong>Viewer</strong> - View only access</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Add Member Modal */}
      {showAddMemberModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold mb-4">Add Team Member</h2>
            <form onSubmit={handleAddMember}>
              <div className="mb-4">
                <label htmlFor="user_id" className="block text-sm font-medium text-gray-700 mb-1">
                  User ID
                </label>
                <input
                  type="text"
                  id="user_id"
                  value={newMember.user_id}
                  onChange={(e) => setNewMember({ ...newMember, user_id: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="Enter user ID"
                  required
                />
              </div>
              <div className="mb-6">
                <label htmlFor="role" className="block text-sm font-medium text-gray-700 mb-1">
                  Role
                </label>
                <select
                  id="role"
                  value={newMember.role}
                  onChange={(e) => setNewMember({ ...newMember, role: e.target.value as TeamRole })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                >
                  <option value="member">Member</option>
                  <option value="viewer">Viewer</option>
                  <option value="manager">Manager</option>
                  <option value="owner">Owner</option>
                </select>
              </div>
              <div className="flex gap-3 justify-end">
                <button
                  type="button"
                  onClick={() => setShowAddMemberModal(false)}
                  className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isAdding || !newMember.user_id.trim()}
                  className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 transition-colors"
                >
                  {isAdding ? 'Adding...' : 'Add Member'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
