import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { renderApp, stubApi } from '../../test/renderApp'

const ana = { id: '1', email: 'a@x.com', displayName: 'Ana', isAdmin: false, createdAt: '' }

afterEach(() => vi.unstubAllGlobals())

describe('auth flow', () => {
  it('redirects anonymous visitors to the login screen', async () => {
    stubApi((_m, path) => (path === '/auth/me' ? { status: 401 } : { status: 200 }))
    renderApp('/admin')
    expect(await screen.findByRole('heading', { name: 'Entrar' })).toBeInTheDocument()
  })

  it('shows the generic error when login fails', async () => {
    stubApi((method, path) => {
      if (path === '/auth/login' && method === 'POST') {
        return { status: 401, body: { error: 'E-mail ou senha inválidos' } }
      }
      return { status: 401 }
    })
    renderApp('/login')
    await userEvent.type(await screen.findByLabelText('E-mail'), 'a@x.com')
    await userEvent.type(screen.getByLabelText('Senha'), 'wrong')
    await userEvent.click(screen.getByRole('button', { name: 'Entrar' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('E-mail ou senha inválidos')
  })

})

describe('auth layout', () => {
  it('shows the user name in the layout when logged in', async () => {
    stubApi(() => ({ status: 200, body: ana }))
    renderApp('/')
    expect(await screen.findByRole('button', { name: 'Sair' })).toBeInTheDocument()
    expect(screen.getAllByText('Ana').length).toBeGreaterThan(0)
  })

  it('keeps non-admins out of the admin page', async () => {
    stubApi(() => ({ status: 200, body: ana }))
    renderApp('/admin')
    expect(await screen.findByRole('heading', { name: 'Olá, Ana' })).toBeInTheDocument()
  })
})
