import { describe, expect, it } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'

import type { SendMessageResponse } from '@/features/trips/types'
import { API_BASE_URL, testItineraryDay, testTrip } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { GeneratePage } from './GeneratePage'

type UserEvent = ReturnType<typeof userEvent.setup>

function renderGeneratePage() {
  const { wrapper } = createQueryWrapper()
  return render(<GeneratePage />, { wrapper })
}

function mockCreateTrip(status: 201 | 422 = 201) {
  server.use(
    http.post(`${API_BASE_URL}/trips`, () =>
      status === 201
        ? HttpResponse.json({ trip: testTrip }, { status })
        : HttpResponse.json(
            {
              error: 'validation_failed',
              message: 'Title is required',
              request_id: 'req_trip_422',
            },
            { status }
          )
    )
  )
}

function mockGetItinerary(days: (typeof testItineraryDay)[] = []) {
  server.use(
    http.get(`${API_BASE_URL}/trips/${testTrip.id}/itinerary`, () =>
      HttpResponse.json({ trip_id: testTrip.id, days }, { status: 200 })
    )
  )
}

function mockSendMessage(payload: SendMessageResponse, status = 200) {
  server.use(
    http.post(`${API_BASE_URL}/trips/${testTrip.id}/conversation`, () =>
      status === 200
        ? new HttpResponse(`data: ${JSON.stringify(payload)}\n\n`, {
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

async function createTripAndReachConversation(user: UserEvent) {
  mockCreateTrip()

  await user.type(screen.getByLabelText(/trip title/i), testTrip.title)
  await user.click(screen.getByRole('button', { name: /start planning/i }))

  await waitFor(() => {
    expect(screen.getByRole('log')).toBeInTheDocument()
  })
}

describe('<GeneratePage />', () => {
  describe('when first rendered', () => {
    it('should render the new-trip form', () => {
      renderGeneratePage()

      expect(screen.getByLabelText(/trip title/i)).toBeInTheDocument()
      expect(
        screen.getByRole('button', { name: /start planning/i })
      ).toBeInTheDocument()
    })
  })

  describe('when the trip is created successfully', () => {
    it('should hide the form and show the conversation panel', async () => {
      const user = userEvent.setup()
      renderGeneratePage()

      await createTripAndReachConversation(user)

      expect(
        screen.queryByRole('button', { name: /start planning/i })
      ).not.toBeInTheDocument()
    })
  })

  describe('when trip creation fails', () => {
    it('should keep the form visible and show an error', async () => {
      mockCreateTrip(422)
      const user = userEvent.setup()
      renderGeneratePage()

      await user.type(screen.getByLabelText(/trip title/i), testTrip.title)
      await user.click(screen.getByRole('button', { name: /start planning/i }))

      await waitFor(() => {
        expect(screen.getByRole('alert')).toBeInTheDocument()
      })
      expect(
        screen.getByRole('button', { name: /start planning/i })
      ).toBeInTheDocument()
    })
  })

  describe('once a conversation is active', () => {
    it('should append the sent message and the assistant reply to the thread', async () => {
      const user = userEvent.setup()
      renderGeneratePage()
      await createTripAndReachConversation(user)
      mockSendMessage({
        session_id: 's1',
        role: 'assistant',
        message: 'How many days?',
        itinerary_ready: false,
      })

      await user.type(screen.getByLabelText('Message'), 'Two weeks in Japan')
      await user.click(screen.getByRole('button', { name: 'Send' }))

      await waitFor(() => {
        expect(screen.getByText('How many days?')).toBeInTheDocument()
      })
      expect(screen.getByText('Two weeks in Japan')).toBeInTheDocument()
    })

    it('should surface a send failure via the conversation panel error', async () => {
      const user = userEvent.setup()
      renderGeneratePage()
      await createTripAndReachConversation(user)
      mockSendMessage(
        {
          session_id: 's1',
          role: 'assistant',
          message: '',
          itinerary_ready: false,
        },
        422
      )

      await user.type(screen.getByLabelText('Message'), 'Two weeks in Japan')
      await user.click(screen.getByRole('button', { name: 'Send' }))

      await waitFor(() => {
        expect(screen.getByRole('alert')).toBeInTheDocument()
      })
    })

    it('should switch to the itinerary view once itinerary_ready is true', async () => {
      const user = userEvent.setup()
      renderGeneratePage()
      await createTripAndReachConversation(user)
      mockGetItinerary()
      mockSendMessage({
        session_id: 's1',
        role: 'assistant',
        message: 'Generating your itinerary now.',
        itinerary_ready: true,
      })

      await user.type(screen.getByLabelText('Message'), 'That sounds perfect')
      await user.click(screen.getByRole('button', { name: 'Send' }))

      await waitFor(() => {
        expect(
          screen.getByRole('heading', { name: /itinerary not yet generated/i })
        ).toBeInTheDocument()
      })
      expect(screen.queryByRole('log')).not.toBeInTheDocument()
    })

    it('should render the generated itinerary days once they are available', async () => {
      const user = userEvent.setup()
      renderGeneratePage()
      await createTripAndReachConversation(user)
      mockGetItinerary([testItineraryDay])
      mockSendMessage({
        session_id: 's1',
        role: 'assistant',
        message: 'Generating your itinerary now.',
        itinerary_ready: true,
      })

      await user.type(screen.getByLabelText('Message'), 'That sounds perfect')
      await user.click(screen.getByRole('button', { name: 'Send' }))

      await waitFor(() => {
        expect(screen.getByText('Fushimi Inari Shrine')).toBeInTheDocument()
      })
    })
  })
})
