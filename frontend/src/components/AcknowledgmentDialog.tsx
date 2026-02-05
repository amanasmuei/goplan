import { useState } from 'react'
import { X, CheckCircle2, Edit3, AlertCircle, Clock } from 'lucide-react'
import { clsx } from 'clsx'
import type { Task, AcknowledgmentAction, AcknowledgmentRequest } from '../types'

interface AcknowledgmentDialogProps {
  task: Task
  isOpen: boolean
  onClose: () => void
  onAcknowledge: (data: AcknowledgmentRequest) => void
  isPending: boolean
}

export default function AcknowledgmentDialog({
  task,
  isOpen,
  onClose,
  onAcknowledge,
  isPending,
}: AcknowledgmentDialogProps) {
  const [action, setAction] = useState<AcknowledgmentAction>('accept')
  const [modifiedEstimate, setModifiedEstimate] = useState<string>(
    task.estimated_days?.toString() || ''
  )
  const [disagreementNotes, setDisagreementNotes] = useState('')

  if (!isOpen) return null

  const handleSubmit = () => {
    const data: AcknowledgmentRequest = { action }

    if (action === 'modify') {
      data.modified_estimate = parseFloat(modifiedEstimate)
    } else if (action === 'disagree') {
      data.disagreement_notes = disagreementNotes
    }

    onAcknowledge(data)
  }

  const isValid = () => {
    if (action === 'modify') {
      const estimate = parseFloat(modifiedEstimate)
      return !isNaN(estimate) && estimate > 0
    }
    if (action === 'disagree') {
      return disagreementNotes.length >= 20
    }
    return true
  }

  const predictionRange = task.predicted_days_low && task.predicted_days_high
    ? `${task.predicted_days_low.toFixed(1)}-${task.predicted_days_high.toFixed(1)} days`
    : 'No prediction available'

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl shadow-xl max-w-lg w-full mx-4 overflow-hidden">
        {/* Header */}
        <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-4 text-white">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Acknowledge Task Predictions</h2>
            <button
              onClick={onClose}
              className="p-1 hover:bg-white/20 rounded-lg transition-colors"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        </div>

        <div className="p-6">
          {/* Prediction Summary */}
          <div className="bg-blue-50 rounded-lg p-4 mb-6">
            <div className="flex items-center gap-2 text-blue-700 mb-2">
              <Clock className="h-5 w-5" />
              <span className="font-medium">System Prediction</span>
            </div>
            <p className="text-2xl font-bold text-blue-900">{predictionRange}</p>
            {task.prediction_confidence && (
              <p className="text-sm text-blue-600 mt-1">
                {(task.prediction_confidence * 100).toFixed(0)}% confidence based on similar tasks
              </p>
            )}
            {task.estimated_days && (
              <p className="text-sm text-gray-600 mt-2">
                Your estimate: <span className="font-medium">{task.estimated_days} days</span>
              </p>
            )}
          </div>

          {/* Action Options */}
          <div className="space-y-3 mb-6">
            <p className="text-sm font-medium text-gray-700">How would you like to proceed?</p>

            {/* Accept Option */}
            <button
              onClick={() => setAction('accept')}
              className={clsx(
                'w-full p-4 rounded-lg border-2 text-left transition-all',
                action === 'accept'
                  ? 'border-green-500 bg-green-50'
                  : 'border-gray-200 hover:border-gray-300'
              )}
            >
              <div className="flex items-start gap-3">
                <CheckCircle2
                  className={clsx(
                    'h-6 w-6 mt-0.5',
                    action === 'accept' ? 'text-green-600' : 'text-gray-400'
                  )}
                />
                <div>
                  <p className="font-medium text-gray-900">Accept Prediction</p>
                  <p className="text-sm text-gray-600">
                    I agree with the system's timeline prediction and my original estimate.
                  </p>
                </div>
              </div>
            </button>

            {/* Modify Option */}
            <button
              onClick={() => setAction('modify')}
              className={clsx(
                'w-full p-4 rounded-lg border-2 text-left transition-all',
                action === 'modify'
                  ? 'border-yellow-500 bg-yellow-50'
                  : 'border-gray-200 hover:border-gray-300'
              )}
            >
              <div className="flex items-start gap-3">
                <Edit3
                  className={clsx(
                    'h-6 w-6 mt-0.5',
                    action === 'modify' ? 'text-yellow-600' : 'text-gray-400'
                  )}
                />
                <div className="flex-1">
                  <p className="font-medium text-gray-900">Modify Estimate</p>
                  <p className="text-sm text-gray-600 mb-3">
                    I want to provide a different time estimate based on my analysis.
                  </p>
                  {action === 'modify' && (
                    <div className="mt-2">
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Your revised estimate (days)
                      </label>
                      <input
                        type="number"
                        min="0.5"
                        step="0.5"
                        value={modifiedEstimate}
                        onChange={(e) => setModifiedEstimate(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg
                                 focus:ring-2 focus:ring-yellow-500 focus:border-yellow-500"
                        placeholder="e.g., 5"
                      />
                    </div>
                  )}
                </div>
              </div>
            </button>

            {/* Disagree Option */}
            <button
              onClick={() => setAction('disagree')}
              className={clsx(
                'w-full p-4 rounded-lg border-2 text-left transition-all',
                action === 'disagree'
                  ? 'border-red-500 bg-red-50'
                  : 'border-gray-200 hover:border-gray-300'
              )}
            >
              <div className="flex items-start gap-3">
                <AlertCircle
                  className={clsx(
                    'h-6 w-6 mt-0.5',
                    action === 'disagree' ? 'text-red-600' : 'text-gray-400'
                  )}
                />
                <div className="flex-1">
                  <p className="font-medium text-gray-900">Disagree with Prediction</p>
                  <p className="text-sm text-gray-600 mb-3">
                    The prediction doesn't apply well to this task, but I'll proceed anyway.
                  </p>
                  {action === 'disagree' && (
                    <div className="mt-2">
                      <label className="block text-sm font-medium text-gray-700 mb-1">
                        Please explain why (min 20 characters)
                      </label>
                      <textarea
                        value={disagreementNotes}
                        onChange={(e) => setDisagreementNotes(e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg
                                 focus:ring-2 focus:ring-red-500 focus:border-red-500"
                        rows={3}
                        placeholder="The similar tasks don't account for..."
                      />
                      <p className="text-xs text-gray-500 mt-1">
                        {disagreementNotes.length}/20 characters
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </button>
          </div>

          {/* Actions */}
          <div className="flex gap-3">
            <button
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg
                       font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={!isValid() || isPending}
              className={clsx(
                'flex-1 px-4 py-2 rounded-lg font-medium transition-colors',
                isValid() && !isPending
                  ? 'bg-blue-600 text-white hover:bg-blue-700'
                  : 'bg-gray-300 text-gray-500 cursor-not-allowed'
              )}
            >
              {isPending ? 'Processing...' : 'Acknowledge & Continue'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
