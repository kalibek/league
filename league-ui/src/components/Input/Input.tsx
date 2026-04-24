import type { InputHTMLAttributes } from 'react'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export function Input({ label, error, id, className = '', style, ...rest }: InputProps) {
  const inputId = id ?? label?.toLowerCase().replace(/\s+/g, '-')
  return (
    <div className="flex flex-col gap-1">
      {label && (
        <label htmlFor={inputId} style={{ fontSize: 12, fontWeight: 600, color: '#64748b', letterSpacing: '0.03em', textTransform: 'uppercase' }}>
          {label}
        </label>
      )}
      <input
        id={inputId}
        {...rest}
        style={{
          border: `1.5px solid ${error ? '#ef4444' : 'var(--border)'}`,
          borderRadius: 8,
          padding: '9px 12px',
          fontSize: 14,
          color: 'var(--dark)',
          backgroundColor: '#fff',
          outline: 'none',
          width: '100%',
          ...style,
        }}
        className={`focus:ring-2 focus:ring-[#FF7A00] focus:border-[#FF7A00] disabled:bg-gray-50 disabled:text-gray-400 placeholder:text-gray-400 ${className}`}
      />
      {error && <p style={{ fontSize: 12, color: '#ef4444', marginTop: 2 }}>{error}</p>}
    </div>
  )
}
