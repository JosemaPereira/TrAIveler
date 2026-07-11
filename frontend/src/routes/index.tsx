import { createBrowserRouter } from 'react-router'
import type { RouteObject } from 'react-router'

import { HomePage } from './HomePage'
import { RootLayout } from './RootLayout'

export const routes: RouteObject[] = [
  {
    path: '/',
    element: <RootLayout />,
    children: [
      {
        index: true,
        element: <HomePage />,
      },
    ],
  },
]

export const router = createBrowserRouter(routes)
