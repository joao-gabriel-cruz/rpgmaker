import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Link, Outlet, useNavigate } from 'react-router'
import { HealthStatus } from '../../health/HealthStatus'
import { logout, meKey, useMe } from './auth'

export function Layout() {
  const { data: user } = useMe()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const signOut = useMutation({
    mutationFn: logout,
    onSuccess: () => {
      queryClient.setQueryData(meKey, null)
      queryClient.removeQueries({ predicate: (q) => q.queryKey[0] !== 'me' })
      navigate('/login', { replace: true })
    },
  })

  return (
    <div className="flex min-h-screen flex-col bg-zinc-950 text-zinc-100">
      <header className="flex items-center gap-4 border-b border-zinc-800 px-6 py-3">
        <Link to="/" className="text-lg font-bold tracking-tight">rpgmaker</Link>
        {user?.isAdmin && <Link to="/admin" className="text-sm text-emerald-400">Administração</Link>}
        <span className="ml-auto text-sm text-zinc-300">{user?.displayName}</span>
        <button onClick={() => signOut.mutate()}
          className="rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700">
          Sair
        </button>
      </header>
      <div className="flex-1 p-6"><Outlet /></div>
      <footer className="px-6 py-2 text-xs"><HealthStatus /></footer>
    </div>
  )
}
