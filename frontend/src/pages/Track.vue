<template>
  <div class="track-page">
    <h1>Track Match</h1>

    <!-- Season Selector -->
    <div class="season-selector">
      <select v-model.number="selectedSeasonId" class="season-select">
        <option v-for="season in seasons" :key="season.id" :value="season.id">
          Season {{ season.year }}
        </option>
      </select>
    </div>

    <!-- Step 1: Team Selection and Score Entry -->
    <div v-if="!reviewMode" class="step-1">
      <div class="teams-row">
        <!-- Team A Card -->
        <div class="team-card" @click="showTeamModal('a')">
          <div v-if="teamA" class="team-selected">
            <h2>{{ teamA.name }}</h2>
            <input 
              v-model.number="scoreA" 
              type="number" 
              class="score-input" 
              placeholder="Score"
              min="0"
              @click.stop
            />
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
            <input 
              v-model.number="scoreB" 
              type="number" 
              class="score-input" 
              placeholder="Score"
              min="0"
              @click.stop
            />
          </div>
          <div v-else class="team-placeholder">
            <p>Click to select Team B</p>
          </div>
        </div>
      </div>

      <!-- Submit Button -->
      <div v-if="teamA && teamB" class="submit-section">
        <button @click="proceedToReview" class="submit-btn" :disabled="isSubmitting">
          {{ isSubmitting ? 'Processing...' : 'Next' }}
        </button>
      </div>
    </div>

    <!-- Step 2: Review and Winner Selection -->
    <div v-if="reviewMode" class="step-2">
      <div class="review-card">
        <h2>Review Match</h2>
        
        <div class="review-content">
          <div class="review-row">
            <span class="label">Season:</span>
            <span class="value">{{ getSeasonYear(selectedSeasonId) }}</span>
          </div>
          
          <div class="review-matchup">
            <div class="team-info">
              <div class="team-name">{{ teamA?.name }}</div>
              <div class="team-score">{{ scoreA ?? 0 }}</div>
            </div>
            <div class="vs-text">vs</div>
            <div class="team-info">
              <div class="team-name">{{ teamB?.name }}</div>
              <div class="team-score">{{ scoreB ?? 0 }}</div>
            </div>
          </div>

          <!-- Winner Selection (only if scores are 0-0 or not provided) -->
          <div v-if="(scoreA === null || scoreA === 0) && (scoreB === null || scoreB === 0) && !autoWinner" class="winner-selection">
            <p>Select the winner:</p>
            <div class="winner-buttons">
              <button 
                @click="setWinner(teamA!.id)" 
                class="winner-btn"
                :class="{ selected: selectedWinnerId === teamA?.id }"
              >
                {{ teamA?.name }}
              </button>
              <button 
                @click="setWinner(teamB!.id)" 
                class="winner-btn"
                :class="{ selected: selectedWinnerId === teamB?.id }"
              >
                {{ teamB?.name }}
              </button>
            </div>
          </div>

          <div v-else-if="autoWinner" class="winner-display">
            <p>🏆 {{ getTeamName(autoWinner) }} wins</p>
          </div>
        </div>

        <div class="review-actions">
          <button @click="reviewMode = false" class="btn-back">← Back</button>
          <button 
            @click="submitMatch" 
            class="btn-submit" 
            :disabled="isSubmitting || ((scoreA === null || scoreA === 0) && (scoreB === null || scoreB === 0) && !selectedWinnerId)"
          >
            {{ isSubmitting ? 'Submitting...' : 'Confirm & Submit' }}
          </button>
        </div>
      </div>
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

    <!-- Message Display -->
    <div v-if="message" :class="['message', message.type]">
      {{ message.text }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, computed, onMounted } from 'vue'
import { useAuth } from '@/composables/useAuth'

interface Team {
  id: number
  name: string
}

interface Season {
  id: number
  year: number
}

interface Message {
  type: 'success' | 'error'
  text: string
}

const { getAuthHeaders } = useAuth()

// API
const api = API_URL

// Data
const teams = ref<Team[]>([])
const seasons = ref<Season[]>([])
const teamA = ref<Team | null>(null)
const teamB = ref<Team | null>(null)
const scoreA = ref<number | null>(null)
const scoreB = ref<number | null>(null)
const selectedSeasonId = ref<number>(0)

// Modal
const showModal = ref(false)
const selectedTeamSlot = ref<'a' | 'b'>('a')
const searchQuery = ref('')

// Review
const reviewMode = ref(false)
const selectedWinnerId = ref<number | null>(null)
const autoWinner = ref<number | null>(null)

