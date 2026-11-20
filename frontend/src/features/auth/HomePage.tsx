import { useMe } from './auth'

export function HomePage() {
  const { data: user } = useMe()
  return <h1 className="text-2xl font-bold tracking-tight">Olá, {user?.displayName}</h1>
}
