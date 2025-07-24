import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { servicesAPI } from '@/utils/api'

export const useServicesStore = defineStore('services', () => {
  const services = ref([])
  const selectedService = ref(null)
  const loading = ref(false)
  const error = ref(null)
  const refreshInterval = ref(null)

  const loadServices = async () => {
    loading.value = true
    error.value = null
    
    try {
      const response = await servicesAPI.getServices()
      if (response.data.success) {
        services.value = response.data.data
      } else {
        error.value = response.data.message || '加载服务列表失败'
      }
    } catch (err) {
      console.error('加载服务列表失败:', err)
      if (err.response?.status === 401) {
        error.value = '认证失败，请重新登录'
      } else {
        error.value = '网络错误，请检查服务器连接'
      }
    } finally {
      loading.value = false
    }
  }

  const selectService = (service) => {
    selectedService.value = service
  }

  const loadServiceDetail = async (serviceId, days = 7) => {
    try {
      const response = await servicesAPI.getServiceDetail(serviceId, days)
      if (response.data.success) {
        selectedService.value = {
          ...selectedService.value,
          ...response.data.data
        }
      }
    } catch (error) {
      console.error('加载服务详情失败:', error)
    }
  }

  const deleteService = async (serviceId) => {
    try {
      const response = await servicesAPI.deleteService(serviceId)
      if (response.data.success) {
        services.value = services.value.filter(s => s.id !== serviceId)
        return { success: true }
      } else {
        return { success: false, error: response.data.message }
      }
    } catch (error) {
      console.error('删除服务失败:', error)
      return { success: false, error: '删除失败，请重试' }
    }
  }

  // 开始自动刷新
  const startAutoRefresh = () => {
    if (refreshInterval.value) {
      clearInterval(refreshInterval.value)
    }
    // 每60秒刷新一次
    refreshInterval.value = setInterval(() => {
      loadServices()
    }, 60000)
  }

  // 停止自动刷新
  const stopAutoRefresh = () => {
    if (refreshInterval.value) {
      clearInterval(refreshInterval.value)
      refreshInterval.value = null
    }
  }

  // 强制刷新
  const forceRefresh = () => {
    loadServices()
    if (selectedService.value) {
      loadServiceDetail(selectedService.value.id)
    }
  }

  return {
    services: computed(() => services.value),
    selectedService: computed(() => selectedService.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    loadServices,
    selectService,
    loadServiceDetail,
    deleteService,
    startAutoRefresh,
    stopAutoRefresh,
    forceRefresh
  }
}) 