import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import type { ConversationPanelProps } from './ConversationPanel'
import { ConversationPanel } from './ConversationPanel'

const messages: ConversationPanelProps['messages'] = [
  { id: 'msg-1', role: 'user', content: 'Plan a trip to Kyoto.' },
  { id: 'msg-2', role: 'assistant', content: 'Sure! How many days?' },
]

function renderPanel(overrides: Partial<ConversationPanelProps> = {}) {
  const props: ConversationPanelProps = {
    messages,
    onSendMessage: vi.fn(),
    ...overrides,
  }
  return render(<ConversationPanel {...props} />)
}

describe('<ConversationPanel />', () => {
  describe('when rendered with a message history', () => {
    it('should render each message content inside a live region', () => {
      renderPanel()

      const log = screen.getByRole('log')
      expect(log).toHaveTextContent('Plan a trip to Kyoto.')
      expect(log).toHaveTextContent('Sure! How many days?')
    })

    it('should distinguish user and assistant turns with a text label, not color alone', () => {
      renderPanel()

      expect(screen.getByText('You:')).toBeInTheDocument()
      expect(screen.getByText('Assistant:')).toBeInTheDocument()
    })

    it('should render the message input and send button', () => {
      renderPanel()

      expect(screen.getByRole('textbox', { name: /message/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /send/i })).toBeInTheDocument()
    })
  })

  describe('when the send button is disabled', () => {
    it('should disable it while the draft message is empty', () => {
      renderPanel()

      expect(screen.getByRole('button', { name: /send/i })).toBeDisabled()
    })

    it('should enable it once a message is typed', async () => {
      const user = userEvent.setup()
      renderPanel()

      await user.type(
        screen.getByRole('textbox', { name: /message/i }),
        'Hello'
      )

      expect(screen.getByRole('button', { name: /send/i })).toBeEnabled()
    })
  })

  describe('when the user submits a message', () => {
    it('should call onSendMessage with the trimmed text and clear the field', async () => {
      const onSendMessage = vi.fn()
      const user = userEvent.setup()
      renderPanel({ onSendMessage })

      const input = screen.getByRole('textbox', { name: /message/i })
      await user.type(input, '  Book a ryokan  ')
      await user.click(screen.getByRole('button', { name: /send/i }))

      expect(onSendMessage).toHaveBeenCalledWith('Book a ryokan')
      expect(input).toHaveValue('')
    })
  })

  describe('when a submit event reaches the form with nothing to send', () => {
    it('should not call onSendMessage for an empty draft', () => {
      const onSendMessage = vi.fn()
      renderPanel({ onSendMessage })

      const form = screen
        .getByRole('textbox', { name: /message/i })
        .closest('form')
      expect(form).not.toBeNull()
      fireEvent.submit(form as HTMLFormElement)

      expect(onSendMessage).not.toHaveBeenCalled()
    })

    it('should not call onSendMessage while a send is already in flight', () => {
      const onSendMessage = vi.fn()
      renderPanel({ onSendMessage, isSending: true })

      const form = screen
        .getByRole('textbox', { name: /message/i })
        .closest('form')
      expect(form).not.toBeNull()
      fireEvent.submit(form as HTMLFormElement)

      expect(onSendMessage).not.toHaveBeenCalled()
    })
  })

  describe('when isSending is true', () => {
    it('should disable the input and send button', () => {
      renderPanel({ isSending: true })

      expect(screen.getByRole('textbox', { name: /message/i })).toBeDisabled()
      expect(screen.getByRole('button', { name: /send/i })).toBeDisabled()
    })

    it('should show a loading indicator when there is no streaming content yet', () => {
      renderPanel({ isSending: true })

      expect(screen.getByRole('status')).toBeInTheDocument()
    })

    it('should not show the loading indicator once streaming content has arrived', () => {
      renderPanel({ isSending: true, streamingMessage: 'Working on it' })

      expect(screen.queryByRole('status')).not.toBeInTheDocument()
    })
  })

  describe('when streamingMessage is present', () => {
    it('should append it as an in-progress assistant bubble after the messages', () => {
      renderPanel({ streamingMessage: 'Here is a first draft itinerary...' })

      expect(
        screen.getByText('Here is a first draft itinerary...')
      ).toBeInTheDocument()
    })
  })

  describe('when error is set', () => {
    it('should render an inline error message', () => {
      renderPanel({ error: 'Failed to send message. Please try again.' })

      expect(
        screen.getByText('Failed to send message. Please try again.')
      ).toBeInTheDocument()
    })
  })
})
