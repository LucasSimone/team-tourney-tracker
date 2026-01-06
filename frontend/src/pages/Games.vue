<template>
  <div class="games-page">
    <h1>Games</h1>

    <div class="filters">
      <div class="filter-group">
        <select v-model="selectedTeamId" class="filter-select">
          <option value="">All Teams</option>
          <option v-for="team in teams" :key="team.id" :value="team.id">
            {{ team.name }}
          </option>
        </select>
      </div>

      <div class="filter-group">
        <input 
          v-model="selectedDate" 
          type="date" 
          class="filter-input"
          placeholder="Filter by date"
        />
      </div>

      <button v-if="selectedTeamId || selectedDate" @click="clearFilters" class="clear-button">
        Clear Filters
      </button>
    </div>

    <div v-if="filteredGames.length === 0" class="empty-state">
      No games found. Check back soon!
    </div>

    <div v-else class="games-list">
      <div v-for="game in filteredGames" :key="game.id" class="game-card">
        <div class="game-header">
          <span class="game-date">{{ formatDate(game.createdAt) }}</span>
          <span class="season-badge">Season {{ getSeasonYear(game.season_id) }}</span>
        </div>

        <div class="game-matchup">
          <div class="team-section team-a">
            <router-link :to="`/teams/${game.team_a_id}`" class="team-name">
              {{ getTeamName(game.team_a_id) }}
            </router-link>
            <div class="score" v-if="game.score_a !== undefined && game.score_b !== undefined">
              {{ game.score_a }}
            </div>
          </div>

          <div class="vs-section">
            <span class="vs">vs</span>
            <span v-if="game.winner_id" class="winner-badge">
              🏆 {{ getTeamName(game.winner_id) }} won
            </span>
          </div>

          <div class="team-section team-b">
            <router-link :to="`/teams/${game.team_b_id}`" class="team-name">
              {{ getTeamName(game.team_b_id) }}
            </router-link>
            <div class="score" v-if="game.score_a !== undefined && game.score_b !== undefined">
              {{ game.score_b }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, computed, onMounted } from 'vue'

interface Team { id?: number; name: string }
interface Season { id?: number; year: number }
interface Game {
  id?: number
  season_id: number
  team_a_id: number
  team_b_id: number
  score_a?: number
  score_b?: number
  winner_id?: number
  created_at?: string
  createdAt?: string
}

const api = API_URL
const teams = ref<Team[]>([])
const seasons = ref<Season[]>([])
const games = ref<Game[]>([])
const selectedTeamId = ref('')
const selectedDate = ref('')

const filteredGames = computed(() => {
  return games.value.filter(game => {
    // Filter by team
    if (selectedTeamId.value) {
      const teamId = Number(selectedTeamId.value)
      if (game.team_a_id !== teamId && game.team_b_id !== teamId) {
        return false
      }
    }

    // Filter by date
    if (selectedDate.value) {
      const gameDate = new Date(game.createdAt || game.created_at || '')
      const filterDate = new Date(selectedDate.value)
      
      // Compare only the date part (YYYY-MM-DD)
      const gameDateStr = gameDate.toISOString().split('T')[0]
      const filterDateStr = filterDate.toISOString().split('T')[0]
      
      if (gameDateStr !== filterDateStr) {
        return false
      }
    }

    return true
  })
})

const clearFilters = () => {
  selectedTeamId.value = ''
  selectedDate.value = ''
}

onMounted(() => {
  fetchTeams()
  fetchSeasons()
  fetchGames()
})

const fetchTeams = async () => {
  try {
    const res = await fetch(`${api}/teams`)
    if (res.ok) {
      const data = await res.json()
      teams.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch teams', e)
  }
}

const fetchSeasons = async () => {
  try {
    const res = await fetch(`${api}/seasons`)
    if (res.ok) {
      const data = await res.json()
      seasons.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch seasons', e)
  }
}

const fetchGames = async () => {
  try {
    const res = await fetch(`${api}/matches`)
    if (res.ok) {
      const data = await res.json()
      // Normalize the data - some fields might be named differently
      games.value = (data || []).map((game: any) => ({
        ...game,
        createdAt: game.created_at || game.createdAt || new Date().toISOString()
      }))
      // Sort by date descending (newest first)
      games.value.sort((a, b) => {
        const dateA = new Date(a.createdAt || '').getTime()
        const dateB = new Date(b.createdAt || '').getTime()
        return dateB - dateA
      })
    }
  } catch (e) {
    console.error('Failed to fetch games', e)
  }
}

const getTeamName = (id?: number) => teams.value.find(t => t.id === id)?.name || 'Unknown'
const getSeasonYear = (id?: number) => seasons.value.find(s => s.id === id)?.year || '?'

const formatDate = (dateStr?: string) => {
  if (!dateStr) return 'N/A'
  const date = new Date(dateStr)
  return date.toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })
}
</script>

<style scoped>
.games-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

h1 {
  margin: 0;
  font-size: 32px;
}

.filters {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.filter-select,
.filter-input {
  padding: var(--spacing-md);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background-color: var(--clr-surface-a10);
  color: var(--text-primary);
  font-size: 14px;
  min-width: 200px;
}

.filter-select:focus,
.filter-input:focus {
  outline: none;
  border-color: var(--clr-primary-a50);
}

.clear-button {
  padding: var(--spacing-md) var(--spacing-lg);
  background-color: var(--clr-surface-a10);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.clear-button:hover {
  background-color: var(--clr-surface-a20);
  border-color: var(--clr-primary-a50);
}

.empty-state {
  padding: var(--spacing-xl);
  text-align: center;
  color: var(--text-secondary);
  background-color: var(--clr-surface-a10);
  border-radius: var(--radius);
}

.games-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.game-card {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  color: white;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.game-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

.game-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
  padding-bottom: var(--spacing-md);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.game-date {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
  font-weight: 500;
}

.season-badge {
  font-size: 11px;
  padding: 4px 8px;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  color: rgba(255, 255, 255, 0.6);
}

.game-matchup {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-lg);
}

.team-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
}

.team-section.team-a {
  align-items: flex-end;
}

.team-section.team-b {
  align-items: flex-start;
}

.team-name {
  font-size: 16px;
  font-weight: 600;
  color: white;
  text-decoration: none;
  transition: color 0.2s ease;
}

.team-name:hover {
  color: var(--clr-primary-a50);
}

.score {
  font-size: 24px;
  font-weight: 700;
  color: var(--clr-primary-a50);
}

.vs-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.vs {
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.winner-badge {
  font-size: 11px;
  padding: 4px 8px;
  background-color: rgba(74, 222, 128, 0.2);
  border-radius: 4px;
  color: #4ade80;
  font-weight: 600;
  white-space: nowrap;
}

@media (max-width: 768px) {
  h1 {
    font-size: 24px;
  }

  .filters {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-group {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-select,
  .filter-input {
    min-width: unset;
    width: 100%;
  }

  .clear-button {
    width: 100%;
  }

  .game-matchup {
    flex-direction: column;
    text-align: center;
  }

  .team-section {
    width: 100%;
    align-items: center !important;
  }

  .game-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }
}
</style>
