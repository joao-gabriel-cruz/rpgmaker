import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'
import { renderApp, stubApi } from '../../test/renderApp'

const admin = { id: '1', email: 'a@x.com', displayName: 'Ana', isAdmin: true, createdAt: '' }

afterEach(() => vi.unstubAllGlobals())

it('lists users and shows a temporary password after reset', async () => {
  stubApi((method, path) => {
    if (path === '/auth/me') return { status: 200, body: admin }
    if (path === '/admin/users') return { status: 200, body: [admin] }
    if (method === 'POST' && path === '/admin/users/1/reset-password') {
      return { status: 200, body: { temporaryPassword: 'Abc123xyz789' } }
    }
    return { status: 200, body: { api: 'ok', database: 'ok' } }
  })
  renderApp('/admin')
  await userEvent.click(await screen.findByRole('button', { name: 'Resetar senha' }))
  expect(await screen.findByText('Abc123xyz789')).toBeInTheDocument()
})
