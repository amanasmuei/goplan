import { useState } from 'react'
import type { Predictions, Assessment, BlockerRisk } from '../types'
import {
  Clock,
  AlertTriangle,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Lightbulb,
  TrendingUp,
  Shield,
} from 'lucide-react'
import { clsx } from 'clsx'

interface PredictionDisplayProps {
  predictions?: Predictions
  assessment?: Assessment
  userEstimate?: number
  compact?: boolean
}

export default function PredictionDisplay({
  predictions,
  assessment,
  userEstimate,
  compact = false,
}: PredictionDisplayProps) {
  const [expandedRisks, setExpandedRisks] = useState(false)
  const [expandedSuggestions, setExpandedSuggestions] = useState(false)

  if (!predictions && !assessment) {
    return null
  }

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.8) return 'text-green-600'
    if (confidence >= 0.6) return 'text-yellow-600'
    return 'text-red-600'
  }

  const getConfidenceLabel = (confidence: number) => {
    if (confidence >= 0.8) return 'High'
    if (confidence >= 0.6) return 'Medium'
    return 'Low'
  }

  const getScoreColor = (score: number) => {
    if (score >= 80) return 'bg-green-500'
    if (score >= 60) return 'bg-yellow-500'
    if (score >= 40) return 'bg-orange-500'
    return 'bg-red-500'
  }

  const isEstimateLow =
    predictions && userEstimate && userEstimate < predictions.predicted_days_low

  const isEstimateHigh =
    predictions &&
    userEstimate &&
    userEstimate > predictions.predicted_days_high * 1.5

  if (compact) {
    return (
      <div className="flex items-center gap-4 text-sm">
        {predictions && (
          <div className="flex items-center gap-2">
            <Clock className="h-4 w-4 text-blue-500" />
            <span className="font-medium">
              {predictions.predicted_days_low.toFixed(1)}-
              {predictions.predicted_days_high.toFixed(1)} days
            </span>
            <span className={clsx('text-xs', getConfidenceColor(predictions.confidence))}>
              ({getConfidenceLabel(predictions.confidence)})
            </span>
          </div>
        )}
        {assessment && (
          <div className="flex items-center gap-2">
            <Shield className="h-4 w-4 text-purple-500" />
            <span className="font-medium">{assessment.score}/100</span>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Timeline Prediction */}
      {predictions && (
        <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl p-6 border border-blue-200">
          <div className="flex items-center gap-2 mb-4">
            <Clock className="h-5 w-5 text-blue-600" />
            <h3 className="font-semibold text-blue-900">Timeline Prediction</h3>
          </div>

          <div className="grid grid-cols-2 gap-6">
            <div>
              <p className="text-sm text-blue-700 mb-1">Predicted Range</p>
              <p className="text-3xl font-bold text-blue-900">
                {predictions.predicted_days_low.toFixed(1)} -{' '}
                {predictions.predicted_days_high.toFixed(1)}
              </p>
              <p className="text-sm text-blue-600">days</p>
            </div>

            <div>
              <p className="text-sm text-blue-700 mb-1">Confidence</p>
              <div className="flex items-center gap-2">
                <div className="flex-1 bg-blue-200 rounded-full h-2">
                  <div
                    className={clsx(
                      'h-2 rounded-full transition-all',
                      predictions.confidence >= 0.8
                        ? 'bg-green-500'
                        : predictions.confidence >= 0.6
                        ? 'bg-yellow-500'
                        : 'bg-red-500'
                    )}
                    style={{ width: `${predictions.confidence * 100}%` }}
                  />
                </div>
                <span
                  className={clsx(
                    'text-sm font-medium',
                    getConfidenceColor(predictions.confidence)
                  )}
                >
                  {(predictions.confidence * 100).toFixed(0)}%
                </span>
              </div>
              <p className={clsx('text-xs mt-1', getConfidenceColor(predictions.confidence))}>
                {getConfidenceLabel(predictions.confidence)} confidence
              </p>
            </div>
          </div>

          {/* User Estimate Comparison */}
          {userEstimate && (
            <div className="mt-4 pt-4 border-t border-blue-200">
              <div className="flex items-center justify-between">
                <span className="text-sm text-blue-700">Your Estimate</span>
                <div className="flex items-center gap-2">
                  <span className="font-bold text-blue-900">{userEstimate} days</span>
                  {isEstimateLow && (
                    <span className="flex items-center gap-1 text-xs text-orange-600 bg-orange-100 px-2 py-0.5 rounded-full">
                      <AlertTriangle className="h-3 w-3" />
                      Below range
                    </span>
                  )}
                  {isEstimateHigh && (
                    <span className="flex items-center gap-1 text-xs text-orange-600 bg-orange-100 px-2 py-0.5 rounded-full">
                      <AlertTriangle className="h-3 w-3" />
                      Above range
                    </span>
                  )}
                  {!isEstimateLow && !isEstimateHigh && (
                    <span className="flex items-center gap-1 text-xs text-green-600 bg-green-100 px-2 py-0.5 rounded-full">
                      <CheckCircle2 className="h-3 w-3" />
                      Within range
                    </span>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Blocker Risks */}
      {predictions && predictions.blocker_risks && predictions.blocker_risks.length > 0 && (
        <div className="bg-gradient-to-r from-red-50 to-orange-50 rounded-xl border border-red-200 overflow-hidden">
          <button
            onClick={() => setExpandedRisks(!expandedRisks)}
            className="w-full p-4 flex items-center justify-between hover:bg-red-100/50 transition-colors"
          >
            <div className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-600" />
              <h3 className="font-semibold text-red-900">
                Blocker Risks ({predictions.blocker_risks.length})
              </h3>
            </div>
            {expandedRisks ? (
              <ChevronUp className="h-5 w-5 text-red-400" />
            ) : (
              <ChevronDown className="h-5 w-5 text-red-400" />
            )}
          </button>

          {expandedRisks && (
            <div className="px-4 pb-4 space-y-3">
              {predictions.blocker_risks.map((risk, index) => (
                <BlockerRiskCard key={index} risk={risk} />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Planning Quality Assessment */}
      {assessment && (
        <div className="bg-gradient-to-r from-purple-50 to-pink-50 rounded-xl border border-purple-200 overflow-hidden">
          <div className="p-4">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-purple-600" />
                <h3 className="font-semibold text-purple-900">Planning Quality</h3>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-2xl font-bold text-purple-900">
                  {assessment.score}
                </span>
                <span className="text-purple-600">/100</span>
              </div>
            </div>

            {/* Score Bar */}
            <div className="mb-4">
              <div className="bg-purple-200 rounded-full h-3">
                <div
                  className={clsx('h-3 rounded-full transition-all', getScoreColor(assessment.score))}
                  style={{ width: `${assessment.score}%` }}
                />
              </div>
            </div>

            {/* Breakdown */}
            <div className="grid grid-cols-5 gap-2 mb-4">
              {Object.entries(assessment.breakdown).map(([key, value]) => (
                <div key={key} className="text-center">
                  <div
                    className={clsx(
                      'text-xs font-medium mb-1',
                      value >= 15 ? 'text-green-600' : value >= 10 ? 'text-yellow-600' : 'text-red-600'
                    )}
                  >
                    {value}/20
                  </div>
                  <div className="text-xs text-gray-500 capitalize">
                    {key.replace(/_/g, ' ').slice(0, 10)}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Suggestions */}
          {assessment.suggestions && assessment.suggestions.length > 0 && (
            <>
              <button
                onClick={() => setExpandedSuggestions(!expandedSuggestions)}
                className="w-full px-4 py-3 flex items-center justify-between border-t border-purple-200 hover:bg-purple-100/50 transition-colors"
              >
                <div className="flex items-center gap-2">
                  <Lightbulb className="h-4 w-4 text-yellow-500" />
                  <span className="text-sm font-medium text-purple-900">
                    Improvement Suggestions ({assessment.suggestions.length})
                  </span>
                </div>
                {expandedSuggestions ? (
                  <ChevronUp className="h-4 w-4 text-purple-400" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-purple-400" />
                )}
              </button>

              {expandedSuggestions && (
                <div className="px-4 pb-4 space-y-2">
                  {assessment.suggestions.map((suggestion, index) => (
                    <div
                      key={index}
                      className="flex items-start gap-2 p-3 bg-white rounded-lg border border-purple-100"
                    >
                      <span className="text-yellow-500 mt-0.5">•</span>
                      <p className="text-sm text-gray-700">{suggestion}</p>
                    </div>
                  ))}
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
}

function BlockerRiskCard({ risk }: { risk: BlockerRisk }) {
  const [expanded, setExpanded] = useState(false)

  const getProbabilityColor = (prob: number) => {
    if (prob >= 0.7) return 'bg-red-500'
    if (prob >= 0.5) return 'bg-orange-500'
    return 'bg-yellow-500'
  }

  const getProbabilityLabel = (prob: number) => {
    if (prob >= 0.7) return 'High Risk'
    if (prob >= 0.5) return 'Medium Risk'
    return 'Low Risk'
  }

  return (
    <div className="bg-white rounded-lg border border-red-100 overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full p-3 flex items-center justify-between hover:bg-red-50 transition-colors"
      >
        <div className="flex items-center gap-3">
          <div
            className={clsx(
              'w-2 h-2 rounded-full',
              getProbabilityColor(risk.probability)
            )}
          />
          <span className="font-medium text-gray-900 capitalize">
            {risk.type.replace(/_/g, ' ')}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <span
            className={clsx(
              'text-xs font-medium px-2 py-0.5 rounded-full',
              risk.probability >= 0.7
                ? 'bg-red-100 text-red-700'
                : risk.probability >= 0.5
                ? 'bg-orange-100 text-orange-700'
                : 'bg-yellow-100 text-yellow-700'
            )}
          >
            {(risk.probability * 100).toFixed(0)}% - {getProbabilityLabel(risk.probability)}
          </span>
          {risk.examples && risk.examples.length > 0 && (
            expanded ? (
              <ChevronUp className="h-4 w-4 text-gray-400" />
            ) : (
              <ChevronDown className="h-4 w-4 text-gray-400" />
            )
          )}
        </div>
      </button>

      {expanded && risk.examples && risk.examples.length > 0 && (
        <div className="px-3 pb-3 border-t border-red-100">
          <p className="text-xs text-gray-500 mb-2 mt-2">Historical examples:</p>
          <div className="space-y-1">
            {risk.examples.map((example, index) => (
              <p key={index} className="text-xs text-gray-600 italic pl-4 border-l-2 border-red-200">
                {example}
              </p>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
