<template>
  <div class="live-score-container">
    <!-- Header with Selectors -->
    <div v-if="!reviewMode" class="header-selectors">
      <select v-model="matchType" class="type-select">
        <option value="team">Team</option>
        <option value="single">Single</option>
      </select>
    </div>

    <!-- Step 1: Score Entry -->
    <div v-if="!reviewMode" class="step-1">
      <!-- Team Match -->
      <div v-if="matchType === 'team'" class="teams-row">
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

      <!-- Single Match -->
      <div v-else class="teams-row">
        <!-- Player A Card -->
        <div class="team-card" @click="showPlayerModal('a')">
          <div v-if="playerA" class="team-selected">
            <h2>{{ playerA.name }}</h2>
            <div class="score-display" :class="getScoreClass('a')">{{ scoreA }}</div>
            <div class="score-controls">
              <button @click.stop="decreaseScore('a')" class="score-btn minus">−</button>
              <button @click.stop="increaseScore('a')" class="score-btn plus">+</button>
            </div>
          </div>
          <div v-else class="team-placeholder">
            <p>Click to select Player A</p>
          </div>
        </div>

        <!-- VS Separator -->
        <div class="vs-separator">
          <span>VS</span>
        </div>

        <!-- Player B Card -->
        <div class="team-card" @click="showPlayerModal('b')">
          <div v-if="playerB" class="team-selected">
            <h2>{{ playerB.name }}</h2>
            <div class="score-display" :class="getScoreClass('b')">{{ scoreB }}</div>
            <div class="score-controls">
              <button @click.stop="decreaseScore('b')" class="score-btn minus">−</button>
              <button @click.stop="increaseScore('b')" class="score-btn plus">+</button>
            </div>
          </div>
          <div v-else class="team-placeholder">
            <p>Click to select Player B</p>
          </div>
        </div>
      </div>

      <!-- Submit Button -->
      <div v-if="(matchType === 'team' && teamA && teamB) || (matchType === 'single' && playerA && playerB)" class="submit-section">
        <button @click="proceedToReview" class="submit-btn" :disabled="isSubmitting">
          {{ isSubmitting ? 'Processing...' : 'Next' }}
        </button>
      </div>
    </div>

    <!-- Step 2: Review and Confirmation -->
    <div v-if="reviewMode" class="step-2">
      <div class="review-card">
        <h2>Review Match</h2>
        
        <div class="review-content">
          <div class="review-matchup">
            <div class="team-info">
              <div class="team-name">{{ matchType === 'team' ? teamA?.name : playerA?.name }}</div>
              <div class="team-score" :class="getScoreClass('a')">{{ scoreA }}</div>
            </div>
            <div class="vs-text">vs</div>
            <div class="team-info">
              <div class="team-name">{{ matchType === 'team' ? teamB?.name : playerB?.name }}</div>
              <div class="team-score" :class="getScoreClass('b')">{{ scoreB }}</div>
            </div>
          </div>

          <div v-if="autoWinner" class="winner-display">
            <p>🏆 {{ getWinnerName(autoWinner) }} wins</p>
          </div>
        </div>

        <div class="review-actions">
          <button @click="reviewMode = false" class="btn-back">← Back</button>
          <button 
            @click="submitMatch" 
            class="btn-submit" 
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? 'Submitting...' : 'Confirm & Submit' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Team/Player Selection Modal -->
    <div v-if="showModal" class="modal-overlay" @click="closeModal">
      <div class="modal-content" @click.stop>
        <h2>{{ modalType === 'team' ? `Select ${selectedTeamSlot === 'a' ? 'Team A' : 'Team B'}` : `Select ${selectedTeamSlot === 'a' ? 'Player A' : 'Player B'}` }}</h2>
        
        <input 
          v-model="searchQuery" 
          type="text" 
          :placeholder="`Search ${modalType === 'team' ? 'teams' : 'players'}...`"
          class="search-input"
        />
        
        <div class="teams-list">
          <div 
            v-for="item in (modalType === 'team' ? filteredTeams : filteredPlayers)" 
            :key="item.id"
            class="team-option"
            @click="selectItem(item)"
          >
            {{ item.name }}
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
import { useAuth } from '@/composables/useAuth'
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

