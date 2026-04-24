import { Outlet, Link, useLocation, NavLink } from 'react-router-dom'
import { AuthProvider, useAuthContext } from '../context/AuthContext'
import { useMyProfile } from '../hooks/useProfile'

function ProfileBannerInner() {
  const location = useLocation()
  const { profile } = useMyProfile()

  if (!profile || profile.isComplete) return null
  if (location.pathname === '/profile/edit') return null

  return (
    <div style={{ backgroundColor: '#fff8ed', borderBottom: '1px solid #fed7aa' }}>
      <div className="max-w-7xl mx-auto px-4 py-2 flex items-center justify-between text-sm">
        <span style={{ color: '#92400e' }}>
          Your profile is incomplete. Fill in your details to get the most out of the platform.
        </span>
        <Link
          to="/profile/edit"
          style={{ color: '#7c2d12', fontWeight: 600, textDecoration: 'underline' }}
          className="ml-4 whitespace-nowrap hover:opacity-80"
        >
          Complete Profile →
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

  if (loading) return null

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    `text-sm font-medium transition-opacity ${isActive ? 'text-white opacity-100' : 'text-white opacity-70 hover:opacity-100'}`

  return (
    <nav style={{ backgroundColor: 'var(--navy)' }} className="shadow-lg">
      <div className="max-w-7xl mx-auto px-4 flex items-center justify-between h-14">
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
              Table Tennis
            </span>
          </Link>
          <NavLink to="/leagues" className={navLinkClass}>
            Leagues
          </NavLink>
          <NavLink to="/players" className={navLinkClass}>
            Players
          </NavLink>
        </div>

        <div className="flex items-center gap-3">
          {user ? (
            <>
              <Link
                to="/profile/edit"
                className="text-sm font-medium text-white opacity-80 hover:opacity-100"
              >
                {user.firstName} {user.lastName}
              </Link>
              <button
                onClick={logout}
                className="text-sm text-white opacity-60 hover:opacity-90"
              >
                Logout
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
              Sign In
            </Link>
          )}
        </div>
      </div>
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