// UI
const isSubmitting = ref(false)
const message = ref<Message | null>(null)

// Team players mapping
const teamPlayers = ref<{ [teamId: number]: number[] }>({})

const currentYear = new Date().getFullYear()

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
    const response = await fetch(`${api}/teams`, {
      headers: getAuthHeaders()
    })
    if (response.ok) {
      const data = await response.json()
      teams.value = data || []
      // Fetch players for each team
      for (const team of teams.value) {
        await fetchTeamPlayers(team.id)
      }
    }
  } catch (error) {
    console.error('Failed to fetch teams:', error)
  }
}

const fetchTeamPlayers = async (teamId: number) => {
  try {
    const response = await fetch(`${api}/teams/${teamId}/players`, {
      headers: getAuthHeaders()
    })
    if (response.ok) {
      const data = await response.json()
      teamPlayers.value[teamId] = (data || []).map((p: any) => p.id)
    }
  } catch (error) {
    console.error(`Failed to fetch players for team ${teamId}:`, error)
  }
}

const fetchSeasons = async () => {
  try {
    const response = await fetch(`${api}/seasons`, {
      headers: getAuthHeaders()
    })
    if (response.ok) {
      const data = await response.json()
      seasons.value = data || []
      // Set default to current year
      const currentSeason = data.find((s: Season) => s.year === currentYear)
      if (currentSeason) {
        selectedSeasonId.value = currentSeason.id
      } else if (data.length > 0) {
        selectedSeasonId.value = data[data.length - 1].id
      }
    }
  } catch (error) {
    console.error('Failed to fetch seasons:', error)
  }
}

const getTeamName = (id: number): string => {
  return teams.value.find(t => t.id === id)?.name || 'Unknown Team'
}

const getSeasonYear = (id: number): number => {
  return seasons.value.find(s => s.id === id)?.year || 0
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
}

const closeModal = () => {
  showModal.value = false
  searchQuery.value = ''
}

const proceedToReview = () => {
  if (!teamA.value || !teamB.value) {
    showMessage('Please select both teams', 'error')
    return
  }

  // Determine if we have valid scores that aren't 0-0
  const scoreAValue = scoreA.value ?? 0
  const scoreBValue = scoreB.value ?? 0
  const hasRealScores = scoreA.value !== null && scoreB.value !== null && (scoreAValue !== 0 || scoreBValue !== 0)

  if (hasRealScores) {
    // Auto-determine winner from scores
    if (scoreAValue > scoreBValue) {
      autoWinner.value = teamA.value.id
    } else if (scoreBValue > scoreAValue) {
      autoWinner.value = teamB.value.id
    } else {
      // Tie - need manual selection
      autoWinner.value = null
      selectedWinnerId.value = null
    }
  } else {
    // No scores or 0-0 - need manual selection
    autoWinner.value = null
    selectedWinnerId.value = null
  }

  reviewMode.value = true
}

const setWinner = (teamId: number) => {
  selectedWinnerId.value = teamId
}

