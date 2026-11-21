import { ChevronDown, ChevronUp } from 'lucide-react'
import { useState } from 'react'
import { SURFACE } from './surface'
import { hasOptions, toolDef } from './tools'
import { useTableTool } from './useTableTool'

/**
 * Floating options panel next to the rail. The map portals the active tool's controls into
 * the host node (see OptionsPortal / options/MapOptions).
 */
export function OptionsPanel() {
  const { tool, draftOpen, setOptionsHost } = useTableTool()
  const [collapsed, setCollapsed] = useState(false)
  const visible = hasOptions(tool) || draftOpen
  const title = hasOptions(tool) ? toolDef(tool).label : 'Rascunho'
  return (
    <section aria-label="Opções da ferramenta" hidden={!visible}
      className={`absolute left-3 top-3 z-10 flex max-h-[calc(100%-1.5rem)] w-72 flex-col rounded-xl p-1 ${SURFACE}`}>
      <header className="flex items-center justify-between px-2 py-1">
        <h2 className="text-sm font-semibold text-zinc-100">{title}</h2>
        <button type="button" aria-label={collapsed ? 'Expandir opções' : 'Recolher opções'} aria-expanded={!collapsed}
          onClick={() => setCollapsed((c) => !c)}
          className="rounded-md p-1 text-zinc-400 transition-colors hover:bg-zinc-800 hover:text-white active:scale-[0.96]">
          {collapsed ? <ChevronDown size={16} aria-hidden /> : <ChevronUp size={16} aria-hidden />}
        </button>
      </header>
      <div ref={setOptionsHost} hidden={collapsed} className="min-h-0 overflow-y-auto rounded-lg p-2" />
    </section>
  )
}
