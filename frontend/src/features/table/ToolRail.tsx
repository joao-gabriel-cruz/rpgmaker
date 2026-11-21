import { FilePenLine } from 'lucide-react'
import { ICON_BTN } from './surface'
import { Kbd, Tip } from './Tip'
import { toolsFor, type TableToolDef } from './tools'
import { useTableTool } from './useTableTool'

function RailButton({ def, active, onClick }: { def: TableToolDef; active: boolean; onClick: () => void }) {
  const Icon = def.icon
  return (
    <button type="button" aria-label={def.label} aria-pressed={active} aria-keyshortcuts={def.key.toUpperCase()} onClick={onClick}
      className={`group relative ${ICON_BTN} ${active ? 'bg-emerald-600 text-white hover:bg-emerald-500' : ''}`}>
      <Icon size={20} strokeWidth={active ? 2.25 : 1.5} aria-hidden />
      <Tip>{def.label}<Kbd>{def.key.toUpperCase()}</Kbd></Tip>
    </button>
  )
}

function DraftToggle() {
  const { draftOpen, toggleDraft, draftActive } = useTableTool()
  return (
    <button type="button" aria-label="Painel de rascunho" aria-pressed={draftOpen} onClick={toggleDraft}
      className={`group relative ${ICON_BTN} ${draftOpen ? 'bg-amber-600 text-white hover:bg-amber-500' : ''}`}>
      <FilePenLine size={20} strokeWidth={draftOpen ? 2.25 : 1.5} aria-hidden />
      {draftActive && <span aria-hidden className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-amber-400" />}
      <Tip>Rascunho (estados salvos)</Tip>
    </button>
  )
}

/** Left rail: one active tool at a time for everyone; the Mestre's tools come after a separator. */
export function ToolRail({ isGm }: { isGm: boolean }) {
  const { tool, setTool } = useTableTool()
  const all = toolsFor(isGm)
  const group = (gm: boolean) => all.filter((t) => t.gmOnly === gm)
  return (
    <nav aria-label="Ferramentas" className="z-20 flex w-14 shrink-0 flex-col items-center gap-1 border-r border-zinc-800 bg-zinc-950 py-2">
      {group(false).map((d) => <RailButton key={d.id} def={d} active={tool === d.id} onClick={() => setTool(d.id)} />)}
      {isGm && (
        <>
          <hr aria-hidden className="my-1 w-8 border-zinc-800" />
          {group(true).map((d) => <RailButton key={d.id} def={d} active={tool === d.id} onClick={() => setTool(d.id)} />)}
          <DraftToggle />
        </>
      )}
    </nav>
  )
}
