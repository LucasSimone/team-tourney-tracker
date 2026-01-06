<template>
  <div class="home-page">
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-number">{{ teams.length }}</div>
        <div class="stat-label">Teams</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ players.length }}</div>
        <div class="stat-label">Players</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ seasons.length }}</div>
        <div class="stat-label">Seasons</div>
      </div>
      <div class="stat-card">
        <div class="stat-number">{{ matches.length }}</div>
        <div class="stat-label">Matches</div>
      </div>
    </div>

    <div class="actions-section">
      <h2>Get Started</h2>
      <div class="action-buttons">
        <router-link to="/track" class="action-button primary">
          📊 Track a Match
        </router-link>
        <router-link to="/admin/teams" class="action-button secondary">
          ⚙️ Admin Panel
        </router-link>
      </div>
    </div>

    <div v-if="matches.length > 0" class="recent-matches">
      <h2>Recent Matches</h2>
      <div class="matches-list">
        <div v-for="match in recentMatches" :key="match.id" class="match-card">
          <div class="match-teams">
            <span class="team-a">{{ getTeamName(match.team_a_id) }}</span>
            <span class="vs">vs</span>
            <span class="team-b">{{ getTeamName(match.team_b_id) }}</span>
          </div>
          <div class="match-details">
            <span class="score">{{ match.score || '—' }}</span>
            <span class="winner">🏆 {{ getTeamName(match.winner_id) || '—' }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, computed, onMounted } from 'vue'
import { useAuth } from '@/composables/useAuth'

interface Team { id?: number; name: string }
interface Player { id?: number; name: string }
interface Season { id?: number; year: number }
interface Match { id?: number; season_id: number; team_a_id: number; team_b_id: number; winner_id?: number; score?: string }

const api = API_URL
const { getAuthHeaders } = useAuth()
const teams = ref<Team[]>([])
const players = ref<Player[]>([])
const seasons = ref<Season[]>([])
const matches = ref<Match[]>([])

const recentMatches = computed(() => matches.value.slice(0, 5))

onMounted(() => {
  fetchTeams()
  fetchPlayers()
  fetchSeasons()
  fetchMatches()
})

const fetchTeams = async () => {
  try {
    const res = await fetch(`${api}/teams`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      teams.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch teams', e)
  }
}

const fetchPlayers = async () => {
  try {
    const res = await fetch(`${api}/players`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      players.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch players', e)
  }
}

const fetchSeasons = async () => {
  try {
    const res = await fetch(`${api}/seasons`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      seasons.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch seasons', e)
  }
}

const fetchMatches = async () => {
  try {
    const res = await fetch(`${api}/matches`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      matches.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch matches', e)
  }
}

const getTeamName = (id?: number) => teams.value.find(t => t.id === id)?.name || 'Unknown'
</script>

<style scoped>
.home-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: var(--spacing-lg);
}

.stat-card {
  background-color: var(--clr-surface-a10);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  text-align: center;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: var(--clr-primary-a0);
  margin-bottom: var(--spacing-md);
}

.stat-label {
  color: var(--text-secondary);
  font-size: 14px;
}

.actions-section h2 {
  margin-top: 0;
}

.action-buttons {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-lg);
}

.action-button {
  padding: var(--spacing-lg);
  text-decoration: none;
  border-radius: var(--radius);
  font-weight: 500;
  text-align: center;
  transition: all 0.2s;
  display: block;
}

.action-button.primary {
  background: var(--clr-primary-a0);
  color: white;
}

.action-button.primary:hover {
  background: var(--clr-primary-a10);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(91, 24, 199, 0.2);
}

.action-button.secondary {
  background: var(--clr-surface-a10);
  color: var(--text-primary);
  border: 1px solid var(--border);
}

.action-button.secondary:hover {
  background: var(--clr-surface-a20);
}

.recent-matches h2 {
  margin-top: 0;
}

.matches-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.match-card {
  background-color: var(--clr-surface-a10);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.match-teams {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex: 1;
  font-weight: 500;
}

.team-a, .team-b {
  flex: 1;
}

.vs {
  color: var(--text-secondary);
  font-size: 12px;
  text-transform: uppercase;
  padding: 0 var(--spacing-md);
}

.match-details {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  text-align: right;
}

.score {
  font-weight: bold;
  color: var(--clr-primary-a50);
}

.winner {
  color: var(--text-secondary);
  font-size: 14px;
}

@media (max-width: 600px) {
  .hero h1 {
    font-size: 24px;
  }

  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .match-card {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .match-details {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
