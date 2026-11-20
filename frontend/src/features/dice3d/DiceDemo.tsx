import { useState } from 'react'
import { DiceTray } from './DiceTray'
import { countSymbols, faceCount, type DieKind, type DieResult } from './faces'

const KINDS: DieKind[] = ['d6', 'd10', 'd12']

function randomDice(counts: Record<DieKind, number>): DieResult[] {
  const dice: DieResult[] = []
  for (const kind of KINDS) {
    for (let i = 0; i < counts[kind]; i++) {
      dice.push({ kind, face: 1 + Math.floor(Math.random() * faceCount(kind)) })
    }
  }
  return dice
}

function useDemoRoll() {
  const [counts, setCounts] = useState<Record<DieKind, number>>({ d6: 3, d10: 2, d12: 1 })
  const [dice, setDice] = useState<DieResult[]>([])
  const [rollId, setRollId] = useState(0)
  const [seed, setSeed] = useState(1)
  const [selected, setSelected] = useState<number[]>([])
  const [done, setDone] = useState(false)
  const roll = () => {
    setDice(randomDice(counts))
    setSeed(Math.floor(Math.random() * 1e9))
    setRollId((n) => n + 1)
    setSelected([])
    setDone(false)
  }
  const toggle = (i: number) => setSelected((s) => (s.includes(i) ? s.filter((x) => x !== i) : [...s, i]))
  return { counts, setCounts, dice, rollId, seed, selected, done, setDone, roll, toggle }
}

interface ControlsProps {
  counts: Record<DieKind, number>
  onChange: (counts: Record<DieKind, number>) => void
  onRoll: () => void
}

function Controls({ counts, onChange, onRoll }: ControlsProps) {
  return (
    <div className="flex flex-wrap items-end gap-4">
      {KINDS.map((kind) => (
        <label key={kind} className="flex flex-col text-sm">
          {kind}
          <input
            type="number"
            min={0}
            max={12}
            value={counts[kind]}
            onChange={(e) => onChange({ ...counts, [kind]: Math.max(0, Math.min(12, Number(e.target.value))) })}
            className="w-20 rounded border border-zinc-700 bg-zinc-900 px-2 py-1"
          />
        </label>
      ))}
      <button onClick={onRoll} className="rounded bg-emerald-600 px-4 py-1.5 font-semibold hover:bg-emerald-500">
        Rolar
      </button>
    </div>
  )
}

// Temporary visual test page (?dice-demo); replaced by the real roll flow in #16.
export function DiceDemo() {
  const r = useDemoRoll()
  const total = countSymbols(r.dice)
  return (
    <main className="mx-auto flex min-h-screen max-w-3xl flex-col gap-4 bg-zinc-950 p-6 text-zinc-100">
      <h1 className="text-2xl font-bold">Demo de dados 3D</h1>
      <Controls counts={r.counts} onChange={r.setCounts} onRoll={r.roll} />
      <DiceTray
        dice={r.dice}
        rollId={r.rollId}
        seed={r.seed}
        selectable
        selected={r.selected}
        onToggle={r.toggle}
        onSettled={() => r.setDone(true)}
      />
      {r.done && (
        <p className="text-sm">
          Sucesso: {total.success} · Adaptação: {total.adaptation} · Pressão: {total.pressure} (selecionados:{' '}
          {r.selected.length})
        </p>
      )}
    </main>
  )
}
