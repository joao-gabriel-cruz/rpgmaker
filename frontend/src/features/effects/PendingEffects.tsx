import { useEffect, useState } from 'react'
import { api } from '../../lib/api'
import { useRealtime } from '../realtime/realtimeContext'
import { describe, TRIGGER_LABEL, type Suggestion } from './effects'

const btn = 'rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700'

function Item({ s, act }: { s: Suggestion; act: (type: string, id: string) => void }) {
  return (
    <li className="flex flex-wrap items-center gap-3 rounded bg-zinc-800 px-3 py-2">
      <div className="min-w-0 flex-1 text-sm">
        <p><strong>{s.characterName}</strong> · {s.source} · {TRIGGER_LABEL[s.trigger]}</p>
        <p className="text-zinc-300">{describe(s)}{s.message ? ` — ${s.message}` : ''}</p>
      </div>
      <button className={`${btn} bg-emerald-700 hover:bg-emerald-600`} onClick={() => act('effect.confirm', s.id)}
        aria-label={`Aplicar efeito em ${s.characterName}`}>Aplicar</button>
      <button className={btn} onClick={() => act('effect.dismiss', s.id)}
        aria-label={`Descartar efeito de ${s.characterName}`}>Descartar</button>
    </li>
  )
}

function merge(loaded: Suggestion[], cur: Suggestion[]) {
  return [...loaded, ...cur.filter((c) => !loaded.some((x) => x.id === c.id))]
}

function usePending(campaignId: string) {
  const { status, subscribe } = useRealtime()
  const [list, setList] = useState<Suggestion[]>([])
  const [fresh, setFresh] = useState(false)

  useEffect(() => subscribe('effect.suggested', (p) => {
    const s = (p as { suggestion: Suggestion }).suggestion
    setList((cur) => (cur.some((c) => c.id === s.id) ? cur : [...cur, s]))
    setFresh(true)
  }), [subscribe])
  useEffect(() => subscribe('effect.resolved', (p) => {
    const { id } = p as { id: string }
    setList((cur) => cur.filter((c) => c.id !== id))
  }), [subscribe])
  useEffect(() => {
    if (status !== 'open') return
    api.get<{ suggestions: Suggestion[] }>(`/campaigns/${campaignId}/effects/pending`)
      .then((r) => setList((cur) => merge(Array.isArray(r?.suggestions) ? r.suggestions : [], cur)), () => {})
  }, [campaignId, status])
  return { list, fresh, setFresh }
}

/** The Mestre's inbox of terrain effects waiting to be applied or dismissed. */
export function PendingEffects({ campaignId }: { campaignId: string }) {
  const { send } = useRealtime()
  const { list, fresh, setFresh } = usePending(campaignId)
  const [error, setError] = useState('')

  const act = (type: string, id: string) => {
    setError('')
    setFresh(false)
    send(type, { id }).catch((e: Error) => setError(e.message))
  }

  return (
    <section aria-label="Efeitos pendentes" onMouseEnter={() => setFresh(false)}
      className={`flex flex-col gap-2 rounded bg-zinc-900 px-4 py-3 ${fresh ? 'ring-2 ring-amber-400' : ''}`}>
      <h2 className="flex items-center gap-2 text-lg font-semibold">
        Efeitos pendentes
        <span data-testid="pending-count" className={`rounded-full px-2 text-sm ${list.length ? 'bg-amber-500 text-black' : 'bg-zinc-700'}`}>
          {list.length}
        </span>
      </h2>
      {list.length === 0 && <p className="text-sm text-zinc-400">Nenhum efeito aguardando.</p>}
      <ul className="flex flex-col gap-2">{list.map((s) => <Item key={s.id} s={s} act={act} />)}</ul>
      {error && <p role="alert" className="text-sm text-red-400">{error}</p>}
    </section>
  )
}
