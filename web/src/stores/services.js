import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { servicesAPI } from '@/utils/api'

export const useServicesStore = defineStore('services', () => {
  const services = ref([])
  const selectedService = ref(null)
  const loading = ref(false)
  const error = ref(null)
  // 统一缓存管理
  const detailsCache = ref({})
  const portDetailsCache = ref({}) // 新增：端口详情缓存
  const userDetailsCache = ref({}) // 新增：用户详情缓存

  const clearAllCaches = () => {
    detailsCache.value = {};
    portDetailsCache.value = {};
    userDetailsCache.value = {};
    console.log("所有缓存已清空。");
  };

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

  // 查找服务
  const findServiceById = (serviceId) => {
    return services.value.find(s => s.id === serviceId) || {};
  }

  // 加载服务详情, 增加 isPreload 参数用于区分是用户导航还是后台预加载
  const loadServiceDetail = async (serviceId, days = 7, force = false, isPreload = false) => {
    const cacheKey = `${serviceId}-${days}d`;
    if (detailsCache.value[cacheKey] && !force) {
      if (!isPreload) {
        selectedService.value = detailsCache.value[cacheKey];
      }
      return detailsCache.value[cacheKey];
    }
    try {
      const response = await servicesAPI.getServiceDetail(serviceId, days)
      if (response.data.success) {
        const fullDetail = {
          ...findServiceById(serviceId),
          ...response.data.data
        };
        detailsCache.value[cacheKey] = fullDetail;
        if (!isPreload) {
          selectedService.value = fullDetail;
        }
        return fullDetail;
      }
    } catch (error) {
      console.error('加载服务详情失败:', error)
    }
    return null;
  }

  // 获取端口详情，带缓存
  const getPortDetail = async (serviceId, tag, days = 7, force = false) => {
    const cacheKey = `${serviceId}-${tag}-${days}d`;
    if (portDetailsCache.value[cacheKey] && !force) {
      return portDetailsCache.value[cacheKey];
    }
    try {
      const response = await servicesAPI.getPortDetail(serviceId, tag, days);
      if (response.data.success) {
        portDetailsCache.value[cacheKey] = response.data.data;
        return response.data.data;
      }
    } catch (error) {
      console.error(`加载端口详情失败 for ${tag}:`, error);
    }
    return null;
  };

  // 获取用户详情，带缓存
  const getUserDetail = async (serviceId, email, days = 7, force = false) => {
    const cacheKey = `${serviceId}-${email}-${days}d`;
    if (userDetailsCache.value[cacheKey] && !force) {
      return userDetailsCache.value[cacheKey];
    }
    try {
      const response = await servicesAPI.getUserDetail(serviceId, email, days);
      if (response.data.success) {
        userDetailsCache.value[cacheKey] = response.data.data;
        return response.data.data;
      }
    } catch (error) {
      console.error(`加载用户详情失败 for ${email}:`, error);
    }
    return null;
  };

  // 核心预加载函数
  const preloadAllDetails = async (isForced = false) => {
    if (isForced) {
      clearAllCaches();
    }
    console.log("启动后台数据预加载...");
    if (services.value.length === 0 || isForced) {
      await loadServices(true);
    }

    const preloadPromises = services.value.map(async (service) => {
      const serviceDetail = await loadServiceDetail(service.id, 7, isForced, true);
      if (serviceDetail) {
        const subDetailPromises = [];
        if (serviceDetail.inbound_traffics) {
          serviceDetail.inbound_traffics.forEach(inbound => {
            subDetailPromises.push(getPortDetail(service.id, inbound.tag, 7, isForced));
          });
        }
        if (serviceDetail.client_traffics) {
          serviceDetail.client_traffics.forEach(client => {
            subDetailPromises.push(getUserDetail(service.id, client.email, 7, isForced));
          });
        }
        await Promise.allSettled(subDetailPromises);
      }
    });

    Promise.allSettled(preloadPromises).then(() => {
        console.log("后台预加载完成。");
    });
  };

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

  // 之前的 forceRefresh 名字有歧义，改为只刷新当前选中的
  const forceRefreshSelected = async () => {
    if (selectedService.value) {
      await loadServiceDetail(selectedService.value.id, 7, true)
    }
  }

  // 全局强制刷新
  const forceRefreshAllData = async () => {
    console.log("全局数据刷新...");
    await preloadAllDetails(true);
  };

  return {
    services: computed(() => services.value),
    selectedService: computed(() => selectedService.value),
    loading: computed(() => loading.value),
    error: computed(() => error.value),
    loadServices,
    selectService,
    loadServiceDetail,
    getPortDetail,
    getUserDetail,
    preloadAllDetails,
    forceRefreshSelected,
    forceRefreshAllData,
    deleteService,
  }
}) 