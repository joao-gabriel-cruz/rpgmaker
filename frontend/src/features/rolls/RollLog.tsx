import { useQuery } from '@tanstack/react-query'
import { useEffect, useState, type FormEvent } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import { modifiersText } from './modifiers'
import { poolLabel, type FinalizedEvent, type RollEntry } from './rolls'
import { AssignToObjective } from '../conflict/AssignToObjective'
import { SymbolCounts } from './SymbolCounts'

function useRollLog(campaignId: string, isGm: boolean) {
  const { subscribe } = useRealtime()
  const { data } = useQuery({
    queryKey: ['campaigns', campaignId, 'rolls'],
    queryFn: () => api.get<RollEntry[]>(`/campaigns/${campaignId}/rolls?limit=50`),
  })
  const [live, setLive] = useState<RollEntry[]>([])
  const [finals, setFinals] = useState<Record<string, FinalizedEvent['result']>>({})
  useEffect(() => subscribe('roll.rolled', (p) => setLive((l) => [p as RollEntry, ...l])), [subscribe])
  useEffect(() => subscribe('roll.finalized', (p) => {
    const ev = p as FinalizedEvent
    setFinals((f) => ({ ...f, [ev.rollId]: ev.result }))
  }), [subscribe])
  const seen = new Set<string>()
  const entries = [...live, ...(Array.isArray(data) ? data : [])].filter((e) => {
    if (seen.has(e.rollId) || (e.visibility === 'gm' && !isGm)) return false
    seen.add(e.rollId)
    return true
  })
  return entries.map((e) => (finals[e.rollId] ? { ...e, status: 'final' as const, result: finals[e.rollId] } : e))
}

function Outcome({ e }: { e: RollEntry }) {
  if (e.kind === 'generic') {
    return <p className="text-sm">[{e.dice.map((d) => d.face).join(', ')}] = <strong>{e.total}</strong></p>
  }
  if (!e.result) return <p className="text-sm text-zinc-400">Escolhendo dados…</p>
  return <SymbolCounts result={e.result} />
}

function Entry({ e }: { e: RollEntry }) {
  const mods = modifiersText(e.request)
  return (
    <li className="flex flex-col gap-0.5 rounded-xl bg-zinc-800/60 px-3 py-2">
      <p className="text-sm">
        <span className="font-medium">{e.userName}</span>
        {e.characterName && <span className="text-zinc-400"> · {e.characterName}</span>}
        {e.visibility === 'gm' && <span className="ml-2 rounded bg-purple-900 px-1.5 text-xs text-purple-200">Secreta</span>}
      </p>
      <p className="text-xs text-zinc-400">{poolLabel(e)}</p>
      {mods && <p className="text-xs text-amber-300">{mods}</p>}
      <Outcome e={e} />
      {e.kind === 'assimilacao' && e.result && <AssignToObjective rollId={e.rollId} ownerId={e.userId} />}
    </li>
  )
}

function CommandBox() {
  const { send } = useRealtime()
  const [text, setText] = useState('')
  const [error, setError] = useState('')
  const submit = (ev: FormEvent) => {
    ev.preventDefault()
    const m = /^\/r\s+(.+)$/i.exec(text.trim())
    if (!m) return setError('Comando desconhecido. Use /r 2d6+3')
    setError('')
    send('roll.generic', { expression: m[1] }).then(() => setText('')).catch((e: Error) => setError(e.message))
  }
  return (
    <form onSubmit={submit} className="flex flex-col gap-1">
      <input aria-label="Comando de rolagem" value={text} onChange={(e) => setText(e.target.value)} placeholder="/r 2d6+3"
        className="rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm" />
      {error && <p role="alert" className="text-xs text-red-400">{error}</p>}
    </form>
  )
}

/** Side panel with the recent rolls (visibility-filtered by the server) and the /r command. */
export function RollLog({ campaignId, isGm }: { campaignId: string; isGm: boolean }) {
  const entries = useRollLog(campaignId, isGm)
  return (
    <aside aria-label="Log de rolagens" className="flex min-h-0 flex-col gap-2">
      <h2 className="text-lg font-semibold">Rolagens</h2>
      <CommandBox />
      <ul className="flex flex-col gap-2">
        {entries.map((e) => <Entry key={e.rollId} e={e} />)}
        {entries.length === 0 && <li className="text-sm text-zinc-400">Nenhuma rolagem ainda.</li>}
      </ul>
    </aside>
  )
}
