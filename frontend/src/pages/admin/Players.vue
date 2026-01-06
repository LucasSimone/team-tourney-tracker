<template>
  <div class="admin-page">
    <h2>Manage Players</h2>

    <div class="form-group">
      <input v-model="newPlayer.name" type="text" placeholder="Player name" @keyup.enter="addPlayer" />
      <button @click="addPlayer" class="btn-primary">Add Player</button>
    </div>

    <div v-if="players.length === 0" class="empty-state">No players yet. Create one to get started!</div>

    <div v-else class="items-grid">
      <div v-for="player in players" :key="player.id" class="item-card">
        <div class="item-header">
          <h3>{{ player.name }}</h3>
        </div>
        <div class="item-actions">
          <button @click="editPlayer(player)" class="btn-secondary">Edit</button>
          <button @click="deletePlayer(player.id)" class="btn-delete">Delete</button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <div v-if="editingPlayer" class="modal-overlay" @click="editingPlayer = null">
      <div class="modal" @click.stop>
        <h3>Edit Player</h3>
        <input v-model="editingPlayer.name" type="text" placeholder="Player name" />
        <div class="modal-actions">
          <button @click="savePlayer" class="btn-primary">Save</button>
          <button @click="editingPlayer = null" class="btn-secondary">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, onMounted } from 'vue'
import { useAuth } from '@/composables/useAuth'

interface Player { id?: number; name: string }

const api = API_URL
const { getAuthHeaders } = useAuth()
const players = ref<Player[]>([])
const newPlayer = ref<Player>({ name: '' })
const editingPlayer = ref<Player | null>(null)

onMounted(() => {
  fetchPlayers()
})

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

const addPlayer = async () => {
  if (!newPlayer.value.name.trim()) return
  try {
    const res = await fetch(`${api}/players`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(newPlayer.value)
    })
    if (res.ok) {
      newPlayer.value.name = ''
      fetchPlayers()
    }
  } catch (e) {
    console.error('Failed to add player', e)
  }
}

const editPlayer = (player: Player) => {
  editingPlayer.value = { ...player }
}

const savePlayer = async () => {
  if (!editingPlayer.value || !editingPlayer.value.name.trim()) return
  try {
    const res = await fetch(`${api}/players/${editingPlayer.value.id}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(editingPlayer.value)
    })
    if (res.ok) {
      editingPlayer.value = null
      fetchPlayers()
    }
  } catch (e) {
    console.error('Failed to save player', e)
  }
}

const deletePlayer = async (id?: number) => {
  if (!id || !confirm('Delete this player?')) return
  try {
    const res = await fetch(`${api}/players/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    })
    if (res.ok) {
      fetchPlayers()
    }
  } catch (e) {
    console.error('Failed to delete player', e)
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
}

input {
  flex: 1;
  min-width: 200px;
  padding: var(--spacing-md);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 14px;
  background-color: var(--clr-surface-a0);
  color: var(--text-primary);
}

input:focus {
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
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
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
  max-width: 400px;
  width: 90%;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.modal h3 {
  margin: 0;
}

.modal input {
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

  input {
    width: 100%;
  }

  .items-grid {
    grid-template-columns: 1fr;
  }
}
</style>
