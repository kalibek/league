import type { ButtonHTMLAttributes, ReactNode } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger'
  loading?: boolean
  children: ReactNode
}

const variantStyles: Record<string, React.CSSProperties> = {
  primary: {
    backgroundColor: 'var(--orange)',
    color: '#fff',
    border: '1px solid transparent',
  },
  secondary: {
    backgroundColor: 'transparent',
    color: 'var(--navy)',
    border: '1.5px solid var(--navy)',
  },
  danger: {
    backgroundColor: '#dc2626',
    color: '#fff',
    border: '1px solid transparent',
  },
}

const variantHoverClass: Record<string, string> = {
  primary: 'hover:opacity-90 disabled:opacity-50',
  secondary: 'hover:bg-[#0B3C5D] hover:text-white disabled:opacity-40',
  danger: 'hover:opacity-90 disabled:opacity-50',
}

export function Button({
  variant = 'primary',
  loading = false,
  disabled = false,
  children,
  className = '',
  style,
  ...rest
}: ButtonProps) {
  return (
    <button
      {...rest}
      disabled={disabled || loading}
      style={{ ...variantStyles[variant], ...style }}
      className={`inline-flex items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-semibold cursor-pointer disabled:cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[#FF7A00] ${variantHoverClass[variant]} ${className}`}
    >
      {loading && (
        <svg
          className="h-4 w-4"
          style={{ animation: 'spin 1s linear infinite' }}
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
        </svg>
      )}
      {children}
    </button>
  )
}
