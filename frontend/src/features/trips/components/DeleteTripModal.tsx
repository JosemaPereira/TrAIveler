import { useEffect, useRef } from 'react'
import type { KeyboardEvent } from 'react'

import { Button } from '@/components/primitives/Button'
import { useDeleteTrip } from '@/features/trips/hooks/useDeleteTrip'
import styles from './DeleteTripModal.module.css'

export interface DeleteTripModalProps {
  tripId: string
  isOpen: boolean
  onClose: () => void
}

const DIALOG_TITLE = 'Delete this trip?'
const DIALOG_MESSAGE =
  'Are you sure you want to delete this trip? This action cannot be undone.'
const CONFIRM_LABEL = 'Confirm Delete'
const CANCEL_LABEL = 'Cancel'
const TITLE_ID = 'delete-trip-modal-title'

const FOCUSABLE_SELECTOR =
  'button:not(:disabled), [href], input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex]:not([tabindex="-1"])'

/**
 * Confirmation dialog for deleting a trip. Self-contained and accessible
 * (WCAG 2.1 AA) — this codebase has no shared `Modal` composite yet
 * (008-T084), so the dialog role, focus trap, Escape handling, and
 * focus-restore-on-close all live here rather than in a generic wrapper.
 *
 * Owns only the confirm/cancel UI and the delete mutation trigger — trip
 * fetching and any post-delete navigation are out of scope (the latter is
 * already handled inside `useDeleteTrip` itself).
 */
export function DeleteTripModal({
  tripId,
  isOpen,
  onClose,
}: DeleteTripModalProps) {
  const dialogRef = useRef<HTMLDivElement>(null)
  const previouslyFocusedElementRef = useRef<HTMLElement | null>(null)
  const deleteTrip = useDeleteTrip()

  useEffect(() => {
    if (!isOpen) {
      return
    }

    previouslyFocusedElementRef.current =
      document.activeElement as HTMLElement | null
    dialogRef.current?.focus()

    return () => {
      previouslyFocusedElementRef.current?.focus()
    }
  }, [isOpen])

  if (!isOpen) {
    return null
  }

  function handleConfirm() {
    deleteTrip.mutate(tripId)
  }

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      event.stopPropagation()
      onClose()
      return
    }

    if (event.key !== 'Tab' || !dialogRef.current) {
      return
    }

    const focusableElements = Array.from(
      dialogRef.current.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)
    )
    const firstElement = focusableElements[0]
    const lastElement = focusableElements[focusableElements.length - 1]
    if (!firstElement || !lastElement) {
      return
    }

    if (event.shiftKey && document.activeElement === firstElement) {
      event.preventDefault()
      lastElement.focus()
    } else if (!event.shiftKey && document.activeElement === lastElement) {
      event.preventDefault()
      firstElement.focus()
    }
  }

  return (
    <div className={styles.overlay}>
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={TITLE_ID}
        tabIndex={-1}
        className={styles.dialog}
        onKeyDown={handleKeyDown}
      >
        <h2 id={TITLE_ID} className={styles.title}>
          {DIALOG_TITLE}
        </h2>
        <p className={styles.message}>{DIALOG_MESSAGE}</p>
        {deleteTrip.isError && (
          <p role="alert" className={styles.error}>
            Something went wrong deleting this trip. Please try again.
          </p>
        )}
        <div className={styles.actions}>
          <Button
            variant="secondary"
            onClick={onClose}
            disabled={deleteTrip.isPending}
          >
            {CANCEL_LABEL}
          </Button>
          <Button
            variant="danger"
            onClick={handleConfirm}
            disabled={deleteTrip.isPending}
          >
            {CONFIRM_LABEL}
          </Button>
        </div>
      </div>
    </div>
  )
}
