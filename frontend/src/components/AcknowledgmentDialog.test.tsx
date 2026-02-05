import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AcknowledgmentDialog from './AcknowledgmentDialog'
import type { Task } from '../types'

const mockTask: Task = {
  id: '550e8400-e29b-41d4-a716-446655440000',
  title: 'Test Task',
  description: 'This is a test task description',
  status: 'pending_acknowledgment',
  created_by: '550e8400-e29b-41d4-a716-446655440001',
  organization_id: '550e8400-e29b-41d4-a716-446655440002',
  created_at: '2024-01-20T10:00:00Z',
  updated_at: '2024-01-20T10:00:00Z',
  estimated_days: 5,
  predicted_days_low: 4.5,
  predicted_days_high: 6.5,
  prediction_confidence: 0.85,
  planning_quality_score: 75,
}

describe('AcknowledgmentDialog', () => {
  const mockOnClose = vi.fn()
  const mockOnAcknowledge = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when not open', () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={false}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    expect(screen.queryByText('Acknowledge Task')).not.toBeInTheDocument()
  })

  it('renders dialog when open', () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    expect(screen.getByText('Acknowledge Task Predictions')).toBeInTheDocument()
  })

  it('displays prediction summary', () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    expect(screen.getByText(/4.5-6.5 days/)).toBeInTheDocument()
    expect(screen.getByText(/85%/)).toBeInTheDocument()
  })

  it('displays three acknowledgment options', () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    expect(screen.getByText('Accept Prediction')).toBeInTheDocument()
    expect(screen.getByText('Modify Estimate')).toBeInTheDocument()
    expect(screen.getByText('Disagree with Prediction')).toBeInTheDocument()
  })

  it('calls onClose when cancel is clicked', async () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    const cancelButton = screen.getByText('Cancel')
    fireEvent.click(cancelButton)

    expect(mockOnClose).toHaveBeenCalledTimes(1)
  })

  it('submits accept action correctly', async () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    // Accept should be selected by default, click acknowledge
    const confirmButton = screen.getByText('Acknowledge & Continue')
    fireEvent.click(confirmButton)

    expect(mockOnAcknowledge).toHaveBeenCalledWith({
      action: 'accept',
    })
  })

  it('shows estimate input when modify is selected', async () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    const modifyButton = screen.getByText('Modify Estimate')
    fireEvent.click(modifyButton)

    expect(screen.getByText(/Your revised estimate/i)).toBeInTheDocument()
  })

  it('shows notes input when disagree is selected', async () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    const disagreeButton = screen.getByText('Disagree with Prediction')
    fireEvent.click(disagreeButton)

    expect(screen.getByText(/Please explain why/i)).toBeInTheDocument()
  })

  it('disables confirm button when pending', () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={true}
      />
    )

    const confirmButton = screen.getByText('Processing...')
    expect(confirmButton).toBeDisabled()
  })

  it('validates modify action requires estimate', async () => {
    // Create task without existing estimate
    const taskNoEstimate = { ...mockTask, estimated_days: undefined }

    render(
      <AcknowledgmentDialog
        task={taskNoEstimate}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    const modifyButton = screen.getByText('Modify Estimate')
    fireEvent.click(modifyButton)

    // Button should be disabled because no estimate provided
    const confirmButton = screen.getByText('Acknowledge & Continue')
    expect(confirmButton).toBeDisabled()
  })

  it('validates disagree action requires notes', async () => {
    render(
      <AcknowledgmentDialog
        task={mockTask}
        isOpen={true}
        onClose={mockOnClose}
        onAcknowledge={mockOnAcknowledge}
        isPending={false}
      />
    )

    const disagreeButton = screen.getByText('Disagree with Prediction')
    fireEvent.click(disagreeButton)

    // Button should be disabled because no notes
    const confirmButton = screen.getByText('Acknowledge & Continue')
    expect(confirmButton).toBeDisabled()
  })
})
