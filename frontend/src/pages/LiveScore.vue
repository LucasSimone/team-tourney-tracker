<template>
  <div class="live-score-container">
    <h1>Live Score Tracker</h1>
    
    <div class="teams-row">
      <!-- Team A Card -->
      <div class="team-card" @click="showTeamModal('a')">
        <div v-if="teamA" class="team-selected">
          <h2>{{ teamA.name }}</h2>
          <div class="score-display" :class="getScoreClass('a')">{{ scoreA }}</div>
          <div class="score-controls">
            <button @click.stop="decreaseScore('a')" class="score-btn minus">−</button>
            <button @click.stop="increaseScore('a')" class="score-btn plus">+</button>
          </div>
        </div>
        <div v-else class="team-placeholder">
          <p>Click to select Team A</p>
        </div>
      </div>

      <!-- VS Separator -->
      <div class="vs-separator">
        <span>VS</span>
      </div>

      <!-- Team B Card -->
      <div class="team-card" @click="showTeamModal('b')">
        <div v-if="teamB" class="team-selected">
          <h2>{{ teamB.name }}</h2>
          <div class="score-display" :class="getScoreClass('b')">{{ scoreB }}</div>
          <div class="score-controls">
            <button @click.stop="decreaseScore('b')" class="score-btn minus">−</button>
            <button @click.stop="increaseScore('b')" class="score-btn plus">+</button>
          </div>
        </div>
        <div v-else class="team-placeholder">
          <p>Click to select Team B</p>
        </div>
      </div>
    </div>

    <!-- Submit Button -->
    <div v-if="teamA && teamB" class="submit-section">
      <button @click="submitMatch" class="submit-btn" :disabled="isSubmitting">
        {{ isSubmitting ? 'Submitting...' : 'Submit Match' }}
      </button>
    </div>

    <!-- Team Selection Modal -->
    <div v-if="showModal" class="modal-overlay" @click="closeModal">
      <div class="modal-content" @click.stop>
        <h2>Select {{ selectedTeamSlot === 'a' ? 'Team A' : 'Team B' }}</h2>
        
        <input 
          v-model="searchQuery" 
          type="text" 
          placeholder="Search teams..."
          class="search-input"
        />
        
        <div class="teams-list">
          <div 
            v-for="team in filteredTeams" 
            :key="team.id"
            class="team-option"
            @click="selectTeam(team)"
          >
            {{ team.name }}
          </div>
        </div>
        
        <button @click="closeModal" class="modal-close-btn">Cancel</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

interface Team {
  id: number
  name: string
}

interface Match {
  season_id: number
  team_a_id: number
  score_a: number
  team_b_id: number
  score_b: number
  winner_id: number
}

const router = useRouter()

// Teams
const teams = ref<Team[]>([])
const teamA = ref<Team | null>(null)
const teamB = ref<Team | null>(null)

// Scores
const scoreA = ref(0)
const scoreB = ref(0)

// Season
const currentSeasonId = ref<number | null>(null)

// Modal
const showModal = ref(false)
const selectedTeamSlot = ref<'a' | 'b'>('a')
const searchQuery = ref('')

// Team players mapping
const teamPlayers = ref<{ [teamId: number]: number[] }>({})

// Submission
const isSubmitting = ref(false)

const filteredTeams = computed(() => {
  return teams.value.filter(team => {
    const query = searchQuery.value.toLowerCase()
    const matchesSearch = team.name.toLowerCase().includes(query)
    const isNotSelectedTeam = team.id !== teamA.value?.id && team.id !== teamB.value?.id
    
    // Check if team shares players with the opposite team
    let noSharedPlayers = true
    if (selectedTeamSlot.value === 'a' && teamB.value) {
      const teamBPlayers = teamPlayers.value[teamB.value.id] || []
      const currentTeamPlayers = teamPlayers.value[team.id] || []
      const hasSharedPlayers = currentTeamPlayers.some(p => teamBPlayers.includes(p))
      noSharedPlayers = !hasSharedPlayers
    } else if (selectedTeamSlot.value === 'b' && teamA.value) {
      const teamAPlayers = teamPlayers.value[teamA.value.id] || []
      const currentTeamPlayers = teamPlayers.value[team.id] || []
      const hasSharedPlayers = currentTeamPlayers.some(p => teamAPlayers.includes(p))
      noSharedPlayers = !hasSharedPlayers
    }
    
    return matchesSearch && isNotSelectedTeam && noSharedPlayers
  })
})

