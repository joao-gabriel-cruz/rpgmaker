import { Navigate, Outlet } from 'react-router'
import { useMe } from './auth'

export function RequireAuth() {
  const { data: user, isPending } = useMe()
  if (isPending) return <p className="p-8 text-zinc-400">Carregando…</p>
  if (!user) return <Navigate to="/login" replace />
  return <Outlet />
}

export function RequireAdmin() {
  const { data: user } = useMe()
  if (!user?.isAdmin) return <Navigate to="/" replace />
  return <Outlet />
}
