interface RatingDeltaProps {
  delta: number
}

export function RatingDelta({ delta }: RatingDeltaProps) {
  if (delta === 0) {
    return <span style={{ color: '#94a3b8', fontSize: 13, fontWeight: 500 }}>±0</span>
  }
  const isPositive = delta > 0
  return (
    <span
      style={{
        fontSize: 13,
        fontWeight: 700,
        color: isPositive ? '#16a34a' : '#dc2626',
        backgroundColor: isPositive ? '#dcfce7' : '#fee2e2',
        padding: '2px 8px',
        borderRadius: 4,
      }}
    >
      {isPositive ? '+' : ''}{delta}
    </span>
  )
}
