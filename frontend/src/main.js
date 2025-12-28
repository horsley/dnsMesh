import m from 'mithril'
import Dashboard from './views/Dashboard'

// Routes - All routes go to dashboard since auth is handled by reverse proxy
m.route(document.getElementById('app'), '/dashboard', {
  '/dashboard': Dashboard,
})
