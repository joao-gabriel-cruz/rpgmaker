import { act, fireEvent, render, renderHook, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it } from 'vitest'
import { runShortcut, useTableShortcuts } from './useTableShortcuts'
import { TableToolProvider, useTableTool } from './useTableTool'

const wrap = (isGm: boolean) => ({ children }: { children: React.ReactNode }) => (
  <TableToolProvider isGm={isGm}>{children}</TableToolProvider>
)

function Harness({ isGm, onHelp = () => {} }: { isGm: boolean; onHelp?: () => void }) {
  const { tool, setTool } = useTableTool()
  useTableShortcuts({ isGm, setTool, setSidebarOpen: () => {}, openHelp: onHelp })
  return <><output aria-label="tool">{tool}</output><input aria-label="campo" /></>
}

const mount = (isGm: boolean, onHelp?: () => void) =>
  render(<TableToolProvider isGm={isGm}><Harness isGm={isGm} onHelp={onHelp} /></TableToolProvider>)

it('starts on Move and switches tool with the rail letters', () => {
  mount(true)
  expect(screen.getByLabelText('tool')).toHaveTextContent('move')
  fireEvent.keyDown(window, { key: 'p' })
  expect(screen.getByLabelText('tool')).toHaveTextContent('ping')
  fireEvent.keyDown(window, { key: 'T' })
  expect(screen.getByLabelText('tool')).toHaveTextContent('terrain')
  fireEvent.keyDown(window, { key: 'f' })
  expect(screen.getByLabelText('tool')).toHaveTextContent('fog')
})

it('returns to Move on Escape', () => {
  mount(false)
  fireEvent.keyDown(window, { key: 'd' })
  expect(screen.getByLabelText('tool')).toHaveTextContent('draw')
  fireEvent.keyDown(window, { key: 'Escape' })
  expect(screen.getByLabelText('tool')).toHaveTextContent('move')
})

it('ignores shortcuts while typing in a field and with modifiers', async () => {
  mount(true)
  await userEvent.type(screen.getByLabelText('campo'), 'rtp')
  expect(screen.getByLabelText('tool')).toHaveTextContent('move')
  fireEvent.keyDown(window, { key: 'r', ctrlKey: true })
  fireEvent.keyDown(window, { key: 'R', shiftKey: true })
  expect(screen.getByLabelText('tool')).toHaveTextContent('move')
})

it('keeps the Mestre tools away from players', () => {
  mount(false)
  fireEvent.keyDown(window, { key: 't' })
  fireEvent.keyDown(window, { key: 'k' })
  expect(screen.getByLabelText('tool')).toHaveTextContent('move')
  const { result } = renderHook(() => useTableTool(), { wrapper: wrap(false) })
  act(() => result.current.setTool('fog'))
  expect(result.current.tool).toBe('move')
})

it('opens the shortcuts help on ? and drives the sidebar with [ and ]', () => {
  let help = 0
  mount(false, () => { help++ })
  fireEvent.keyDown(window, { key: '?', shiftKey: true })
  expect(help).toBe(1)
  const calls: boolean[] = []
  const actions = { isGm: false, setTool: () => {}, setSidebarOpen: (o: boolean) => calls.push(o), openHelp: () => {} }
  expect(runShortcut('[', actions)).toBe(true)
  expect(runShortcut(']', actions)).toBe(true)
  expect(runShortcut('q', actions)).toBe(false)
  expect(calls).toEqual([false, true])
})

it('keeps the terrain editor tool coherent with the rail tool', () => {
  const { result } = renderHook(() => useTableTool(), { wrapper: wrap(true) })
  act(() => result.current.setTool('wall'))
  expect(result.current.editor.tool).toBe('wall')
  act(() => result.current.setTool('objects'))
  expect(result.current.editor.tool).toBe('object')
  act(() => result.current.setTool('terrain'))
  expect(result.current.editor.tool).toBe('paint')
})
