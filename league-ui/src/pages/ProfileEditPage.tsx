import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMyProfile, useUpsertProfile, useCountries, useCities, useBlades, useRubbers } from '../hooks/useProfile'
import { Button } from '../components/Button/Button'
import type { Country, City, Blade, Rubber } from '../types'

const inputStyle: React.CSSProperties = {
  width: '100%',
  border: '1.5px solid var(--border)',
  borderRadius: 8,
  padding: '9px 12px',
  fontSize: 14,
  color: 'var(--dark)',
  backgroundColor: '#fff',
  outline: 'none',
}

const labelStyle: React.CSSProperties = {
  display: 'block',
  fontSize: 11,
  fontWeight: 700,
  color: '#64748b',
  letterSpacing: '0.05em',
  textTransform: 'uppercase',
  marginBottom: 5,
}

const selectStyle: React.CSSProperties = {
  flex: 1,
  border: '1.5px solid var(--border)',
  borderRadius: 8,
  padding: '9px 12px',
  fontSize: 14,
  color: 'var(--dark)',
  backgroundColor: '#fff',
  outline: 'none',
  appearance: 'none',
  backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2394a3b8'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'/%3E%3C/svg%3E")`,
  backgroundRepeat: 'no-repeat',
  backgroundPosition: 'right 10px center',
  backgroundSize: '16px',
  paddingRight: 34,
}

type AddNewState = { show: boolean; value: string }

function AddNewInput({
  state,
  onChange,
  onConfirm,
  onCancel,
  placeholder,
}: {
  state: AddNewState
  onChange: (v: string) => void
  onConfirm: () => void
  onCancel: () => void
  placeholder: string
}) {
  if (!state.show) return null
  return (
    <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
      <input
        autoFocus
        type="text"
        style={{ flex: 1, border: '1.5px solid var(--border)', borderRadius: 8, padding: '7px 10px', fontSize: 13, outline: 'none' }}
        className="focus:border-[#FF7A00]"
        placeholder={placeholder}
        value={state.value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') onConfirm()
          if (e.key === 'Escape') onCancel()
        }}
      />
      <button
        type="button"
        style={{ fontSize: 12, padding: '7px 14px', backgroundColor: 'var(--orange)', color: '#fff', borderRadius: 8, border: 'none', cursor: 'pointer', fontWeight: 600 }}
        onClick={onConfirm}
      >
        Add
      </button>
      <button
        type="button"
        style={{ fontSize: 12, padding: '7px 12px', border: '1.5px solid var(--border)', borderRadius: 8, cursor: 'pointer', backgroundColor: '#fff', color: '#64748b' }}
        onClick={onCancel}
      >
        Cancel
      </button>
    </div>
  )
}

function SelectWithAdd<T extends { name: string }>({
  label,
  items,
  valueKey,
  value,
  onChange,
  onAdd,
  addPlaceholder,
  disabled,
}: {
  label: string
  items: T[]
  valueKey: keyof T
  value: number | null
  onChange: (id: number | null) => void
  onAdd?: (name: string) => Promise<T | null>
  addPlaceholder: string
  disabled?: boolean
}) {
  const [addState, setAddState] = useState<AddNewState>({ show: false, value: '' })

  const handleConfirm = async () => {
    if (!addState.value.trim() || !onAdd) return
    const result = await onAdd(addState.value.trim())
    if (result) {
      onChange(result[valueKey] as number)
      setAddState({ show: false, value: '' })
    }
  }

  return (
    <div>
      <label style={labelStyle}>{label}</label>
      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <select
          style={{ ...selectStyle, opacity: disabled ? 0.6 : 1, cursor: disabled ? 'not-allowed' : 'default' }}
          value={value ?? ''}
          onChange={(e) => onChange(e.target.value ? Number(e.target.value) : null)}
          disabled={disabled}
        >
          <option value="">— select —</option>
          {items.map((item) => (
            <option key={String(item[valueKey])} value={String(item[valueKey])}>
              {item.name}
            </option>
          ))}
        </select>
        {onAdd && !addState.show && (
          <button
            type="button"
            style={{ fontSize: 12, color: 'var(--orange)', fontWeight: 600, background: 'none', border: 'none', cursor: 'pointer', whiteSpace: 'nowrap' }}
            onClick={() => setAddState({ show: true, value: '' })}
          >
            + Add new
          </button>
        )}
      </div>
      {onAdd && (
        <AddNewInput
          state={addState}
          onChange={(v) => setAddState((s) => ({ ...s, value: v }))}
          onConfirm={handleConfirm}
          onCancel={() => setAddState({ show: false, value: '' })}
          placeholder={addPlaceholder}
        />
      )}
    </div>
  )
}

