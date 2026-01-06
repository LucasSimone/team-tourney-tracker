<template>
  <div class="auth-container">
    <div class="auth-box">
      <h1>Tourney Tracker</h1>
      <p class="subtitle">Sign in to your account</p>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label>Username</label>
          <input 
            v-model="username" 
            type="text" 
            placeholder="Enter username"
            required
          />
        </div>

        <div class="form-group">
          <label>Password</label>
          <input 
            v-model="password" 
            type="password" 
            placeholder="Enter password"
            required
          />
        </div>

        <button type="submit" class="btn-login">Sign In</button>
      </form>

      <div v-if="error" class="error-message">{{ error }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { login } = useAuth()

const username = ref('')
const password = ref('')
const error = ref('')

const handleLogin = async () => {
  error.value = ''
  try {
    await login(username.value, password.value)
    router.push('/')
  } catch (e) {
    error.value = 'Invalid username or password'
  }
}
</script>

<style scoped>
.auth-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--clr-primary-a20) 0%, var(--clr-primary-a0) 100%);
  padding: var(--spacing-lg);
}

.auth-box {
  background-color: var(--clr-surface-a0);
  border-radius: var(--radius);
  padding: var(--spacing-xl);
  width: 100%;
  max-width: 400px;
  box-shadow: 0 10px 40px var(--shadow);
}

h1 {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 28px;
  text-align: center;
}

.subtitle {
  text-align: center;
  color: var(--text-secondary);
  margin: 0 0 var(--spacing-xl) 0;
}

.form-group {
  margin-bottom: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

label {
  font-weight: 500;
  color: var(--text-primary);
}

input {
  padding: var(--spacing-md);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background-color: var(--clr-surface-a10);
  color: var(--text-primary);
  font-size: 14px;
}

input:focus {
  outline: none;
  border-color: var(--clr-primary-a50);
  box-shadow: 0 0 0 3px var(--clr-primary-a0);
}

.btn-login {
  width: 100%;
  padding: var(--spacing-lg);
  background: var(--clr-primary-a0);
  color: white;
  border: none;
  border-radius: var(--radius);
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-login:hover {
  background: var(--clr-primary-a10);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--clr-primary-a20);
}

.error-message {
  margin-top: var(--spacing-lg);
  padding: var(--spacing-md);
  background-color: rgba(255, 68, 68, 0.1);
  color: #ff4444;
  border: 1px solid rgba(255, 68, 68, 0.3);
  border-radius: var(--radius);
  font-size: 14px;
}

@media (max-width: 600px) {
  .auth-box {
    padding: var(--spacing-lg);
  }

  h1 {
    font-size: 24px;
  }
}
</style>
