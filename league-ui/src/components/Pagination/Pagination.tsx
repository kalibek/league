export interface PaginationProps {
  page: number
  pageSize: number
  total: number
  onPageChange: (page: number) => void
}

export function Pagination({ page, pageSize, total, onPageChange }: PaginationProps) {
  const pageCount = Math.ceil(total / pageSize)

  if (pageCount <= 1) {
    return null
  }

  const getPageNumbers = () => {
    const pages: (number | string)[] = []
    const maxButtons = 7
    const halfWindow = 3

    if (pageCount <= maxButtons) {
      for (let i = 1; i <= pageCount; i++) {
        pages.push(i)
      }
    } else {
      pages.push(1)

      if (page > halfWindow + 1) {
        pages.push('...')
      }

      const start = Math.max(2, page - halfWindow)
      const end = Math.min(pageCount - 1, page + halfWindow)

      for (let i = start; i <= end; i++) {
        if (i !== 1 && i !== pageCount) {
          pages.push(i)
        }
      }

      if (page < pageCount - halfWindow - 1) {
        pages.push('...')
      }

      pages.push(pageCount)
    }

    return pages
  }

  const pageNumbers = getPageNumbers()
  const isPrevDisabled = page === 1
  const isNextDisabled = page === pageCount

  const btnStyle = (active: boolean, disabled: boolean): React.CSSProperties => ({
    fontSize: 13,
    fontWeight: active ? 700 : 500,
    color: active ? '#fff' : disabled ? '#cbd5e1' : '#64748b',
    backgroundColor: active ? 'var(--orange)' : disabled ? '#f1f5f9' : 'transparent',
    border: '1.5px solid var(--border)',
    borderRadius: 6,
    padding: '6px 10px',
    cursor: disabled ? 'not-allowed' : 'pointer',
    minWidth: 32,
    height: 32,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    transition: 'all 0.2s',
  })

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 4, justifyContent: 'center' }}>
      <button
        style={btnStyle(false, isPrevDisabled)}
        onClick={() => onPageChange(page - 1)}
        disabled={isPrevDisabled}
        aria-label="Previous page"
      >
        ←
      </button>

      {pageNumbers.map((num, idx) =>
        num === '...' ? (
          <span
            key={`ellipsis-${idx}`}
            style={{
              fontSize: 13,
              color: '#94a3b8',
              padding: '6px 4px',
            }}
          >
            …
          </span>
        ) : (
          <button
            key={num}
            style={btnStyle(num === page, false)}
            onClick={() => onPageChange(num as number)}
          >
            {num}
          </button>
        )
      )}

      <button
        style={btnStyle(false, isNextDisabled)}
        onClick={() => onPageChange(page + 1)}
        disabled={isNextDisabled}
        aria-label="Next page"
      >
        →
      </button>
    </div>
  )
}
