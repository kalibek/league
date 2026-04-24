type BadgeVariant = 'DRAFT' | 'IN_PROGRESS' | 'DONE' | 'DNS'

interface BadgeProps {
  variant: BadgeVariant
  label?: string
}

const variantStyles: Record<BadgeVariant, React.CSSProperties> = {
  DRAFT: { backgroundColor: '#fef3c7', color: '#92400e', border: '1px solid #fcd34d' },
  IN_PROGRESS: { backgroundColor: '#dbeafe', color: '#1e40af', border: '1px solid #93c5fd' },
  DONE: { backgroundColor: '#dcfce7', color: '#166534', border: '1px solid #86efac' },
  DNS: { backgroundColor: '#fee2e2', color: '#991b1b', border: '1px solid #fca5a5' },
}

const defaultLabels: Record<BadgeVariant, string> = {
  DRAFT: 'Draft',
  IN_PROGRESS: 'In Progress',
  DONE: 'Done',
  DNS: 'DNS',
}

export function Badge({ variant, label }: BadgeProps) {
  return (
    <span
      style={{
        ...variantStyles[variant],
        display: 'inline-flex',
        alignItems: 'center',
        borderRadius: 9999,
        padding: '2px 10px',
        fontSize: 11,
        fontWeight: 600,
        letterSpacing: '0.02em',
        whiteSpace: 'nowrap',
      }}
    >
      {label ?? defaultLabels[variant]}
    </span>
  )
}
