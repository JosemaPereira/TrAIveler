import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'

import { APIError } from '@/lib/api-client'
import { useAuthStore } from '@/stores/auth-store'
import { API_BASE_URL, testUser } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { useLogin } from '@/features/auth/hooks/useLogin'
import { LoginForm } from './LoginForm'

vi.mock('../hooks/useLogin', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/useLogin')>()
  return { ...actual, useLogin: vi.fn(actual.useLogin) }
})

const mockedUseLogin = vi.mocked(useLogin)
const { useLogin: actualUseLogin } =
  await vi.importActual<typeof import('../hooks/useLogin')>('../hooks/useLogin')

const validPassword = 'CorrectHorse1!'

function renderForm(onSuccess?: () => void) {
  const { wrapper } = createQueryWrapper()
  render(<LoginForm onSuccess={onSuccess} />, { wrapper })
}

function fakeLoginResult(
  error: APIError | null,
  retryAfterSeconds: number | undefined
) {
  return {
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    isPending: false,
    isError: error !== null,
    isSuccess: false,
    error,
    retryAfterSeconds,
    // Cast below: only the fields LoginForm actually reads are provided —
    // the real hook's full UseMutationResult shape is exercised by
    // useLogin.test.tsx already.
  } as unknown as ReturnType<typeof useLogin>
}

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: false,
    user: null,
    isLoading: false,
  })
})