interface Team {
  id: number
  name: string
}

interface Player {
  id: number
  name: string
}

interface Match {
  season_id: number
  match_type: string
  participant_a_id?: number
  participant_b_id?: number
  score_a: number
  score_b: number
  winner_id: number
}

const router = useRouter()

// Auth
const { token } = useAuth()

// Teams and Players
const teams = ref<Team[]>([])
const players = ref<Player[]>([])
const teamA = ref<Team | null>(null)
const teamB = ref<Team | null>(null)
const playerA = ref<Player | null>(null)
const playerB = ref<Player | null>(null)

// Scores
const scoreA = ref(0)
const scoreB = ref(0)

// Match Type
const matchType = ref<'team' | 'single'>('team')

// Season
const currentSeasonId = ref<number | null>(null)

// Modal
const showModal = ref(false)
const selectedTeamSlot = ref<'a' | 'b'>('a')
const modalType = ref<'team' | 'player'>('team')
const searchQuery = ref('')

// Team players mapping
const teamPlayers = ref<{ [teamId: number]: number[] }>({})

// Submission
const isSubmitting = ref(false)

// Review mode
const reviewMode = ref(false)
const autoWinner = ref<number | null>(null)

// Messages
interface Message {
  type: 'success' | 'error'
  text: string
}
const message = ref<Message | null>(null)

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

const filteredPlayers = computed(() => {
  return players.value.filter(player => {
    const query = searchQuery.value.toLowerCase()
    const matchesSearch = player.name.toLowerCase().includes(query)
    const isNotSelectedPlayer = player.id !== playerA.value?.id && player.id !== playerB.value?.id
    return matchesSearch && isNotSelectedPlayer
  })
})

const fetchTeams = async () => {
  try {
    const response = await fetch(`${API_URL}/teams`)
    const data = await response.json()
    teams.value = data || []
    // Populate teamPlayers from the team objects (now includes players)
    for (const team of teams.value) {
      teamPlayers.value[team.id] = team.players || []
    }
  } catch (error) {
    console.error('Failed to fetch teams:', error)
  }
}

const fetchPlayers = async () => {
  try {
    const response = await fetch(`${API_URL}/players`)
    const data = await response.json()
    players.value = data || []
  } catch (error) {
    console.error('Failed to fetch players:', error)
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
      // Get the current year
      const currentYear = new Date().getFullYear()
      // Find the season matching the current year
      const currentSeason = data.find((season: any) => season.year === currentYear)
      if (currentSeason) {
        currentSeasonId.value = currentSeason.id
      } else {
        // If current year season doesn't exist, fall back to the most recent season
        currentSeasonId.value = data[data.length - 1].id
        console.warn(`No season found for year ${currentYear}, using most recent season`)
      }
    }
  } catch (error) {
    console.error('Failed to fetch seasons:', error)
  }
}

const showTeamModal = (slot: 'a' | 'b') => {
  selectedTeamSlot.value = slot
  modalType.value = 'team'
  searchQuery.value = ''
  showModal.value = true
}

const showPlayerModal = (slot: 'a' | 'b') => {
  selectedTeamSlot.value = slot
  modalType.value = 'player'
  searchQuery.value = ''
  showModal.value = true
}

