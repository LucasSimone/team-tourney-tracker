<template>
  <div class="admin-page">
    <h2>Manage Matches</h2>

    <div class="form-group">
      <div class="form-row">
        <div class="form-column">
          <label>Season</label>
          <select v-model.number="newMatch.season_id">
            <option value="0">Select Season</option>
            <option v-for="season in seasons" :key="season.id" :value="season.id">
              Season {{ season.year }}
            </option>
          </select>
        </div>
      </div>

      <div class="form-row">
        <div class="form-column">
          <label>{{ newMatch.team_a_id ? getTeamName(newMatch.team_a_id) : 'Team 1' }}</label>
          <select v-model.number="newMatch.team_a_id">
            <option value="0">Select Team 1</option>
            <option v-for="team in teams" :key="team.id" :value="team.id">
              {{ team.name }}
            </option>
          </select>
          <input v-model="newMatch.score_a" type="number" placeholder="Score" class="score-input" />
        </div>

        <div class="form-column">
          <label>{{ newMatch.team_b_id ? getTeamName(newMatch.team_b_id) : 'Team 2' }}</label>
          <select v-model.number="newMatch.team_b_id">
            <option value="0">Select Team 2</option>
            <option v-for="team in teams" :key="team.id" :value="team.id">
              {{ team.name }}
            </option>
          </select>
          <input v-model="newMatch.score_b" type="number" placeholder="Score" class="score-input" />
        </div>
      </div>

      <div class="form-row">
        <div class="form-column">
          <label>Winner</label>
          <select v-model.number="newMatch.winner_id" :disabled="!newMatch.team_a_id || !newMatch.team_b_id">
            <option value="0">{{ getWinnerLabel(newMatch) }}</option>
            <option v-if="newMatch.team_a_id" :value="newMatch.team_a_id">
              {{ getTeamName(newMatch.team_a_id) }}
            </option>
            <option v-if="newMatch.team_b_id" :value="newMatch.team_b_id">
              {{ getTeamName(newMatch.team_b_id) }}
            </option>
          </select>
        </div>
      </div>

      <div class="form-row">
        <button @click="addMatch" class="btn-record">Record Match</button>
      </div>
    </div>

    <div v-if="matches.length === 0" class="empty-state">No matches yet. Create one to get started!</div>

    <div v-else class="items-grid">
      <div v-for="match in matches" :key="match.id" class="item-card">
        <div class="item-header">
          <h3>{{ getTeamName(match.team_a_id) }} vs {{ getTeamName(match.team_b_id) }}</h3>
          <p class="match-score">{{ match.score_a !== undefined && match.score_b !== undefined ? `${match.score_a} - ${match.score_b}` : '—' }}</p>
          <p class="match-winner">Winner: {{ getTeamName(match.winner_id) || '—' }}</p>
          <p class="match-season">Season {{ getSeasonYear(match.season_id) }}</p>
          <p class="match-date">{{ formatDateTime(match.created_at) }}</p>
        </div>
        <div class="item-actions">
          <button @click="editMatch(match)" class="btn-secondary">Edit</button>
          <button @click="deleteMatch(match.id)" class="btn-delete">Delete</button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="editingMatch" class="modal-overlay" @click="editingMatch = null">
      <div class="modal" @click.stop>
        <h3>Edit Match</h3>
        <div class="modal-form-group">
          <label>Season</label>
          <select v-model.number="editingMatch.season_id">
            <option value="0">Select Season</option>
            <option v-for="season in seasons" :key="season.id" :value="season.id">
              Season {{ season.year }}
            </option>
          </select>
        </div>

        <div class="modal-form-group">
          <label>{{ editingMatch.team_a_id ? getTeamName(editingMatch.team_a_id) : 'Team 1' }}</label>
          <select v-model.number="editingMatch.team_a_id">
            <option value="0">Select Team 1</option>
            <option v-for="team in teams" :key="team.id" :value="team.id">
              {{ team.name }}
            </option>
          </select>
          <input v-model.number="editingMatch.score_a" type="number" placeholder="Score" />
        </div>

        <div class="modal-form-group">
          <label>{{ editingMatch.team_b_id ? getTeamName(editingMatch.team_b_id) : 'Team 2' }}</label>
          <select v-model.number="editingMatch.team_b_id">
            <option value="0">Select Team 2</option>
            <option v-for="team in teams" :key="team.id" :value="team.id">
              {{ team.name }}
            </option>
          </select>
          <input v-model.number="editingMatch.score_b" type="number" placeholder="Score" />
        </div>

        <div class="modal-form-group">
          <label>Winner</label>
          <select v-model.number="editingMatch.winner_id" :disabled="!editingMatch.team_a_id || !editingMatch.team_b_id">
            <option value="0">{{ getWinnerLabel(editingMatch) }}</option>
            <option v-if="editingMatch.team_a_id" :value="editingMatch.team_a_id">
              {{ getTeamName(editingMatch.team_a_id) }}
            </option>
            <option v-if="editingMatch.team_b_id" :value="editingMatch.team_b_id">
              {{ getTeamName(editingMatch.team_b_id) }}
            </option>
          </select>
        </div>

        <div class="modal-form-group">
          <label>Date & Time</label>
          <input 
            :value="formatToDatetimeLocal(editingMatch.created_at)" 
            @input="(e) => { editingMatch!.created_at = parseFromDatetimeLocal((e.target as HTMLInputElement).value) }"
            type="datetime-local" 
          />
        </div>

        <div class="modal-actions">
          <button @click="saveMatch" class="btn-primary">Save</button>
          <button @click="editingMatch = null" class="btn-secondary">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, onMounted } from 'vue'
