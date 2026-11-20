export type Health =
  | { api: 'ok'; database: 'ok' | 'error' }
  | { api: 'error' }

// A 503 still carries a valid body: the API answered, the database is what's down.
export async function fetchHealth(): Promise<Health> {
  try {
    const res = await fetch('/api/health')
    if (res.status !== 200 && res.status !== 503) return { api: 'error' }
    const body = (await res.json()) as { database?: string }
    return { api: 'ok', database: body.database === 'ok' ? 'ok' : 'error' }
  } catch {
    return { api: 'error' }
  }
}