export function ProfileEditPage() {
  const navigate = useNavigate()
  const { profile, loading: profileLoading } = useMyProfile()
  const { save, loading: saving, error: saveError } = useUpsertProfile()

  const { countries } = useCountries()
  const [selectedCountryId, setSelectedCountryId] = useState<number | null>(null)
  const { cities, add: addCity } = useCities(selectedCountryId)
  const { blades, add: addBlade } = useBlades()
  const { rubbers, add: addRubber } = useRubbers()

  const [form, setForm] = useState({
    firstName: '',
    lastName: '',
    countryId: null as number | null,
    cityId: null as number | null,
    birthdate: '',
    grip: '' as '' | 'penhold' | 'shakehand',
    gender: '' as '' | 'male' | 'female' | 'other',
    bladeId: null as number | null,
    fhRubberId: null as number | null,
    bhRubberId: null as number | null,
  })

  useEffect(() => {
    if (!profile) return
    setForm({
      firstName: profile.firstName ?? '',
      lastName: profile.lastName ?? '',
      countryId: profile.country?.countryId ?? null,
      cityId: profile.city?.cityId ?? null,
      birthdate: profile.birthdate ? profile.birthdate.slice(0, 10) : '',
      grip: (profile.grip as typeof form.grip) ?? '',
      gender: (profile.gender as typeof form.gender) ?? '',
      bladeId: profile.blade?.bladeId ?? null,
      fhRubberId: profile.fhRubber?.rubberId ?? null,
      bhRubberId: profile.bhRubber?.rubberId ?? null,
    })
    if (profile.country?.countryId) {
      setSelectedCountryId(profile.country.countryId)
    }
  }, [profile])

  const handleCountryChange = (id: number | null) => {
    setForm((f) => ({ ...f, countryId: id, cityId: null }))
    setSelectedCountryId(id)
  }

  const handleAddCity = async (name: string): Promise<City | null> => addCity(name)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const result = await save({
      firstName: form.firstName,
      lastName: form.lastName,
      countryId: form.countryId,
      cityId: form.cityId,
      birthdate: form.birthdate || null,
      grip: form.grip || null,
      gender: form.gender || null,
      bladeId: form.bladeId,
      fhRubberId: form.fhRubberId,
      bhRubberId: form.bhRubberId,
    })
    if (result) navigate('/')
  }

  if (profileLoading) return (
    <div style={{ padding: '48px 16px', textAlign: 'center', color: '#94a3b8', fontSize: 14 }}>
      Loading profile…
    </div>
  )

  const sectionLabel: React.CSSProperties = {
    fontSize: 10,
    fontWeight: 700,
    color: '#94a3b8',
    letterSpacing: '0.08em',
    textTransform: 'uppercase',
    marginBottom: 14,
    marginTop: 4,
  }

  const radioLabel: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 6,
    fontSize: 13,
    fontWeight: 500,
    color: 'var(--dark)',
    cursor: 'pointer',
  }

  return (
    <div className="max-w-2xl mx-auto py-10 px-4">
      <h1 style={{ fontSize: 26, fontWeight: 800, color: 'var(--navy)', letterSpacing: '-0.5px', marginBottom: 24 }}>
        Edit Profile
      </h1>

      <form
        onSubmit={handleSubmit}
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: 16,
          backgroundColor: '#fff',
          borderRadius: 16,
          border: '1px solid var(--border)',
          padding: '24px 28px',
          boxShadow: '0 2px 8px rgba(11,60,93,0.06)',
        }}
      >
        <p style={sectionLabel}>Personal Info</p>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
          <div>
            <label style={labelStyle}>First Name</label>
            <input
              type="text"
              style={inputStyle}
              className="focus:border-[#FF7A00] focus:ring-2 focus:ring-[#FF7A00]/20"
              value={form.firstName}
              onChange={(e) => setForm((f) => ({ ...f, firstName: e.target.value }))}
            />
          </div>
          <div>
            <label style={labelStyle}>Last Name</label>
            <input
              type="text"
              style={inputStyle}
              className="focus:border-[#FF7A00] focus:ring-2 focus:ring-[#FF7A00]/20"
              value={form.lastName}
              onChange={(e) => setForm((f) => ({ ...f, lastName: e.target.value }))}
            />
          </div>
        </div>

        <SelectWithAdd<Country>
          label="Country"
          items={countries}
          valueKey="countryId"
          value={form.countryId}
          onChange={handleCountryChange}
          addPlaceholder="Country name"
          disabled={false}
        />

        <SelectWithAdd<City>
          label="City"
          items={cities}
          valueKey="cityId"
          value={form.cityId}
          onChange={(id) => setForm((f) => ({ ...f, cityId: id }))}
          onAdd={form.countryId ? handleAddCity : undefined}
          addPlaceholder="City name"
          disabled={!form.countryId}
        />

        <div>
          <label style={labelStyle}>Birthdate</label>
          <input
            type="date"
            style={inputStyle}
            className="focus:border-[#FF7A00] focus:ring-2 focus:ring-[#FF7A00]/20"
            value={form.birthdate}
            onChange={(e) => setForm((f) => ({ ...f, birthdate: e.target.value }))}
          />
        </div>

        <div>
          <label style={labelStyle}>Grip</label>
          <div style={{ display: 'flex', gap: 20 }}>
            {(['penhold', 'shakehand'] as const).map((g) => (
              <label key={g} style={radioLabel}>
                <input
                  type="radio"
                  name="grip"
                  value={g}
                  checked={form.grip === g}
                  onChange={() => setForm((f) => ({ ...f, grip: g }))}
                  style={{ accentColor: 'var(--orange)' }}
                />
                {g.charAt(0).toUpperCase() + g.slice(1)}
              </label>
            ))}
          </div>
        </div>

        <div>
          <label style={labelStyle}>Gender</label>
          <div style={{ display: 'flex', gap: 20 }}>
            {(['male', 'female', 'other'] as const).map((g) => (
              <label key={g} style={radioLabel}>
                <input
                  type="radio"
                  name="gender"
                  value={g}
                  checked={form.gender === g}
                  onChange={() => setForm((f) => ({ ...f, gender: g }))}
                  style={{ accentColor: 'var(--orange)' }}
                />
                {g.charAt(0).toUpperCase() + g.slice(1)}
              </label>
            ))}
          </div>
        </div>

        <div style={{ borderTop: '1px solid var(--border)', paddingTop: 16, marginTop: 4 }}>
          <p style={sectionLabel}>Equipment Setup</p>
        </div>

        <SelectWithAdd<Blade>
          label="Blade"
          items={blades}
          valueKey="bladeId"
          value={form.bladeId}
          onChange={(id) => setForm((f) => ({ ...f, bladeId: id }))}
          onAdd={addBlade}
          addPlaceholder="Blade name (e.g. Timo Boll ALC)"
        />

        <SelectWithAdd<Rubber>
          label="Forehand Rubber"
          items={rubbers}
          valueKey="rubberId"
          value={form.fhRubberId}
          onChange={(id) => setForm((f) => ({ ...f, fhRubberId: id }))}
          onAdd={addRubber}
          addPlaceholder="Rubber name (e.g. Tenergy 05)"
        />

        <SelectWithAdd<Rubber>
          label="Backhand Rubber"
          items={rubbers}
          valueKey="rubberId"
          value={form.bhRubberId}
          onChange={(id) => setForm((f) => ({ ...f, bhRubberId: id }))}
          onAdd={addRubber}
          addPlaceholder="Rubber name (e.g. MXP)"
        />

        {saveError && (
          <div style={{ color: '#dc2626', backgroundColor: '#fee2e2', borderRadius: 8, padding: '10px 14px', fontSize: 13 }}>
            {saveError}
          </div>
        )}

        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, paddingTop: 8 }}>
          <Button variant="secondary" type="button" onClick={() => navigate(-1)}>
            Cancel
          </Button>
          <Button variant="primary" type="submit" loading={saving}>
            Save Profile
          </Button>
        </div>
      </form>
    </div>
  )
}
