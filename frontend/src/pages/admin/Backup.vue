<template>
  <div class="admin-page">
    <h2>Database Backups</h2>

    <div class="backup-section">
      <p>Manually trigger a database backup to be sent to the configured email address.</p>
      <button 
        @click="triggerBackup" 
        :disabled="isBackingUp"
        :class="['btn-primary', { 'btn-loading': isBackingUp }]"
      >
        {{ isBackingUp ? 'Backing up...' : 'Trigger Backup Now' }}
      </button>
    </div>

    <div v-if="backupMessage" :class="['message', backupMessage.type]">
      {{ backupMessage.text }}
    </div>

    <div class="info-section">
      <h3>Weekly Backups</h3>
      <p>Automatic backups are configured to run weekly and will be sent via email.</p>
      <ul>
        <li>Frequency: Every 7 days</li>
        <li>First backup: Runs immediately on backend startup</li>
        <li>Email configuration: Set via SENDER_EMAIL, SENDER_APP_PASSWORD, and BACKUP_EMAIL environment variables</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref } from 'vue'
import { useAuth } from '@/composables/useAuth'

interface BackupMessage {
  type: 'success' | 'error'
  text: string
}

const api = API_URL
const { getAuthHeaders } = useAuth()
const isBackingUp = ref(false)
const backupMessage = ref<BackupMessage | null>(null)

const triggerBackup = async () => {
  isBackingUp.value = true
  backupMessage.value = null

  try {
    const res = await fetch(`${api}/admin/backup`, {
      method: 'POST',
      headers: getAuthHeaders()
    })

    if (res.ok) {
      backupMessage.value = {
        type: 'success',
        text: 'Backup triggered successfully! Check your email for the database file.'
      }
    } else {
      const errorData = await res.text()
      backupMessage.value = {
        type: 'error',
        text: `Backup failed: ${errorData || 'Unknown error'}`
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

.btn-primary {
  background-color: var(--clr-primary-a0);
  color: white;
  border: none;
  padding: var(--spacing-md) var(--spacing-lg);
  border-radius: var(--radius);
  cursor: pointer;
  font-size: 1rem;
  transition: background-color 0.2s;
  width: fit-content;
}

.btn-primary:hover:not(:disabled) {
  background-color: var(--clr-primary-a10);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-loading {
  pointer-events: none;
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

.info-section {
  background: var(--clr-surface-a10);
  border-left: 4px solid var(--clr-primary-a0);
  padding: var(--spacing-lg);
  border-radius: var(--radius);
  border: 1px solid var(--border);
  margin-top: var(--spacing-lg);
}

.info-section h3 {
  margin-top: 0;
  margin-bottom: var(--spacing-md);
  color: var(--text-primary);
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
