<template>
  <div class="team-detail-page">
    <div v-if="loading" class="loading">Loading team details...</div>
    <div v-else-if="team && matchups">
      <div class="header">
        <router-link to="/standings" class="back-button">← Back to Standings</router-link>
        <h1>{{ team.name }}</h1>
      </div>

      <div class="season-filter">
        <select v-model.number="selectedSeasonId">
          <option value="0">All Seasons</option>
          <option v-for="season in seasons" :key="season.id" :value="season.id">
            Season {{ season.year }}{{ isCurrentSeason(season.id) ? ' (Current)' : '' }}
          </option>
        </select>
      </div>

      <div class="matchup-sections">
        <div class="matchup-section">
          <h2>Best Against</h2>
          <div class="matchup-cards">
            <div v-for="matchup in bestAgainst" :key="matchup.opponentId" class="matchup-card">
              <div class="opponent-name">{{ matchup.opponentName }}</div>
              <div class="matchup-stats">
                <div class="stat-badge wins">
                  <span class="label">W</span>
                  <span class="value">{{ matchup.wins }}</span>
                </div>
                <div class="stat-badge losses">
                  <span class="label">L</span>
                  <span class="value">{{ matchup.losses }}</span>
                </div>
                <div class="winrate">
                  {{ matchup.winRate }}% Win Rate
                </div>
              </div>
            </div>
            <div v-if="bestAgainst.length === 0" class="empty-message">No matchups yet</div>
          </div>
        </div>

        <div class="matchup-section">
          <h2>Worst Against</h2>
          <div class="matchup-cards">
            <div v-for="matchup in worstAgainst" :key="matchup.opponentId" class="matchup-card">
              <div class="opponent-name">{{ matchup.opponentName }}</div>
              <div class="matchup-stats">
                <div class="stat-badge wins">
                  <span class="label">W</span>
                  <span class="value">{{ matchup.wins }}</span>
                </div>
                <div class="stat-badge losses">
                  <span class="label">L</span>
                  <span class="value">{{ matchup.losses }}</span>
                </div>
                <div class="winrate">
                  {{ matchup.winRate }}% Win Rate
                </div>
              </div>
            </div>
            <div v-if="worstAgainst.length === 0" class="empty-message">No matchups yet</div>
          </div>
        </div>
      </div>

      <div class="all-matchups-section">
        <h2>All Matchups</h2>
        <div class="matchup-cards">
          <div v-for="matchup in allMatchups" :key="matchup.opponentId" class="matchup-card">
            <div class="opponent-name">{{ matchup.opponentName }}</div>
            <div class="matchup-stats">
              <div class="stat-badge wins">
                <span class="label">W</span>
                <span class="value">{{ matchup.wins }}</span>
              </div>
              <div class="stat-badge losses">
                <span class="label">L</span>
                <span class="value">{{ matchup.losses }}</span>
              </div>
              <div class="winrate">
                {{ matchup.winRate }}% Win Rate
              </div>
            </div>
          </div>
          <div v-if="allMatchups.length === 0" class="empty-message">No matchups yet</div>
        </div>
      </div>
    </div>
    <div v-else class="error">Team not found</div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'

interface Team { id?: number; name: string }
interface Season { id?: number; year: number }
interface Match {
  id?: number
  season_id: number
  team_a_id: number
  team_b_id: number
  winner_id?: number
}
interface Matchup {
  opponentId: number
  opponentName: string
  wins: number
  losses: number
  winRate: number
}

const api = API_URL
const route = useRoute()
const teamId = Number(route.params.id)

const team = ref<Team | null>(null)
const teams = ref<Team[]>([])
const seasons = ref<Season[]>([])
const matches = ref<Match[]>([])
const selectedSeasonId = ref(0)
const loading = ref(true)

const filteredMatches = computed(() => {
  if (selectedSeasonId.value === 0) {
    return matches.value
  }
  return matches.value.filter(m => m.season_id === selectedSeasonId.value)
})

const matchups = computed(() => {
  const matchupMap: Record<number, Matchup> = {}

  filteredMatches.value.forEach(match => {
    let opponentId = 0
    let isWin = false

    // Determine if this team was team_a or team_b
    if (match.team_a_id === teamId) {
      opponentId = match.team_b_id
      isWin = match.winner_id === teamId
    } else if (match.team_b_id === teamId) {
      opponentId = match.team_a_id
      isWin = match.winner_id === teamId
    } else {
      return // This match doesn't involve our team
    }

    if (!matchupMap[opponentId]) {
      const opponentName = teams.value.find(t => t.id === opponentId)?.name || 'Unknown'
      matchupMap[opponentId] = {
        opponentId,
        opponentName,
        wins: 0,
        losses: 0,
        winRate: 0
      }
    }

    if (isWin) {
      matchupMap[opponentId].wins++
    } else {
      matchupMap[opponentId].losses++
    }
  })

  // Calculate win rates
  Object.values(matchupMap).forEach(matchup => {
    const total = matchup.wins + matchup.losses
    if (total > 0) {
      matchup.winRate = Math.round((matchup.wins / total) * 100)
    }
  })

  return matchupMap
})

