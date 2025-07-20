<template>
  <div class="container">
    <button class="back-button" @click="backToDetail">
      ← 返回详情页
    </button>

    <div class="header">
      <h1>
        {{ portDetail?.port_info?.custom_name || portDetail?.port_info?.tag }}
        <button 
          class="edit-icon" 
          @click="startEditPortName"
          title="编辑端口名称"
        >
          ✏️
        </button>
      </h1>
      <p>端口历史流量详情</p>
    </div>

    <div class="detail-container" v-if="portDetail">
      <div class="detail-header">
        <div class="detail-title">端口信息</div>
        <button class="refresh-button" @click="refreshPortDetail">
          刷新数据
        </button>
      </div>

      <div class="port-info">
        <div class="info-grid">
          <div class="info-item">
            <div class="info-label">服务IP</div>
            <div class="info-value">{{ portDetail.port_info.ip_address }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">端口号</div>
            <div class="info-value">{{ portDetail.port_info.port }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">历史上传</div>
            <div class="info-value">{{ formatBytes(portDetail.port_info.total_up) }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">历史下载</div>
            <div class="info-value">{{ formatBytes(portDetail.port_info.total_down) }}</div>
          </div>
          <div class="info-item">
            <div class="info-label">最后活跃</div>
            <div class="info-value">{{ formatDateTime(portDetail.port_info.last_seen) }}</div>
          </div>
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
            (共 {{ portDetail.history.length }} 条记录)
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
  
  <!-- 编辑端口名称弹窗 -->
  <EditNameModal
    v-model:visible="showEditModal"
    :value="currentEditingValue"
    title="编辑端口名称"
    label="端口名称"
    placeholder="请输入端口名称"
    @save="savePortName"
    @close="closeModal"
  />
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServicesStore } from '../stores/services'
import { formatBytes, formatDate, formatDateTime } from '../utils/formatters'
import { servicesAPI } from '../utils/api'
import EditNameModal from '../components/EditNameModal.vue'

const route = useRoute()
const router = useRouter()
const servicesStore = useServicesStore()

const portDetail = ref(null)
const currentHistoryPage = ref(1)
const historyPageSize = 20

// 弹窗相关状态
const showEditModal = ref(false)
const currentEditingValue = ref('')

const selectedService = computed(() => servicesStore.selectedService)

// 分页后的历史数据
const paginatedHistory = computed(() => {
  if (!portDetail.value || !portDetail.value.history) {
    return []
  }
  
  const start = (currentHistoryPage.value - 1) * historyPageSize
  const end = start + historyPageSize
  return portDetail.value.history.slice(start, end)
})

// 总页数
const totalHistoryPages = computed(() => {
  if (!portDetail.value || !portDetail.value.history) {
    return 0
  }
  return Math.ceil(portDetail.value.history.length / historyPageSize)
})

const loadPortDetail = async () => {
  try {
    const serviceId = route.params.serviceId
    const tag = route.params.tag
    const response = await servicesAPI.getPortDetail(serviceId, tag)
    
    if (response.data.success) {
      portDetail.value = response.data.data
      // 重置分页
      currentHistoryPage.value = 1
    }
  } catch (error) {
    console.error('获取端口详情失败:', error)
  }
}

const refreshPortDetail = async () => {
  await loadPortDetail()
}

const backToDetail = () => {
  router.push(`/detail/${route.params.serviceId}`)
}

// 编辑端口名称
const startEditPortName = () => {
  currentEditingValue.value = portDetail.value?.port_info?.custom_name || portDetail.value?.port_info?.tag
  showEditModal.value = true
}

const savePortName = async (newName) => {
  try {
    const response = await servicesAPI.updateInboundCustomName(
      route.params.serviceId,
      portDetail.value.port_info.tag,
      newName
    )
    if (response.data.success) {
      portDetail.value.port_info.custom_name = newName
      showEditModal.value = false
    } else {
      alert('保存失败: ' + response.data.error)
    }
  } catch (error) {
    console.error('保存端口名称失败:', error)
    alert('保存失败: ' + error.message)
  }
}

const closeModal = () => {
  showEditModal.value = false
}



const changeHistoryPage = (page) => {
  currentHistoryPage.value = page
}

onMounted(async () => {
  await loadPortDetail()
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
  background: #2c3e50;
  padding: 12px 16px;
  font-weight: 600;
  color: white;
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
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
</style> 