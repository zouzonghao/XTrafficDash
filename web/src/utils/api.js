import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 15000, // 增加超时时间
  retry: 3, // 重试次数
  retryDelay: 1000 // 重试延迟
})

// 请求拦截器 - 添加认证token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    console.error('请求拦截器错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理认证错误和重试
api.interceptors.response.use(
  (response) => {
    return response
  },
  async (error) => {
    const { config } = error
    
    // 处理401认证错误
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
      return Promise.reject(error)
    }
    
    // 重试逻辑
    if (config && config.retry > 0) {
      config.retry -= 1
      
      if (config.retry > 0) {
        await new Promise(resolve => setTimeout(resolve, config.retryDelay))
        return api(config)
      }
    }
    
    console.error('API请求失败:', error)
    return Promise.reject(error)
  }
)

export const servicesAPI = {
  // 获取服务列表
  getServices: () => api.get('/db/services'),
  
  // 获取服务详情
  getServiceDetail: (serviceId) => api.get(`/db/services/${serviceId}/traffic`),
  
  // 删除服务
  deleteService: (serviceId) => api.delete(`/db/services/${serviceId}`),
  
  // 获取7天流量数据
  getWeeklyTraffic: (serviceId) => api.get(`/db/traffic/weekly/${serviceId}`),
  
  // 获取端口详情
  getPortDetail: (serviceId, tag) => api.get(`/db/port-detail/${serviceId}/${tag}`),
  
  // 获取用户详情
  getUserDetail: (serviceId, email) => api.get(`/db/user-detail/${serviceId}/${email}`),
  
  // 更新服务自定义名称
  updateServiceCustomName: (serviceId, customName) => api.put(`/db/services/${serviceId}/custom-name`, { custom_name: customName }),
  
  // 更新入站端口自定义名称
  updateInboundCustomName: (serviceId, tag, customName) => api.put(`/db/inbound/${serviceId}/${tag}/custom-name`, { custom_name: customName }),
  
  // 更新客户端自定义名称
  updateClientCustomName: (serviceId, email, customName) => api.put(`/db/client/${serviceId}/${email}/custom-name`, { custom_name: customName })
}

export const authAPI = {
  // 登录
  login: (password) => api.post('/auth/login', { password }),
  
  // 验证token
  verifyToken: () => api.get('/auth/verify')
}

export default api 