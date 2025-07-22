<template>
  <div class="container">
    <div class="header">
      <h1>XTrafficDash</h1>
      <button @click="handleLogout" class="logout-button">退出登录</button>
      <button @click="goHy2Setting" class="hy2-setting-button">HY2设置</button>
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
        v-for="service in sortedServices"
        :key="service.id"
        :service="service"
        :trafficData="trafficDataMap[service.id]"
        @select="handleSelectService"
        @delete="handleDeleteService"
      />
    </div>

    <!-- 删除确认对话框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="hideDeleteConfirm">
      <div class="modal" @click.stop>
        <h3>确认删除</h3>
        <p>您确定要删除节点 <strong>{{ serviceToDelete?.custom_name || serviceToDelete?.ip }}</strong> 吗？</p>
        <p v-if="serviceToDelete?.custom_name" class="ip-info">IP：{{ serviceToDelete?.ip }}</p>
        <p class="warning-text">此操作将删除该节点的所有数据，包括流量记录和历史数据，且无法恢复。</p>
        <div class="modal-buttons">
          <button class="modal-button cancel" @click="hideDeleteConfirm">取消</button>
          <button class="modal-button confirm" @click="confirmDelete">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useServicesStore } from '../stores/services'
import { useAuthStore } from '../stores/auth'
import ServiceCard from '../components/ServiceCard.vue'
import { servicesAPI } from '../utils/api'

const router = useRouter()
const servicesStore = useServicesStore()
const authStore = useAuthStore()

const showDeleteModal = ref(false)
const serviceToDelete = ref(null)
const trafficDataMap = ref({})

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

const goHy2Setting = () => {
  router.push('/hy2-setting')
}

const loadAllTrafficData = async () => {
  const map = {}
  for (const service of servicesStore.services) {
    const res = await servicesAPI.getWeeklyTraffic(service.id)
    if (res.data.success) {
      map[service.id] = res.data.data
    }
  }
  trafficDataMap.value = map
}

// 排序后的服务列表
const sortedServices = computed(() => {
  // 先拷贝一份，避免影响原数据
  const arr = [...servicesStore.services]
  arr.sort((a, b) => {
    const aName = a.custom_name?.trim()
    const bName = b.custom_name?.trim()
    if (aName && bName) {
      return aName.localeCompare(bName, 'zh-Hans-CN', { sensitivity: 'base' })
    } else if (aName) {
      return -1 // a有名，b没名，a排前
    } else if (bName) {
      return 1 // b有名，a没名，b排前
    } else {
      // 都没名，按ip排序
      return a.ip.localeCompare(b.ip, 'zh-Hans-CN', { sensitivity: 'base' })
    }
  })
  return arr
})

onMounted(async () => {
  await servicesStore.loadServices()
  await loadAllTrafficData()
  servicesStore.startAutoRefresh()
})

onUnmounted(() => {
  servicesStore.stopAutoRefresh()
})
</script>

<style scoped>
.logout-button {
  background: #FF6B81;
  color: #fff;
  border: none;
  padding: 8px 16px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background 0.2s, color 0.2s, box-shadow 0.2s;
  margin-top: 10px;
  box-shadow: 0 2px 8px rgba(255,107,129,0.10);
}

.logout-button:hover {
  background: #FF4757;
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(255,107,129,0.18);
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

.hy2-setting-button {
  background: #70A1FF;
  color: #fff;
  border: none;
  padding: 8px 16px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.9rem;
  margin-left: 12px;
  margin-top: 10px;
  transition: background 0.2s, color 0.2s, box-shadow 0.2s;
  box-shadow: 0 2px 8px rgba(112,161,255,0.10);
}
.hy2-setting-button:hover {
  background: #1E90FF;
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(112,161,255,0.18);
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
  border-radius: 20px;
  cursor: pointer;
  font-size: 1rem;
  transition: background 0.2s, color 0.2s, box-shadow 0.2s, transform 0.18s;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}

.modal-button.cancel {
  background: #f1f2f6;
  color: #222;
}

.modal-button.cancel:hover {
  background: #e1e2e6;
  color: #222;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(0,0,0,0.10);
}

.modal-button.confirm {
  background: #FF6B81;
  color: #fff;
}

.modal-button.confirm:hover {
  background: #FF4757;
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(255,107,129,0.18);
}

.header h1 {
  color: #222;
  text-shadow: none;
}
.detail-title {
  color: #222;
}
.section-title {
  color: #222;
}
</style> 