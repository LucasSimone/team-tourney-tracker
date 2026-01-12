<template>
  <div class="teams-page">
    <div v-if="seasons.length > 0" class="season-filter">
      <select v-model.number="selectedSeasonId">
        <option value="0">All Seasons</option>
        <option v-for="season in seasons" :key="season.id" :value="season.id">
          Season {{ season.year }}{{ isCurrentSeason(season.id) ? ' (Current)' : '' }}
        </option>
      </select>
    </div>

    <div v-if="teamStats.length > 0" class="stats-section">
      <h2>All Teams</h2>
      <div class="stats-cards">
        <div v-for="stat in teamStats" :key="stat.id" class="stat-card" :style="getCardStyle(stat.id)">
          <router-link :to="`/teams/${stat.id}`" class="card-name">{{ stat.name }}</router-link>
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

    <div v-else class="empty-state">
      No teams found yet.
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
}
interface TeamStat {
  id: number
  name: string
  gamesPlayed: number
  wins: number
  losses: number
  winPercentage: number
}

const api = API_URL
const teams = ref<Team[]>([])
const seasons = ref<Season[]>([])
const matches = ref<Match[]>([])
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

    // Count wins and losses
    if (match.winner_id === match.participant_a_id) {
      if (stats[match.participant_a_id]) {
        stats[match.participant_a_id].wins++
      }
      if (stats[match.participant_b_id]) {
        stats[match.participant_b_id].losses++
      }
    } else if (match.winner_id === match.participant_b_id) {
      if (stats[match.participant_b_id]) {
        stats[match.participant_b_id].wins++
      }
      if (stats[match.participant_a_id]) {
        stats[match.participant_a_id].losses++
      }
    }
  })

  // Calculate win percentages and sort alphabetically
  return Object.values(stats)
    .map(stat => {
      const total = stat.wins + stat.losses
      stat.winPercentage = total > 0 ? Math.round((stat.wins / total) * 100) : 0
      return stat
    })
    .sort((a, b) => a.name.localeCompare(b.name))
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
  fetchTeams()
  fetchSeasons()
  fetchMatches()
})

watch(selectedSeasonId, () => {
  // Stats will automatically recompute
})

const fetchTeams = async () => {
  try {
    const res = await fetch(`${api}/teams`)
    if (res.ok) {
      const data = await res.json()
      teams.value = data || []
      // Load horizontal images for each team
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

const loadTeamImage = async (teamId: number) => {
  try {
    const res = await fetch(`${api}/teams/${teamId}/image/horizontal`)
    if (res.ok) {
      const blob = await res.blob()
      teamImages.value[teamId] = URL.createObjectURL(blob)
    }
  } catch (e) {
    // 404 is expected for teams without images
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

const getCardStyle = (teamId: number) => {
  const imageUrl = teamImages.value[teamId]
  if (!imageUrl) {
    // Return default blueish background if no image
    return {
      background: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)'
    }
  }
  return {
    backgroundImage: `linear-gradient(135deg, rgba(26, 26, 46, 0.85) 0%, rgba(22, 33, 62, 0.85) 100%), url(${imageUrl})`,
    backgroundSize: 'contain',
    backgroundPosition: 'center'
  }
}
</script>

<style scoped>
.teams-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
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

.empty-state {
  padding: var(--spacing-xl);
  text-align: center;
  color: var(--text-secondary);
  background-color: var(--clr-surface-a10);
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

.card-name {
  font-size: 18px;
  font-weight: 600;
  color: white;
  text-decoration: none;
  transition: color 0.2s ease;
  white-space: normal;
  word-wrap: break-word;
  text-align: left;
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

@media (max-width: 768px) {
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

  .card-name {
    text-align: left;
    min-width: unset;
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
