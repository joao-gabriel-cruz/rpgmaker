import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { vi } from 'vitest'
import App from '../App'

type Handler = (method: string, path: string, body?: string) => { status: number; body?: unknown }

export function stubApi(handler: Handler) {
  vi.stubGlobal('fetch', vi.fn((url: string, init?: RequestInit) => {
    const r = handler(init?.method ?? 'GET', url.replace('/api', ''), init?.body as string | undefined)
    return Promise.resolve(new Response(r.status === 204 ? null : JSON.stringify(r.body ?? {}), { status: r.status }))
  }))
}

export function renderApp(path: string) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[path]}>
        <App />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}
