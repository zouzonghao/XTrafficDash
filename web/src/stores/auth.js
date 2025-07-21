import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '@/utils/api'

export const useAuthStore = defineStore('auth', () => {
  const isAuthenticated = ref(false)
  const isLoading = ref(false)
  const error = ref(null)

  const login = async (password) => {
    isLoading.value = true
    error.value = null
    
    try {
      const response = await authAPI.login(password)
      if (response.data.success) {
        isAuthenticated.value = true
        localStorage.setItem('auth_token', response.data.token)
        return { success: true }
      } else {
        error.value = response.data.message || '登录失败'
        return { success: false, error: error.value }
      }
    } catch (err) {
      error.value = err.response?.data?.message || '网络错误'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const logout = () => {
    isAuthenticated.value = false
    localStorage.removeItem('auth_token')
  }

  const checkAuth = () => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      isAuthenticated.value = true
    }
  }

  return {
    isAuthenticated: computed(() => isAuthenticated.value),
    isLoading: computed(() => isLoading.value),
    error: computed(() => error.value),
    login,
    logout,
    checkAuth
  }
}) 