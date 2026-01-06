import { ref } from 'vue'
import { API_URL } from '../config'

export interface User {
  id: number
  username: string
  role: 'admin' | 'user'
}

const api = API_URL
const user = ref<User | null>(null)
const token = ref<string | null>(null)

export const useAuth = () => {
  const initialize = () => {
    const savedToken = localStorage.getItem('token')
    const savedUser = localStorage.getItem('user')
    
    if (savedToken && savedUser) {
      token.value = savedToken
      user.value = JSON.parse(savedUser)
    }
  }

  const login = async (username: string, password: string) => {
    try {
      const res = await fetch(`${api}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      })

      if (!res.ok) {
        throw new Error('Login failed')
      }

      const data = await res.json()
      token.value = data.token
      user.value = data.user

      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))

      return data
    } catch (e) {
      console.error('Login error:', e)
      throw e
    }
  }

  const register = async (username: string, password: string) => {
    try {
      const res = await fetch(`${api}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      })

      if (!res.ok) {
        throw new Error('Registration failed')
      }

      const data = await res.json()
      token.value = data.token
      user.value = data.user

      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))

      return data
    } catch (e) {
      console.error('Register error:', e)
      throw e
    }
  }

  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const getAuthHeaders = () => {
    return {
      'Content-Type': 'application/json',
      ...(token.value && { 'Authorization': `Bearer ${token.value}` })
    }
  }

  const isAdmin = () => user.value?.role === 'admin'
  const isLoggedIn = () => !!user.value

  return {
    user,
    token,
    initialize,
    login,
    register,
    logout,
    getAuthHeaders,
    isAdmin,
    isLoggedIn
  }
}