const selectItem = (item: Team | Player) => {
  if (modalType.value === 'team') {
    const team = item as Team
    if (selectedTeamSlot.value === 'a') {
      teamA.value = team
    } else {
      teamB.value = team
    }
  } else {
    const player = item as Player
    if (selectedTeamSlot.value === 'a') {
      playerA.value = player
    } else {
      playerB.value = player
    }
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

const getTeamName = (id: number): string => {
  return teams.value.find(t => t.id === id)?.name || 'Unknown Team'
}

const getPlayerName = (id: number): string => {
  return players.value.find(p => p.id === id)?.name || 'Unknown Player'
}

const getWinnerName = (id: number): string => {
  if (matchType.value === 'team') {
    return getTeamName(id)
  } else {
    return getPlayerName(id)
  }
}

const proceedToReview = () => {
  if (matchType.value === 'team') {
    if (!teamA.value || !teamB.value) {
      showMessage('Please select both teams', 'error')
      return
    }
  } else {
    if (!playerA.value || !playerB.value) {
      showMessage('Please select both players', 'error')
      return
    }
  }

  // Prevent 0-0 scores
  if (scoreA.value === 0 && scoreB.value === 0) {
    showMessage('Scores cannot be 0-0. At least one must score.', 'error')
    return
  }

  // Auto-determine winner from scores
  if (scoreA.value > scoreB.value) {
    autoWinner.value = matchType.value === 'team' ? teamA.value!.id : playerA.value!.id
  } else if (scoreB.value > scoreA.value) {
    autoWinner.value = matchType.value === 'team' ? teamB.value!.id : playerB.value!.id
  } else {
    // Tie - set to null, backend will handle
    autoWinner.value = null
  }

  reviewMode.value = true
}

const showMessage = (text: string, type: 'success' | 'error') => {
  message.value = { text, type }
}

const submitMatch = async () => {
  if (matchType.value === 'team') {
    if (!teamA.value || !teamB.value) {
      showMessage('Please select both teams', 'error')
      return
    }
  } else {
    if (!playerA.value || !playerB.value) {
      showMessage('Please select both players', 'error')
      return
    }
  }

  if (!currentSeasonId.value) {
    showMessage('No season found. Please create a season first.', 'error')
    return
  }

  isSubmitting.value = true

  try {
    // Use the winner determined in review mode
    let winnerId: number | null = autoWinner.value

    const match: any = {
      season_id: currentSeasonId.value,
      match_type: matchType.value,
      score_a: scoreA.value,
      score_b: scoreB.value,
      winner_id: winnerId || 0
    }

    if (matchType.value === 'team') {
      match.participant_a_id = teamA.value!.id
      match.participant_b_id = teamB.value!.id
    } else {
      match.participant_a_id = playerA.value!.id
      match.participant_b_id = playerB.value!.id
    }

    const response = await fetch(`${API_URL}/matches`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token.value && { 'Authorization': `Bearer ${token.value}` })
      },
      body: JSON.stringify(match)
    })

    if (response.ok) {
      showMessage('Match recorded successfully! 🎉', 'success')
      // Reset form
      teamA.value = null
      teamB.value = null
      playerA.value = null
      playerB.value = null
      scoreA.value = 0
      scoreB.value = 0
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

onMounted(() => {
  fetchTeams()
  fetchPlayers()
  fetchCurrentSeason()
})
</script>

<style scoped>
.live-score-container {
  max-width: 900px;
  margin: 0 auto;
}

/* Header Selectors */
.header-selectors {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  justify-content: flex-end;
}

.type-select {
  background-color: #0f3460;
  color: #ffffff;
  border: 1px solid #5b18c7;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: border-color 0.2s;
}

.type-select:hover,
.type-select:focus {
  outline: none;
  border-color: #7237ce;
}

/* Step 1: Team Cards */
.step-1 {
  display: flex;
  flex-direction: column;
}

.teams-row {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 4rem;
  align-items: center;
  margin-bottom: 1rem;
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
  margin-top: 0.5rem;
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
    gap: 1rem;
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

.team-score.score-winner {
  color: #22c55e;
}

.team-score.score-loser {
  color: #ef4444;
}

.vs-text {
  color: #888;
  font-weight: bold;
  text-transform: uppercase;
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
  position: fixed;
  top: 20px;
  right: 20px;
  padding: 1rem 1.5rem;
  border-radius: 8px;
  font-weight: bold;
  z-index: 2000;
  max-width: 300px;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

.message.success {
  background-color: #4ade80;
  color: #000;
}

.message.error {
  background-color: #f87171;
  color: #fff;
}

@media (max-width: 768px) {
  .teams-row {
    grid-template-columns: 1fr;
    gap: 1rem;
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
