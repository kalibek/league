import { Link, NavLink, Outlet, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import { AuthProvider, useAuthContext } from '../context/AuthContext'
import { useMyProfile } from '../hooks/useProfile'

function LanguageSwitcher() {
  const { i18n, t } = useTranslation()
  const langs = ['en', 'ru', 'kk'] as const

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
      {langs.map((lang) => (
        <button
          key={lang}
          onClick={() => i18n.changeLanguage(lang)}
          style={{
            fontSize: 11,
            fontWeight: i18n.language === lang ? 700 : 400,
            color: i18n.language === lang ? '#fff' : 'rgba(255,255,255,0.55)',
            background: i18n.language === lang ? 'rgba(255,255,255,0.15)' : 'none',
            border: 'none',
            cursor: 'pointer',
            padding: '3px 7px',
            borderRadius: 4,
            letterSpacing: '0.02em',
          }}
          aria-label={t(`language.${lang}`)}
        >
          {lang.toUpperCase()}
        </button>
      ))}
    </div>
  )
}

function ProfileBannerInner() {
  const location = useLocation()
  const { profile } = useMyProfile()
  const { t } = useTranslation()

  if (!profile || profile.isComplete) return null
  if (location.pathname === '/profile/edit') return null

  return (
    <div style={{ backgroundColor: '#fff8ed', borderBottom: '1px solid #fed7aa' }}>
      <div className="max-w-7xl mx-auto px-4 py-2 flex items-center justify-between text-sm">
        <span style={{ color: '#92400e' }}>{t('profileBanner.message')}</span>
        <Link
          to="/profile/edit"
          style={{ color: '#7c2d12', fontWeight: 600, textDecoration: 'underline' }}
          className="ml-4 whitespace-nowrap hover:opacity-80"
        >
          {t('profileBanner.completeProfile')} →
        </Link>
      </div>
    </div>
  )
}

function ProfileBanner() {
  const { user } = useAuthContext()
  if (!user) return null
  return <ProfileBannerInner />
}

function NavBar() {
  const { user, logout, loading } = useAuthContext()
  const { t } = useTranslation()
  const location = useLocation()
  const [menuOpen, setMenuOpen] = useState(false)

  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => setMenuOpen(false), [location.pathname])

  if (loading) return null

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    `text-sm font-medium transition-opacity ${isActive ? 'text-white opacity-100' : 'text-white opacity-70 hover:opacity-100'}`

  return (
    <nav style={{ backgroundColor: 'var(--navy)' }} className="shadow-lg relative">
      <div className="max-w-7xl mx-auto px-4 flex items-center justify-between h-14">
        {/* Left: logo always visible */}
        <div className="flex items-center gap-7">
          <Link to="/" className="flex items-center gap-2 group">
            <div
              style={{
                backgroundColor: 'var(--orange)',
                width: 32,
                height: 32,
                borderRadius: 6,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontWeight: 800,
                fontSize: 13,
                color: '#fff',
                letterSpacing: '-0.5px',
              }}
            >
              TT
            </div>
            <span
              style={{ color: '#fff', fontWeight: 700, fontSize: 15, letterSpacing: '-0.3px' }}
              className="group-hover:opacity-90"
            >
              {t('nav.tableTennis')}
            </span>
          </Link>

          {/* Nav links: hidden on mobile, visible on sm+ */}
          <div className="hidden sm:flex items-center gap-7">
            <NavLink to="/leagues" className={navLinkClass}>
              {t('nav.leagues')}
            </NavLink>
            <NavLink to="/players" className={navLinkClass}>
              {t('nav.players')}
            </NavLink>
          </div>
        </div>

        {/* Right */}
        <div className="flex items-center gap-3">
          {/* Language + user/signin: hidden on mobile */}
          <div className="hidden sm:flex items-center gap-3">
            <LanguageSwitcher />
            {user ? (
              <>
                <Link
                  to="/profile/edit"
                  className="text-sm font-medium text-white opacity-80 hover:opacity-100"
                >
                  {user.firstName} {user.lastName}
                </Link>
                <button onClick={logout} className="text-sm text-white opacity-60 hover:opacity-90">
                  {t('nav.logout')}
                </button>
              </>
            ) : (
              <Link
                to="/login"
                style={{
                  backgroundColor: 'var(--orange)',
                  color: '#fff',
                  fontWeight: 600,
                  fontSize: 13,
                  padding: '6px 16px',
                  borderRadius: 6,
                }}
                className="hover:opacity-90"
              >
                {t('nav.signIn')}
              </Link>
            )}
          </div>

          {/* Hamburger: only on mobile */}
          <button
            className="sm:hidden flex flex-col justify-center items-center w-9 h-9 gap-1.5"
            onClick={() => setMenuOpen((o) => !o)}
            aria-label={t('nav.toggleMenu')}
            aria-expanded={menuOpen}
          >
            <span
              className={`block w-5 h-0.5 bg-white transition-transform duration-200 ${
                menuOpen ? 'translate-y-2 rotate-45' : ''
              }`}
            />
            <span
              className={`block w-5 h-0.5 bg-white transition-opacity duration-200 ${
                menuOpen ? 'opacity-0' : ''
              }`}
            />
            <span
              className={`block w-5 h-0.5 bg-white transition-transform duration-200 ${
                menuOpen ? '-translate-y-2 -rotate-45' : ''
              }`}
            />
          </button>
        </div>
      </div>

      {/* Mobile dropdown menu */}
      {menuOpen && (
        <div
          className="sm:hidden absolute top-14 left-0 right-0 z-50 shadow-lg"
          style={{
            backgroundColor: 'var(--navy)',
            borderTop: '1px solid rgba(255,255,255,0.1)',
          }}
        >
          <div className="px-4 py-3 flex flex-col gap-1">
            <NavLink to="/leagues" className={navLinkClass} onClick={() => setMenuOpen(false)}>
              {t('nav.leagues')}
            </NavLink>
            <NavLink to="/players" className={navLinkClass} onClick={() => setMenuOpen(false)}>
              {t('nav.players')}
            </NavLink>
            <div className="pt-2 border-t border-white/10 mt-1 flex items-center justify-between">
              <LanguageSwitcher />
              {user ? (
                <div className="flex items-center gap-3">
                  <Link
                    to="/profile/edit"
                    className="text-sm font-medium text-white opacity-80"
                    onClick={() => setMenuOpen(false)}
                  >
                    {user.firstName} {user.lastName}
                  </Link>
                  <button
                    onClick={() => {
                      logout()
                      setMenuOpen(false)
                    }}
                    className="text-sm text-white opacity-60"
                  >
                    {t('nav.logout')}
                  </button>
                </div>
              ) : (
                <Link
                  to="/login"
                  style={{
                    backgroundColor: 'var(--orange)',
                    color: '#fff',
                    fontWeight: 600,
                    fontSize: 13,
                    padding: '6px 16px',
                    borderRadius: 6,
                  }}
                  onClick={() => setMenuOpen(false)}
                >
                  {t('nav.signIn')}
                </Link>
              )}
            </div>
          </div>
        </div>
      )}
    </nav>
  )
}

function Layout() {
  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--surface)' }}>
      <NavBar />
      <ProfileBanner />
      <main>
        <Outlet />
      </main>
    </div>
  )
}

export function RootLayout() {
  return (
    <AuthProvider>
      <Layout />
    </AuthProvider>
  )
}
