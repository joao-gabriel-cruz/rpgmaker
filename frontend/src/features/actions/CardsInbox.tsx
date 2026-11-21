import { useEffect, useState } from 'react'
import { useRealtime } from '../realtime/realtimeContext'

interface Card {
  suit: 'hearts' | 'diamonds' | 'spades'
  name: string
  description: string
}

interface Drawn {
  actionRunId: string
  characterId: string
  characterName: string
  cards: Card[]
}

const SUIT: Record<Card['suit'], { label: string; tone: string }> = {
  hearts: { label: '♥ Evolutiva', tone: 'text-rose-300' },
  diamonds: { label: '♦ Adaptativa', tone: 'text-sky-300' },
  spades: { label: '♠ Inoportuna', tone: 'text-zinc-300' },
}

/** Mestre only: Assimilação cards drawn by a rest test; confirming adds one to the character's sheet. */
export function CardsInbox() {
  const { subscribe, send } = useRealtime()
  const [drawn, setDrawn] = useState<Drawn[]>([])
  const [error, setError] = useState('')
  useEffect(() => subscribe('action.cards', (p) => setDrawn((l) => [...l, p as Drawn])), [subscribe])
  const drop = (run: string, name: string) => setDrawn((l) => l
    .map((d) => (d.actionRunId === run ? { ...d, cards: d.cards.filter((c) => c.name !== name) } : d)).filter((d) => d.cards.length > 0))
  const accept = (d: Drawn, c: Card) => {
    setError('')
    send('action.card.accept', { characterId: d.characterId, ...c }).then(() => drop(d.actionRunId, c.name), (e: Error) => setError(e.message))
  }
  if (drawn.length === 0) return null
  return (
    <aside aria-label="Cartas de Assimilação sorteadas" className="fixed bottom-28 left-20 z-40 flex w-72 flex-col gap-2 rounded-2xl bg-zinc-900/90 p-3 shadow-xl ring-1 ring-white/10 backdrop-blur">
      <h2 className="text-sm font-semibold">Teste de Descanso: cartas sorteadas</h2>
      {drawn.map((d) => d.cards.map((c) => (
        <div key={`${d.actionRunId}:${c.name}`} className="flex flex-col gap-1 rounded-lg bg-zinc-800/70 p-2">
          <p className="text-xs text-zinc-400">{d.characterName} · <span className={SUIT[c.suit].tone}>{SUIT[c.suit].label}</span></p>
          <p className="text-sm font-medium">{c.name}</p>
          <p className="text-xs text-zinc-300">{c.description}</p>
          <div className="mt-1 flex gap-1.5">
            <button type="button" onClick={() => accept(d, c)} className="rounded-md bg-emerald-600 px-2 py-1 text-xs font-semibold text-white hover:bg-emerald-500">Adicionar à ficha</button>
            <button type="button" onClick={() => drop(d.actionRunId, c.name)} className="rounded-md px-2 py-1 text-xs text-zinc-300 hover:bg-zinc-700">Descartar</button>
          </div>
        </div>
      )))}
      {error && <p role="alert" className="text-xs text-red-400">{error}</p>}
    </aside>
  )
}