import { useAuth } from '@/composables/useAuth'

interface Team { id?: number; name: string }
interface Season { id?: number; year: number }
interface Match { id?: number; season_id: number; team_a_id: number; team_b_id: number; winner_id?: number; score_a?: number; score_b?: number; created_at?: string }

const api = API_URL
const { getAuthHeaders } = useAuth()
const teams = ref<Team[]>([])
const seasons = ref<Season[]>([])
const matches = ref<Match[]>([])
const newMatch = ref<Match>({ season_id: 0, team_a_id: 0, team_b_id: 0, winner_id: 0, score_a: undefined, score_b: undefined })
const editingMatch = ref<Match | null>(null)
const currentYear = new Date().getFullYear()

onMounted(() => {
  fetchTeams()
  fetchSeasons()
  fetchMatches()
  // Set default season to current year
  fetchSeasons().then(() => {
    const currentSeason = seasons.value.find(s => s.year === currentYear)
    if (currentSeason) {
      newMatch.value.season_id = currentSeason.id || 0
    }
  })
})

const fetchTeams = async () => {
  try {
    const res = await fetch(`${api}/teams`, {
      headers: getAuthHeaders()
    })
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
    const res = await fetch(`${api}/seasons`, {
      headers: getAuthHeaders()
    })
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
    const res = await fetch(`${api}/matches`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      matches.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch matches', e)
  }
}

const getTeamName = (id?: number) => teams.value.find(t => t.id === id)?.name || 'Unknown'
const getSeasonYear = (id?: number) => seasons.value.find(s => s.id === id)?.year || '—'

const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return '—'
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return '—'
  }
}

const formatToDatetimeLocal = (dateStr?: string) => {
  if (!dateStr) return ''
  try {
    const date = new Date(dateStr)
    return date.toISOString().slice(0, 16)
  } catch {
    return ''
  }
}

const parseFromDatetimeLocal = (value: string) => {
  if (!value) return ''
  return new Date(value).toISOString()
}

const getWinnerLabel = (match: Match) => {
  // If both scores are provided, show the automatically determined winner
  if (match.score_a !== undefined && match.score_b !== undefined) {
    if (match.score_a > match.score_b) {
      return `Winner: ${getTeamName(match.team_a_id)}`
    } else if (match.score_b > match.score_a) {
      return `Winner: ${getTeamName(match.team_b_id)}`
    } else {
      return 'Draw - Select Winner'
    }
  }
  return 'Select Winner'
}

const addMatch = async () => {
  if (!newMatch.value.season_id || !newMatch.value.team_a_id || !newMatch.value.team_b_id) return

  // Auto-determine winner from scores if both are provided
  if (newMatch.value.score_a !== undefined && newMatch.value.score_b !== undefined) {
    if (newMatch.value.score_a > newMatch.value.score_b) {
      newMatch.value.winner_id = newMatch.value.team_a_id
    } else if (newMatch.value.score_b > newMatch.value.score_a) {
      newMatch.value.winner_id = newMatch.value.team_b_id
    }
  }

  if (!newMatch.value.winner_id) return

  try {
    const res = await fetch(`${api}/matches`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(newMatch.value)
    })
    if (res.ok) {
      const currentSeason = seasons.value.find(s => s.year === currentYear)
      newMatch.value = { season_id: currentSeason?.id || 0, team_a_id: 0, team_b_id: 0, winner_id: 0, score_a: undefined, score_b: undefined }
      fetchMatches()
    }
  } catch (e) {
    console.error('Failed to add match', e)
  }
}

