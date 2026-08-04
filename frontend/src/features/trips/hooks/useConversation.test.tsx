import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { useConversation } from './useConversation'

function mockSendMessage(
  framePayload: {
    session_id: string
    role: 'assistant'
    message: string
    itinerary_ready: boolean
  },
  status = 200
) {
  server.use(
    http.post(`${API_BASE_URL}/trips/${testTrip.id}/conversation`, () =>
      status === 200
        ? new HttpResponse(`data: ${JSON.stringify(framePayload)}\n\n`, {
            status,
            headers: { 'Content-Type': 'text/event-stream' },
          })
        : HttpResponse.json(
            {
              error: 'validation_failed',
              message: 'Message is required',
              request_id: 'req_conv_422',
            },
            { status }
          )
    )
  )
}

describe('useConversation', () => {
  describe('when initialized', () => {
    it('should start with no messages and idle state', () => {
      const { result } = renderHook(() => useConversation(testTrip.id))

      expect(result.current.messages).toEqual([])
      expect(result.current.isSending).toBe(false)
      expect(result.current.itineraryReady).toBe(false)
      expect(result.current.error).toBeUndefined()
    })
  })

  describe('when sending a message', () => {
    it('should optimistically append the user message before the response resolves', () => {
      mockSendMessage({
        session_id: 's1',
        role: 'assistant',
        message: 'How many days?',
        itinerary_ready: false,
      })
      const { result } = renderHook(() => useConversation(testTrip.id))

      act(() => {
        result.current.sendMessage('I want to go to Tokyo')
      })

      expect(result.current.messages).toEqual([
        expect.objectContaining({
          role: 'user',
          content: 'I want to go to Tokyo',
        }),
      ])
      expect(result.current.isSending).toBe(true)
    })

    describe('when the request succeeds', () => {
      it('should append the assistant reply and clear the sending flag', async () => {
        mockSendMessage({
          session_id: 's1',
          role: 'assistant',
          message: 'How many days?',
          itinerary_ready: false,
        })
        const { result } = renderHook(() => useConversation(testTrip.id))

        act(() => {
          result.current.sendMessage('I want to go to Tokyo')
        })

        await waitFor(() => {
          expect(result.current.isSending).toBe(false)
        })
        expect(result.current.messages).toEqual([
          expect.objectContaining({ role: 'user' }),
          expect.objectContaining({
            role: 'assistant',
            content: 'How many days?',
          }),
        ])
      })

      it('should set itineraryReady to true when the response says so', async () => {
        mockSendMessage({
          session_id: 's1',
          role: 'assistant',
          message: 'Generating your itinerary now.',
          itinerary_ready: true,
        })
        const { result } = renderHook(() => useConversation(testTrip.id))

        act(() => {
          result.current.sendMessage('That sounds perfect')
        })

        await waitFor(() => {
          expect(result.current.itineraryReady).toBe(true)
        })
      })
    })

    describe('when the request fails', () => {
      it('should surface an error while keeping the optimistic user message', async () => {
        mockSendMessage(
          {
            session_id: 's1',
            role: 'assistant',
            message: '',
            itinerary_ready: false,
          },
          422
        )
        const { result } = renderHook(() => useConversation(testTrip.id))

        act(() => {
          result.current.sendMessage('I want to go to Tokyo')
        })

        await waitFor(() => {
          expect(result.current.isSending).toBe(false)
        })
        expect(result.current.error).toBeTruthy()
        expect(result.current.messages).toEqual([
          expect.objectContaining({
            role: 'user',
            content: 'I want to go to Tokyo',
          }),
        ])
      })
    })
  })
})
