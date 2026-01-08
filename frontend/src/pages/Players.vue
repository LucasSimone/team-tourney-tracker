<template>
  <div class="players-page">
    <div v-if="seasons.length > 0" class="season-filter">
      <select v-model.number="selectedSeasonId">
        <option value="0">All Seasons</option>
        <option v-for="season in seasons" :key="season.id" :value="season.id">
          Season {{ season.year }}{{ isCurrentSeason(season.id) ? ' (Current)' : '' }}
        </option>
      </select>
    </div>

    <div v-if="playerStats.length > 0" class="stats-section">
      <h2>All Players</h2>
      <div class="stats-cards">
        <div v-for="stat in playerStats" :key="stat.id" class="stat-card">
          <router-link :to="`/players/${stat.id}`" class="card-name">{{ stat.name }}</router-link>
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

    <div v-else class="empty-state">
      No players found yet.
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { API_URL } from '@/config'

interface Season { id?: number; year: number }
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
const seasons = ref<Season[]>([])
const playerStats = ref<PlayerStat[]>([])
const allPlayers = ref<{ id: number; name: string }[]>([])
const selectedSeasonId = ref(0)

onMounted(() => {
  fetchSeasons()
  fetchAllPlayers()
  fetchPlayerStats()
})

// Watch for season changes and refetch player stats
watch(selectedSeasonId, () => {
  fetchPlayerStats()
})

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

const fetchAllPlayers = async () => {
  try {
    const res = await fetch(`${api}/players`)
    if (res.ok) {
      const data = await res.json()
      allPlayers.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch all players', e)
  }
}

const fetchPlayerStats = async () => {
  try {
    const seasonParam = selectedSeasonId.value === 0 ? '0' : selectedSeasonId.value
    const res = await fetch(`${api}/players/stats?season_id=${seasonParam}`)
    if (res.ok) {
      const statsData = await res.json()
      const statsMap = new Map((statsData || []).map((stat: any) => [
        stat.id,
        {
          id: stat.id,
          name: stat.name,
          games_played: stat.games_played,
          wins: stat.wins,
          losses: stat.losses,
          win_percentage: stat.win_percentage,
          gamesPlayed: stat.games_played,
          winPercentage: stat.win_percentage
        }
      ]))

      // Create list from all players, using stats if available, otherwise 0 stats
      playerStats.value = (allPlayers.value || [])
        .map(player => {
          if (statsMap.has(player.id)) {
            return statsMap.get(player.id)!
          }
          // Player hasn't played yet
          return {
            id: player.id,
            name: player.name,
            games_played: 0,
            wins: 0,
            losses: 0,
            win_percentage: 0,
            gamesPlayed: 0,
            winPercentage: 0
          }
        })
        // Sort alphabetically by name
        .sort((a: PlayerStat, b: PlayerStat) => a.name.localeCompare(b.name))
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
</script>

<style scoped>
.players-page {
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
