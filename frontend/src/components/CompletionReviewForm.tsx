import { useState } from 'react'
import { Star, Clock, AlertTriangle, Lightbulb, RefreshCw } from 'lucide-react'
import { clsx } from 'clsx'
import type { Task, CreateReviewRequest } from '../types'

interface CompletionReviewFormProps {
  task: Task
  onSubmit: (data: CreateReviewRequest) => void
  onCancel: () => void
  isPending: boolean
}

export default function CompletionReviewForm({
  task,
  onSubmit,
  onCancel,
  isPending,
}: CompletionReviewFormProps) {
  const [reviewData, setReviewData] = useState<CreateReviewRequest>({
    prediction_accuracy_rating: 3,
    prediction_feedback: '',
    lessons_learned: '',
    would_approach_differently: '',
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit(reviewData)
  }

  // Calculate actual vs predicted variance
  const predictionVariance = (() => {
    if (!task.actual_days || !task.predicted_days_low || !task.predicted_days_high) {
      return null
    }
    const midPrediction = (task.predicted_days_low + task.predicted_days_high) / 2
    const variance = ((task.actual_days - midPrediction) / midPrediction) * 100
    return variance
  })()

  const getVarianceLabel = () => {
    if (predictionVariance === null) return null
    if (Math.abs(predictionVariance) <= 10) return { text: 'On target', color: 'text-green-600' }
    if (predictionVariance > 0) return { text: `${predictionVariance.toFixed(0)}% longer`, color: 'text-red-600' }
    return { text: `${Math.abs(predictionVariance).toFixed(0)}% faster`, color: 'text-blue-600' }
  }

  const varianceLabel = getVarianceLabel()

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Timing Summary */}
      <div className="bg-gray-50 rounded-lg p-4">
        <h3 className="font-medium text-gray-900 mb-3 flex items-center gap-2">
          <Clock className="h-5 w-5 text-gray-500" />
          Timing Summary
        </h3>
        <div className="grid grid-cols-3 gap-4 text-center">
          <div>
            <p className="text-sm text-gray-500">Estimated</p>
            <p className="text-lg font-semibold text-gray-900">
              {task.estimated_days ? `${task.estimated_days} days` : '-'}
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Predicted</p>
            <p className="text-lg font-semibold text-gray-900">
              {task.predicted_days_low && task.predicted_days_high
                ? `${task.predicted_days_low.toFixed(1)}-${task.predicted_days_high.toFixed(1)} days`
                : '-'}
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Actual</p>
            <p className="text-lg font-semibold text-gray-900">
              {task.actual_days ? `${task.actual_days.toFixed(1)} days` : '-'}
            </p>
          </div>
        </div>
        {varianceLabel && (
          <div className="mt-3 pt-3 border-t border-gray-200 text-center">
            <span className={clsx('font-medium', varianceLabel.color)}>
              {varianceLabel.text} than predicted
            </span>
          </div>
        )}
      </div>

      {/* Prediction Accuracy Rating */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          How accurate was the system's prediction?
        </label>
        <p className="text-xs text-gray-500 mb-3">
          Rate from 1 (completely wrong) to 5 (spot on)
        </p>
        <div className="flex gap-2 justify-center">
          {[1, 2, 3, 4, 5].map((rating) => (
            <button
              key={rating}
              type="button"
              onClick={() => setReviewData((d) => ({ ...d, prediction_accuracy_rating: rating }))}
              className={clsx(
                'p-3 rounded-xl border-2 transition-all',
                reviewData.prediction_accuracy_rating === rating
                  ? 'border-primary-500 bg-primary-50 scale-110'
                  : 'border-gray-200 hover:border-gray-300'
              )}
            >
              <Star
                className={clsx(
                  'h-8 w-8 transition-colors',
                  reviewData.prediction_accuracy_rating >= rating
                    ? 'text-primary-500 fill-current'
                    : 'text-gray-300'
                )}
              />
            </button>
          ))}
        </div>
        <div className="flex justify-between text-xs text-gray-500 mt-2 px-2">
          <span>Way off</span>
          <span>Perfect</span>
        </div>
      </div>

      {/* Prediction Feedback */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
          <AlertTriangle className="h-4 w-4 text-yellow-500" />
          What factors affected the timeline?
        </label>
        <textarea
          value={reviewData.prediction_feedback || ''}
          onChange={(e) => setReviewData((d) => ({ ...d, prediction_feedback: e.target.value }))}
          className="w-full border border-gray-300 rounded-lg p-3
                   focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          rows={3}
          placeholder="Were there unexpected blockers? Technical challenges? Dependencies that weren't anticipated?"
        />
      </div>

      {/* Lessons Learned */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
          <Lightbulb className="h-4 w-4 text-yellow-500" />
          Lessons learned for similar tasks
        </label>
        <textarea
          value={reviewData.lessons_learned || ''}
          onChange={(e) => setReviewData((d) => ({ ...d, lessons_learned: e.target.value }))}
          className="w-full border border-gray-300 rounded-lg p-3
                   focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          rows={3}
          placeholder="What insights would help someone working on a similar task in the future?"
        />
      </div>

      {/* Would Approach Differently */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2 flex items-center gap-2">
          <RefreshCw className="h-4 w-4 text-blue-500" />
          What would you do differently?
        </label>
        <textarea
          value={reviewData.would_approach_differently || ''}
          onChange={(e) => setReviewData((d) => ({ ...d, would_approach_differently: e.target.value }))}
          className="w-full border border-gray-300 rounded-lg p-3
                   focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          rows={3}
          placeholder="Any changes to approach, planning, or execution you'd recommend?"
        />
      </div>

      {/* Actions */}
      <div className="flex gap-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          onClick={onCancel}
          className="flex-1 px-4 py-2 border border-gray-300 rounded-lg
                   font-medium text-gray-700 hover:bg-gray-50 transition-colors"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isPending}
          className={clsx(
            'flex-1 px-4 py-2 rounded-lg font-medium transition-colors',
            isPending
              ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
              : 'bg-primary-600 text-white hover:bg-primary-700'
          )}
        >
          {isPending ? 'Submitting...' : 'Submit Review & Complete'}
        </button>
      </div>
    </form>
  )
}