const editMatch = (match: Match) => {
  editingMatch.value = { ...match }
}

const saveMatch = async () => {
  if (!editingMatch.value || !editingMatch.value.season_id || !editingMatch.value.team_a_id || !editingMatch.value.team_b_id) return
  try {
    const res = await fetch(`${api}/matches/${editingMatch.value.id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(editingMatch.value)
    })
    if (res.ok) {
      editingMatch.value = null
      fetchMatches()
    }
  } catch (e) {
    console.error('Failed to save match', e)
  }
}

const deleteMatch = async (id?: number) => {
  if (!id || !confirm('Delete this match?')) return
  try {
    const res = await fetch(`${api}/matches/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    })
    if (res.ok) {
      fetchMatches()
    }
  } catch (e) {
    console.error('Failed to delete match', e)
  }
}
</script>

<style scoped>
.admin-page {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

h2 {
  font-size: 28px;
  margin: 0;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  background-color: var(--clr-surface-a10);
  padding: var(--spacing-lg);
  border-radius: var(--radius);
  border: 1px solid var(--border);
}

.form-row {
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.form-column {
  flex: 1;
  min-width: 180px;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.form-column label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.score-input {
  flex: 1;
  min-width: 100px;
}

input, select {
  flex: 1;
  min-width: 150px;
  padding: var(--spacing-md);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 14px;
  background-color: var(--clr-surface-a0);
  color: var(--text-primary);
}

input:focus, select:focus {
  outline: none;
  border-color: var(--clr-primary-a50);
  box-shadow: 0 0 0 3px var(--clr-primary-a0);
}

.btn-primary {
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--clr-primary-a50);
  color: white;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: var(--clr-primary-a40);
}

.btn-record {
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--clr-primary-a10);
  color: white;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.2s;
}

.btn-record:hover {
  background: var(--clr-primary-a0);
}

.btn-secondary {
  padding: 6px 12px;
  background: var(--clr-surface-a20);
  color: var(--text-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 12px;
  transition: background 0.2s;
}

.btn-secondary:hover {
  background: var(--border);
}

.btn-delete {
  padding: 6px 12px;
  background: var(--clr-danger-a20);
  color: white;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 12px;
  transition: background 0.2s;
}

.btn-delete:hover {
  background: var(--clr-danger-a10);
}

.empty-state {
  padding: var(--spacing-xl);
  text-align: center;
  color: var(--text-secondary);
  background-color: var(--clr-surface-a10);
  border-radius: var(--radius);
}

.items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: var(--spacing-lg);
}

.item-card {
  background-color: var(--clr-surface-a10);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.item-header {
  flex: 1;
}

.item-header h3 {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 16px;
}

.match-score {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
}

.match-winner {
  margin: 0;
  color: var(--clr-primary-a50);
  font-size: 14px;
}

.match-season {
  margin: 0;
  color: var(--text-secondary);
  font-size: 12px;
}

.match-date {
  margin: 0;
  color: var(--text-secondary);
  font-size: 11px;
  margin-top: var(--spacing-sm);
}

.item-actions {
  display: flex;
  gap: var(--spacing-md);
}

.item-actions button {
  flex: 1;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background-color: var(--clr-surface-a0);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--spacing-xl);
  max-width: 500px;
  width: 90%;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  max-height: 80vh;
  overflow-y: auto;
}

.modal h3 {
  margin: 0;
}

.modal-form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.modal-form-group label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.modal input, .modal select {
  padding: var(--spacing-md);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background-color: var(--clr-surface-a10);
  color: var(--text-primary);
}

.modal-actions {
  display: flex;
  gap: var(--spacing-md);
}

.modal-actions button {
  flex: 1;
}

@media (max-width: 600px) {
  .form-group {
    flex-direction: column;
  }

  input, select {
    width: 100%;
  }

  .items-grid {
    grid-template-columns: 1fr;
  }
}
</style>
