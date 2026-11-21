import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import { ShortcutsDialog } from '../table/ShortcutsDialog'
import { runShortcut, type ShortcutActions } from '../table/useTableShortcuts'

const actions = (isGm: boolean): ShortcutActions & { speak: () => void; react: () => void; setTool: (t: string) => void } =>
  ({ isGm, setTool: vi.fn(), setSidebarOpen: vi.fn(), openHelp: vi.fn(), speak: vi.fn(), react: vi.fn() })

it("' opens the speak box, G the reaction wheel, A the Mestre's Ambiente tool", () => {
  const gm = actions(true)
  expect(runShortcut("'", gm)).toBe(true)
  expect(gm.speak).toHaveBeenCalled()
  expect(runShortcut('g', gm)).toBe(true)
  expect(gm.react).toHaveBeenCalled()
  expect(runShortcut('a', gm)).toBe(true)
  expect(gm.setTool).toHaveBeenCalledWith('ambience')
  const player = actions(false)
  expect(runShortcut('a', player)).toBe(false)
})

it('lists the new shortcuts in the help dialog', () => {
  render(<ShortcutsDialog isGm onClose={() => {}} />)
  expect(screen.getByText(/Falar em personagem/)).toBeInTheDocument()
  expect(screen.getByText(/roda de reações/)).toBeInTheDocument()
  expect(screen.getByText('Ambiente')).toBeInTheDocument()
})
