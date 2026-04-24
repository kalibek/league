import { useState, type ReactNode } from 'react'

export interface Column<T> {
  key: string
  header: string
  render: (row: T) => ReactNode
  sortable?: boolean
  sortValue?: (row: T) => string | number
}

interface TableProps<T> {
  columns: Column<T>[]
  rows: T[]
  rowKey: (row: T) => string | number
  emptyMessage?: string
}

type SortDir = 'asc' | 'desc'

export function Table<T>({ columns, rows, rowKey, emptyMessage = 'No data' }: TableProps<T>) {
  const [sortKey, setSortKey] = useState<string | null>(null)
  const [sortDir, setSortDir] = useState<SortDir>('asc')

  const handleSort = (col: Column<T>) => {
    if (!col.sortable) return
    if (sortKey === col.key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'))
    } else {
      setSortKey(col.key)
      setSortDir('asc')
    }
  }

  const sortedRows = [...rows].sort((a, b) => {
    if (!sortKey) return 0
    const col = columns.find((c) => c.key === sortKey)
    if (!col?.sortValue) return 0
    const av = col.sortValue(a)
    const bv = col.sortValue(b)
    if (av < bv) return sortDir === 'asc' ? -1 : 1
    if (av > bv) return sortDir === 'asc' ? 1 : -1
    return 0
  })

  return (
    <div className="w-full overflow-x-auto rounded-xl shadow-sm" style={{ border: '1px solid var(--border)' }}>
      <table className="w-full text-sm text-left">
        <thead>
          <tr style={{ backgroundColor: 'var(--navy)' }}>
            {columns.map((col) => (
              <th
                key={col.key}
                style={{ color: 'rgba(255,255,255,0.85)', fontWeight: 600, fontSize: 11, letterSpacing: '0.06em', padding: '12px 16px', textTransform: 'uppercase' }}
                className={col.sortable ? 'cursor-pointer select-none hover:text-white' : ''}
                onClick={() => handleSort(col)}
              >
                {col.header}
                {col.sortable && sortKey === col.key && (
                  <span className="ml-1 text-[#FF7A00]">{sortDir === 'asc' ? '↑' : '↓'}</span>
                )}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {sortedRows.length === 0 ? (
            <tr>
              <td
                colSpan={columns.length}
                style={{ padding: '32px 16px', textAlign: 'center', color: '#94a3b8', backgroundColor: '#fff' }}
              >
                {emptyMessage}
              </td>
            </tr>
          ) : (
            sortedRows.map((row, i) => (
              <tr
                key={rowKey(row)}
                style={{
                  backgroundColor: i % 2 === 0 ? '#fff' : '#fafbfd',
                  borderBottom: '1px solid var(--border)',
                }}
                className="hover:bg-[#f0f4f8] transition-colors"
              >
                {columns.map((col) => (
                  <td key={col.key} style={{ padding: '12px 16px', color: 'var(--dark)' }}>
                    {col.render(row)}
                  </td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}
