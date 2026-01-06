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
            <router-link :to="`/teams/${team.id}`" class="team-name">{{ team.name }}</router-link>
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

      <div class="matchup-sections">
        <div class="matchup-section">
          <h2>Best Against</h2>
          <div class="matchup-cards">
            <div v-for="matchup in bestAgainst" :key="matchup.opponentId" class="matchup-card">
              <router-link :to="`/players/${matchup.opponentId}`" class="opponent-name">{{ matchup.opponentName }}</router-link>
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
              <router-link :to="`/players/${matchup.opponentId}`" class="opponent-name">{{ matchup.opponentName }}</router-link>
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
            <router-link :to="`/players/${matchup.opponentId}`" class="opponent-name">{{ matchup.opponentName }}</router-link>
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
interface Matchup {
  opponentId: number
  opponentName: string
  wins: number
  losses: number
  winRate: number
}

const api = API_URL
const route = useRoute()

const playerId = computed(() => Number(route.params.id))

const player = ref<Player | null>(null)
const teams = ref<Team[]>([])
const players = ref<Player[]>([])
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
    if (playerIds.includes(playerId.value)) {
      playerTeamIds.push(Number(teamId))
    }
  })

  // Create a Set to track matches we've already counted to avoid duplicates
  const countedMatches = new Set<number>()

  // Count stats from matches where player's teams played
  filteredMatches.value.forEach(match => {
    // Only count each match once
    if (countedMatches.has(match.id || 0)) {
      return
    }

    const playerOnTeamA = playerTeamIds.includes(match.team_a_id)
    const playerOnTeamB = playerTeamIds.includes(match.team_b_id)

    if (playerOnTeamA || playerOnTeamB) {
      countedMatches.add(match.id || 0)
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
    if (playerIds.includes(playerId.value)) {
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

const matchups = computed(() => {
  const matchupMap: Record<number, Matchup> = {}

  // Get all players this player has faced
  filteredMatches.value.forEach(match => {
    // Find all teams this player is on
    const playerTeamIds: number[] = []
    Object.entries(teamPlayerMap.value).forEach(([teamId, playerIds]) => {
      if (playerIds.includes(playerId.value)) {
        playerTeamIds.push(Number(teamId))
      }
    })

    let playerTeamId = 0
    let opponentTeamId = 0
    let isWin = false

    // Determine which side the player was on
    if (playerTeamIds.includes(match.team_a_id)) {
      playerTeamId = match.team_a_id
      opponentTeamId = match.team_b_id
      isWin = match.winner_id === match.team_a_id
    } else if (playerTeamIds.includes(match.team_b_id)) {
      playerTeamId = match.team_b_id
      opponentTeamId = match.team_a_id
      isWin = match.winner_id === match.team_b_id
    } else {
      return // Player not involved in this match
    }

    // Get all players on the opponent team
    const opponentPlayerIds = teamPlayerMap.value[opponentTeamId] || []

    // Record stats against each opponent player
    opponentPlayerIds.forEach(opponentPlayerId => {
      if (!matchupMap[opponentPlayerId]) {
        const opponentPlayer = players.value.find(p => p.id === opponentPlayerId)
        matchupMap[opponentPlayerId] = {
          opponentId: opponentPlayerId,
          opponentName: opponentPlayer?.name || 'Unknown',
          wins: 0,
          losses: 0,
          winRate: 0
        }
      }

      if (isWin) {
        matchupMap[opponentPlayerId].wins++
      } else {
        matchupMap[opponentPlayerId].losses++
      }
    })
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
  fetchPlayer()
  fetchTeams()
  fetchPlayers()
  fetchSeasons()
  fetchMatches()
})

watch(playerId, () => {
  loading.value = true
  fetchPlayer()
  fetchMatches()
})

watch(selectedSeasonId, () => {
  // This will trigger the computed properties to update
})

const fetchPlayer = async () => {
  try {
    const res = await fetch(`${api}/players/${playerId.value}`)
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

const fetchPlayers = async () => {
  try {
    const res = await fetch(`${api}/players`)
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
  flex-direction: column;
  gap: var(--spacing-lg);
  align-items: flex-start;
  margin-bottom: var(--spacing-lg);
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
  color: inherit;
  text-decoration: none;
  transition: color 0.2s ease;
}

.team-name:hover {
  color: #60a5fa;
  text-decoration: underline;
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
  color: inherit;
  text-decoration: none;
  transition: color 0.2s ease;
}

.opponent-name:hover {
  color: #60a5fa;
  text-decoration: underline;
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
  font-size: 20px;
  font-weight: 700;
}

.winrate {
  font-size: 12px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.7);
  text-align: center;
}

.all-matchups-section {
  margin-top: var(--spacing-xl);
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

  .matchup-sections {
    grid-template-columns: 1fr;
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
    margin-left: 0;
    width: 100%;
  }
}
</style>
