<template>
  <div class="container">
    <button class="back-button" @click="backToDetail">
      ← 返回节点详情
    </button>

    <div class="header">
      <h1>
        {{ userDetail?.user_info?.custom_name || userDetail?.user_info?.email }}
        <button 
          class="edit-icon" 
          @click="startEditUserName"
          title="编辑用户名称"
        >
          ✏️
        </button>
      </h1>
    </div>

    <div class="detail-container" v-if="userDetail">
      <div class="detail-header">
        <div class="detail-title">用户信息</div>
        <button class="refresh-button" @click="refreshUserDetail">
          刷新数据
        </button>
      </div>

      <div class="user-info">
        <div class="info-grid">
          <div class="info-item">
            <div class="info-label">服务IP</div>
            <div class="info-value">{{ userDetail.user_info.ip }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">所属端口</div>
            <div 
              class="info-value clickable" 
              @click="viewPortDetail(selectedService.id, userDetail.user_info.inbound_tag)"
            >
              {{ userDetail.user_info.inbound_tag }}
            </div>
          </div>
          <div class="info-item">
            <div class="info-label">历史上传</div>
            <div class="info-value">{{ formatBytes(userDetail.user_info.total_up) }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">历史下载</div>
            <div class="info-value">{{ formatBytes(userDetail.user_info.total_down) }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">最后活跃</div>
            <div class="info-value">{{ formatSmartTime(userDetail.user_info.last_seen) }}</div>
          </div>
        </div>
      </div>

      <div class="chart-section">
        <div class="chart-header">
          <div class="section-title">历史流量趋势</div>
          <div class="chart-controls">
            <button 
              class="chart-btn" 
              :class="{ active: chartPeriod === '7d' }"
              @click="switchChartPeriod('7d')"
            >
              7天
            </button>
            <button 
              class="chart-btn" 
              :class="{ active: chartPeriod === '30d' }"
              @click="switchChartPeriod('30d')"
            >
              30天
            </button>
          </div>
        </div>
        <div class="chart-container">
          <canvas id="user-chart"></canvas>
        </div>
      </div>

      <div class="history-section">
        <div class="history-container">
          <div class="section-title">历史流量数据</div>
          <div class="history-table">
          <div class="table-header">
            <div class="header-cell date-col sortable" @click="sortBy('date')">
              日期
              <span class="sort-icon" v-if="sortField === 'date'">
                {{ sortOrder === 'asc' ? '↑' : '↓' }}
              </span>
            </div>
            <div class="header-cell traffic-col sortable" @click="sortBy('daily_up')">
              上传流量
              <span class="sort-icon" v-if="sortField === 'daily_up'">
                {{ sortOrder === 'asc' ? '↑' : '↓' }}
              </span>
            </div>
            <div class="header-cell traffic-col sortable" @click="sortBy('daily_down')">
              下载流量
              <span class="sort-icon" v-if="sortField === 'daily_down'">
                {{ sortOrder === 'asc' ? '↑' : '↓' }}
              </span>
            </div>
            <div class="header-cell traffic-col sortable" @click="sortBy('total_daily')">
              总流量
              <span class="sort-icon" v-if="sortField === 'total_daily'">
                {{ sortOrder === 'asc' ? '↑' : '↓' }}
              </span>
            </div>
          </div>
          <div v-for="item in paginatedHistory" :key="item.date" class="table-row">
            <div class="table-cell date-col">{{ formatDate(item.date) }}</div>
            <div class="table-cell traffic-col upload">
              <span class="traffic-icon">↑</span>
              {{ formatBytes(item.daily_up) }}
            </div>
            <div class="table-cell traffic-col download">
              <span class="traffic-icon">↓</span>
              {{ formatBytes(item.daily_down) }}
            </div>
            <div class="table-cell traffic-col total">
              <span class="traffic-icon">📊</span>
              {{ formatBytes(item.total_daily) }}
            </div>
          </div>
        </div>
        
        <!-- 分页控件 -->
        <div class="pagination" v-if="totalHistoryPages > 1">
          <button 
            class="pagination-btn" 
            :disabled="currentHistoryPage === 1" 
            @click="changeHistoryPage(currentHistoryPage - 1)"
          >
            上一页
          </button>
          <span class="pagination-info">
            第 {{ currentHistoryPage }} 页，共 {{ totalHistoryPages }} 页
            (共 {{ sortedHistory.length }} 条记录)
          </span>
          <button 
            class="pagination-btn" 
            :disabled="currentHistoryPage === totalHistoryPages" 
            @click="changeHistoryPage(currentHistoryPage + 1)"
          >
            下一页
          </button>
        </div>
        </div>
      </div>

      <!-- 下载历史数据按钮 -->
      <div class="download-section">
        <button class="download-button" @click="downloadHistoryData">
          📥 下载历史数据 (CSV)
        </button>
        <p class="download-hint">下载当前用户的所有历史流量数据，包含格式化的流量信息</p>
      </div>

    </div>
  </div>
  
  <!-- 编辑用户名称弹窗 -->
  <EditNameModal
    v-model:visible="showEditModal"
    :value="currentEditingValue"
    title="编辑用户名称"
    label="用户名称"
    placeholder="请输入用户名称"
    @save="saveUserName"
    @close="closeModal"
  />
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServicesStore } from '../stores/services'
import { formatBytes as rawFormatBytes, formatDate, formatSmartTime } from '../utils/formatters'
import { servicesAPI } from '../utils/api'
import Chart from 'chart.js/auto'
import EditNameModal from '../components/EditNameModal.vue'

const route = useRoute()
const router = useRouter()
const servicesStore = useServicesStore()

const userDetail = ref(null)
const currentHistoryPage = ref(1)
const historyPageSize = 10
let userChart = null
const chartPeriod = ref('7d') // 图表周期：7d 或 30d

// 排序相关状态
const sortField = ref('date')
const sortOrder = ref('desc')

// 弹窗相关状态
const showEditModal = ref(false)
const currentEditingValue = ref('')

const selectedService = computed(() => servicesStore.selectedService)

const loadUserDetail = async (days = 7) => {
  try {
    const serviceId = route.params.serviceId
    const email = route.params.email
    const response = await servicesAPI.getUserDetail(serviceId, email, days)
    
    if (response.data.success) {
      userDetail.value = response.data.data
      // 重置分页
      currentHistoryPage.value = 1
    }
  } catch (error) {
    console.error('获取用户详情失败:', error)
  }
}

const refreshUserDetail = async () => {
  const days = chartPeriod.value === '7d' ? 7 : 30
  await loadUserDetail(days)
  await createUserChart()
}

// 切换图表周期
const switchChartPeriod = async (period) => {
  if (chartPeriod.value === period) return
  
  chartPeriod.value = period
  const days = period === '7d' ? 7 : 30
  await loadUserDetail(days)
  await createUserChart()
}

// 创建用户图表
const createUserChart = async () => {
  try {
    if (!userDetail.value || !userDetail.value.history) {
      return
    }

    const ctx = document.getElementById('user-chart')
    if (!ctx) {
      return
    }

    // 销毁旧图表
    if (userChart) {
      userChart.destroy()
    }

    // 准备数据
    const history = [...userDetail.value.history] // 不再reverse，保持API顺序
    const labels = history.map(item => formatDate(item.date))
    const uploadData = history.map(item => item.daily_up)
    const downloadData = history.map(item => item.daily_down)

    // 创建新图表
    userChart = new Chart(ctx, {
      type: 'line',
      data: {
        labels: labels,
        datasets: [
          {
            label: '上传流量',
            data: uploadData,
            borderColor: '#74b9ff',
            backgroundColor: 'rgba(116, 185, 255, 0.1)',
            tension: 0.4,
            fill: true
          },
          {
            label: '下载流量',
            data: downloadData,
            borderColor: '#00b894',
            backgroundColor: 'rgba(0, 184, 148, 0.1)',
            tension: 0.4,
            fill: true
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: true,
            position: 'top',
            labels: {
              color: '#2c3e50',
              font: {
                size: 14,
                weight: 'bold'
              }
            }
          },
          tooltip: {
            callbacks: {
              label: function(context) {
                return context.dataset.label + ': ' + formatBytes(context.parsed.y)
              }
            }
          }
        },
        scales: {
          x: {
            display: true,
            title: {
              display: true,
              text: '日期',
              color: '#2c3e50',
              font: {
                size: 14,
                weight: 'bold'
              }
            },
            ticks: {
              color: '#2c3e50',
              font: {
                size: 12
              }
            },
            grid: {
              color: 'rgba(44, 62, 80, 0.1)'
            }
          },
          y: {
            display: true,
            title: {
              display: true,
              text: '流量',
              color: '#2c3e50',
              font: {
                size: 14,
                weight: 'bold'
              }
            },
            ticks: {
              color: '#2c3e50',
              font: {
                size: 12
              },
              callback: function(value) {
                return formatBytes(value)
              }
            },
            grid: {
              color: 'rgba(44, 62, 80, 0.1)'
            }
          }
        }
      }
    })
  } catch (error) {
    console.error('创建用户图表失败:', error)
  }
}

const backToDetail = () => {
  router.push(`/detail/${route.params.serviceId}`)
}

// 排序后的历史数据
const sortedHistory = computed(() => {
  if (!userDetail.value || !userDetail.value.history) {
    return []
  }
  
  const history = [...userDetail.value.history]
  
  history.sort((a, b) => {
    let aValue, bValue
    
    if (sortField.value === 'date') {
      aValue = new Date(a.date)
      bValue = new Date(b.date)
    } else {
      aValue = a[sortField.value] || 0
      bValue = b[sortField.value] || 0
    }
    
    if (sortOrder.value === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })
  
  return history
})

// 分页后的历史数据
const paginatedHistory = computed(() => {
  const start = (currentHistoryPage.value - 1) * historyPageSize
  const end = start + historyPageSize
  return sortedHistory.value.slice(start, end)
})

// 总页数
const totalHistoryPages = computed(() => {
  return Math.ceil(sortedHistory.value.length / historyPageSize)
})

const viewPortDetail = (serviceId, tag) => {
  router.push(`/port/${serviceId}/${tag}`)
}

const changeHistoryPage = (page) => {
  currentHistoryPage.value = page
}

// 排序功能
const sortBy = (field) => {
  if (sortField.value === field) {
    // 如果点击的是当前排序字段，则切换排序顺序
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    // 如果点击的是新字段，则设置为该字段，默认降序
    sortField.value = field
    sortOrder.value = 'desc'
  }
  // 重置到第一页
  currentHistoryPage.value = 1
}

// 下载历史数据
const downloadHistoryData = async () => {
  try {
    const response = await servicesAPI.downloadUserHistory(route.params.serviceId, route.params.email)
    
    // 创建下载链接
    const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    
    link.setAttribute('href', url)
    link.setAttribute('download', `用户历史数据_${userDetail.value?.user_info?.custom_name || userDetail.value?.user_info?.email}_${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error('下载历史数据失败:', error)
    alert('下载失败: ' + (error.response?.data?.error || error.message))
  }
}

// 编辑用户名称
const startEditUserName = () => {
  // 如果custom_name为空或null，则显示空字符串，让用户可以输入新名称
  // 如果custom_name有值，则显示当前的自定义名称
  currentEditingValue.value = userDetail.value?.user_info?.custom_name || ''
  showEditModal.value = true
}

const saveUserName = async (newName) => {
  try {
    const response = await servicesAPI.updateClientCustomName(
      route.params.serviceId,
      userDetail.value.user_info.email,
      newName
    )
    if (response.data.success) {
      userDetail.value.user_info.custom_name = newName
      showEditModal.value = false
    } else {
      alert('保存失败: ' + response.data.error)
    }
  } catch (error) {
    console.error('保存用户名称失败:', error)
    alert('保存失败: ' + error.message)
  }
}

const closeModal = () => {
  showEditModal.value = false
}

// 页面挂载时默认加载7天
onMounted(async () => {
  await loadUserDetail(7)
  await createUserChart()
})

onUnmounted(() => {
  // 清理图表
  if (userChart) {
    userChart.destroy()
  }
})

function formatBytes(num) {
  if (typeof num !== 'number' || isNaN(num)) return '-';
  if (num >= 1024 * 1024 * 1024) {
    return (num / (1024 * 1024 * 1024)).toPrecision(5) + ' GB';
  } else if (num >= 1024 * 1024) {
    return (num / (1024 * 1024)).toPrecision(5) + ' MB';
  } else if (num >= 1024) {
    return (num / 1024).toPrecision(5) + ' KB';
  } else {
    return num + ' B';
  }
}
</script>

<style scoped>
.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 16px;
  margin-bottom: 25px;
}

/* 保留 info-item 的竖线 */
.info-item {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 6px;
  border-left: 3px solid #70A1FF;
  box-shadow: 0 2px 8px rgba(112,161,255,0.10);
}

.info-label {
  font-size: 0.85rem;
  color: #6c757d;
  margin-bottom: 4px;
}

.info-value {
  font-size: 1rem;
  font-weight: 600;
  color: #495057;
}

.history-section {
  margin-top: 25px;
  margin-bottom: 25px;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 20px;
  color: #495057;
}

.history-container .section-title {
  margin-bottom: 24px;
}

.history-container {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
}

.history-table {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  margin-bottom: 16px;
  border: 1px solid #e9ecef;
}

.table-header {
  display: grid;
  grid-template-columns: 120px 1fr 1fr 1fr;
  background: white;
  padding: 12px 16px;
  font-weight: 600;
  color: #2c3e50;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 2px solid #e9ecef;
}

.header-cell {
  display: flex;
  align-items: center;
}

.header-cell.date-col {
  font-weight: 600;
}

.header-cell.date-col.sortable {
  cursor: pointer;
  user-select: none;
  transition: all 0.2s ease;
  position: relative;
}

.header-cell.date-col.sortable:hover {
  background: rgba(52, 152, 219, 0.1);
  border-radius: 4px;
}

.header-cell.traffic-col {
  justify-content: flex-end;
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
}

.header-cell.sortable {
  cursor: pointer;
  user-select: none;
  transition: all 0.2s ease;
  position: relative;
}

.header-cell.sortable:hover {
  background: rgba(52, 152, 219, 0.1);
  border-radius: 4px;
}

.sort-icon {
  margin-left: 4px;
  font-weight: bold;
  color: #3498db;
}

.table-row {
  display: grid;
  grid-template-columns: 120px 1fr 1fr 1fr;
  padding: 10px 16px;
  border-bottom: 1px solid #f1f3f4;
  transition: all 0.2s ease;
  align-items: center;
  border-left: 3px solid #70A1FF;
}

.table-row:hover {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  transform: translateX(2px);
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.table-row:last-child {
  border-bottom: none;
}

.table-cell {
  display: flex;
  align-items: center;
  color: #495057;
  font-size: 0.85rem;
  font-weight: 500;
}

.date-col {
  font-weight: 600;
  color: #2c3e50;
}

.traffic-col {
  justify-content: flex-end;
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
  gap: 6px;
}

.traffic-icon {
  font-size: 0.8rem;
  opacity: 0.8;
}

.upload {
  color: #74b9ff;
}

.download {
  color: #00b894;
}

.total {
  color: #6c5ce7;
  font-weight: 600;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
}

.pagination-btn {
  background: #007bff;
  color: white;
  border: none;
  padding: 6px 14px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s;
}

.pagination-btn:hover:not(:disabled) {
  background: #0056b3;
}

.pagination-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
}

.pagination-info {
  font-size: 0.9rem;
  color: #2c3e50;
  font-weight: 500;
  background: rgba(255, 255, 255, 0.9);
  padding: 6px 12px;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.edit-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
  margin-left: 8px;
  vertical-align: middle;
}

.edit-icon:hover {
  background: rgba(52, 152, 219, 0.1);
}

/* 图表相关样式 */
.chart-section {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
  margin-top: 25px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 1.2rem;
  font-weight: bold;
  color: #222;
  margin: 0;
}

.chart-controls {
  display: flex;
  gap: 8px;
}

.chart-btn {
  padding: 6px 12px;
  border: 1.5px solid #70A1FF;
  background: #fff;
  color: #70A1FF;
  border-radius: 20px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.2s, color 0.2s, border-color 0.2s, box-shadow 0.2s, transform 0.18s;
  margin-right: 8px;
}
.chart-btn:last-child { margin-right: 0; }

.chart-btn:hover {
  background: #EAF3FF;
  color: #1E90FF;
  border-color: #1E90FF;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(112,161,255,0.10);
}

.chart-btn.active {
  background: #70A1FF;
  color: #fff;
  border-color: #70A1FF;
  box-shadow: 0 2px 8px rgba(112,161,255,0.18);
}

.chart-container {
  height: 400px;
  position: relative;
}

/* 下载按钮样式 */
.download-section {
  margin-top: 30px;
  text-align: center;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
}

.download-button {
  background: #70A1FF;
  color: #fff;
  border: none;
  padding: 12px 24px;
  border-radius: 20px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s, color 0.2s, box-shadow 0.2s, transform 0.18s;
  box-shadow: 0 2px 8px rgba(112,161,255,0.10);
}
.download-button:hover {
  background: #1E90FF;
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(112,161,255,0.18);
}

.download-button:active {
  transform: translateY(0);
}

.download-hint {
  margin-top: 8px;
  color: #6c757d;
  font-size: 14px;
  margin-bottom: 0;
}

.refresh-button {
  background: #70A1FF;
  color: #fff;
  border: none;
  padding: 10px 20px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(112,161,255,0.10);
  transition: background 0.2s, color 0.2s, box-shadow 0.2s, transform 0.18s;
  position: relative;
  overflow: hidden;
}

.refresh-button:hover:not(:disabled) {
  background: #1E90FF;
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(112,161,255,0.18);
}

.refresh-button:disabled {
  background: #d1d1d6;
  color: #fff;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
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

.container {
  max-width: 900px;
  margin: 0 auto;
  padding: 0 24px;
}
</style> 