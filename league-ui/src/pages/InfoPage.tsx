import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'

export function InfoPage() {
  const { slug } = useParams<{ slug: string }>()
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const [content, setContent] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!slug) {
      navigate('/')
      return
    }

    const fetchMarkdown = async () => {
      setLoading(true)
      setError(null)
      setContent(null)

      // Determine language with fallback to English
      let lang = i18n.language
      if (!['en', 'ru', 'kk'].includes(lang)) {
        lang = 'en'
      }

      try {
        const response = await fetch(`/docs/${lang}/${slug}.md`)
        if (!response.ok) {
          if (response.status === 404) {
            setError(t('info.notFound'))
          } else {
            setError(t('common.error'))
          }
          setContent(null)
        } else {
          const text = await response.text()
          setContent(text)
        }
      } catch (err) {
        setError(t('common.error'))
        setContent(null)
      } finally {
        setLoading(false)
      }
    }

    fetchMarkdown()
  }, [slug, i18n.language, navigate, t])

  const customMarkdownComponents = {
    h1: ({ children }: any) => (
      <h1 style={{ fontSize: 28, fontWeight: 800, color: 'var(--navy)', marginBottom: 8 }}>
        {children}
      </h1>
    ),
    h2: ({ children }: any) => (
      <h2 style={{ fontSize: 20, fontWeight: 700, color: 'var(--navy)', marginTop: 28, marginBottom: 10 }}>
        {children}
      </h2>
    ),
    p: ({ children }: any) => (
      <p style={{ fontSize: 15, lineHeight: 1.7, color: '#374151', marginBottom: 12 }}>
        {children}
      </p>
    ),
    ul: ({ children }: any) => (
      <ul style={{ paddingLeft: 20, marginBottom: 12 }}>
        {children}
      </ul>
    ),
    ol: ({ children }: any) => (
      <ol style={{ paddingLeft: 20, marginBottom: 12 }}>
        {children}
      </ol>
    ),
    li: ({ children }: any) => (
      <li style={{ fontSize: 15, lineHeight: 1.7, color: '#374151', marginBottom: 4 }}>
        {children}
      </li>
    ),
    strong: ({ children }: any) => (
      <strong style={{ fontWeight: 700, color: 'var(--navy)' }}>
        {children}
      </strong>
    ),
    code: ({ inline, children }: any) => (
      <code
        style={{
          backgroundColor: '#f1f5f9',
          padding: inline ? '2px 6px' : '12px',
          borderRadius: 4,
          fontSize: 13,
          fontFamily: 'monospace',
          display: inline ? 'inline' : 'block',
          overflow: 'auto',
        }}
      >
        {children}
      </code>
    ),
    table: ({ children }: any) => (
      <table
        style={{
          width: '100%',
          borderCollapse: 'collapse',
          marginBottom: 16,
        }}
      >
        {children}
      </table>
    ),
    thead: ({ children }: any) => <thead>{children}</thead>,
    tbody: ({ children }: any) => <tbody>{children}</tbody>,
    th: ({ children }: any) => (
      <th
        style={{
          padding: '8px 12px',
          border: '1px solid var(--border)',
          fontSize: 14,
          fontWeight: 700,
          backgroundColor: '#f8fafc',
        }}
      >
        {children}
      </th>
    ),
    td: ({ children }: any) => (
      <td
        style={{
          padding: '8px 12px',
          border: '1px solid var(--border)',
          fontSize: 14,
        }}
      >
        {children}
      </td>
    ),
    blockquote: ({ children }: any) => (
      <blockquote
        style={{
          borderLeft: '3px solid var(--orange)',
          paddingLeft: 16,
          color: '#64748b',
          fontStyle: 'italic',
        }}
      >
        {children}
      </blockquote>
    ),
  }

  return (
    <div className="max-w-2xl mx-auto py-10 px-4">
      <button
        onClick={() => navigate('/')}
        style={{
          background: 'none',
          border: 'none',
          color: 'var(--orange)',
          fontSize: 13,
          fontWeight: 600,
          cursor: 'pointer',
          padding: 0,
          marginBottom: 24,
          textDecoration: 'none',
        }}
        className="hover:opacity-80 transition-opacity"
      >
        {t('info.backToHome')}
      </button>

      {loading && (
        <div style={{ color: '#94a3b8', fontSize: 14, padding: '40px 0', textAlign: 'center' }}>
          {t('info.loading')}
        </div>
      )}

      {error && (
        <div
          style={{
            color: '#dc2626',
            backgroundColor: '#fee2e2',
            borderRadius: 8,
            padding: '10px 14px',
            fontSize: 13,
          }}
        >
          {error}
        </div>
      )}

      {content && !loading && !error && (
        <ReactMarkdown components={customMarkdownComponents}>
          {content}
        </ReactMarkdown>
      )}
    </div>
  )
}
