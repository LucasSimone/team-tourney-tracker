<template>
  <div class="dark">
    <header class="navbar" :class="{ 'nav-open': menuOpen }">
      <div class="navbar-header">
        <h1 v-if="pageTitle">{{ pageTitle }}</h1>
        <button class="hamburger" :class="{ active: menuOpen }" @click="menuOpen = !menuOpen">
          <span></span>
          <span></span>
          <span></span>
        </button>
      </div>

      <nav class="nav-menu" :class="{ active: menuOpen }">
        <div class="nav-public">
          <router-link to="/standings" @click="menuOpen = false">Standings</router-link>
          <router-link to="/players" @click="menuOpen = false">Players</router-link>
          <router-link to="/games" @click="menuOpen = false">Games</router-link>
          <router-link to="/seasons" @click="menuOpen = false">Seasons</router-link>
        </div>

        <div v-if="isLoggedInUser" class="nav-authenticated">
          <router-link to="/track" @click="menuOpen = false">Track Match</router-link>
          <router-link to="/live-score" @click="menuOpen = false">Live Score</router-link>
          
          <div v-if="isAdminUser" class="admin-dropdown" :class="{ open: adminMenuOpen }">
            <button class="admin-btn" @click="adminMenuOpen = !adminMenuOpen">Admin</button>
            <div class="dropdown-menu">
              <router-link to="/admin/players" @click="closeMenus">Players</router-link>
              <router-link to="/admin/teams" @click="closeMenus">Teams</router-link>
              <router-link to="/admin/matches" @click="closeMenus">Matches</router-link>
              <router-link to="/admin/seasons" @click="closeMenus">Seasons</router-link>
              <router-link to="/admin/users" @click="closeMenus">Users</router-link>
              <router-link to="/admin/backup" @click="closeMenus">Backups</router-link>
            </div>
          </div>
        </div>

        <div class="nav-footer">
          <div v-if="user" class="user-info">
            {{ user.username }} ({{ user.role }})
          </div>
          <button v-if="isLoggedInUser" @click="handleLogout" class="btn-logout">Logout</button>
          <router-link v-else to="/login" class="btn-login" @click="menuOpen = false">Login</router-link>
        </div>
      </nav>
    </header>

    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, provide } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const route = useRoute()
const menuOpen = ref(false)
const adminMenuOpen = ref(false)
const detailPageTitle = ref<string>('')

const { user, isAdmin, isLoggedIn, initialize, logout } = useAuth()

const isAdminUser = computed(() => isAdmin())
const isLoggedInUser = computed(() => isLoggedIn())

// Provide a function for detail pages to update the navbar title
const updateDetailTitle = (title: string) => {
  detailPageTitle.value = title
}
provide('updateDetailTitle', updateDetailTitle)

const pageTitle = computed(() => {
  const routeName = route.name as string
  
  // Map route names to display titles
  const titleMap: Record<string, string> = {
    'Login': 'Login',
    'Standings': 'Standings',
    'StandingsAlias': 'Standings',
    'TeamDetail': detailPageTitle.value,
    'PlayerDetail': detailPageTitle.value,
    'Games': 'Games',
    'Players': 'Players',
    'PublicSeasons': 'Seasons',
    'Track': 'Track Match',
    'LiveScore': 'Live Score',
    'AdminTeams': 'Admin: Teams',
    'AdminPlayers': 'Admin: Players',
    'AdminSeasons': 'Admin: Seasons',
    'AdminMatches': 'Admin: Matches',
    'AdminUsers': 'Admin: Users',
    'AdminBackup': 'Admin: Backups'
  }
  
  return titleMap[routeName] ?? 'Tourney Tracker'
})

onMounted(() => {
  initialize()
})

const handleLogout = () => {
  logout()
  router.push('/login')
}

const closeMenus = () => {
  menuOpen.value = false
  adminMenuOpen.value = false
}
</script>

