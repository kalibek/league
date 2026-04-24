import { createBrowserRouter, Navigate } from 'react-router-dom'
import { RootLayout } from './RootLayout'
import { HomePage } from '../pages/HomePage'
import { LoginPage } from '../pages/LoginPage'
import { PlayersPage } from '../pages/PlayersPage'
import { PlayerProfilePage } from '../pages/PlayerProfilePage'
import { PlayerCreatePage } from '../pages/PlayerCreatePage'
import { PlayerImportPage } from '../pages/PlayerImportPage'
import { LeaguesPage } from '../pages/LeaguesPage'
import { LeaguePage } from '../pages/LeaguePage'
import { LeagueConfigPage } from '../pages/LeagueConfigPage'
import { CreateLeaguePage } from '../pages/CreateLeaguePage'
import { LiveViewPage } from '../pages/LiveViewPage'
import { EventSetupPage } from '../pages/EventSetupPage'
import { ProfileEditPage } from '../pages/ProfileEditPage'

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: <RootLayout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: 'players', element: <PlayersPage /> },
      { path: 'players/new', element: <PlayerCreatePage /> },
      { path: 'players/import', element: <PlayerImportPage /> },
      { path: 'players/:id', element: <PlayerProfilePage /> },
      { path: 'leagues', element: <LeaguesPage /> },
      { path: 'leagues/new', element: <CreateLeaguePage /> },
      { path: 'leagues/:id', element: <LeaguePage /> },
      { path: 'leagues/:id/config', element: <LeagueConfigPage /> },
      { path: 'leagues/:id/events/:eid', element: <LiveViewPage /> },
      { path: 'leagues/:id/events/:eid/setup', element: <EventSetupPage /> },
      { path: 'profile/edit', element: <ProfileEditPage /> },
      { path: '*', element: <Navigate to="/" replace /> },
    ],
  },
])
