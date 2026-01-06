<template>
  <div class="seasons-page">
    <h1>Seasons</h1>

    <div v-if="seasons.length === 0" class="empty-state">
      No seasons available yet. Check back soon!
    </div>

    <div v-else class="seasons-grid">
      <div v-for="season in seasonStats" :key="season.id" class="season-card">
        <div class="season-year">{{ season.year }}</div>
        <div class="season-stats">
          <div class="stat">
            <span class="stat-label">Teams</span>
            <span class="stat-value">{{ season.teamCount }}</span>
          </div>
          <div class="stat">
            <span class="stat-label">Matches</span>
            <span class="stat-value">{{ season.matchCount }}</span>
          </div>
        </div>
        <router-link 
          :to="`/standings?season=${season.id}`"
          class="view-standings-btn"
        >
          View Standings →
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, computed, onMounted } from 'vue'

interface Season { id?: number; year: number }
interface Match { 
  id?: number
  season_id: number
  team_a_id: number
  team_b_id: number
  winner_id?: number
}
interface SeasonStat {
  id?: number
  year: number
  teamCount: number
  matchCount: number
}

const api = API_URL
const seasons = ref<Season[]>([])
const matches = ref<Match[]>([])

const seasonStats = computed(() => {
  return seasons.value
    .map(season => {
      const seasonMatches = matches.value.filter(m => m.season_id === season.id)
      const teamsInSeason = new Set<number>()
      
      seasonMatches.forEach(match => {
        teamsInSeason.add(match.team_a_id)
        teamsInSeason.add(match.team_b_id)
      })

      return {
        id: season.id,
        year: season.year,
        teamCount: teamsInSeason.size,
        matchCount: seasonMatches.length
      } as SeasonStat
    })
    .sort((a, b) => b.year - a.year)
})

onMounted(() => {
  fetchSeasons()
  fetchMatches()
})

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
.seasons-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

h1 {
  margin: 0;
  font-size: 32px;
}

.empty-state {
  padding: var(--spacing-xl);
  text-align: center;
  color: var(--text-secondary);
  background-color: var(--clr-surface-a10);
  border-radius: var(--radius);
}

.seasons-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-lg);
}

.season-card {
  background-color: var(--clr-surface-a10);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  transition: transform 0.2s, box-shadow 0.2s;
}

.season-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.season-year {
  font-size: 36px;
  font-weight: bold;
  color: var(--clr-primary-a50);
}

.season-stats {
  display: flex;
  gap: var(--spacing-lg);
  padding: var(--spacing-lg) 0;
  border-top: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
}

.stat {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-sm);
}

.stat-label {
  font-size: 12px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--text-primary);
}

.view-standings-btn {
  display: inline-block;
  padding: var(--spacing-md) var(--spacing-lg);
  background-color: var(--clr-primary-a0);
  color: white;
  border-radius: var(--radius);
  text-decoration: none;
  text-align: center;
  font-weight: 500;
  transition: background-color 0.2s;
}

.view-standings-btn:hover {
  background-color: #7d2aed;
}

@media (max-width: 768px) {
  h1 {
    font-size: 24px;
  }

  .seasons-grid {
    grid-template-columns: 1fr;
  }

  .season-card {
    padding: var(--spacing-lg);
  }

  .season-year {
    font-size: 28px;
  }

  .season-stats {
    gap: var(--spacing-md);
    padding: var(--spacing-md) 0;
  }

  .stat-value {
    font-size: 20px;
  }
}
</style>
