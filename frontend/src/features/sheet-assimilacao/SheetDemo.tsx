import { useState } from 'react'
import { AssimilacaoSheet } from './AssimilacaoSheet'
import { newSheet, type RollRequest } from './types'

export function SheetDemo() {
  const [sheet, setSheet] = useState(newSheet)
  const [last, setLast] = useState<RollRequest | null>(null)
  return (
    <div className="grid min-h-screen grid-cols-[1fr_24rem] gap-4 bg-zinc-950 p-4">
      <AssimilacaoSheet sheet={sheet} onChange={setSheet} onRollRequest={setLast} />
      <aside className="sticky top-4 h-[calc(100vh-2rem)] overflow-auto rounded-lg border border-zinc-800 bg-zinc-900 p-3 text-xs text-emerald-300">
        <p className="mb-2 text-zinc-400">Última rolagem pedida: {last ? JSON.stringify(last) : '—'}</p>
        <pre>{JSON.stringify(sheet, null, 2)}</pre>
      </aside>
    </div>
  )
}
