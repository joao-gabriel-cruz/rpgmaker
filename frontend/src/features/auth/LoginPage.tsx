import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type FormEvent } from 'react'
import { Navigate, useNavigate } from 'react-router'
import { login, meKey, useMe } from './auth'

type FieldProps = { label: string; type: string; value: string; onChange: (v: string) => void }

function Field({ label, type, value, onChange }: FieldProps) {
  return (
    <label className="flex flex-col gap-1 text-sm">
      {label}
      <input type={type} required value={value} onChange={(e) => onChange(e.target.value)}
        className="rounded bg-zinc-800 px-3 py-2" />
    </label>
  )
}

function useLogin() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) => login(email, password),
    onSuccess: (user) => {
      queryClient.setQueryData(meKey, user)
      navigate('/', { replace: true })
    },
  })
}

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const { data: current } = useMe()
  const mutation = useLogin()

  if (current) return <Navigate to="/" replace />

  const submit = (e: FormEvent) => {
    e.preventDefault()
    mutation.mutate({ email, password })
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-zinc-950 text-zinc-100">
      <form onSubmit={submit} className="flex w-80 flex-col gap-4 rounded-lg bg-zinc-900 p-6">
        <h1 className="text-2xl font-bold tracking-tight">Entrar</h1>
        <Field label="E-mail" type="email" value={email} onChange={setEmail} />
        <Field label="Senha" type="password" value={password} onChange={setPassword} />
        {mutation.isError && (
          <p role="alert" className="text-sm text-red-400">{mutation.error.message}</p>
        )}
        <button type="submit" disabled={mutation.isPending}
          className="rounded bg-emerald-600 px-3 py-2 font-medium hover:bg-emerald-500 disabled:opacity-50">
          Entrar
        </button>
      </form>
    </main>
  )
}