const submitMatch = async () => {
  if (!teamA.value || !teamB.value || !selectedSeasonId.value) {
    showMessage('Missing required information', 'error')
    return
  }

  let winnerId = autoWinner.value || selectedWinnerId.value
  if (!winnerId) {
    showMessage('Please select a winner', 'error')
    return
  }

  isSubmitting.value = true

  try {
    const matchData = {
      season_id: selectedSeasonId.value,
      team_a_id: teamA.value.id,
      team_b_id: teamB.value.id,
      score_a: scoreA.value ?? 0,
      score_b: scoreB.value ?? 0,
      winner_id: winnerId
    }

    const response = await fetch(`${api}/matches`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(matchData)
    })

    if (response.ok) {
      showMessage('Match recorded successfully! 🎉', 'success')
      // Reset form
      teamA.value = null
      teamB.value = null
      scoreA.value = null
      scoreB.value = null
      selectedWinnerId.value = null
      autoWinner.value = null
      reviewMode.value = false
      setTimeout(() => {
        message.value = null
      }, 3000)
    } else {
      showMessage('Failed to record match', 'error')
    }
  } catch (error) {
    console.error('Error submitting match:', error)
    showMessage('Error recording match', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const showMessage = (text: string, type: 'success' | 'error') => {
  message.value = { text, type }
}

onMounted(() => {
  fetchTeams()
  fetchSeasons()
})
</script>

<style scoped>
.track-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

h1 {
  text-align: center;
  color: #ffffff;
  margin-bottom: 2rem;
  font-size: 2rem;
}

/* Season Selector */
.season-selector {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 2rem;
}

.season-select {
  background-color: #0f3460;
  color: #ffffff;
  border: 1px solid #5b18c7;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: border-color 0.2s;
}

.season-select:hover,
.season-select:focus {
  outline: none;
  border-color: #7237ce;
}

/* Step 1: Team Cards */
.step-1 {
  display: flex;
  flex-direction: column;
  gap: 2rem;
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
  min-height: 250px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.2s;
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

.score-input {
  width: 100%;
  max-width: 120px;
  padding: 0.75rem;
  font-size: 2rem;
  text-align: center;
  background-color: #0f3460;
  color: #5b18c7;
  border: 2px solid #5b18c7;
  border-radius: 8px;
  font-weight: bold;
  transition: border-color 0.2s;
}

.score-input:focus {
  outline: none;
  border-color: #7237ce;
}

.vs-separator {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: bold;
  color: #5b18c7;
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

/* Step 2: Review */
.step-2 {
  display: flex;
  justify-content: center;
}

.review-card {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  border: 2px solid #5b18c7;
  border-radius: 12px;
  padding: 2rem;
  width: 100%;
  max-width: 600px;
}

.review-card h2 {
  color: #ffffff;
  margin-top: 0;
  margin-bottom: 1.5rem;
  text-align: center;
}

.review-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.review-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background-color: #0f3460;
  border-radius: 6px;
}

.review-row .label {
  color: #888;
  font-weight: 500;
}

.review-row .value {
  color: #ffffff;
  font-weight: bold;
}

.review-matchup {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 1rem;
  align-items: center;
  padding: 1.5rem;
  background-color: #0f3460;
  border-radius: 8px;
}

.team-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.team-name {
  color: #ffffff;
  font-weight: bold;
  font-size: 1.1rem;
  text-align: center;
}

.team-score {
  color: #5b18c7;
  font-size: 2rem;
  font-weight: bold;
}

.vs-text {
  color: #888;
  font-weight: bold;
  text-transform: uppercase;
}

.winner-selection {
  padding: 1rem;
  background-color: #0f3460;
  border-radius: 8px;
}

.winner-selection p {
  color: #ffffff;
  margin: 0 0 1rem 0;
  text-align: center;
}

.winner-buttons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.winner-btn {
  padding: 0.75rem;
  background-color: #1a4d6d;
  color: #ffffff;
  border: 2px solid #5b18c7;
  border-radius: 6px;
  cursor: pointer;
  font-weight: bold;
  transition: all 0.2s;
}

.winner-btn:hover {
  background-color: #2a5d7d;
  border-color: #7237ce;
}

.winner-btn.selected {
  background-color: #5b18c7;
  border-color: #7237ce;
}

.winner-display {
  padding: 1rem;
  background-color: #1a4d6d;
  border-radius: 8px;
  text-align: center;
}

.winner-display p {
  color: #4ade80;
  font-size: 1.3rem;
  font-weight: bold;
  margin: 0;
}

.review-actions {
  display: flex;
  gap: 1rem;
  justify-content: space-between;
}

.btn-back,
.btn-submit {
  flex: 1;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-back {
  background-color: #f87171;
  color: white;
}

.btn-back:hover {
  background-color: #fb5757;
}

.btn-submit {
  background: linear-gradient(135deg, #5b18c7 0%, #7237ce 100%);
  color: white;
}

.btn-submit:hover:not(:disabled) {
  box-shadow: 0 0 15px rgba(91, 24, 199, 0.4);
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Modal */
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

/* Message */
.message {
  margin-top: 2rem;
  padding: 1rem;
  border-radius: 8px;
  text-align: center;
  font-weight: bold;
  animation: slideIn 0.3s ease;
}

.message.success {
  background-color: rgba(74, 222, 128, 0.1);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.3);
}

.message.error {
  background-color: rgba(248, 113, 113, 0.1);
  color: #f87171;
  border: 1px solid rgba(248, 113, 113, 0.3);
}

@keyframes slideIn {
  from {
    transform: translateY(-10px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@media (max-width: 768px) {
  .track-page {
    padding: 1rem;
  }

  .teams-row {
    grid-template-columns: 1fr;
    gap: 2rem;
  }

  .vs-separator {
    display: none;
  }

  .team-card {
    min-height: 200px;
  }

  .score-input {
    font-size: 1.5rem;
  }

  .modal-content {
    width: 95%;
  }
}
</style>
