import axios from 'axios'

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 15000
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

// 响应拦截器 - 处理认证错误
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    // 处理401认证错误
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
      return Promise.reject(error)
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
  
  // 获取30天流量数据
  getMonthlyTraffic: (serviceId) => api.get(`/db/traffic/monthly/${serviceId}`),
  
  // 获取端口详情
  getPortDetail: (serviceId, tag, days = 7) => api.get(`/db/port-detail/${serviceId}/${tag}?days=${days}`),
  
  // 获取用户详情
  getUserDetail: (serviceId, email, days = 7) => api.get(`/db/user-detail/${serviceId}/${email}?days=${days}`),
  
  // 更新服务自定义名称
  updateServiceCustomName: (serviceId, customName) => api.put(`/db/services/${serviceId}/custom-name`, { custom_name: customName }),
  
  // 更新入站端口自定义名称
  updateInboundCustomName: (serviceId, tag, customName) => api.put(`/db/inbound/${serviceId}/${tag}/custom-name`, { custom_name: customName }),
  
  // 更新客户端自定义名称
  updateClientCustomName: (serviceId, email, customName) => api.put(`/db/client/${serviceId}/${email}/custom-name`, { custom_name: customName }),
  
  // 下载端口历史数据
  downloadPortHistory: (serviceId, tag) => api.get(`/db/download/port-history/${serviceId}/${tag}`, { responseType: 'blob' }),
  
  // 下载用户历史数据
  downloadUserHistory: (serviceId, email) => api.get(`/db/download/user-history/${serviceId}/${email}`, { responseType: 'blob' })
}

export const authAPI = {
  // 登录
  login: (password) => api.post('/auth/login', { password }),
  
  // 验证token
  verifyToken: () => api.get('/auth/verify')
}

export default api 