import { beforeEach, describe, expect, it } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { http, HttpResponse } from 'msw'

import { useAuthStore } from '@/stores/auth-store'
import { API_BASE_URL, testTrip, testUser } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { TripDetailPage } from './TripDetailPage'

function renderAt(path: string) {
  const memoryRouter = createMemoryRouter(
    [{ path: '/trips/:id', element: <TripDetailPage /> }],
    { initialEntries: [path] }
  )
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={memoryRouter} />
    </QueryClientProvider>
  )
}

function mockGetTrip(trip = testTrip) {
  server.use(
    http.get(`${API_BASE_URL}/trips/${trip.id}`, () =>
      HttpResponse.json({ trip }, { status: 200 })
    )
  )
}

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: true,
    user: testUser,
    isLoading: false,
  })
})

describe('<TripDetailPage />', () => {
  describe('while the trip is loading', () => {
    it('should render a loading spinner', () => {
      mockGetTrip()

      renderAt(`/trips/${testTrip.id}`)

      expect(screen.getByRole('status')).toBeInTheDocument()
    })
  })

  describe('when fetching the trip fails', () => {
    it('should render an error message', async () => {
      server.use(
        http.get(`${API_BASE_URL}/trips/${testTrip.id}`, () =>
          HttpResponse.json(
            {
              error: 'not_found',
              message: 'Trip not found',
              request_id: 'req_trip_404',
            },
            { status: 404 }
          )
        )
      )

      renderAt(`/trips/${testTrip.id}`)

      await waitFor(() => {
        expect(screen.getByRole('alert')).toBeInTheDocument()
      })
      expect(screen.getByText('Trip not found')).toBeInTheDocument()
    })
  })

  describe('when the trip loads successfully', () => {
    it('should render the Navigation and the trip title', async () => {
      mockGetTrip()

      renderAt(`/trips/${testTrip.id}`)

      await waitFor(() => {
        expect(
          screen.getByRole('heading', { name: testTrip.title })
        ).toBeInTheDocument()
      })
      expect(screen.getByRole('navigation')).toBeInTheDocument()
    })

    describe('when the signed-in user is the creator with an active subscription', () => {
      it('should render the Edit Trip and Delete Trip actions', async () => {
        useAuthStore.setState({
          user: { ...testUser, has_subscription: true },
        })
        mockGetTrip()

        renderAt(`/trips/${testTrip.id}`)

        await waitFor(() => {
          expect(
            screen.getByRole('button', { name: 'Edit Trip' })
          ).toBeInTheDocument()
        })
        expect(
          screen.getByRole('button', { name: 'Delete Trip' })
        ).toBeInTheDocument()
      })
    })

    describe('when the signed-in user is not the creator', () => {
      it('should hide the action bar', async () => {
        useAuthStore.setState({
          user: { ...testUser, id: 'someone-else', has_subscription: true },
        })
        mockGetTrip()

        renderAt(`/trips/${testTrip.id}`)

        await waitFor(() => {
          expect(
            screen.getByRole('heading', { name: testTrip.title })
          ).toBeInTheDocument()
        })
        expect(
          screen.queryByRole('button', { name: 'Edit Trip' })
        ).not.toBeInTheDocument()
      })
    })

    describe('when the signed-in creator has no active subscription', () => {
      it('should hide the action bar', async () => {
        useAuthStore.setState({
          user: { ...testUser, has_subscription: false },
        })
        mockGetTrip()

        renderAt(`/trips/${testTrip.id}`)

        await waitFor(() => {
          expect(
            screen.getByRole('heading', { name: testTrip.title })
          ).toBeInTheDocument()
        })
        expect(
          screen.queryByRole('button', { name: 'Edit Trip' })
        ).not.toBeInTheDocument()
      })
    })

    describe('when the creator saves an edit', () => {
      it('should PUT the updated trip with the If-Match version header', async () => {
        useAuthStore.setState({
          user: { ...testUser, has_subscription: true },
        })
        mockGetTrip()
        let receivedBody: unknown
        let receivedIfMatch: string | null = null
        server.use(
          http.put(
            `${API_BASE_URL}/trips/${testTrip.id}`,
            async ({ request }) => {
              receivedBody = await request.json()
              receivedIfMatch = request.headers.get('If-Match')
              return HttpResponse.json(
                { trip: { ...testTrip, title: 'Updated title' } },
                { status: 200 }
              )
            }
          )
        )
        const user = userEvent.setup()
        renderAt(`/trips/${testTrip.id}`)
        await waitFor(() => {
          expect(
            screen.getByRole('button', { name: 'Edit Trip' })
          ).toBeInTheDocument()
        })

        await user.click(screen.getByRole('button', { name: 'Edit Trip' }))
        const titleInput = screen.getByLabelText('Title')
        await user.clear(titleInput)
        await user.type(titleInput, 'Updated title')
        await user.click(screen.getByRole('button', { name: 'Save' }))

        await waitFor(() => {
          expect(receivedBody).toMatchObject({
            title: 'Updated title',
            status: testTrip.status,
          })
        })
        expect(receivedIfMatch).toBe(String(testTrip.version))
      })
    })

    describe('when the creator clicks Delete Trip', () => {
      it('should open the delete confirmation modal', async () => {
        useAuthStore.setState({
          user: { ...testUser, has_subscription: true },
        })
        mockGetTrip()
        const user = userEvent.setup()
        renderAt(`/trips/${testTrip.id}`)
        await waitFor(() => {
          expect(
            screen.getByRole('button', { name: 'Delete Trip' })
          ).toBeInTheDocument()
        })

        await user.click(screen.getByRole('button', { name: 'Delete Trip' }))

        expect(screen.getByRole('dialog')).toBeInTheDocument()
      })
    })
  })
})