<style>
:root {
  /* Base colors */
  --clr-dark-a0: #000000;
  --clr-light-a0: #ffffff;

  /* Theme primary colors */
  --clr-primary-a0: #5b18c7;
  --clr-primary-a10: #7237ce;
  --clr-primary-a20: #8750d5;
  --clr-primary-a30: #9a69dc;
  --clr-primary-a40: #ac81e2;
  --clr-primary-a50: #be9ae8;

  /* Theme surface colors */
  --clr-surface-a0: #121212;
  --clr-surface-a10: #282828;
  --clr-surface-a20: #3f3f3f;
  --clr-surface-a30: #575757;
  --clr-surface-a40: #717171;
  --clr-surface-a50: #8b8b8b;

  /* Theme tonal surface colors */
  --clr-surface-tonal-a0: #1b1522;
  --clr-surface-tonal-a10: #302a37;
  --clr-surface-tonal-a20: #47414d;
  --clr-surface-tonal-a30: #5e5964;
  --clr-surface-tonal-a40: #77727c;
  --clr-surface-tonal-a50: #918d94;

  /* Success colors */
  --clr-success-a0: #22946e;
  --clr-success-a10: #47d5a6;
  --clr-success-a20: #9ae8ce;

  /* Warning colors */
  --clr-warning-a0: #a87a2a;
  --clr-warning-a10: #d7ac61;
  --clr-warning-a20: #ecd7b2;

  /* Danger colors */
  --clr-danger-a0: #9c2121;
  --clr-danger-a10: #d94a4a;
  --clr-danger-a20: #eb9e9e;

  /* Info colors */
  --clr-info-a0: #21498a;
  --clr-info-a10: #4077d1;
  --clr-info-a20: #92b2e5;

  /* Spacing */
  --spacing-sm: 8px;
  --spacing-md: 12px;
  --spacing-lg: 16px;
  --spacing-xl: 20px;
  --radius: 4px;
}

div.dark {
  --text-primary: #ffffff;
  --text-secondary: #e0e0e0;
  --border: #404040;
  --shadow: rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  height: 100%;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  background-color: var(--clr-surface-a0);
  color: var(--text-primary);
  transition: background-color 0.2s, color 0.2s;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: fixed;
  width: 100%;
  overflow: hidden;
}

h1, h2, h3, h4, h5, h6 {
  color: var(--text-primary);
}

.navbar {
  background-color: var(--clr-surface-a10);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 100;
  box-shadow: 0 2px 4px var(--shadow);
  display: flex;
  flex-direction: column;
}

.navbar-header {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 0px var(--spacing-lg);
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--clr-surface-a20);
}

.navbar h1 {
  font-size: 24px;
  margin: 0;
  color: var(--text-primary);
}

.hamburger {
  display: flex;
  flex-direction: column;
  gap: 5px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  width: 49px;
  height: 49px;
  overflow: hidden;
  justify-content: center;
  align-items: center;
}

.hamburger span {
  width: 25px;
  height: 3px;
  background-color: var(--text-primary);
  border-radius: 2px;
  transition: all 0.3s;
}

.hamburger.active span:nth-child(1) {
  transform: rotate(45deg) translate(6px, 6px);
}

.hamburger.active span:nth-child(2) {
  opacity: 0;
}

.hamburger.active span:nth-child(3) {
  transform: rotate(-45deg) translate(6px, -6px);
}

.nav-menu {
  position: absolute;
  top: calc(100% + 1px);
  left: 0;
  right: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  align-items: flex-start;
  background-color: var(--clr-surface-a10);
  max-height: 0;
  overflow: hidden;
  transition: max-height 0.3s ease-in-out;
  padding: 0;
  margin: 0;
  width: 100%;
  z-index: 101;
}

