<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>XTrafficDash</h1>
        <p>请输入密码登录</p>
      </div>
      
      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="请输入密码"
            required
            :disabled="authStore.isLoading"
          />
        </div>
        
        <div v-if="authStore.error" class="error-message">
          {{ authStore.error }}
        </div>
        
        <button 
          type="submit" 
          class="login-button"
          :disabled="authStore.isLoading"
        >
          <span v-if="authStore.isLoading">登录中...</span>
          <span v-else>登录</span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const password = ref('')

const handleLogin = async () => {
  if (!password.value.trim()) {
    return
  }
  
  const result = await authStore.login(password.value)
  if (result.success) {
    router.push('/home')
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #F5F6FA;
  padding: 20px;
}

.login-card {
  background: #fff;
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 4px 24px rgba(44,62,80,0.10);
  width: 100%;
  max-width: 400px;
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h1 {
  color: #222;
  font-size: 2rem;
  margin-bottom: 10px;
}

.login-header p {
  color: #7f8c8d;
  font-size: 1rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  color: #2c3e50;
  font-weight: 600;
  font-size: 0.9rem;
}

.form-group input {
  padding: 12px 16px;
  border: 2px solid #ecf0f1;
  border-radius: 8px;
  font-size: 1rem;
  transition: border-color 0.3s ease;
}

.form-group input:focus {
  outline: none;
  border-color: #70A1FF;
}

.form-group input:disabled {
  background-color: #f8f9fa;
  cursor: not-allowed;
}

.error-message {
  color: #e74c3c;
  background: #ffeaea;
  padding: 12px;
  border-radius: 6px;
  font-size: 0.9rem;
  text-align: center;
}

.login-button {
  background: #70A1FF;
  color: #fff;
  border: none;
  padding: 14px 12px;
  border-radius: 20px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s, color 0.2s, box-shadow 0.2s, transform 0.18s;
  box-shadow: 0 2px 8px rgba(112,161,255,0.10);
  min-width: 200px;
  display: block;
  margin: 0 auto;
}

.login-button:hover:not(:disabled) {
  background: #1E90FF;
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(112,161,255,0.18);
}

.login-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}
</style> 