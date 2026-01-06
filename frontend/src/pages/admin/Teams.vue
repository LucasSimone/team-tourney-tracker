<template>
  <div class="admin-page">
    <h2>Manage Teams</h2>

    <div class="form-group">
      <input v-model="newTeam.name" type="text" placeholder="Team name" />
      <select v-model.number="selectedPlayerForNew">
        <option value="0">Add Player...</option>
        <option v-for="player in players" :key="player.id" :value="player.id">
          {{ player.name }}
        </option>
      </select>
      <button @click="addPlayerToNewTeam" class="btn-secondary" v-if="selectedPlayerForNew > 0">Add Player</button>
      <button @click="addTeam" class="btn-primary">Create Team</button>
    </div>

    <div v-if="newTeamPlayers.length > 0" class="selected-players">
      <h4>Team Players:</h4>
      <div class="player-tags">
        <div v-for="playerId in newTeamPlayers" :key="playerId" class="player-tag">
          {{ getPlayerName(playerId) }}
          <button @click="removePlayerFromNewTeam(playerId)" class="remove-btn">×</button>
        </div>
      </div>
    </div>

    <div v-if="teams.length === 0" class="empty-state">No teams yet. Create one to get started!</div>

    <div v-else class="items-grid">
      <div v-for="team in teams" :key="team.id" class="item-card">
        <div class="item-header">
          <h3>{{ team.name }}</h3>
        </div>
        <div v-if="teamPlayers[team.id]" class="team-players">
          <p class="players-label">Players:</p>
          <div v-if="teamPlayers[team.id].length > 0" class="player-list">
            <span v-for="player in teamPlayers[team.id]" :key="player.id" class="player-badge">
              {{ player.name }}
            </span>
          </div>
          <p v-else class="no-players">No players assigned</p>
        </div>
        <div class="item-actions">
          <button @click="editTeam(team)" class="btn-secondary">Edit</button>
          <button @click="deleteTeam(team.id)" class="btn-delete">Delete</button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="editingTeam" class="modal-overlay" @click="editingTeam = null">
      <div class="modal" @click.stop>
        <h3>Edit Team</h3>
        <input v-model="editingTeam.name" type="text" placeholder="Team name" />
        
        <div class="edit-players">
          <h4>Team Players</h4>
          <div v-if="editingTeamPlayers.length > 0" class="player-list">
            <div v-for="player in editingTeamPlayers" :key="player.id" class="player-item">
              <span>{{ player.name }}</span>
              <button @click="removePlayerFromEditTeam(player.id)" class="btn-small-delete">Remove</button>
            </div>
          </div>
          <p v-else class="no-players">No players assigned</p>
          
          <div class="add-player-row">
            <select v-model.number="selectedPlayerForEdit">
              <option value="0">Add Player...</option>
              <option v-for="player in availablePlayersForEdit" :key="player.id" :value="player.id">
                {{ player.name }}
              </option>
            </select>
            <button @click="addPlayerToEditTeam" class="btn-secondary" v-if="selectedPlayerForEdit > 0">Add</button>
          </div>
        </div>

        <div class="modal-actions">
          <button @click="saveTeam" class="btn-primary">Save</button>
          <button @click="editingTeam = null" class="btn-secondary">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, onMounted, computed } from 'vue'
import { useAuth } from '@/composables/useAuth'

interface Team { id?: number; name: string }
interface Player { id?: number; name: string }

const api = API_URL
const { getAuthHeaders } = useAuth()
const teams = ref<Team[]>([])
const players = ref<Player[]>([])
const newTeam = ref<Team>({ name: '' })
const newTeamPlayers = ref<number[]>([])
const editingTeam = ref<Team | null>(null)
const editingTeamPlayers = ref<Player[]>([])
const selectedPlayerForNew = ref(0)
const selectedPlayerForEdit = ref(0)
const teamPlayers = ref<{ [key: number]: Player[] }>({})

onMounted(() => {
  fetchTeams()
  fetchPlayers()
})

const fetchTeams = async () => {
  try {
    const res = await fetch(`${api}/teams`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      teams.value = data || []
      // Fetch players for each team
      for (const team of teams.value) {
        await fetchTeamPlayers(team.id!)
      }
    }
  } catch (e) {
    console.error('Failed to fetch teams', e)
  }
}