const allMatchups = computed(() => {
  return Object.values(matchups.value).sort((a, b) => b.opponentName.localeCompare(a.opponentName))
})

const bestAgainst = computed(() => {
  return Object.values(matchups.value)
    .sort((a, b) => b.winRate - a.winRate)
    .slice(0, 5)
})

const worstAgainst = computed(() => {
  return Object.values(matchups.value)
    .sort((a, b) => a.winRate - b.winRate)
    .slice(0, 5)
})

const currentSeasonId = computed(() => {
  if (seasons.value.length === 0) return 0
  const latest = seasons.value.reduce((latest, current) =>
    (current.year > latest.year) ? current : latest
  )
  return latest.id || 0
})

const isCurrentSeason = (seasonId?: number) => {
  return seasonId === currentSeasonId.value
}

onMounted(() => {
  fetchTeam()
  fetchTeams()
  fetchSeasons()
  fetchMatches()
})

const fetchTeam = async () => {
  try {
    const res = await fetch(`${api}/teams/${teamId}`)
    if (res.ok) {
      const data = await res.json()
      team.value = data
    }
  } catch (e) {
    console.error('Failed to fetch team', e)
  } finally {
    loading.value = false
  }
}

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
      // Set the default to the latest season
      if (seasons.value.length > 0) {
        const latestSeason = seasons.value.reduce((latest, current) =>
          (current.year > latest.year) ? current : latest
        )
        selectedSeasonId.value = latestSeason.id || 0
      }
    }
  } catch (e) {
    console.error('Failed to fetch seasons', e)
  }
}

const fetchMatches = async () => {
  try {
    const res = await fetch(`${api}/matches`)
    if (res.ok) {
      const data = await res.json()
      matches.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch matches', e)
  }
}
</script>

<style scoped>
.team-detail-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.loading,
.error {
  padding: var(--spacing-xl);
  text-align: center;
  color: var(--text-secondary);
  background-color: var(--clr-surface-a10);
  border-radius: var(--radius);
}

.error {
  color: #f87171;
}

.header {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.back-button {
  padding: var(--spacing-md) var(--spacing-lg);
  background-color: var(--clr-surface-a10);
  border-radius: var(--radius);
  color: var(--text-primary);
  text-decoration: none;
  transition: background-color 0.2s ease;
  font-size: 14px;
  font-weight: 500;
}

.back-button:hover {
  background-color: var(--clr-surface-a20);
}

h1 {
  margin: 0;
  font-size: 32px;
}

h2 {
  margin: 0 0 var(--spacing-lg) 0;
  font-size: 24px;
}

.season-filter {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  justify-content: flex-end;
}

.season-filter select {
  padding: var(--spacing-md);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background-color: var(--clr-surface-a10);
  color: var(--text-primary);
  font-size: 14px;
}

.matchup-sections {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-xl);
  margin-top: var(--spacing-xl);
}

.matchup-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.matchup-cards {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.matchup-card {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: white;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.matchup-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

.opponent-name {
  font-size: 16px;
  font-weight: 600;
  flex: 1;
}

.matchup-stats {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  margin-left: var(--spacing-lg);
}

.stat-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  min-width: 50px;
}

.stat-badge.wins {
  color: #4ade80;
}

.stat-badge.losses {
  color: #f87171;
}

.stat-badge .label {
  font-size: 11px;
  font-weight: 600;
  opacity: 0.7;
  text-transform: uppercase;
}

.stat-badge .value {
  font-size: 18px;
  font-weight: 700;
}

.winrate {
  font-size: 12px;
  color: #60a5fa;
  font-weight: 600;
  white-space: nowrap;
}

.all-matchups-section {
  margin-top: var(--spacing-xl);
}

.empty-message {
  padding: var(--spacing-lg);
  text-align: center;
  color: rgba(255, 255, 255, 0.5);
  background: rgba(0, 0, 0, 0.2);
  border-radius: var(--radius);
}

@media (max-width: 768px) {
  .matchup-sections {
    grid-template-columns: 1fr;
  }

  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .matchup-card {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .opponent-name {
    width: 100%;
  }

  .matchup-stats {
    width: 100%;
    margin-left: 0;
  }
}
</style>
