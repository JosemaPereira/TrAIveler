import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'

import { APIError } from '../../../lib/api-client'
import {
  API_BASE_URL,
  testSubscription,
  testUser,
} from '../../../test/msw/handlers'
import { server } from '../../../test/msw/server'
import { authApi } from './authApi'

describe('authApi.register', () => {
  describe('when the registration succeeds', () => {
    it('should send the registration payload to POST /auth/register', async () => {
      let receivedBody: unknown
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, async ({ request }) => {
          receivedBody = await request.json()
          return HttpResponse.json({ user: testUser }, { status: 201 })
        })
      )

      await authApi.register({
        email: 'traveler@example.com',
        password: 'CorrectHorse1!',
        full_name: 'Ada Traveler',
      })

      expect(receivedBody).toEqual({
        email: 'traveler@example.com',
        password: 'CorrectHorse1!',
        full_name: 'Ada Traveler',
      })
    })

    it('should send the session cookie along with the request', async () => {
      let credentials: RequestCredentials | undefined
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, ({ request }) => {
          credentials = request.credentials
          return HttpResponse.json({ user: testUser }, { status: 201 })
        })
      )

      await authApi.register({
        email: 'traveler@example.com',
        password: 'CorrectHorse1!',
        full_name: 'Ada Traveler',
      })

      expect(credentials).toBe('include')
    })

    it('should resolve with the created user for a free account', async () => {
      const result = await authApi.register({
        email: 'traveler@example.com',
        password: 'CorrectHorse1!',
        full_name: 'Ada Traveler',
      })

      expect(result.user).toEqual(testUser)
      expect(result.subscription).toBeUndefined()
    })

    it('should resolve with the subscription when a payment token is supplied', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, () =>
          HttpResponse.json(
            {
              user: { ...testUser, has_subscription: true },
              subscription: testSubscription,
            },
            { status: 201 }
          )
        )
      )

      const result = await authApi.register({
        email: 'traveler@example.com',
        password: 'CorrectHorse1!',
        full_name: 'Ada Traveler',
        payment_method_token: 'tok_visa_demo',
      })

      expect(result.user.has_subscription).toBe(true)
      expect(result.subscription).toEqual(testSubscription)
    })
  })

  describe('when the email is already registered', () => {
    it('should reject with a 409 conflict APIError', async () => {
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

      await expect(
        authApi.register({
          email: 'taken@example.com',
          password: 'CorrectHorse1!',
          full_name: 'Ada Traveler',
        })
      ).rejects.toMatchObject({
        status: 409,
        code: 'conflict',
        message: 'Email already registered',
      })
    })
  })

  describe('when the payload fails validation', () => {
    it('should reject with a 422 APIError carrying per-field errors', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/register`, () =>
          HttpResponse.json(
            {
              error: 'validation_failed',
              message: 'One or more fields failed validation',
              request_id: 'req_validation',
              fields: [{ field: 'password', error: 'Password is too short' }],
            },
            { status: 422 }
          )
        )
      )

      await expect(
        authApi.register({
          email: 'traveler@example.com',
          password: 'short',
          full_name: 'Ada Traveler',
        })
      ).rejects.toMatchObject({
        status: 422,
        code: 'validation_failed',
        fields: [{ field: 'password', error: 'Password is too short' }],
      })
    })
  })
})

describe('authApi.login', () => {
  describe('when the credentials are valid', () => {
    it('should send the credentials to POST /auth/login', async () => {
      let receivedBody: unknown
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, async ({ request }) => {
          receivedBody = await request.json()
          return HttpResponse.json({ user: testUser }, { status: 200 })
        })
      )

      await authApi.login({
        email: 'traveler@example.com',
        password: 'CorrectHorse1!',
      })

      expect(receivedBody).toEqual({
        email: 'traveler@example.com',
        password: 'CorrectHorse1!',
      })
    })

    it('should resolve with the authenticated user', async () => {
      const result = await authApi.login({
        email: 'traveler@example.com',
        password: 'CorrectHorse1!',
      })

      expect(result.user).toEqual(testUser)
    })
  })

  describe('when the credentials are rejected', () => {
    it('should reject with a 401 authentication_required APIError', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, () =>
          HttpResponse.json(
            {
              error: 'authentication_required',
              message: 'Invalid email or password',
              request_id: 'req_bad_creds',
            },
            { status: 401 }
          )
        )
      )

      await expect(
        authApi.login({ email: 'traveler@example.com', password: 'wrong' })
      ).rejects.toMatchObject({
        status: 401,
        code: 'authentication_required',
      })
    })
  })

  describe('when the caller is rate limited', () => {
    it('should reject with a 429 APIError exposing retry_after_seconds', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/login`, () =>
          HttpResponse.json(
            {
              error: 'rate_limit_exceeded',
              message: 'Too many login attempts. Please try again later.',
              request_id: 'req_rate',
              details: { retry_after_seconds: 45 },
            },
            { status: 429 }
          )
        )
      )

      const error = await authApi
        .login({ email: 'traveler@example.com', password: 'wrong' })
        .catch((caught: unknown) => caught)

      expect(error).toBeInstanceOf(APIError)
      expect(error).toMatchObject({
        status: 429,
        code: 'rate_limit_exceeded',
        details: { retry_after_seconds: 45 },
      })
    })
  })
})