const fetchPlayers = async () => {
  try {
    const res = await fetch(`${api}/players`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      players.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch players', e)
  }
}

const fetchTeamPlayers = async (teamId: number) => {
  try {
    const res = await fetch(`${api}/teams/${teamId}/players`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      teamPlayers.value[teamId] = data || []
    }
  } catch (e) {
    console.error('Failed to fetch team players', e)
  }
}

const getPlayerName = (playerId: number) => {
  return players.value.find(p => p.id === playerId)?.name || 'Unknown'
}

const addPlayerToNewTeam = () => {
  if (selectedPlayerForNew.value > 0 && !newTeamPlayers.value.includes(selectedPlayerForNew.value)) {
    newTeamPlayers.value.push(selectedPlayerForNew.value)
    selectedPlayerForNew.value = 0
  }
}

const removePlayerFromNewTeam = (playerId: number) => {
  newTeamPlayers.value = newTeamPlayers.value.filter(id => id !== playerId)
}

const availablePlayersForEdit = computed(() => {
  return players.value.filter(p => !editingTeamPlayers.value.find(ep => ep.id === p.id))
})

const addPlayerToEditTeam = () => {
  if (selectedPlayerForEdit.value > 0) {
    const player = players.value.find(p => p.id === selectedPlayerForEdit.value)
    if (player && !editingTeamPlayers.value.find(p => p.id === player.id)) {
      editingTeamPlayers.value.push(player)
      selectedPlayerForEdit.value = 0
    }
  }
}

const removePlayerFromEditTeam = (playerId: number) => {
  editingTeamPlayers.value = editingTeamPlayers.value.filter(p => p.id !== playerId)
}

const addTeam = async () => {
  if (!newTeam.value.name.trim()) return
  try {
    const res = await fetch(`${api}/teams`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({
        name: newTeam.value.name,
        players: newTeamPlayers.value
      })
    })
    if (res.ok) {
      newTeam.value.name = ''
      newTeamPlayers.value = []
      selectedPlayerForNew.value = 0
      fetchTeams()
    }
  } catch (e) {
    console.error('Failed to add team', e)
  }
}

const editTeam = async (team: Team) => {
  editingTeam.value = { ...team }
  await fetchTeamPlayers(team.id!)
  editingTeamPlayers.value = teamPlayers.value[team.id!] || []
  selectedPlayerForEdit.value = 0
}

const saveTeam = async () => {
  if (!editingTeam.value || !editingTeam.value.name.trim()) return
  try {
    const res = await fetch(`${api}/teams/${editingTeam.value.id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify({
        name: editingTeam.value.name,
        players: editingTeamPlayers.value.map(p => p.id!)
      })
    })
    if (res.ok) {
      editingTeam.value = null
      editingTeamPlayers.value = []
      fetchTeams()
    }
  } catch (e) {
    console.error('Failed to save team', e)
  }
}

const deleteTeam = async (id?: number) => {
  if (!id || !confirm('Delete this team?')) return
  try {
    const res = await fetch(`${api}/teams/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    })
    if (res.ok) {
      fetchTeams()
    }
  } catch (e) {
    console.error('Failed to delete team', e)
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
  gap: var(--spacing-md);
  flex-wrap: wrap;
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
  white-space: nowrap;
}

.btn-primary:hover {
  background: var(--clr-primary-a40);
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
  white-space: nowrap;
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

.btn-small-delete {
  padding: 4px 8px;
  background: var(--clr-danger-a20);
  color: white;
  border: none;
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 11px;
  transition: background 0.2s;
}

.btn-small-delete:hover {
  background: var(--clr-danger-a10);
}

.selected-players {
  background-color: var(--clr-surface-a10);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
}

.selected-players h4 {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 14px;
  color: var(--text-secondary);
}

.player-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.player-tag {
  display: flex;
  align-items: center;
  gap: 6px;
  background-color: var(--clr-primary-a0);
  color: var(--clr-primary-a50);
  padding: 4px 8px;
  border-radius: var(--radius);
  font-size: 12px;
  font-weight: 500;
}

.remove-btn {
  background: none;
  border: none;
  color: var(--clr-primary-a50);
  cursor: pointer;
  font-size: 16px;
  padding: 0;
  line-height: 1;
}

.remove-btn:hover {
  opacity: 0.7;
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
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
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
  margin: 0;
  font-size: 18px;
}

.team-players {
  background-color: var(--clr-surface-a0);
  padding: var(--spacing-md);
  border-radius: var(--radius);
}

.players-label {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.player-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.player-badge {
  background-color: var(--clr-primary-a0);
  color: var(--clr-primary-a50);
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
}

.no-players {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
  font-style: italic;
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
  max-height: 80vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.modal h3 {
  margin: 0;
}

.modal input, .modal select {
  padding: var(--spacing-md);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background-color: var(--clr-surface-a10);
  color: var(--text-primary);
}

.edit-players {
  background-color: var(--clr-surface-a10);
  padding: var(--spacing-lg);
  border-radius: var(--radius);
}

.edit-players h4 {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 14px;
}

.player-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-sm);
  background-color: var(--clr-surface-a0);
  border-radius: var(--radius);
  margin-bottom: var(--spacing-sm);
  font-size: 14px;
  color: var(--text-primary);
}

.add-player-row {
  display: flex;
  gap: var(--spacing-md);
  margin-top: var(--spacing-md);
}

.add-player-row select {
  flex: 1;
  min-width: 150px;
}

.add-player-row button {
  white-space: nowrap;
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

  .modal {
    width: 95%;
  }
}
</style>
