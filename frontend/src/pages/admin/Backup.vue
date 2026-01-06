<template>
  <div class="admin-page">
    <div class="backup-section">
      <p>Manually trigger a database backup. Backups are stored locally on the server.</p>
      <button 
        @click="triggerBackup" 
        :disabled="isBackingUp"
        :class="['btn-primary', { 'btn-loading': isBackingUp }]"
      >
        {{ isBackingUp ? 'Creating backup...' : 'Create Backup Now' }}
      </button>
    </div>

    <div v-if="backupMessage" :class="['message', backupMessage.type]">
      {{ backupMessage.text }}
    </div>

    <div class="backups-section">
      <h3>Available Backups</h3>
      <p>Download or manage existing backup files. Backups older than 3 are automatically deleted.</p>
      
      <button 
        @click="loadBackups"
        :disabled="isLoadingBackups"
        class="btn-primary"
      >
        {{ isLoadingBackups ? 'Loading...' : 'Refresh Backups' }}
      </button>

      <div v-if="backups.length > 0" class="backups-table">
        <table>
          <thead>
            <tr>
              <th>Filename</th>
              <th>Size</th>
              <th>Created</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="backup in backups" :key="backup.name">
              <td>{{ backup.name }}</td>
              <td>{{ formatFileSize(backup.size) }}</td>
              <td>{{ formatDate(backup.modified) }}</td>
              <td>
                <button 
                  @click="downloadBackup(backup.name)"
                  class="btn-small"
                  :disabled="isDownloading"
                >
                  {{ isDownloading ? 'Downloading...' : 'Download' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="no-backups">
        No backups available yet.
      </div>
    </div>

    <div class="info-section">
      <h3>Weekly Backups</h3>
      <p>Automatic backups are created weekly and stored in the backups directory.</p>
      <ul>
        <li>Frequency: Every 7 days</li>
        <li>First backup: Runs immediately on backend startup</li>
        <li>Retention: Only the 3 most recent backups are kept (older ones auto-deleted)</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, onMounted } from 'vue'
import { useAuth } from '@/composables/useAuth'

interface BackupMessage {
  type: 'success' | 'error'
  text: string
}

interface Backup {
  name: string
  size: number
  modified: string
}

const api = API_URL
const { getAuthHeaders } = useAuth()
const isBackingUp = ref(false)
const isLoadingBackups = ref(false)
const isDownloading = ref(false)
const backupMessage = ref<BackupMessage | null>(null)
const backups = ref<Backup[]>([])

const triggerBackup = async () => {
  isBackingUp.value = true
  backupMessage.value = null

  try {
    const res = await fetch(`${api}/admin/backup`, {
      method: 'POST',
      headers: getAuthHeaders()
    })

    const data = await res.json()

    if (res.ok) {
      backupMessage.value = {
        type: 'success',
        text: data.message || 'Backup created successfully!'
      }
      // Refresh backup list
      await loadBackups()
    } else {
      backupMessage.value = {
        type: 'error',
        text: data.message || 'Backup failed'
      }
    }
  } catch (e) {
    backupMessage.value = {
      type: 'error',
      text: `Backup error: ${e instanceof Error ? e.message : 'Unknown error'}`
    }
  } finally {
    isBackingUp.value = false
  }
}

const loadBackups = async () => {
  isLoadingBackups.value = true
  backups.value = []

  try {
    const res = await fetch(`${api}/admin/backup`, {
      method: 'GET',
      headers: getAuthHeaders()
    })

    if (res.ok) {
      const data = await res.json()
      backups.value = data.backups || []
    }
  } catch (e) {
    console.error('Failed to load backups:', e)
  } finally {
    isLoadingBackups.value = false
  }
}

const downloadBackup = async (filename: string) => {
  isDownloading.value = true

  try {
    const res = await fetch(`${api}/admin/backup/download?file=${encodeURIComponent(filename)}`, {
      method: 'GET',
      headers: getAuthHeaders()
    })

    if (res.ok) {
      const blob = await res.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } else {
      backupMessage.value = {
        type: 'error',
        text: 'Failed to download backup'
      }
    }
  } catch (e) {
    backupMessage.value = {
      type: 'error',
      text: `Download error: ${e instanceof Error ? e.message : 'Unknown error'}`
    }
  } finally {
    isDownloading.value = false
  }
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
}

onMounted(() => {
  loadBackups()
})
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

h3 {
  margin-top: 0;
  color: var(--text-primary);
}

.backup-section {
  background: var(--clr-surface-a10);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  border: 1px solid var(--border);
}

.backup-section p {
  margin: 0 0 var(--spacing-md) 0;
  color: var(--text-secondary);
}

.backups-section {
  background: var(--clr-surface-a10);
  border-radius: var(--radius);
  padding: var(--spacing-lg);
  border: 1px solid var(--border);
}

.backups-section p {
  color: var(--text-secondary);
  margin: 0 0 var(--spacing-md) 0;
}

.backups-section h3 {
  margin-bottom: var(--spacing-md);
}

.btn-primary,
.btn-secondary {
  border: none;
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 1rem;
  transition: background-color 0.2s;
  width: fit-content;
  margin-bottom: var(--spacing-md);
}

.btn-primary {
  background-color: var(--clr-primary-a0);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: var(--clr-primary-a10);
}

.btn-secondary {
  background-color: var(--clr-primary-a10);
  color: var(--clr-primary-a0);
}

.btn-secondary:hover:not(:disabled) {
  background-color: var(--clr-primary-a20);
}

.btn-small {
  background-color: var(--clr-primary-a0);
  color: white;
  border: none;
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 0.9rem;
  transition: background-color 0.2s;
}

.btn-small:hover:not(:disabled) {
  background-color: var(--clr-primary-a10);
}

.btn-primary:disabled,
.btn-secondary:disabled,
.btn-small:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.message {
  padding: var(--spacing-md);
  border-radius: var(--radius);
  font-weight: 500;
  border-left: 4px solid;
}

.message.success {
  background-color: rgba(76, 175, 80, 0.1);
  color: #4caf50;
  border-left-color: #4caf50;
}

.message.error {
  background-color: rgba(244, 67, 54, 0.1);
  color: #f44336;
  border-left-color: #f44336;
}

.backups-table {
  overflow-x: auto;
  margin-top: var(--spacing-md);
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.95rem;
}

table th {
  background-color: var(--clr-primary-a5);
  padding: var(--spacing-md);
  text-align: left;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 2px solid var(--border);
}

table td {
  padding: var(--spacing-md);
  border-bottom: 1px solid var(--border);
  color: var(--text-secondary);
}

table tr:hover {
  background-color: var(--clr-surface-a5);
}

.no-backups {
  color: var(--text-secondary);
  padding: var(--spacing-lg);
  text-align: center;
  font-style: italic;
}

.info-section {
  background: var(--clr-surface-a10);
  border-left: 4px solid var(--clr-primary-a0);
  padding: var(--spacing-lg);
  border-radius: var(--radius);
  border: 1px solid var(--border);
  margin-top: var(--spacing-lg);
}

.info-section h3 {
  margin-bottom: var(--spacing-md);
}

.info-section p {
  color: var(--text-primary);
  margin: 0 0 var(--spacing-md) 0;
}

.info-section ul {
  margin: var(--spacing-md) 0;
  padding-left: 1.5rem;
  color: var(--text-secondary);
}

.info-section li {
  margin: var(--spacing-sm) 0;
}
</style>
