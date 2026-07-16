import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'

import { routes } from './index'

describe('routes', () => {
  describe('when navigating to the root route', () => {
    it('should render the TrAIveler placeholder heading', () => {
      const memoryRouter = createMemoryRouter(routes, { initialEntries: ['/'] })

      render(<RouterProvider router={memoryRouter} />)

      expect(
        screen.getByRole('heading', { name: 'TrAIveler' })
      ).toBeInTheDocument()
    })
  })
})
