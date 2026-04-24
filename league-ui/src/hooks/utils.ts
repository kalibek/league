export { useDebounce } from './useDebounce'

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

export function extractErrorMessage(e: unknown): string {
  if (e && typeof e === 'object') {
    const err = e as { response?: { data?: { message?: string } }; message?: string }
    return err.response?.data?.message ?? err.message ?? 'Unknown error'
  }
  return String(e)
}
