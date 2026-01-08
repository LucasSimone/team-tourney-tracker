import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from './composables/useAuth'
import Login from './pages/Login.vue'
import AdminTeams from './pages/admin/Teams.vue'
import AdminPlayers from './pages/admin/Players.vue'
import AdminSeasons from './pages/admin/Seasons.vue'
import AdminMatches from './pages/admin/Matches.vue'
import AdminUsers from './pages/admin/Users.vue'
import AdminBackup from './pages/admin/Backup.vue'
import Track from './pages/Track.vue'
import Standings from './pages/Standings.vue'
import Games from './pages/Games.vue'
import Players from './pages/Players.vue'
import LiveScore from './pages/LiveScore.vue'
import TeamDetail from './pages/TeamDetail.vue'
import PlayerDetail from './pages/PlayerDetail.vue'
import PublicSeasons from './pages/PublicSeasons.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    name: 'Standings',
    component: Standings,
    meta: { requiresAuth: false }
  },
  {
    path: '/standings',
    name: 'StandingsAlias',
    component: Standings,
    meta: { requiresAuth: false }
  },
  {
    path: '/teams/:id',
    name: 'TeamDetail',
    component: TeamDetail,
    meta: { requiresAuth: false }
  },
  {
    path: '/players/:id',
    name: 'PlayerDetail',
    component: PlayerDetail,
    meta: { requiresAuth: false }
  },
  {
    path: '/games',
    name: 'Games',
    component: Games,
    meta: { requiresAuth: false }
  },
  {
    path: '/players',
    name: 'Players',
    component: Players,
    meta: { requiresAuth: false }
  },
  {
    path: '/seasons',
    name: 'PublicSeasons',
    component: PublicSeasons,
    meta: { requiresAuth: false }
  },
  {
    path: '/track',
    name: 'Track',
    component: Track,
    meta: { requiresAuth: true }
  },
  {
    path: '/live-score',
    name: 'LiveScore',
    component: LiveScore,
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/teams',
    name: 'AdminTeams',
    component: AdminTeams,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/admin/players',
    name: 'AdminPlayers',
    component: AdminPlayers,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/admin/seasons',
    name: 'AdminSeasons',
    component: AdminSeasons,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/admin/matches',
    name: 'AdminMatches',
    component: AdminMatches,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/admin/users',
    name: 'AdminUsers',
    component: AdminUsers,
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/admin/backup',
    name: 'AdminBackup',
    component: AdminBackup,
    meta: { requiresAuth: true, requiresAdmin: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const { isLoggedIn, isAdmin } = useAuth()

  if (to.meta.requiresAuth === false) {
    // Public pages - redirect to standings if already logged in and on login page
    if (isLoggedIn() && to.path === '/login') {
      next('/standings')
    } else {
      next()
    }
  } else if (to.meta.requiresAuth) {
    // Protected pages
    if (!isLoggedIn()) {
      next('/login')
    } else if (to.meta.requiresAdmin && !isAdmin()) {
      next('/')
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router
