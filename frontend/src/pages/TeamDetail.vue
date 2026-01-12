<template>
  <div class="team-detail-page" :style="getBackgroundStyle()">
    <div v-if="loading" class="loading">Loading team details...</div>
    <div v-else-if="team && matchups" class="team-detail-content">
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

      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-label">Games Played</div>
          <div class="stat-value">{{ teamStats.gamesPlayed }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Wins</div>
          <div class="stat-value stat-wins">{{ teamStats.wins }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Losses</div>
          <div class="stat-value stat-losses">{{ teamStats.losses }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">Win Rate</div>
          <div class="stat-value stat-winpct">{{ teamStats.winRate }}%</div>
        </div>
      </div>

      <div class="matchup-sections">
        <div class="matchup-section">
          <h2>Best Matchup</h2>
          <div class="matchup-cards">
            <div v-for="matchup in bestAgainst" :key="matchup.opponentId" class="matchup-card">
              <router-link :to="`/teams/${matchup.opponentId}`" class="card-name">{{ matchup.opponentName }}</router-link>
              <div class="card-stats">
                <div class="stat-item">
                  <span class="stat-label">GP</span>
                  <span class="stat-value">{{ matchup.gamesPlayed }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">W</span>
                  <span class="stat-value stat-wins">{{ matchup.wins }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">L</span>
                  <span class="stat-value stat-losses">{{ matchup.losses }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Win %</span>
                  <span class="stat-value stat-winpct">{{ matchup.winRate }}%</span>
                </div>
              </div>
            </div>
            <div v-if="bestAgainst.length === 0" class="empty-message">No matchups yet</div>
          </div>
        </div>

        <div class="matchup-section">
          <h2>Worst Matchup</h2>
          <div class="matchup-cards">
            <div v-for="matchup in worstAgainst" :key="matchup.opponentId" class="matchup-card">
              <router-link :to="`/teams/${matchup.opponentId}`" class="card-name">{{ matchup.opponentName }}</router-link>
              <div class="card-stats">
                <div class="stat-item">
                  <span class="stat-label">GP</span>
                  <span class="stat-value">{{ matchup.gamesPlayed }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">W</span>
                  <span class="stat-value stat-wins">{{ matchup.wins }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">L</span>
                  <span class="stat-value stat-losses">{{ matchup.losses }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Win %</span>
                  <span class="stat-value stat-winpct">{{ matchup.winRate }}%</span>
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
            <router-link :to="`/teams/${matchup.opponentId}`" class="card-name">{{ matchup.opponentName }}</router-link>
            <div class="card-stats">
              <div class="stat-item">
                <span class="stat-label">GP</span>
                <span class="stat-value">{{ matchup.gamesPlayed }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">W</span>
                <span class="stat-value stat-wins">{{ matchup.wins }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">L</span>
                <span class="stat-value stat-losses">{{ matchup.losses }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">Win %</span>
                <span class="stat-value stat-winpct">{{ matchup.winRate }}%</span>
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
import { ref, computed, onMounted, watch, provide, inject } from 'vue'
import { useRoute } from 'vue-router'

interface Team { id?: number; name: string }
interface Season { id?: number; year: number }
interface Match {
  id?: number
  season_id: number
  match_type: string
  participant_a_id: number
  participant_b_id: number
  winner_id?: number
}
interface Matchup {
  opponentId: number
  opponentName: string
  wins: number
  losses: number
  gamesPlayed: number
  winRate: number
}

const api = API_URL
const route = useRoute()

const teamId = computed(() => Number(route.params.id))

const team = ref<Team | null>(null)
const teams = ref<Team[]>([])
const seasons = ref<Season[]>([])
const matches = ref<Match[]>([])
const selectedSeasonId = ref(0)
const loading = ref(true)
const teamImageUrl = ref<string | null>(null)

// Get the title update function from App.vue
const updateDetailTitle = inject<(title: string) => void>('updateDetailTitle')

// Watch for team changes and update the navbar title
watch(
  () => team.value?.name,
  (newName) => {
    if (newName && updateDetailTitle) {
      updateDetailTitle(newName)
    }
  }
)

const filteredMatches = computed(() => {
  if (selectedSeasonId.value === 0) {
    return matches.value
  }
  return matches.value.filter(m => m.season_id === selectedSeasonId.value)
})

const teamStats = computed(() => {
  let gamesPlayed = 0
  let wins = 0
  let losses = 0

  filteredMatches.value.forEach(match => {
    // Check if this team was involved in the match (only team matches)
    if (match.match_type === 'team' && (match.participant_a_id === teamId.value || match.participant_b_id === teamId.value)) {
      gamesPlayed++

      if (match.winner_id === teamId.value) {
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

const matchups = computed(() => {
  const matchupMap: Record<number, Matchup> = {}

  filteredMatches.value.forEach(match => {
    let opponentId = 0
    let isWin = false

    // Only process team matches
    if (match.match_type !== 'team') {
      return
    }

    // Determine if this team was participant_a or participant_b
    if (match.participant_a_id === teamId.value) {
      opponentId = match.participant_b_id
      isWin = match.winner_id === teamId.value
    } else if (match.participant_b_id === teamId.value) {
      opponentId = match.participant_a_id
      isWin = match.winner_id === teamId.value
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
        gamesPlayed: 0,
        winRate: 0
      }
    }

    if (isWin) {
      matchupMap[opponentId].wins++
    } else {
      matchupMap[opponentId].losses++
    }
    matchupMap[opponentId].gamesPlayed++
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
  return Object.values(matchups.value).sort((a, b) => b.winRate - a.winRate)
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

const loadTeamImage = async (id: number) => {
  try {
    const res = await fetch(`${api}/teams/${id}/image`)
    if (res.ok) {
      const blob = await res.blob()
      teamImageUrl.value = URL.createObjectURL(blob)
    }
    // 404 is expected for teams without images
  } catch (e) {
    console.error('Failed to load team image', e)
  }
}

const getBackgroundStyle = () => {
  if (!teamImageUrl.value) {
    return {}
  }
  return {
    backgroundImage: `linear-gradient(rgba(26, 26, 46, 0.85), rgba(22, 33, 62, 0.85)), url('${teamImageUrl.value}')`,
    backgroundSize: 'cover',
    backgroundPosition: 'center',
    backgroundAttachment: 'fixed',
    minHeight: '100vh'
  }
}

onMounted(() => {
  fetchTeam()
  fetchTeams()
  fetchSeasons()
  fetchMatches()
  loadTeamImage(teamId.value)
})

watch(teamId, () => {
  loading.value = true
  fetchTeam()
  fetchMatches()
  loadTeamImage(teamId.value)
})

const fetchTeam = async () => {
  try {
    const res = await fetch(`${api}/teams/${teamId.value}`)
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
  background-size: cover;
  background-position: center;
  background-attachment: fixed;
  min-height: 100vh;
  margin: calc(var(--spacing-lg) * -1);
  padding: var(--spacing-lg);
}

.team-detail-content {
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
  gap: var(--spacing-lg);
}

.matchup-card {
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

.matchup-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}

.card-name {
  font-size: 18px;
  font-weight: 600;
  color: white;
  text-decoration: none;
  transition: color 0.2s ease;
  white-space: normal;
  word-wrap: break-word;
  min-width: 150px;
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
  color: rgba(255, 255, 255, 0.6);
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
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

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

  .card-name {
    width: 100%;
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
    background: rgba(255, 255, 255, 0.05);
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