describe('authApi.logout', () => {
  describe('when the session is ended', () => {
    it('should POST to /auth/logout and resolve on 204 No Content', async () => {
      let calledMethod: string | undefined
      server.use(
        http.post(`${API_BASE_URL}/auth/logout`, ({ request }) => {
          calledMethod = request.method
          return new HttpResponse(null, { status: 204 })
        })
      )

      await expect(authApi.logout()).resolves.toBeUndefined()
      expect(calledMethod).toBe('POST')
    })

    it('should resolve against the default no-content response', async () => {
      await expect(authApi.logout()).resolves.toBeUndefined()
    })
  })

  describe('when the session is already gone', () => {
    it('should reject with a 401 APIError', async () => {
      server.use(
        http.post(`${API_BASE_URL}/auth/logout`, () =>
          HttpResponse.json(
            {
              error: 'authentication_required',
              message: 'Session expired. Please log in again.',
              request_id: 'req_logout_401',
            },
            { status: 401 }
          )
        )
      )

      await expect(authApi.logout()).rejects.toMatchObject({
        status: 401,
        code: 'authentication_required',
      })
    })
  })
})

describe('authApi.me', () => {
  describe('when a session is active', () => {
    it('should GET /auth/me and resolve with the signed-in user', async () => {
      let method: string | undefined
      server.use(
        http.get(`${API_BASE_URL}/auth/me`, ({ request }) => {
          method = request.method
          return HttpResponse.json({ user: testUser }, { status: 200 })
        })
      )

      const result = await authApi.me()

      expect(method).toBe('GET')
      expect(result.user).toEqual(testUser)
    })

    it('should send the session cookie along with the request', async () => {
      let credentials: RequestCredentials | undefined
      server.use(
        http.get(`${API_BASE_URL}/auth/me`, ({ request }) => {
          credentials = request.credentials
          return HttpResponse.json({ user: testUser }, { status: 200 })
        })
      )

      await authApi.me()

      expect(credentials).toBe('include')
    })
  })

  describe('when there is no session', () => {
    it('should reject with a 401 APIError', async () => {
      server.use(
        http.get(`${API_BASE_URL}/auth/me`, () =>
          HttpResponse.json(
            {
              error: 'authentication_required',
              message: 'Session expired. Please log in again.',
              request_id: 'req_me401',
            },
            { status: 401 }
          )
        )
      )

      await expect(authApi.me()).rejects.toMatchObject({
        status: 401,
        code: 'authentication_required',
      })
    })
  })
})
