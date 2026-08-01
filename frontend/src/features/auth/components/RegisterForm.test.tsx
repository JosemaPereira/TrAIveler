import { beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'

import { useAuthStore } from '@/stores/auth-store'
import { API_BASE_URL, testUser } from '@/test/msw/handlers'
import { server } from '@/test/msw/server'
import { createQueryWrapper } from '@/test/queryWrapper'
import { DEMO_PAYMENT_TOKEN, RegisterForm } from './RegisterForm'

const validEmail = 'traveler@example.com'
const validPassword = 'CorrectHorse1!'
const validFullName = 'Ada Traveler'

async function fillValidFields(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByRole('textbox', { name: 'Email' }), validEmail)
  await user.type(screen.getByLabelText('Password'), validPassword)
  await user.type(
    screen.getByRole('textbox', { name: 'Full Name' }),
    validFullName
  )
}

function renderForm(onSuccess?: () => void) {
  const { wrapper } = createQueryWrapper()
  render(<RegisterForm onSuccess={onSuccess} />, { wrapper })
}

beforeEach(() => {
  useAuthStore.setState({
    isAuthenticated: false,
    user: null,
    isLoading: false,
  })
})

describe('<RegisterForm />', () => {
  describe('when rendered', () => {
    it('should render the email, password, and full name fields', () => {
      renderForm()

      expect(screen.getByRole('textbox', { name: 'Email' })).toBeInTheDocument()
      expect(screen.getByLabelText('Password')).toBeInTheDocument()
      expect(
        screen.getByRole('textbox', { name: 'Full Name' })
      ).toBeInTheDocument()
    })

    it('should render the free-account submit button', () => {
      renderForm()

      expect(
        screen.getByRole('button', { name: 'Create Free Account' })
      ).toBeInTheDocument()
    })

    it('should not render the payment button until the demo checkbox is checked', () => {
      renderForm()

      expect(
        screen.queryByRole('button', { name: /continue to payment/i })
      ).not.toBeInTheDocument()
    })
  })

  describe('when the demo payment checkbox is checked', () => {
    it('should reveal a clearly labeled demo payment button', async () => {
      const user = userEvent.setup()
      renderForm()

      await user.click(
        screen.getByRole('checkbox', { name: /add a payment method/i })
      )

      expect(
        screen.getByRole('button', { name: /continue to payment.*demo/i })
      ).toBeInTheDocument()
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
      await user.type(
        screen.getByRole('textbox', { name: 'Full Name' }),
        validFullName
      )
      await user.click(
        screen.getByRole('button', { name: 'Create Free Account' })
      )

      expect(
        await screen.findByText(/enter a valid email address/i)
      ).toBeInTheDocument()
    })
  })

  describe('when submitted with a password that fails the backend rules', () => {
    it('should show the failing password rules inline', async () => {
      const user = userEvent.setup()
      renderForm()

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        validEmail
      )
      await user.type(screen.getByLabelText('Password'), 'weak')
      await user.type(
        screen.getByRole('textbox', { name: 'Full Name' }),
        validFullName
      )
      await user.click(
        screen.getByRole('button', { name: 'Create Free Account' })
      )

      expect(
        await screen.findByText(/must be between 8 and 72 characters/i)
      ).toBeInTheDocument()
    })
  })

  describe('when submitted with an empty full name', () => {
    it('should show a full name required error', async () => {
      const user = userEvent.setup()
      renderForm()

      await user.type(
        screen.getByRole('textbox', { name: 'Email' }),
        validEmail
      )
      await user.type(screen.getByLabelText('Password'), validPassword)
      await user.click(
        screen.getByRole('button', { name: 'Create Free Account' })
      )

      expect(
        await screen.findByText(/full name is required/i)
      ).toBeInTheDocument()
    })
  })

  describe('when submitted via the free-account path with valid data', () => {
    it('should register without a payment_method_token and call onSuccess', async () => {
      let receivedBody: unknown
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, async ({ request }) => {
          receivedBody = await request.json()
          return HttpResponse.json({ user: testUser }, { status: 201 })
        })
      )
      const onSuccess = vi.fn()
      const user = userEvent.setup()
      renderForm(onSuccess)

      await fillValidFields(user)
      await user.click(
        screen.getByRole('button', { name: 'Create Free Account' })
      )

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalledTimes(1)
      })
      expect(receivedBody).toEqual({
        email: validEmail,
        password: validPassword,
        full_name: validFullName,
      })
    })
  })

  describe('when submitted via the demo payment path with valid data', () => {
    it('should register with a fixed demo payment_method_token', async () => {
      let receivedBody: unknown
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, async ({ request }) => {
          receivedBody = await request.json()
          return HttpResponse.json({ user: testUser }, { status: 201 })
        })
      )
      const onSuccess = vi.fn()
      const user = userEvent.setup()
      renderForm(onSuccess)

      await fillValidFields(user)
      await user.click(
        screen.getByRole('checkbox', { name: /add a payment method/i })
      )
      await user.click(
        screen.getByRole('button', { name: /continue to payment/i })
      )

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalledTimes(1)
      })
      expect(receivedBody).toEqual({
        email: validEmail,
        password: validPassword,
        full_name: validFullName,
        payment_method_token: DEMO_PAYMENT_TOKEN,
      })
    })
  })

  describe('when the server rejects a duplicate email with 409', () => {
    it('should show a form-level duplicate-email message', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, () =>
          HttpResponse.json(
            {
              error: 'conflict',
              message: 'Email already registered',
              request_id: 'req_conflict',
            },
            { status: 409 }
          )
        )
      )
      const user = userEvent.setup()
      renderForm()

      await fillValidFields(user)
      await user.click(
        screen.getByRole('button', { name: 'Create Free Account' })
      )

      expect(
        await screen.findByText(/email already registered/i)
      ).toBeInTheDocument()
    })
  })

  describe('when the server rejects with a 422 validation_failed', () => {
    it('should show the per-field server error message', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, () =>
          HttpResponse.json(
            {
              error: 'validation_failed',
              message: 'Validation failed',
              request_id: 'req_validation',
              fields: [{ field: 'email', error: 'Email format is invalid' }],
            },
            { status: 422 }
          )
        )
      )
      const user = userEvent.setup()
      renderForm()

      await fillValidFields(user)
      await user.click(
        screen.getByRole('button', { name: 'Create Free Account' })
      )

      expect(
        await screen.findByText(/email format is invalid/i)
      ).toBeInTheDocument()
    })
  })
})
