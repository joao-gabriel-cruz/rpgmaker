import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router'
import { api } from '../../lib/api'
import { meKey, useMe, type User } from '../auth/auth'
import { campaignsKey, type InviteInfo } from './campaigns'

const field = 'flex flex-col gap-1 text-sm'
const input = 'rounded bg-zinc-800 px-3 py-2'

function Field({ label, ...rest }: { label: string } & React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <label className={field}>{label}
      <input required {...rest} className={input} />
    </label>
  )
}

function useSignup(token: string, form: object) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  return useMutation({
    mutationFn: () => api.post<User>(`/invites/${token}/signup`, form),
    onSuccess: (user) => {
      queryClient.setQueryData(meKey, user)
      queryClient.invalidateQueries({ queryKey: campaignsKey })
      navigate('/', { replace: true })
    },
  })
}

function SignupForm({ token }: { token: string }) {
  const [form, setForm] = useState({ email: '', displayName: '', password: '' })
  const signup = useSignup(token, form)
  const set = (k: keyof typeof form) => (e: { target: { value: string } }) =>
    setForm({ ...form, [k]: e.target.value })
  const submit = (e: FormEvent) => {
    e.preventDefault()
    signup.mutate()
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-3">
      <Field label="Nome" value={form.displayName} onChange={set('displayName')} />
      <Field label="E-mail" type="email" value={form.email} onChange={set('email')} />
      <Field label="Senha" type="password" minLength={8} value={form.password} onChange={set('password')} />
      {signup.isError && <p role="alert" className="text-sm text-red-400">{signup.error.message}</p>}
      <button type="submit" disabled={signup.isPending}
        className="rounded bg-emerald-600 px-3 py-2 font-medium hover:bg-emerald-500 disabled:opacity-50">
        Cadastrar e entrar
      </button>
      <Link to="/login" className="text-sm text-emerald-400">Já tenho conta</Link>
    </form>
  )
}

function AcceptButton({ token }: { token: string }) {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const accept = useMutation({
    mutationFn: () => api.post<{ campaignId: string }>(`/invites/${token}/accept`),
    onSuccess: (m) => {
      queryClient.invalidateQueries({ queryKey: campaignsKey })
      navigate(`/campaigns/${m.campaignId}`, { replace: true })
    },
  })
  return (
    <div className="flex flex-col gap-3">
      <button onClick={() => accept.mutate()} disabled={accept.isPending}
        className="rounded bg-emerald-600 px-3 py-2 font-medium hover:bg-emerald-500 disabled:opacity-50">
        Entrar na campanha
      </button>
      {accept.isError && <p role="alert" className="text-sm text-red-400">{accept.error.message}</p>}
    </div>
  )
}

export function InvitePage() {
  const { token = '' } = useParams()
  const { data: user } = useMe()
  const invite = useQuery({
    queryKey: ['invite', token],
    queryFn: () => api.get<InviteInfo>(`/invites/${token}`),
    retry: false,
  })

  return (
    <main className="flex min-h-screen items-center justify-center bg-zinc-950 text-zinc-100">
      <div className="flex w-96 flex-col gap-4 rounded-lg bg-zinc-900 p-6">
        {invite.isPending && <p className="text-zinc-400">Carregando…</p>}
        {invite.isError && <p role="alert" className="text-red-400">{invite.error.message}</p>}
        {invite.data && (
          <>
            <h1 className="text-2xl font-bold tracking-tight">Convite para {invite.data.campaignName}</h1>
            {user ? <AcceptButton token={token} /> : <SignupForm token={token} />}
          </>
        )}
      </div>
    </main>
  )
}
