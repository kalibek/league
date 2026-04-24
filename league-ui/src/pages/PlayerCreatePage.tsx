import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useCreatePlayer } from '../hooks/usePlayers'
import { Button } from '../components/Button/Button'
import { Input } from '../components/Input/Input'

export function PlayerCreatePage() {
  const navigate = useNavigate()
  const { create, loading, error } = useCreatePlayer()
  const [form, setForm] = useState({ firstName: '', lastName: '', email: '' })

  const set = (key: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((prev) => ({ ...prev, [key]: e.target.value }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const player = await create(form)
    if (player) navigate(`/players/${player.userId}`)
  }

  return (
    <div className="max-w-md mx-auto py-8 px-4">
      <Link to="/players" className="text-sm text-blue-600 hover:underline mb-4 block">
        &larr; Back to Players
      </Link>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Add Player</h1>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <Input label="First Name" value={form.firstName} onChange={set('firstName')} required />
        <Input label="Last Name" value={form.lastName} onChange={set('lastName')} required />
        <Input label="Email" type="email" value={form.email} onChange={set('email')} required />
        {error && <p className="text-sm text-red-600">{error}</p>}
        <div className="flex gap-2">
          <Button type="button" variant="secondary" onClick={() => navigate('/players')}>
            Cancel
          </Button>
          <Button type="submit" loading={loading}>
            Create Player
          </Button>
        </div>
      </form>
    </div>
  )
}
