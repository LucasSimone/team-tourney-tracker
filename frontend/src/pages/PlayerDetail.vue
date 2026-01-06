<template>
  <div class="player-detail-page">
    <div v-if="loading" class="loading">Loading player details...</div>
    <div v-else-if="player && playerTeams">
      <div class="header">
        <router-link to="/standings" class="back-button">← Back to Standings</router-link>
        <h1>{{ player.name }}</h1>
      </div>

      <div class="season-filter">
        <select v-model.number="selectedSeasonId">
          <option value="0">All Seasons</option>
          <option v-for="season in seasons" :key="season.id" :value="season.id">
            Season {{ season.year }}{{ isCurrentSeason(season.id) ? ' (Current)' : '' }}
          </option>
        </select>
      </div>

      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-label">Games Played</div>
          <div class="stat-value">{{ playerStats.gamesPlayed }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Wins</div>
          <div class="stat-value stat-wins">{{ playerStats.wins }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Losses</div>
          <div class="stat-value stat-losses">{{ playerStats.losses }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Win Rate</div>
          <div class="stat-value stat-winpct">{{ playerStats.winRate }}%</div>
        </div>
      </div>

      <div class="teams-section">
        <h2>Teams</h2>
        <div class="team-cards">
          <div v-for="team in playerTeams" :key="team.id" class="team-card">
            <div class="team-name">{{ team.name }}</div>
            <div class="team-stats">
              <div class="stat-item">
                <span class="stat-label">W</span>
                <span class="stat-value stat-wins">{{ team.wins }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">L</span>
                <span class="stat-value stat-losses">{{ team.losses }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">Win %</span>
                <span class="stat-value stat-winpct">{{ team.winRate }}%</span>
              </div>
            </div>
          </div>
          <div v-if="playerTeams.length === 0" class="empty-message">No team data available</div>
        </div>
      </div>
    </div>
    <div v-else class="error">Player not found</div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

interface Player { id?: number; name: string }
interface Team { id?: number; name: string }
interface Season { id?: number; year: number }
interface Match {
  id?: number
  season_id: number
  team_a_id: number
  team_b_id: number
  winner_id?: number
}
interface TeamData {
  id: number
  name: string
  wins: number
  losses: number
  winRate: number
}

const api = API_URL
const route = useRoute()
const playerId = Number(route.params.id)

const player = ref<Player | null>(null)
const teams = ref<Team[]>([])
const seasons = ref<Season[]>([])
const teamPlayerMap = ref<Record<number, number[]>>({}) // team_id -> player_ids
const matches = ref<Match[]>([])
const selectedSeasonId = ref(0)
const loading = ref(true)

const filteredMatches = computed(() => {
  if (selectedSeasonId.value === 0) {
    return matches.value
  }
  return matches.value.filter(m => m.season_id === selectedSeasonId.value)
})

const playerStats = computed(() => {
  let gamesPlayed = 0
  let wins = 0
  let losses = 0

  // Find all teams this player is on
  const playerTeamIds: number[] = []
  Object.entries(teamPlayerMap.value).forEach(([teamId, playerIds]) => {
    if (playerIds.includes(playerId)) {
      playerTeamIds.push(Number(teamId))
    }
  })

  // Count stats from matches where player's teams played
  filteredMatches.value.forEach(match => {
    const playerOnTeamA = playerTeamIds.includes(match.team_a_id)
    const playerOnTeamB = playerTeamIds.includes(match.team_b_id)

    if (playerOnTeamA || playerOnTeamB) {
      gamesPlayed++

      const playerTeamId = playerOnTeamA ? match.team_a_id : match.team_b_id
      if (match.winner_id === playerTeamId) {
        wins++
      } else {
        losses++
      }
    }
  })

  const winRate = gamesPlayed > 0 ? Math.round((wins / gamesPlayed) * 100) : 0

  return {
    gamesPlayed,
    wins,
    losses,
    winRate
  }
})

const playerTeams = computed(() => {
  const teamData: Record<number, TeamData> = {}

  // Find all teams this player is on
  const playerTeamIds: number[] = []
  Object.entries(teamPlayerMap.value).forEach(([teamId, playerIds]) => {
    if (playerIds.includes(playerId)) {
      playerTeamIds.push(Number(teamId))
    }
  })

  // Initialize team data
  playerTeamIds.forEach(teamId => {
    const team = teams.value.find(t => t.id === teamId)
    if (team) {
      teamData[teamId] = {
        id: teamId,
        name: team.name,
        wins: 0,
        losses: 0,
        winRate: 0
      }
    }
  })

  // Count stats per team
  filteredMatches.value.forEach(match => {
    if (playerTeamIds.includes(match.team_a_id)) {
      if (match.winner_id === match.team_a_id) {
        teamData[match.team_a_id].wins++
      } else {
        teamData[match.team_a_id].losses++
      }
    }
    if (playerTeamIds.includes(match.team_b_id)) {
      if (match.winner_id === match.team_b_id) {
        teamData[match.team_b_id].wins++
      } else {
        teamData[match.team_b_id].losses++
      }
    }
  })

  // Calculate win rates
  Object.values(teamData).forEach(team => {
    const total = team.wins + team.losses
    if (total > 0) {
      team.winRate = Math.round((team.wins / total) * 100)
    }
  })

  return Object.values(teamData)
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
  fetchPlayer()
  fetchTeams()
  fetchSeasons()
  fetchMatches()
})

watch(selectedSeasonId, () => {
  // This will trigger the computed properties to update
})

const fetchPlayer = async () => {
  try {
    const res = await fetch(`${api}/players/${playerId}`)
    if (res.ok) {
      const data = await res.json()
      player.value = data
    }
  } catch (e) {
    console.error('Failed to fetch player', e)
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

  // Also fetch team-player mappings for all teams
  try {
    const res = await fetch(`${api}/teams`)
    if (res.ok) {
      const teamsData = await res.json()
      // For each team, fetch its players
      for (const team of teamsData) {
        try {
          const playersRes = await fetch(`${api}/teams/${team.id}/players`)
          if (playersRes.ok) {
            const playersData = await playersRes.json()
            teamPlayerMap.value[team.id] = playersData.map((p: any) => p.id)
          }
        } catch (e) {
          console.error(`Failed to fetch players for team ${team.id}`, e)
        }
      }
    }
  } catch (e) {
    console.error('Failed to fetch team-player mappings', e)
  }
}
</script>

<style scoped>
.player-detail-page {
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: var(--spacing-lg);
  margin-top: var(--spacing-xl);
}

.stat-card {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-md);
  color: white;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

.stat-card .stat-label {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.6);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-card .stat-value {
  font-size: 28px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.9);
}

.stat-card .stat-wins {
  color: #4ade80;
}

.stat-card .stat-losses {
  color: #f87171;
}

.stat-card .stat-winpct {
  color: #60a5fa;
}

.teams-section {
  margin-top: var(--spacing-xl);
}

.team-cards {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.team-card {
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

.team-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

.team-name {
  font-size: 18px;
  font-weight: 600;
  flex: 1;
}

.team-stats {
  display: flex;
  gap: var(--spacing-xl);
  margin-left: var(--spacing-lg);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-item .stat-label {
  font-size: 11px;
  font-weight: 600;
  opacity: 0.7;
  text-transform: uppercase;
}

.stat-item .stat-value {
  font-size: 18px;
  font-weight: 700;
}

.stat-item .stat-wins {
  color: #4ade80;
}

.stat-item .stat-losses {
  color: #f87171;
}

.stat-item .stat-winpct {
  color: #60a5fa;
}

.empty-message {
  padding: var(--spacing-lg);
  text-align: center;
  color: rgba(255, 255, 255, 0.5);
  background: rgba(0, 0, 0, 0.2);
  border-radius: var(--radius);
}

@media (max-width: 768px) {
  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .team-card {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .team-name {
    width: 100%;
  }

  .team-stats {
    margin-left: 0;
    width: 100%;
  }
}
</style>
