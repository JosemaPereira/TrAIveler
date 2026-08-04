import { describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import type { DeleteTripModalProps } from './DeleteTripModal'
import { DeleteTripModal } from './DeleteTripModal'

function renderModal(overrides: Partial<DeleteTripModalProps> = {}) {
  const onClose = vi.fn()
  const { wrapper } = createQueryWrapper()
  const props: DeleteTripModalProps = {
    tripId: testTrip.id,
    isOpen: true,
    onClose,
    ...overrides,
  }

  render(<DeleteTripModal {...props} />, { wrapper })

  return { onClose }
}

describe('<DeleteTripModal />', () => {
  describe('when closed', () => {
    it('should render nothing', () => {
      renderModal({ isOpen: false })

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })
  })

  describe('when open', () => {
    it('should render an accessible dialog with confirm and cancel actions', () => {
      renderModal()

      const dialog = screen.getByRole('dialog')
      expect(dialog).toBeInTheDocument()
      expect(dialog).toHaveAttribute('aria-modal', 'true')
      expect(
        screen.getByText(/are you sure you want to delete this trip/i)
      ).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: 'Confirm Delete' })
      ).toBeInTheDocument()
      expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
    })

    it('should move focus into the dialog', () => {
      renderModal()

      const dialog = screen.getByRole('dialog')
      expect(dialog).toContainElement(document.activeElement as HTMLElement)
    })

    it('should call onClose without deleting when Cancel is clicked', async () => {
      const user = userEvent.setup()
      const { onClose } = renderModal()

      await user.click(screen.getByRole('button', { name: 'Cancel' }))

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('should call onClose when Escape is pressed', async () => {
      const user = userEvent.setup()
      const { onClose } = renderModal()

      await user.keyboard('{Escape}')

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('should trigger the delete mutation for the given trip id when Confirm is clicked', async () => {
      let deleteRequestReceived = false
      server.use(
        http.delete(`${API_BASE_URL}/trips/${testTrip.id}`, () => {
          deleteRequestReceived = true
          return new HttpResponse(null, { status: 204 })
        })
      )
      const user = userEvent.setup()
      renderModal()

      await user.click(screen.getByRole('button', { name: 'Confirm Delete' }))

      await waitFor(() => {
        expect(deleteRequestReceived).toBe(true)
      })
    })
  })
})
