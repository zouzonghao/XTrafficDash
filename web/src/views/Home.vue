<template>
  <div class="container">
    <div class="header">
      <h1>XTrafficDash</h1>
      <p>多服务器流量统计</p>
      <button @click="handleLogout" class="logout-button">退出登录</button>
    </div>

    <div v-if="servicesStore.loading" class="loading">
      加载中...
    </div>

    <div v-else-if="servicesStore.error" class="error">
      <h3>加载失败</h3>
      <p>{{ servicesStore.error }}</p>
      <button @click="servicesStore.loadServices" class="retry-button">
        重试
      </button>
    </div>

    <div v-else class="cards-grid">
      <ServiceCard
        v-for="service in servicesStore.services"
        :key="service.id"
        :service="service"
        @select="handleSelectService"
        @delete="handleDeleteService"
      />
    </div>

    <!-- 删除确认对话框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="hideDeleteConfirm">
      <div class="modal" @click.stop>
        <h3>确认删除</h3>
        <p>您确定要删除服务 <strong>{{ serviceToDelete?.custom_name || serviceToDelete?.ip_address }}</strong> 吗？</p>
        <p v-if="serviceToDelete?.custom_name" class="ip-info">IP：{{ serviceToDelete?.ip_address }}</p>
        <p class="warning-text">此操作将删除该服务的所有数据，包括流量记录和历史数据，且无法恢复。</p>
        <div class="modal-buttons">
          <button class="modal-button cancel" @click="hideDeleteConfirm">取消</button>
          <button class="modal-button confirm" @click="confirmDelete">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useServicesStore } from '../stores/services'
import { useAuthStore } from '../stores/auth'
import ServiceCard from '../components/ServiceCard.vue'

const router = useRouter()
const servicesStore = useServicesStore()
const authStore = useAuthStore()

const showDeleteModal = ref(false)
const serviceToDelete = ref(null)

const handleSelectService = (service) => {
  servicesStore.selectService(service)
  router.push(`/detail/${service.id}`)
}

const handleDeleteService = (service) => {
  serviceToDelete.value = service
  showDeleteModal.value = true
}

const hideDeleteConfirm = () => {
  showDeleteModal.value = false
  serviceToDelete.value = null
}

const confirmDelete = async () => {
  if (!serviceToDelete.value) return
  
  const result = await servicesStore.deleteService(serviceToDelete.value.id)
  if (result.success) {
    hideDeleteConfirm()
  } else {
    alert('删除失败: ' + result.error)
  }
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  servicesStore.loadServices()
  servicesStore.startAutoRefresh()
})

onUnmounted(() => {
  servicesStore.stopAutoRefresh()
})
</script>

<style scoped>
.logout-button {
  background: rgba(255,255,255,0.2);
  color: white;
  border: 1px solid rgba(255,255,255,0.3);
  padding: 8px 16px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.3s ease;
  margin-top: 10px;
}

.logout-button:hover {
  background: rgba(255,255,255,0.3);
  transform: translateY(-1px);
}

.retry-button {
  background: #3498db;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 5px;
  cursor: pointer;
  margin-top: 15px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.modal {
  background: white;
  border-radius: 15px;
  padding: 30px;
  max-width: 400px;
  width: 90%;
  text-align: center;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}

.modal h3 {
  margin-bottom: 20px;
  color: #2c3e50;
}

.modal p {
  margin-bottom: 30px;
  color: #7f8c8d;
  line-height: 1.5;
}

.ip-info {
  margin-bottom: 20px !important;
  color: #95a5a6 !important;
  font-size: 0.9rem;
}

.warning-text {
  font-size: 0.9rem;
  color: #e74c3c !important;
}

.modal-buttons {
  display: flex;
  gap: 15px;
  justify-content: center;
}

.modal-button {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.modal-button.cancel {
  background: #ecf0f1;
  color: #2c3e50;
}

.modal-button.confirm {
  background: #e74c3c;
  color: white;
}

.modal-button:hover {
  transform: translateY(-2px);
}
</style> 