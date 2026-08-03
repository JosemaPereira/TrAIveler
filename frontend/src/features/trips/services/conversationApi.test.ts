import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import { APIError } from '@/lib/api-client'
import { API_BASE_URL, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { conversationApi } from './conversationApi'

const sessionId = '44444444-4444-4444-4444-444444444444'

describe('conversationApi.getHistory', () => {
  describe('when the request succeeds', () => {
    it('should GET /trips/:id/conversation and resolve with the session and messages', async () => {
      let method: string | undefined
      const responseBody = {
        session_id: sessionId,
        status: 'in_progress',
        messages: [
          {
            id: 'msg_1',
            session_id: sessionId,
            role: 'assistant',
            content: 'Where would you like to go?',
            token_count: 6,
            timestamp: '2026-01-01T00:00:00Z',
          },
        ],
      }
      server.use(
        http.get(
          `${API_BASE_URL}/trips/${testTrip.id}/conversation`,
          ({ request }) => {
            method = request.method
            return HttpResponse.json(responseBody, { status: 200 })
          }
        )
      )

      const result = await conversationApi.getHistory(testTrip.id)

      expect(method).toBe('GET')
      expect(result).toEqual(responseBody)
    })

    it('should send the session cookie along with the request', async () => {
      let credentials: RequestCredentials | undefined
      server.use(
        http.get(
          `${API_BASE_URL}/trips/${testTrip.id}/conversation`,
          ({ request }) => {
            credentials = request.credentials
            return HttpResponse.json(
              { session_id: sessionId, status: 'in_progress', messages: [] },
              { status: 200 }
            )
          }
        )
      )

      await conversationApi.getHistory(testTrip.id)

      expect(credentials).toBe('include')
    })
  })

  describe('when the trip does not exist or is not owned by the caller', () => {
    it('should reject with a 404 APIError', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips/not-mine/conversation`, () =>
          HttpResponse.json(
            {
              error: 'not_found',
              message: 'Trip not found',
              request_id: 'req_conv_404',
            },
            { status: 404 }
          )
        )
      )

      await expect(
        conversationApi.getHistory('not-mine')
      ).rejects.toMatchObject({ status: 404, code: 'not_found' })
    })
  })
})

describe('conversationApi.sendMessage', () => {
  describe('when the request succeeds', () => {
    it('should POST the message and parse the single SSE data frame', async () => {
      let receivedBody: unknown
      const framePayload = {
        session_id: sessionId,
        role: 'assistant',
        message: 'How many days are you planning to stay?',
        itinerary_ready: false,
      }
      server.use(
        http.post(
          `${API_BASE_URL}/trips/${testTrip.id}/conversation`,
          async ({ request }) => {
            receivedBody = await request.json()
            return new HttpResponse(
              `data: ${JSON.stringify(framePayload)}\n\n`,
              {
                status: 200,
                headers: { 'Content-Type': 'text/event-stream' },
              }
            )
          }
        )
      )

      const result = await conversationApi.sendMessage(testTrip.id, {
        message: 'I want to visit Japan for two weeks',
      })

      expect(receivedBody).toEqual({
        message: 'I want to visit Japan for two weeks',
      })
      expect(result).toEqual(framePayload)
    })

    it('should send the session cookie along with the request', async () => {
      let credentials: RequestCredentials | undefined
      server.use(
        http.post(
          `${API_BASE_URL}/trips/${testTrip.id}/conversation`,
          ({ request }) => {
            credentials = request.credentials
            return new HttpResponse(
              `data: ${JSON.stringify({
                session_id: sessionId,
                role: 'assistant',
                message: 'Got it.',
                itinerary_ready: false,
              })}\n\n`,
              { status: 200, headers: { 'Content-Type': 'text/event-stream' } }
            )
          }
        )
      )

      await conversationApi.sendMessage(testTrip.id, { message: 'Hi' })

      expect(credentials).toBe('include')
    })
  })

  describe('when the itinerary is ready', () => {
    it('should resolve with itinerary_ready set to true', async () => {
      server.use(
        http.post(
          `${API_BASE_URL}/trips/${testTrip.id}/conversation`,
          () =>
            new HttpResponse(
              `data: ${JSON.stringify({
                session_id: sessionId,
                role: 'assistant',
                message: 'Generating your itinerary now.',
                itinerary_ready: true,
              })}\n\n`,
              { status: 200, headers: { 'Content-Type': 'text/event-stream' } }
            )
        )
      )

      const result = await conversationApi.sendMessage(testTrip.id, {
        message: 'That sounds perfect',
      })

      expect(result.itinerary_ready).toBe(true)
    })
  })

  describe('when the request fails', () => {
    it('should reject with an APIError built from the JSON error envelope', async () => {
      server.use(
        http.post(`${API_BASE_URL}/trips/${testTrip.id}/conversation`, () =>
          HttpResponse.json(
            {
              error: 'validation_failed',
              message: 'Message is required',
              request_id: 'req_conv_422',
            },
            { status: 422 }
          )
        )
      )

      const error = await conversationApi
        .sendMessage(testTrip.id, { message: '' })
        .catch((caught: unknown) => caught)

      expect(error).toBeInstanceOf(APIError)
      expect(error).toMatchObject({
        status: 422,
        code: 'validation_failed',
        message: 'Message is required',
      })
    })
  })
})
