<template>
  <div class="standings-page">
    <div v-if="seasons.length > 0" class="season-filter">
      <select v-model.number="selectedSeasonId">
        <option value="0">All Seasons</option>
        <option v-for="season in seasons" :key="season.id" :value="season.id">
          {{ season.year }}{{ isCurrentSeason(season.id) ? ' (Current)' : '' }}
        </option>
      </select>
    </div>

    <div v-if="teamStats.length > 0" class="stats-section">
      <h2>Team Rankings</h2>
      <div class="stats-cards">
        <div v-for="(stat, index) in teamStats" :key="stat.id" class="stat-card" :style="getCardStyle(stat.id)">
          <div class="card-header">
            <div class="card-rank">{{ index + 1 }}</div>
            <router-link :to="`/teams/${stat.id}`" class="card-name">{{ stat.name }}</router-link>
          </div>
          <div class="card-stats">
            <div class="stat-item">
              <span class="stat-label">GP</span>
              <span class="stat-value">{{ stat.gamesPlayed }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">W</span>
              <span class="stat-value stat-wins">{{ stat.wins }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">L</span>
              <span class="stat-value stat-losses">{{ stat.losses }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Win %</span>
              <span class="stat-value stat-winpct">{{ stat.winPercentage }}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="playerStats.length > 0" class="stats-section">
      <h2>Player Rankings</h2>
      <div class="stats-cards">
        <div v-for="(stat, index) in playerStats" :key="stat.id" class="stat-card">
          <div class="card-header">
            <div class="card-rank">{{ index + 1 }}</div>
            <router-link :to="`/players/${stat.id}`" class="card-name">{{ stat.name }}</router-link>
          </div>
          <div class="card-stats">
            <div class="stat-item">
              <span class="stat-label">GP</span>
              <span class="stat-value">{{ stat.gamesPlayed }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">W</span>
              <span class="stat-value stat-wins">{{ stat.wins }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">L</span>
              <span class="stat-value stat-losses">{{ stat.losses }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Win %</span>
              <span class="stat-value stat-winpct">{{ formatWinPercentage(stat.winPercentage) }}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { API_URL } from '@/config'

interface Team { id?: number; name: string }
interface Season { id?: number; year: number }
interface Match { 
  id?: number
  season_id: number
  match_type: string
  participant_a_id: number
  participant_b_id: number
  winner_id?: number
  score?: string
}
interface TeamStat {
  id: number
  name: string
  gamesPlayed: number
  wins: number
  losses: number
  winPercentage: number
}
interface PlayerStat {
  id: number
  name: string
  games_played: number
  wins: number
  losses: number
  win_percentage: number
  gamesPlayed?: number
  winPercentage?: number
}

const api = API_URL
const teams = ref<Team[]>([])
const seasons = ref<Season[]>([])
const matches = ref<Match[]>([])
const playerStats = ref<PlayerStat[]>([])
const selectedSeasonId = ref(0)
const teamImages = ref<{ [key: number]: string }>({})

const filteredMatches = computed(() => {
  if (selectedSeasonId.value === 0) {
    return matches.value
  }
  return matches.value.filter(m => m.season_id === selectedSeasonId.value)
})

const teamStats = computed(() => {
  const stats: Record<number, TeamStat> = {}

  // Initialize stats for all teams
  teams.value.forEach(team => {
    if (team.id) {
      stats[team.id] = {
        id: team.id,
        name: team.name || 'Unknown',
        gamesPlayed: 0,
        wins: 0,
        losses: 0,
        winPercentage: 0
      }
    }
  })

  // Count wins/losses from filtered matches
  filteredMatches.value.forEach(match => {
    // Only count team matches
    if (match.match_type !== 'team') {
      return
    }

    // Count games played for both teams
    if (stats[match.participant_a_id]) {
      stats[match.participant_a_id].gamesPlayed++
    }
    if (stats[match.participant_b_id]) {
      stats[match.participant_b_id].gamesPlayed++
    }

    if (match.winner_id) {
      if (stats[match.winner_id]) {
        stats[match.winner_id].wins++
      }
      // Loser is the other team
      const loserId = match.participant_a_id === match.winner_id ? match.participant_b_id : match.participant_a_id
      if (stats[loserId]) {
        stats[loserId].losses++
      }
    }
  })

  // Calculate win percentage and sort by win percentage (descending)
  return Object.values(stats)
    .filter(stat => stat.gamesPlayed >= 10) // Only include teams with 10+ games
    .map(stat => ({
      ...stat,
      winPercentage: stat.wins + stat.losses > 0 
        ? Math.round((stat.wins / (stat.wins + stat.losses)) * 100)
        : 0
    }))
    .sort((a, b) => b.winPercentage - a.winPercentage)
})

onMounted(() => {
  fetchTeams()
  fetchSeasons()
  fetchMatches()
  fetchPlayerStats()
})

// Watch for season changes and refetch player stats
watch(selectedSeasonId, () => {
  fetchPlayerStats()
})

const fetchTeams = async () => {
  try {
    const res = await fetch(`${api}/teams`)
    if (res.ok) {
      const data = await res.json()
      teams.value = data || []
      // Load images for each team
      for (const team of teams.value) {
        if (team.id) {
          await loadTeamImage(team.id)
        }
      }
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
      // Set the default to the latest season (highest year)
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

const fetchPlayerStats = async () => {
  try {
    const seasonParam = selectedSeasonId.value === 0 ? '0' : selectedSeasonId.value
    const res = await fetch(`${api}/players/stats?season_id=${seasonParam}`)
    if (res.ok) {
      const data = await res.json()
      // Map the API response to our interface and filter by minimum 10 games
      playerStats.value = (data || [])
        .filter((stat: any) => stat.games_played >= 10) // Only include players with 10+ games
        .map((stat: any) => ({
          id: stat.id,
          name: stat.name,
          games_played: stat.games_played,
          wins: stat.wins,
          losses: stat.losses,
          win_percentage: stat.win_percentage,
          gamesPlayed: stat.games_played,
          winPercentage: stat.win_percentage
        }))
    }
  } catch (e) {
    console.error('Failed to fetch player stats', e)
  }
}

const formatWinPercentage = (percentage: number) => {
  return Math.round(percentage * 100) / 100
}

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

const getTeamName = (id?: number) => teams.value.find(t => t.id === id)?.name || 'Unknown'
const getSeasonYear = (id?: number) => seasons.value.find(s => s.id === id)?.year || '?'

const loadTeamImage = async (teamId: number) => {
  try {
    const res = await fetch(`${api}/teams/${teamId}/image/horizontal`)
    if (res.ok) {
      const blob = await res.blob()
      teamImages.value[teamId] = URL.createObjectURL(blob)
    }
  } catch (e) {
    // 404 is expected for teams without images - don't log error
  }
}

const getCardStyle = (teamId: number) => {
  const imageUrl = teamImages.value[teamId]
  if (!imageUrl) {
    return {}
  }
  return {
    backgroundImage: `linear-gradient(135deg, rgba(26, 26, 46, 0.85) 0%, rgba(22, 33, 62, 0.85) 100%), url(${imageUrl})`,
    backgroundSize: 'cover',
    backgroundPosition: 'center'
  }
}
</script>

<style scoped>
.standings-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
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

.season-filter label {
  font-weight: 500;
}

.season-filter select {
  padding: var(--spacing-md);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background-color: var(--clr-surface-a10);
  color: var(--text-primary);
  font-size: 14px;
}

.empty-state {
  padding: var(--spacing-xl);
  text-align: center;
  color: var(--text-secondary);
  background-color: var(--clr-surface-a10);
  border-radius: var(--radius);
}

.matches-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.matches-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-lg);
}

.match-card {
  background-color: var(--clr-surface-a10);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.match-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.season-badge {
  font-size: 12px;
  padding: 4px 8px;
  background-color: var(--clr-surface-a20);
  border-radius: var(--radius);
  color: var(--text-secondary);
}

.match-body {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  justify-content: space-between;
}

.team {
  flex: 1;
  text-align: center;
}

.team-name {
  font-weight: 500;
  font-size: 16px;
}

.vs {
  color: var(--text-secondary);
  font-size: 12px;
  text-transform: uppercase;
  padding: 0 var(--spacing-md);
}

.score {
  text-align: center;
  font-size: 24px;
  font-weight: bold;
  color: var(--clr-primary-a50);
  padding: var(--spacing-md);
  background-color: var(--clr-surface-a20);
  border-radius: var(--radius);
}

.winner {
  text-align: center;
  font-weight: 500;
  color: var(--clr-primary-a50);
  padding: var(--spacing-md);
  background-color: var(--clr-surface-a20);
  border: 2px solid var(--clr-primary-a50);
  border-radius: var(--radius);
}

.stats-section {
  margin-top: var(--spacing-xl);
}

.stats-cards {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.stat-card {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  color: white;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

.card-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  justify-content: space-between;
}

.card-rank {
  font-size: 28px;
  font-weight: 700;
  color: white;
  min-width: 50px;
  text-align: center;
  text-shadow: 1px 1px 10px rgba(0, 0, 0, 1)

}

.card-name {
  font-size: 18px;
  font-weight: 600;
  color: white;
  text-decoration: none;
  transition: color 0.2s ease;
  white-space: normal;
  word-wrap: break-word;
  text-align: right;
  text-shadow: 1px 1px 10px rgba(0, 0, 0, 1)
}

.card-name:hover {
  color: var(--clr-primary-a50);
}

.card-stats {
  display: flex;
  gap: var(--spacing-xl);
  margin-left: auto;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-label {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.7);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.9);
}

.stat-wins {
  color: #4ade80;
}

.stat-losses {
  color: #f87171;
}

.stat-winpct {
  color: #60a5fa;
}

@media (max-width: 768px) {
  h1 {
    font-size: 24px;
  }

  .season-filter {
    flex-direction: column;
    align-items: flex-start;
  }

  .season-filter select {
    width: 100%;
  }

  .stat-card {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
  }

  .card-header {
    width: 100%;
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    flex-wrap: wrap;
    justify-content: space-between;
  }

  .card-rank {
    min-width: unset;
  }

  .card-name {
    flex: 1;
    min-width: 0;
    text-align: right;
    text-shadow: 1px 1px 3px rgba(0, 0, 0, 1);
  }

  .card-stats {
    margin-left: 0;
    width: 100%;
    gap: var(--spacing-md);
    flex-wrap: wrap;
  }

  .stat-item {
    flex: 1;
    min-width: calc(33.333% - var(--spacing-md));
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    background: rgba(0, 0, 0, 0.6);
    padding: var(--spacing-md);
    border-radius: 4px;
  }

  .stat-item:nth-child(4) {
    flex-basis: 100%;
    min-width: unset;
  }

  .stat-label {
    margin-bottom: 0;
  }
}
</style>
