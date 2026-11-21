import { Route, Routes } from 'react-router'
import { AdminPage } from './features/admin/AdminPage'
import { RequireAdmin, RequireAuth } from './features/auth/guards'
import { HomePage } from './features/auth/HomePage'
import { CampaignPage } from './features/campaigns/CampaignPage'
import { InvitePage } from './features/campaigns/InvitePage'
import { Layout } from './features/auth/Layout'
import { LoginPage } from './features/auth/LoginPage'
import { DiceDemo } from './features/dice3d/DiceDemo'
import { MapDemo } from './features/map3d/MapDemo'

export default function App() {
  if (location.search.includes('map-demo')) return <MapDemo />
  if (location.search.includes('dice-demo')) return <DiceDemo />
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/invite/:token" element={<InvitePage />} />
      <Route element={<RequireAuth />}>
        <Route element={<Layout />}>
          <Route index element={<HomePage />} />
          <Route path="campaigns/:id" element={<CampaignPage />} />
          <Route element={<RequireAdmin />}>
            <Route path="admin" element={<AdminPage />} />
          </Route>
        </Route>
      </Route>
    </Routes>
  )
}
