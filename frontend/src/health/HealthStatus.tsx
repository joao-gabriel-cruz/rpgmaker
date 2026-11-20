import { useEffect, useState } from 'react'
import { fetchHealth, type Health } from './health'

function label(health: Health): string {
  if (health.api === 'error') return 'API fora do ar'
  return health.database === 'ok' ? 'API ok · banco ok' : 'API ok · banco fora do ar'
}

export function HealthStatus() {
  const [health, setHealth] = useState<Health | null>(null)

  useEffect(() => {
    let active = true
    fetchHealth().then((h) => {
      if (active) setHealth(h)
    })
    return () => {
      active = false
    }
  }, [])

  if (!health) return <p className="text-zinc-400">Verificando…</p>

  const healthy = health.api === 'ok' && health.database === 'ok'
  return (
    <p role="status" className={healthy ? 'text-emerald-400' : 'text-red-400'}>
      {label(health)}
    </p>
  )
}
