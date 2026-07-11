import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'

import { routes } from './index'

describe('routes', () => {
  it('renders the TrAIveler placeholder heading at the root route', () => {
    const memoryRouter = createMemoryRouter(routes, { initialEntries: ['/'] })

    render(<RouterProvider router={memoryRouter} />)

    expect(
      screen.getByRole('heading', { name: 'TrAIveler' })
    ).toBeInTheDocument()
  })
})
