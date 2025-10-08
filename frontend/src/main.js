import m from 'mithril'
import Login from './views/Login'
import Dashboard from './views/Dashboard'

// Simple auth check
const requireAuth = {
  onmatch() {
    const user = localStorage.getItem('user')
    if (!user) {
      m.route.set('/login')
      return null
    }
  }
}

// Routes
m.route(document.getElementById('app'), '/login', {
  '/login': Login,
  '/dashboard': {
    onmatch: requireAuth.onmatch,
    render: () => m(Dashboard)
  },
})
