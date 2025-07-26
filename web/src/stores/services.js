import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { servicesAPI } from '@/utils/api'

export const useServicesStore = defineStore('services', () => {
  const services = ref([])
  const selectedService = ref(null)
  const loading = ref(false)
  const error = ref(null)
  // 新增：详情缓存
  const detailsCache = ref({})

  // 加载服务列表，支持强制刷新
  const loadServices = async (force = false) => {
    if (services.value.length > 0 && !force) {
      return;
    }
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

  // 选择服务
  const selectService = (service) => {
    selectedService.value = service
  }

  // 加载服务详情，支持天数和强制刷新
  const loadServiceDetail = async (serviceId, days = 7, force = false) => {
    const cacheKey = `${serviceId}-${days}d`;
    if (detailsCache.value[cacheKey] && !force) {
      selectedService.value = detailsCache.value[cacheKey];
      return;
    }
    try {
      const response = await servicesAPI.getServiceDetail(serviceId, days)
      if (response.data.success) {
        selectedService.value = {
          ...selectedService.value,
          ...response.data.data
        }
        detailsCache.value[cacheKey] = selectedService.value
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

  // 移除定时刷新相关代码

  // 强制刷新
  const forceRefresh = () => {
    loadServices(true)
    if (selectedService.value) {
      loadServiceDetail(selectedService.value.id, 7, true)
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
    forceRefresh
  }
}) 