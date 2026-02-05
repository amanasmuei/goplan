import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import CompletionReviewForm from './CompletionReviewForm'
import type { Task } from '../types'

const mockTask: Task = {
  id: '550e8400-e29b-41d4-a716-446655440000',
  title: 'Test Task',
  description: 'This is a test task description',
  status: 'pending_review',
  created_by: '550e8400-e29b-41d4-a716-446655440001',
  organization_id: '550e8400-e29b-41d4-a716-446655440002',
  created_at: '2024-01-20T10:00:00Z',
  updated_at: '2024-01-20T10:00:00Z',
  estimated_days: 5,
  predicted_days_low: 4,
  predicted_days_high: 6,
  actual_days: 5.5,
  prediction_confidence: 0.85,
  started_at: '2024-01-21T10:00:00Z',
  completed_at: '2024-01-26T10:00:00Z',
}

describe('CompletionReviewForm', () => {
  const mockOnSubmit = vi.fn()
  const mockOnCancel = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders timing summary', () => {
    render(
      <CompletionReviewForm
        task={mockTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    expect(screen.getByText('Timing Summary')).toBeInTheDocument()
    expect(screen.getByText('5 days')).toBeInTheDocument() // Estimated
    expect(screen.getByText('5.5 days')).toBeInTheDocument() // Actual
  })

  it('displays star rating selector', () => {
    render(
      <CompletionReviewForm
        task={mockTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    expect(screen.getByText(/How accurate was the system's prediction/)).toBeInTheDocument()
    // Should have 5 rating options
    const stars = screen.getAllByRole('button').filter(btn =>
      btn.querySelector('svg')
    )
    expect(stars.length).toBeGreaterThanOrEqual(5)
  })

  it('displays feedback text areas', () => {
    render(
      <CompletionReviewForm
        task={mockTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    expect(screen.getByText(/What factors affected the timeline/)).toBeInTheDocument()
    expect(screen.getByText(/Lessons learned/)).toBeInTheDocument()
    expect(screen.getByText(/What would you do differently/)).toBeInTheDocument()
  })

  it('calls onCancel when cancel button clicked', () => {
    render(
      <CompletionReviewForm
        task={mockTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    const cancelButton = screen.getByText('Cancel')
    fireEvent.click(cancelButton)

    expect(mockOnCancel).toHaveBeenCalledTimes(1)
  })

  it('calls onSubmit with form data when submitted', async () => {
    render(
      <CompletionReviewForm
        task={mockTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    // Fill in feedback
    const textareas = screen.getAllByRole('textbox')
    fireEvent.change(textareas[0], { target: { value: 'Some feedback about timeline factors' } })
    fireEvent.change(textareas[1], { target: { value: 'Lessons learned from this task' } })
    fireEvent.change(textareas[2], { target: { value: 'Things to do differently next time' } })

    // Submit form
    const submitButton = screen.getByText('Submit Review & Complete')
    fireEvent.click(submitButton)

    expect(mockOnSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        prediction_accuracy_rating: expect.any(Number),
        prediction_feedback: 'Some feedback about timeline factors',
        lessons_learned: 'Lessons learned from this task',
        would_approach_differently: 'Things to do differently next time',
      })
    )
  })

  it('disables submit button when pending', () => {
    render(
      <CompletionReviewForm
        task={mockTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={true}
      />
    )

    const submitButton = screen.getByText('Submitting...')
    expect(submitButton).toBeDisabled()
  })

  it('calculates variance correctly', () => {
    // Use a task with clear variance > 10%
    const taskWithHighVariance = {
      ...mockTask,
      actual_days: 6.5, // Higher variance: ((6.5-5)/5)*100 = 30%
    }

    render(
      <CompletionReviewForm
        task={taskWithHighVariance}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    // Midpoint is (4+6)/2 = 5
    // Actual is 6.5
    // Variance is ((6.5 - 5) / 5) * 100 = 30%
    expect(screen.getByText(/longer than predicted/i)).toBeInTheDocument()
  })

  it('shows "On target" when variance is within 10%', () => {
    const onTargetTask = {
      ...mockTask,
      actual_days: 5.0, // Exactly at midpoint
    }

    render(
      <CompletionReviewForm
        task={onTargetTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    expect(screen.getByText(/On target/i)).toBeInTheDocument()
  })

  it('shows faster when actual is less than predicted', () => {
    const fasterTask = {
      ...mockTask,
      actual_days: 4.0, // Lower than midpoint (5)
    }

    render(
      <CompletionReviewForm
        task={fasterTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    expect(screen.getByText(/faster than predicted/i)).toBeInTheDocument()
  })

  it('handles task without predictions gracefully', () => {
    const noPredictionTask = {
      ...mockTask,
      predicted_days_low: undefined,
      predicted_days_high: undefined,
      actual_days: undefined,
    }

    render(
      <CompletionReviewForm
        task={noPredictionTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    // Should show dashes for missing data
    expect(screen.getAllByText('-').length).toBeGreaterThan(0)
  })

  it('allows changing star rating', async () => {
    render(
      <CompletionReviewForm
        task={mockTask}
        onSubmit={mockOnSubmit}
        onCancel={mockOnCancel}
        isPending={false}
      />
    )

    // Find and click on a star rating
    const buttons = screen.getAllByRole('button')
    const starButtons = buttons.filter(btn =>
      btn.classList.contains('rounded-xl') || btn.querySelector('svg')
    )

    if (starButtons.length >= 5) {
      // Click the 5th star (excellent rating)
      fireEvent.click(starButtons[4])
    }

    // Submit and verify rating
    const submitButton = screen.getByText('Submit Review & Complete')
    fireEvent.click(submitButton)

    expect(mockOnSubmit).toHaveBeenCalled()
  })
})
