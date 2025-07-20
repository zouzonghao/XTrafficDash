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
      <p>用户历史流量详情</p>
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
            <div class="info-value">{{ userDetail.user_info.ip_address }}</div>
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
            <div class="info-value">{{ formatDateTime(userDetail.user_info.last_seen) }}</div>
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
            <div class="header-cell date-col">日期</div>
            <div class="header-cell traffic-col">上传流量</div>
            <div class="header-cell traffic-col">下载流量</div>
            <div class="header-cell traffic-col">总流量</div>
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
            (共 {{ userDetail.history.length }} 条记录)
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
import { formatBytes, formatDate, formatDateTime } from '../utils/formatters'
import { servicesAPI } from '../utils/api'
import Chart from 'chart.js/auto'
import EditNameModal from '../components/EditNameModal.vue'

const route = useRoute()
const router = useRouter()
const servicesStore = useServicesStore()

const userDetail = ref(null)
const currentHistoryPage = ref(1)
const historyPageSize = 20
let userChart = null
const chartPeriod = ref('7d') // 图表周期：7d 或 30d

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
      // 更新图表
      updateChart()
    }
  } catch (error) {
    console.error('获取用户详情失败:', error)
  }
}

const refreshUserDetail = async () => {
  const days = chartPeriod.value === '7d' ? 7 : 30
  await loadUserDetail(days)
}

// 切换图表周期
const switchChartPeriod = async (period) => {
  chartPeriod.value = period
  const days = period === '7d' ? 7 : 30
  await loadUserDetail(days)
}

// 更新图表
const updateChart = () => {
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
  const history = [...userDetail.value.history].reverse() // 按时间正序
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
          fill: true
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'top',
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
        y: {
          beginAtZero: true,
          ticks: {
            callback: function(value) {
              return formatBytes(value)
            }
          }
        }
      }
    }
  })
}

const backToDetail = () => {
  router.push(`/detail/${route.params.serviceId}`)
}

// 分页后的历史数据
const paginatedHistory = computed(() => {
  if (!userDetail.value || !userDetail.value.history) {
    return []
  }
  
  const start = (currentHistoryPage.value - 1) * historyPageSize
  const end = start + historyPageSize
  return userDetail.value.history.slice(start, end)
})

// 总页数
const totalHistoryPages = computed(() => {
  if (!userDetail.value || !userDetail.value.history) {
    return 0
  }
  return Math.ceil(userDetail.value.history.length / historyPageSize)
})

const viewPortDetail = (serviceId, tag) => {
  router.push(`/port/${serviceId}/${tag}`)
}

const changeHistoryPage = (page) => {
  currentHistoryPage.value = page
}

// 编辑用户名称
const startEditUserName = () => {
  currentEditingValue.value = userDetail.value?.user_info?.custom_name || userDetail.value?.user_info?.email
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

onMounted(async () => {
  await loadUserDetail(7) // 默认加载7天数据
})

onUnmounted(() => {
  // 清理图表
  if (userChart) {
    userChart.destroy()
  }
})
</script>

<style scoped>
.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
  margin-bottom: 25px;
}

.info-item {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 6px;
  border-left: 3px solid #007bff;
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
  margin-bottom: 12px;
  color: #495057;
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

.header-cell.traffic-col {
  justify-content: flex-end;
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
}

.table-row {
  display: grid;
  grid-template-columns: 120px 1fr 1fr 1fr;
  padding: 10px 16px;
  border-bottom: 1px solid #f1f3f4;
  transition: all 0.2s ease;
  align-items: center;
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
  margin-top: 25px;
  margin-bottom: 25px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.chart-controls {
  display: flex;
  gap: 8px;
}

.chart-btn {
  background: #f8f9fa;
  color: #6c757d;
  border: 1px solid #dee2e6;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s ease;
}

.chart-btn:hover {
  background: #e9ecef;
  border-color: #adb5bd;
}

.chart-btn.active {
  background: #007bff;
  color: white;
  border-color: #007bff;
}

.chart-container {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 15px rgba(0,0,0,0.1);
  height: 300px;
  position: relative;
}
</style> 