const fetchTeams = async () => {
  try {
    const response = await fetch(`${API_URL}/teams`)
    const data = await response.json()
    teams.value = data || []
    // Fetch players for each team
    for (const team of teams.value) {
      await fetchTeamPlayers(team.id)
    }
  } catch (error) {
    console.error('Failed to fetch teams:', error)
  }
}

const fetchTeamPlayers = async (teamId: number) => {
  try {
    const response = await fetch(`${API_URL}/teams/${teamId}/players`)
    if (response.ok) {
      const data = await response.json()
      teamPlayers.value[teamId] = (data || []).map((p: any) => p.id)
    }
  } catch (error) {
    console.error(`Failed to fetch players for team ${teamId}:`, error)
  }
}

const fetchCurrentSeason = async () => {
  try {
    const response = await fetch(`${API_URL}/seasons`)
    const data = await response.json()
    if (data && data.length > 0) {
      // Get the most recent season (assuming they're ordered by year)
      currentSeasonId.value = data[data.length - 1].id
    }
  } catch (error) {
    console.error('Failed to fetch seasons:', error)
  }
}

const showTeamModal = (slot: 'a' | 'b') => {
  selectedTeamSlot.value = slot
  searchQuery.value = ''
  showModal.value = true
}

const selectTeam = (team: Team) => {
  if (selectedTeamSlot.value === 'a') {
    teamA.value = team
  } else {
    teamB.value = team
  }
  showModal.value = false
  scoreA.value = 0
  scoreB.value = 0
}

const closeModal = () => {
  showModal.value = false
  searchQuery.value = ''
}

const increaseScore = (team: 'a' | 'b') => {
  if (team === 'a') {
    scoreA.value++
  } else {
    scoreB.value++
  }
}

const decreaseScore = (team: 'a' | 'b') => {
  if (team === 'a' && scoreA.value > 0) {
    scoreA.value--
  } else if (team === 'b' && scoreB.value > 0) {
    scoreB.value--
  }
}

const getScoreClass = (team: 'a' | 'b'): string => {
  const scoreA_val = scoreA.value
  const scoreB_val = scoreB.value

  if (team === 'a') {
    if (scoreA_val > scoreB_val) return 'score-winning'
    if (scoreA_val < scoreB_val) return 'score-losing'
    return 'score-even'
  } else {
    if (scoreB_val > scoreA_val) return 'score-winning'
    if (scoreB_val < scoreA_val) return 'score-losing'
    return 'score-even'
  }
}

const submitMatch = async () => {
  if (!teamA.value || !teamB.value) {
    alert('Please select both teams')
    return
  }

  if (!currentSeasonId.value) {
    alert('No season found. Please create a season first.')
    return
  }

  isSubmitting.value = true

  try {
    // Determine winner based on scores
    let winnerId: number
    if (scoreA.value > scoreB.value) {
      winnerId = teamA.value.id
    } else if (scoreB.value > scoreA.value) {
      winnerId = teamB.value.id
    } else {
      // For ties, we'll use 0 or null - adjust based on your backend requirement
      winnerId = 0
    }

    const match: Match = {
      season_id: currentSeasonId.value,
      team_a_id: teamA.value.id,
      score_a: scoreA.value,
      team_b_id: teamB.value.id,
      score_b: scoreB.value,
      winner_id: winnerId
    }

    const response = await fetch(`${API_URL}/matches`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(match)
    })

    if (response.ok) {
      alert('Match submitted successfully!')
      teamA.value = null
      teamB.value = null
      scoreA.value = 0
      scoreB.value = 0
    } else {
      alert('Failed to submit match')
    }
  } catch (error) {
    console.error('Error submitting match:', error)
    alert('Error submitting match')
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  fetchTeams()
  fetchCurrentSeason()
})
</script>

