import { useCallback, useState } from 'react'

import type { ConversationMessage } from '@/features/trips/components/ConversationPanel'
import { conversationApi } from '@/features/trips/services/conversationApi'

export interface UseConversationResult {
  messages: ConversationMessage[]
  isSending: boolean
  error?: string
  /** True once a `sendMessage` response reports enough was gathered to generate a full itinerary. */
  itineraryReady: boolean
  sendMessage: (message: string) => void
}

const SEND_FAILED_MESSAGE = 'Failed to send message. Please try again.'

/**
 * Message-list/send/loading/error state for a trip's AI-guided planning
 * conversation. `conversationApi.sendMessage` is a raw promise, not a
 * TanStack Query hook (see that file's doc comment for why), so this hook
 * owns the optimistic-append/loading/error bookkeeping GeneratePage needs
 * around it.
 *
 * A failed send intentionally keeps the optimistic user message in the
 * thread rather than rolling it back: the user did type and submit that
 * text, and silently erasing it would force them to retype it blind while
 * also reading an error. `error` is surfaced separately (e.g. via
 * ConversationPanel's `error` prop) so the caller can show the failure
 * without losing the message that was already sent.
 */
export function useConversation(tripId: string): UseConversationResult {
  const [messages, setMessages] = useState<ConversationMessage[]>([])
  const [isSending, setIsSending] = useState(false)
  const [error, setError] = useState<string>()
  const [itineraryReady, setItineraryReady] = useState(false)

  const sendMessage = useCallback(
    (message: string) => {
      const userMessage: ConversationMessage = {
        id: crypto.randomUUID(),
        role: 'user',
        content: message,
      }
      setMessages((previous) => [...previous, userMessage])
      setError(undefined)
      setIsSending(true)

      conversationApi
        .sendMessage(tripId, { message })
        .then((response) => {
          setMessages((previous) => [
            ...previous,
            {
              id: crypto.randomUUID(),
              role: 'assistant',
              content: response.message,
            },
          ])
          if (response.itinerary_ready) {
            setItineraryReady(true)
          }
        })
        .catch(() => {
          setError(SEND_FAILED_MESSAGE)
        })
        .finally(() => {
          setIsSending(false)
        })
    },
    [tripId]
  )

  return { messages, isSending, error, itineraryReady, sendMessage }
}
