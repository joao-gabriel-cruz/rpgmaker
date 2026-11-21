import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'
import { fakeRealtime } from '../../test/fakeRealtime'
import { stubApi } from '../../test/renderApp'
import { ToastSlot } from '../table/slots/ToastSlot'
import { ClaudeChangeToast } from './ClaudeChangeToast'
import { ClaudePanel } from './ClaudePanel'

afterEach(() => vi.unstubAllGlobals())

const change = (id: string, summary: string) => ({ id, sceneId: 's1', summary, at: new Date().toISOString() })

function setup(opts: { exists?: boolean; changes?: unknown[] } = {}) {
  const calls: string[] = []
  stubApi((method, path) => {
    calls.push(`${method} ${path}`)
    if (path === '/campaigns/c1/mcp-link' && method === 'GET') return { status: 200, body: { exists: !!opts.exists, createdAt: '2026-01-01T00:00:00Z', lastUsedAt: null } }
    if (path === '/campaigns/c1/mcp-link' && method === 'POST') return { status: 201, body: { url: 'https://x/api/mcp/secret123', createdAt: '' } }
    if (path === '/campaigns/c1/mcp-link' && method === 'DELETE') return { status: 204 }
    if (path === '/campaigns/c1/mcp-changes') return { status: 200, body: { changes: opts.changes ?? [] } }
    return { status: 404, body: {} }
  })
  const rt = fakeRealtime()
  const view = () => (
    <QueryClientProvider client={client}>
      <rt.Provider><ClaudePanel campaignId="c1" /></rt.Provider>
    </QueryClientProvider>
  )
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const utils = render(view())
  return { calls, rt, utils, view, client }
}

it('shows the generated URL once and hides it after remount', async () => {
  const { utils, view } = setup()
  await userEvent.click(await screen.findByRole('button', { name: 'Gerar link' }))
  expect(await screen.findByDisplayValue('https://x/api/mcp/secret123')).toBeInTheDocument()
  expect(screen.getByText(/Ele só aparece agora/)).toBeInTheDocument()
  utils.unmount()
  render(view())
  await screen.findByText('Mudanças recentes do Claude')
  expect(screen.queryByDisplayValue('https://x/api/mcp/secret123')).not.toBeInTheDocument()
})

it('revokes the link with DELETE', async () => {
  const { calls } = setup({ exists: true })
  expect(await screen.findByText('Nunca usado')).toBeInTheDocument()
  await userEvent.click(screen.getByRole('button', { name: 'Revogar link' }))
  await waitFor(() => expect(calls).toContain('DELETE /campaigns/c1/mcp-link'))
})

it('confirms before regenerating', async () => {
  const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
  const { calls } = setup({ exists: true })
  await userEvent.click(await screen.findByRole('button', { name: 'Gerar novo link' }))
  expect(confirm).toHaveBeenCalledWith('O link atual vai parar de funcionar. Continuar?')
  expect(calls).not.toContain('POST /campaigns/c1/mcp-link')
  confirm.mockRestore()
})

it('updates the list on an mcp.changed event', async () => {
  const { rt } = setup({ changes: [change('1', 'Pintou grama')] })
  expect(await screen.findByText('Pintou grama')).toBeInTheDocument()
  act(() => rt.emit('mcp.changed', { changes: [change('2', 'Trocou o clima'), change('1', 'Pintou grama')] }))
  expect(await screen.findByText('Trocou o clima')).toBeInTheDocument()
})

it('shows the empty state', async () => {
  setup()
  expect(await screen.findByText('Nenhuma mudança ainda.')).toBeInTheDocument()
})

it('sends mcp.undo with the change id and surfaces errors', async () => {
  const { rt } = setup({ changes: [change('7', 'Pintou grama')] })
  await userEvent.click(await screen.findByRole('button', { name: 'Desfazer' }))
  expect(rt.send).toHaveBeenCalledWith('mcp.undo', { changeId: '7' })
  rt.send.mockRejectedValueOnce(new Error('Não dá para desfazer.'))
  await userEvent.click(screen.getByRole('button', { name: 'Desfazer' }))
  expect(await screen.findByRole('alert')).toHaveTextContent('Não dá para desfazer.')
})

function renderToast(isGm: boolean) {
  const rt = fakeRealtime()
  const client = new QueryClient()
  render(<QueryClientProvider client={client}><rt.Provider>{isGm ? <ClaudeChangeToast /> : <ToastSlot isGm={false} />}</rt.Provider></QueryClientProvider>)
  return rt
}

it('toast shows the latest change and undoes it', async () => {
  const rt = renderToast(true)
  const c = change('9', '14 células → Lava perto de (12,8)')
  act(() => rt.emit('mcp.changed', { changes: [c], latest: c }))
  expect(await screen.findByRole('status')).toHaveTextContent('Claude: 14 células → Lava perto de (12,8)')
  await userEvent.click(screen.getByRole('button', { name: 'Desfazer' }))
  expect(rt.send).toHaveBeenCalledWith('mcp.undo', { changeId: '9' })
})

it('toast stays closed without latest and closes when its change is undone', async () => {
  const rt = renderToast(true)
  act(() => rt.emit('mcp.changed', { changes: [change('1', 'a')] }))
  expect(screen.queryByRole('status')).not.toBeInTheDocument()
  const c = change('2', 'b')
  act(() => rt.emit('mcp.changed', { changes: [c], latest: c }))
  expect(await screen.findByRole('status')).toBeInTheDocument()
  act(() => rt.emit('mcp.changed', { changes: [] }))
  expect(screen.queryByRole('status')).not.toBeInTheDocument()
})

it('toast is not rendered for a non-GM slot', () => {
  const rt = renderToast(false)
  const c = change('3', 'c')
  act(() => rt.emit('mcp.changed', { changes: [c], latest: c }))
  expect(screen.queryByText(/Claude:/)).not.toBeInTheDocument()
})

function stubClipboard(writeText: () => Promise<void>) {
  Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
}

it('shows Copiado! after copying the link', async () => {
  stubClipboard(() => Promise.resolve())
  setup()
  await userEvent.click(await screen.findByRole('button', { name: 'Gerar link' }))
  await userEvent.click(await screen.findByRole('button', { name: 'Copiar' }))
  expect(await screen.findByRole('button', { name: 'Copiado!' })).toBeInTheDocument()
})

it('explains how to copy manually when the clipboard fails', async () => {
  stubClipboard(() => Promise.reject(new Error('denied')))
  setup()
  await userEvent.click(await screen.findByRole('button', { name: 'Gerar link' }))
  await userEvent.click(await screen.findByRole('button', { name: 'Copiar' }))
  expect(await screen.findByText('Não foi possível copiar; selecione o link e copie manualmente.')).toBeInTheDocument()
})