<style scoped>
.live-score-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
}

h1 {
  text-align: center;
  color: #ffffff;
  margin-bottom: 2rem;
  font-size: 2rem;
}

.teams-row {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 4rem;
  align-items: center;
  margin-bottom: 2rem;
}

.team-card {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border: 2px solid #5b18c7;
  border-radius: 12px;
  padding: 2rem;
  cursor: pointer;
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.team-card:hover {
  border-color: #7237ce;
}

.team-placeholder {
  text-align: center;
  color: #888;
  font-size: 1.1rem;
  width: 100%;
}

.team-selected {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}

.team-selected h2 {
  color: #ffffff;
  margin: 0;
  font-size: 1.5rem;
  text-align: center;
  word-break: break-word;
}

.score-display {
  font-size: 4rem;
  font-weight: bold;
  color: #5b18c7;
  line-height: 1;
}

.score-display.score-winning {
  color: #4ade80;
}

.score-display.score-losing {
  color: #f87171;
}

.score-display.score-even {
  color: #ffffff;
}

.score-controls {
  display: flex;
  gap: 1rem;
  width: 100%;
  justify-content: space-between;
}

.score-btn {
  background-color: #5b18c7;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 8px;
  font-size: 1.5rem;
  cursor: pointer;
  transition: background-color 0.2s;
  font-weight: bold;
  min-width: 60px;
}

.score-btn:hover {
  background-color: #7237ce;
}

.score-btn.minus {
  background-color: #f87171;
}

.score-btn.minus:hover {
  background-color: #fb5757;
}

.score-btn.plus {
  background-color: #4ade80;
}

.score-btn.plus:hover {
  background-color: #5fd395;
}

.vs-separator {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: bold;
  color: #4ade80;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.submit-section {
  display: flex;
  justify-content: center;
  margin-top: 2rem;
}

.submit-btn {
  background: linear-gradient(135deg, #5b18c7 0%, #7237ce 100%);
  color: white;
  border: none;
  padding: 1rem 3rem;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s;
}

.submit-btn:hover:not(:disabled) {
  transform: scale(1.05);
  box-shadow: 0 0 20px rgba(91, 24, 199, 0.4);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border: 2px solid #5b18c7;
  border-radius: 12px;
  padding: 2rem;
  width: 90%;
  max-width: 400px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-content h2 {
  color: #ffffff;
  margin-top: 0;
  margin-bottom: 1.5rem;
  text-align: center;
}

.search-input {
  background-color: #0f3460;
  color: #ffffff;
  border: 1px solid #5b18c7;
  padding: 0.75rem;
  border-radius: 6px;
  margin-bottom: 1rem;
  font-size: 1rem;
}

.search-input::placeholder {
  color: #888;
}

.search-input:focus {
  outline: none;
  border-color: #7237ce;
  box-shadow: 0 0 10px rgba(91, 24, 199, 0.3);
}

.teams-list {
  flex: 1;
  overflow-y: auto;
  margin-bottom: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.team-option {
  background-color: #0f3460;
  color: #ffffff;
  padding: 0.75rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}

.team-option:hover {
  background-color: #1a4d6d;
  border-color: #5b18c7;
}

.modal-close-btn {
  background-color: #f87171;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
  align-self: center;
}

.modal-close-btn:hover {
  background-color: #fb5757;
}

@media (max-width: 768px) {
  .teams-row {
    grid-template-columns: 1fr;
    gap: 2rem;
  }

  .vs-separator {
    display: none;
  }

  .team-card {
    min-height: 250px;
  }

  .score-display {
    font-size: 3rem;
  }

  .modal-content {
    width: 95%;
  }
}
</style>
