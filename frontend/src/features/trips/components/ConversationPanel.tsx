import { useState } from 'react'
import type { ChangeEvent, SyntheticEvent } from 'react'

import { Button } from '@/components/primitives/Button'
import { Input } from '@/components/primitives/Input'
import { LoadingSpinner } from '@/components/primitives/LoadingSpinner'
import styles from './ConversationPanel.module.css'

export interface ConversationMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
}

export interface ConversationPanelProps {
  messages: ConversationMessage[]
  onSendMessage: (message: string) => void
  isSending?: boolean
  streamingMessage?: string
  error?: string
}

// Visually-hidden turn labels, so the user/assistant distinction survives
// without color (docs/ui-guidelines.md accessibility rules).
const ROLE_LABELS: Record<ConversationMessage['role'], string> = {
  user: 'You:',
  assistant: 'Assistant:',
}

const TYPING_LABEL = 'Assistant is typing…'
const SEND_LABEL = 'Send'
const MESSAGE_INPUT_LABEL = 'Message'

/**
 * Conversation panel for the AI-assisted trip planning chat. Purely
 * presentational — messages, streaming text, and send state are all driven
 * by props; this component owns only the controlled draft-input state.
 *
 * The thread is a role="log" live region so newly appended messages
 * (including in-progress streaming content) are announced to assistive
 * technology, mirroring LoadingSpinner's aria-live pattern.
 */
export function ConversationPanel({
  messages,
  onSendMessage,
  isSending = false,
  streamingMessage,
  error,
}: ConversationPanelProps) {
  const [draft, setDraft] = useState('')

  const trimmedDraft = draft.trim()
  const isSendDisabled = isSending || trimmedDraft.length === 0
  const showTypingIndicator = isSending && !streamingMessage

  function handleSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault()
    if (trimmedDraft.length === 0 || isSending) {
      return
    }
    onSendMessage(trimmedDraft)
    setDraft('')
  }

  return (
    <div className={styles.panel}>
      <div className={styles.thread} role="log" aria-live="polite">
        {messages.map((message) => (
          <div
            key={message.id}
            className={[styles.bubble, styles[message.role]].join(' ')}
          >
            <span className={styles.srOnly}>{ROLE_LABELS[message.role]}</span>
            <p className={styles.content}>{message.content}</p>
          </div>
        ))}
        {streamingMessage && (
          <div className={[styles.bubble, styles.assistant].join(' ')}>
            <span className={styles.srOnly}>{ROLE_LABELS.assistant}</span>
            <p className={styles.content}>{streamingMessage}</p>
          </div>
        )}
        {showTypingIndicator && (
          <div className={[styles.bubble, styles.assistant].join(' ')}>
            <LoadingSpinner size="sm" label={TYPING_LABEL} />
          </div>
        )}
      </div>

      {error && (
        <p role="alert" className={styles.error}>
          {error}
        </p>
      )}

      <form className={styles.form} onSubmit={handleSubmit}>
        <Input
          label={MESSAGE_INPUT_LABEL}
          value={draft}
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            setDraft(event.target.value)
          }}
          disabled={isSending}
        />
        <Button type="submit" disabled={isSendDisabled}>
          {SEND_LABEL}
        </Button>
      </form>
    </div>
  )
}