.nav-menu.active {
  max-height: calc(100vh - 80px);
  padding: var(--spacing-lg);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.nav-public,
.nav-authenticated {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  width: 100%;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.nav-public a,
.nav-authenticated a,
.admin-btn {
  width: 100%;
  padding: var(--spacing-md) var(--spacing-lg);
  text-align: left;
  margin: 0;
}

.nav-menu a {
  color: var(--text-primary);
  text-decoration: none;
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: var(--radius);
  transition: background-color 0.2s, border 0.2s;
  white-space: normal;
  font-size: 16px;
  font-weight: 500;
  border: 2px solid var(--text-primary);
  margin: 0;
  width: 100%;
}

.nav-menu a.btn-login,
.nav-menu a.btn-logout {
  border: none;
  padding: var(--spacing-md) var(--spacing-lg);
  font-size: 14px;
  font-weight: 400;
  white-space: normal;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-menu a:hover {
  background-color: var(--clr-surface-a20);
}

.nav-menu a.router-link-active:not(.btn-login) {
  background-color: transparent;
  color: white;
  border-color: var(--clr-primary-a0);
}

/* Admin Dropdown */
.admin-dropdown {
  position: relative;
  width: 100%;
  margin-bottom: var(--spacing-lg);
}

.admin-btn {
  color: var(--text-primary);
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: var(--radius);
  transition: background-color 0.2s;
  white-space: normal;
  font-size: 16px;
  font-weight: 500;
  border: 2px solid var(--text-primary);
  background-color: transparent;
  cursor: pointer;
  margin: 0;
  width: 100%;
  text-align: left;
}

.admin-btn:hover {
  background-color: var(--clr-surface-a20);
}

.admin-dropdown.open .admin-btn {
  background-color: var(--clr-surface-a20);
}

.dropdown-menu {
  position: static;
  top: auto;
  left: auto;
  background-color: transparent;
  border: none;
  border-radius: var(--radius);
  margin-top: var(--spacing-md);
  margin-left: 0;
  min-width: auto;
  display: none;
  flex-direction: column;
  z-index: 1000;
  gap: var(--spacing-md);
  width: 100%;
}

.admin-dropdown.open .dropdown-menu {
  display: flex;
}

.dropdown-menu a {
  color: var(--text-primary);
  text-decoration: none;
  padding: var(--spacing-md) var(--spacing-lg);
  border: 2px solid var(--text-primary);
  transition: background-color 0.2s, border 0.2s;
  white-space: normal;
  font-size: 16px;
  font-weight: 500;
  border-radius: var(--radius);
  margin: 0;
  width: 100%;
}

.dropdown-menu a:first-child {
  border-radius: var(--radius);
}

.dropdown-menu a:last-child {
  border-radius: var(--radius);
}

.dropdown-menu a:hover {
  background-color: var(--clr-surface-a20);
}

.dropdown-menu a.router-link-active {
  background-color: transparent;
  color: white;
  border-color: var(--clr-primary-a0);
}


.nav-footer {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: var(--spacing-md);
  width: 100%;
  margin-top: auto;
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--clr-surface-a20);
}

.user-info {
  font-size: 12px;
  color: var(--text-secondary);
  text-align: center;
}

.btn-logout,
.btn-login {
  color: white;
  border: none;
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
  text-decoration: none;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  width: 100%;
  margin-left: 0;
}

.btn-login,
.btn-logout {
  background-color: var(--clr-primary-a0);
}

.btn-login:hover,
.btn-logout:hover {
  background-color: var(--clr-primary-a10);
}

.main-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--spacing-lg);
  flex: 1;
  overflow-y: auto;
  width: 100%;
}

@media (max-width: 768px) {
  .navbar-header {
    border-bottom: 1px solid var(--clr-surface-a20);
    width: 100%;
  }

  .dropdown-menu {
    position: static;
    border: none;
    margin-top: var(--spacing-md);
  }

  .admin-dropdown.open .dropdown-menu {
    display: flex;
  }
}
</style>
