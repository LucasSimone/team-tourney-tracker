<template>
  <div class="admin-page">
    <div class="form-group">
      <input v-model="newUser.username" type="text" placeholder="Username" />
      <input v-model="newUser.password" type="password" placeholder="Password" />
      <select v-model="newUser.role">
        <option value="user">User</option>
        <option value="admin">Admin</option>
      </select>
      <button @click="addUser" class="btn-primary">Add User</button>
    </div>

    <div v-if="users.length === 0" class="empty-state">No users yet.</div>

    <div v-else class="items-grid">
      <div v-for="u in users" :key="u.id" class="item-card">
        <div class="item-header">
          <h3>{{ u.username }}</h3>
          <span :class="['role-badge', u.role]">{{ u.role }}</span>
        </div>
        <div class="item-actions">
          <select @change="changeRole($event, u.id)" :value="u.role" class="role-select">
            <option value="user">User</option>
            <option value="admin">Admin</option>
          </select>
          <button @click="deleteUser(u.id)" class="btn-delete">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, onMounted } from 'vue'
import { useAuth } from '@/composables/useAuth'

interface User {
  id: number
  username: string
  role: 'admin' | 'user'
}

interface NewUser {
  username: string
  password: string
  role: 'admin' | 'user'
}

const api = API_URL
const { getAuthHeaders } = useAuth()
const users = ref<User[]>([])
const newUser = ref<NewUser>({ username: '', password: '', role: 'user' })

onMounted(() => {
  fetchUsers()
})

const fetchUsers = async () => {
  try {
    const res = await fetch(`${api}/auth/users`, {
      headers: getAuthHeaders()
    })
    if (res.ok) {
      const data = await res.json()
      users.value = data || []
    }
  } catch (e) {
    console.error('Failed to fetch users', e)
  }
}

const addUser = async () => {
  if (!newUser.value.username.trim() || !newUser.value.password.trim()) return
  try {
    const res = await fetch(`${api}/auth/users`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(newUser.value)
    })
    if (res.ok) {
      newUser.value = { username: '', password: '', role: 'user' }
      fetchUsers()
    }
  } catch (e) {
    console.error('Failed to add user', e)
  }
}

const changeRole = async (event: Event, userId: number) => {
  const newRole = (event.target as HTMLSelectElement).value as 'admin' | 'user'
  try {
    await fetch(`${api}/auth/users/${userId}/role`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify({ role: newRole })
    })
    fetchUsers()
  } catch (e) {
    console.error('Failed to change role', e)
  }
}

const deleteUser = async (id: number) => {
  if (!confirm('Delete this user?')) return
  try {
    await fetch(`${api}/auth/users/${id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    })
    fetchUsers()
  } catch (e) {
    console.error('Failed to delete user', e)
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
}

.btn-primary:hover {
  background: var(--clr-primary-a40);
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-md);
}

.item-header h3 {
  margin: 0;
  font-size: 18px;
}

.role-badge {
  padding: 4px 8px;
  border-radius: var(--radius);
  font-size: 12px;
  font-weight: 500;
}

.role-badge.admin {
  background-color: var(--clr-primary-a0);
  color: var(--clr-primary-a50);
}

.role-badge.user {
  background-color: rgba(176, 176, 176, 0.2);
  color: var(--text-secondary);
}

.item-actions {
  display: flex;
  gap: var(--spacing-md);
}

.role-select {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background-color: var(--clr-surface-a0);
  color: var(--text-primary);
  font-size: 12px;
}

.item-actions button {
  flex-shrink: 0;
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

  .item-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
