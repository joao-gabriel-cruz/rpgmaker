import { useMutation, useQuery } from '@tanstack/react-query'
import { api } from '../../lib/api'
import type { User } from '../auth/auth'

const resetPassword = (id: string) =>
  api.post<{ temporaryPassword: string }>(`/admin/users/${id}/reset-password`)

function UserRow({ user }: { user: User }) {
  const reset = useMutation({ mutationFn: () => resetPassword(user.id) })
  return (
    <tr className="border-t border-zinc-800">
      <td className="py-2 pr-4">{user.displayName}</td>
      <td className="py-2 pr-4 text-zinc-400">{user.email}</td>
      <td className="py-2 pr-4">{user.isAdmin ? 'Admin' : 'Jogador'}</td>
      <td className="py-2">
        <button onClick={() => reset.mutate()} disabled={reset.isPending}
          className="rounded bg-zinc-800 px-3 py-1 text-sm hover:bg-zinc-700">
          Resetar senha
        </button>
        {reset.data && (
          <p role="status" className="mt-1 text-sm text-emerald-400">
            Senha temporária: <code className="font-mono">{reset.data.temporaryPassword}</code>
          </p>
        )}
        {reset.isError && <p role="alert" className="mt-1 text-sm text-red-400">{reset.error.message}</p>}
      </td>
    </tr>
  )
}

export function AdminPage() {
  const { data: users, isPending, isError } = useQuery({
    queryKey: ['admin', 'users'],
    queryFn: () => api.get<User[]>('/admin/users'),
  })

  return (
    <section>
      <h1 className="mb-4 text-2xl font-bold tracking-tight">Usuários</h1>
      {isPending && <p className="text-zinc-400">Carregando…</p>}
      {isError && <p role="alert" className="text-red-400">Não foi possível carregar os usuários</p>}
      {users && (
        <table className="w-full max-w-3xl text-left text-sm">
          <thead className="text-zinc-400">
            <tr><th className="pb-2">Nome</th><th className="pb-2">E-mail</th><th className="pb-2">Papel</th><th /></tr>
          </thead>
          <tbody>{users.map((u) => <UserRow key={u.id} user={u} />)}</tbody>
        </table>
      )}
    </section>
  )
}