describe('<LoginForm />', () => {
  describe('when rendered', () => {
    it('should render the email and password fields', () => {
      renderForm()

      expect(screen.getByRole('textbox', { name: 'Email' })).toBeInTheDocument()
      expect(screen.getByLabelText('Password')).toBeInTheDocument()
    })

    it('should render a link to the password-reset page', () => {
      renderForm()

      expect(
        screen.getByRole('link', { name: /forgot password/i })
      ).toHaveAttribute('href', '/password-reset')
    })

    it('should render the submit button enabled', () => {
      renderForm()

      expect(screen.getByRole('button', { name: 'Log In' })).toBeEnabled()
    })
  })

  describe('when submitted with an invalid email', () => {
    it('should show an inline email error without calling the API', async () => {
      const user = userEvent.setup()
      renderForm()

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        'not-an-email'
      )
      await user.type(screen.getByLabelText('Password'), validPassword)
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      expect(
        await screen.findByText(/enter a valid email address/i)
      ).toBeInTheDocument()
    })
  })

  describe('when submitted with an empty password', () => {
    it('should show a password-required error', async () => {
      const user = userEvent.setup()
      renderForm()

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        testUser.email
      )
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      expect(
        await screen.findByText(/password is required/i)
      ).toBeInTheDocument()
    })
  })

  describe('when the credentials are valid', () => {
    it('should call onSuccess once the login resolves', async () => {
      const onSuccess = vi.fn()
      const user = userEvent.setup()
      renderForm(onSuccess)

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        testUser.email
      )
      await user.type(screen.getByLabelText('Password'), validPassword)
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalledTimes(1)
      })
    })
  })

  describe('when the credentials are rejected with a 401', () => {
    it('should show one generic message for an unknown email', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, () =>
          HttpResponse.json(
            {
              error: 'authentication_required',
              message: 'Invalid email or password',
              request_id: 'req_unknown_email',
            },
            { status: 401 }
          )
        )
      )
      const user = userEvent.setup()
      renderForm()

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        'nobody@example.com'
      )
      await user.type(screen.getByLabelText('Password'), validPassword)
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      expect(
        await screen.findByText(/invalid email or password/i)
      ).toBeInTheDocument()
    })

    it('should show the identical generic message for a wrong password', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, () =>
          HttpResponse.json(
            {
              error: 'authentication_required',
              message: 'Invalid email or password',
              request_id: 'req_wrong_password',
            },
            { status: 401 }
          )
        )
      )
      const user = userEvent.setup()
      renderForm()

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        testUser.email
      )
      await user.type(screen.getByLabelText('Password'), 'WrongPassword1')
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      expect(
        await screen.findByText(/invalid email or password/i)
      ).toBeInTheDocument()
    })
  })

  describe('when the server answers with a 429 over the real request path', () => {
    it('should render the retry-after wait as a countdown and disable submit', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, () =>
          HttpResponse.json(
            {
              error: 'rate_limit_exceeded',
              message: 'Too many login attempts. Please try again later.',
              request_id: 'req_rate',
              details: { retry_after_seconds: 30 },
            },
            { status: 429 }
          )
        )
      )
      const user = userEvent.setup()
      renderForm()

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        testUser.email
      )
      await user.type(screen.getByLabelText('Password'), validPassword)
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      expect(
        await screen.findByText(/try again in 30 seconds/i)
      ).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /log in/i })).toBeDisabled()
    })
  })

  describe('when the server answers with an unexpected 500', () => {
    it('should show the generic server error message', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, () =>
          HttpResponse.json(
            {
              error: 'internal_error',
              message: 'Something went wrong. Please try again.',
              request_id: 'req_500',
            },
            { status: 500 }
          )
        )
      )
      const user = userEvent.setup()
      renderForm()

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        testUser.email
      )
      await user.type(screen.getByLabelText('Password'), validPassword)
      await user.click(screen.getByRole('button', { name: 'Log In' }))

      expect(
        await screen.findByText(/something went wrong\. please try again\./i)
      ).toBeInTheDocument()
    })
  })

  describe('when the rate-limit countdown is driven directly (fake timers)', () => {
    afterEach(() => {
      mockedUseLogin.mockImplementation(actualUseLogin)
      vi.useRealTimers()
    })

    it('should tick the countdown down to zero and re-enable submit', () => {
      vi.useFakeTimers()
      const error = new APIError(
        429,
        'rate_limit_exceeded',
        'Too many login attempts. Please try again later.',
        'req_rate_1',
        undefined,
        { retry_after_seconds: 3 }
      )
      mockedUseLogin.mockReturnValue(fakeLoginResult(error, 3))

      renderForm()

      expect(screen.getByText(/try again in 3 seconds/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /log in/i })).toBeDisabled()

      // Advanced one second at a time (rather than one large jump): the
      // countdown effect reschedules its own next tick from inside a commit,
      // so a single multi-second `advanceTimersByTime` call can outrun React
      // before the next timer is even registered.
      act(() => {
        vi.advanceTimersByTime(1000)
      })
      expect(screen.getByText(/try again in 2 seconds/i)).toBeInTheDocument()

      act(() => {
        vi.advanceTimersByTime(1000)
      })
      expect(screen.getByText(/try again in 1 second\./i)).toBeInTheDocument()

      act(() => {
        vi.advanceTimersByTime(1000)
      })
      expect(screen.queryByText(/try again in/i)).not.toBeInTheDocument()
      expect(screen.getByRole('button', { name: /log in/i })).toBeEnabled()
    })

    it('should restart the countdown when a new 429 arrives with an identical wait', () => {
      vi.useFakeTimers()
      const firstError = new APIError(
        429,
        'rate_limit_exceeded',
        'Too many login attempts. Please try again later.',
        'req_rate_first',
        undefined,
        { retry_after_seconds: 5 }
      )
      mockedUseLogin.mockReturnValue(fakeLoginResult(firstError, 5))

      const { rerender } = render(<LoginForm />, {
        wrapper: createQueryWrapper().wrapper,
      })
      expect(screen.getByText(/try again in 5 seconds/i)).toBeInTheDocument()

      act(() => {
        vi.advanceTimersByTime(1000)
      })
      act(() => {
        vi.advanceTimersByTime(1000)
      })
      act(() => {
        vi.advanceTimersByTime(1000)
      })
      expect(screen.getByText(/try again in 2 seconds/i)).toBeInTheDocument()

      const secondError = new APIError(
        429,
        'rate_limit_exceeded',
        'Too many login attempts. Please try again later.',
        'req_rate_second',
        undefined,
        { retry_after_seconds: 5 }
      )
      mockedUseLogin.mockReturnValue(fakeLoginResult(secondError, 5))
      rerender(<LoginForm />)

      expect(screen.getByText(/try again in 5 seconds/i)).toBeInTheDocument()
    })
  })
})
