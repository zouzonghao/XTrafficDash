<template>
  <div class="service-card" @click="$emit('select', service)">
    <button 
      class="delete-button" 
      @click.stop="$emit('delete', service)"
      title="删除服务"
    >
      ×
    </button>
    
    <div class="card-header">
      <div class="ip-address">
        <div class="name-container">
          <span class="display-name">
            {{ displayName }}
            <button 
              class="edit-icon" 
              @click.stop="openEditModal"
              title="编辑名称"
            >
              ✏️
            </button>
          </span>
        </div>
        <span 
          class="status-badge" 
          :class="service.status === 'active' ? 'status-active' : 'status-inactive'"
        >
          {{ service.status === 'active' ? '活跃' : '离线' }}
        </span>
      </div>
    </div>

    <div class="stats-grid">
      <div class="stat-item">
        <div class="stat-value">{{ formatBytes(service.total_inbound_up) }}</div>
        <div class="stat-label">今日上传</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">{{ formatBytes(service.total_inbound_down) }}</div>
        <div class="stat-label">今日下载</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">{{ service.inbound_count }}</div>
        <div class="stat-label">入站端口</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">{{ service.client_count }}</div>
        <div class="stat-label">用户数量</div>
      </div>
    </div>

    <div class="chart-container">
      <canvas :id="'chart-' + service.id"></canvas>
    </div>
  </div>
  
  <!-- 编辑名称弹窗 -->
  <EditNameModal
    v-model:visible="showEditModal"
    :value="currentEditingValue"
    title="编辑节点名称"
    label="节点名称"
    placeholder="请输入节点名称"
    @save="saveName"
    @close="closeModal"
  />
</template>

<script setup>
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { formatBytes } from '../utils/formatters'
import { servicesAPI } from '../utils/api'
import Chart from 'chart.js/auto'
import EditNameModal from './EditNameModal.vue'
import { useServicesStore } from '../stores/services'

const props = defineProps({
  service: {
    type: Object,
    required: true
  }
})

defineEmits(['select', 'delete'])

let chart = null

// 弹窗相关状态
const showEditModal = ref(false)
const currentEditingValue = ref('')

// 计算显示名称
const displayName = computed(() => {
  return props.service.custom_name || props.service.ip_address
})

// 打开编辑弹窗
const openEditModal = () => {
  currentEditingValue.value = props.service.custom_name || props.service.ip_address
  showEditModal.value = true
}

// 保存名称
const saveName = async (newName) => {
  try {
    const response = await servicesAPI.updateServiceCustomName(props.service.id, newName)
    if (response.data.success) {
      // 更新本地数据
      props.service.custom_name = newName
      // 如果当前选中的服务就是这个服务，也要更新store中的数据
      const servicesStore = useServicesStore()
      if (servicesStore.selectedService && servicesStore.selectedService.id === props.service.id) {
        servicesStore.selectedService.custom_name = newName
      }
      showEditModal.value = false
    } else {
      alert('保存失败: ' + response.data.error)
    }
  } catch (error) {
    console.error('保存名称失败:', error)
    alert('保存失败: ' + error.message)
  }
}

// 关闭弹窗
const closeModal = () => {
  showEditModal.value = false
}

const createChart = async () => {
  try {
    const response = await servicesAPI.getWeeklyTraffic(props.service.id)
    if (response.data.success) {
      const data = response.data.data
      const ctx = document.getElementById(`chart-${props.service.id}`)
      
      if (ctx) {
        chart = new Chart(ctx, {
          type: 'line',
          data: {
            labels: data.dates,
            datasets: [
              {
                label: '上传',
                data: data.upload_data,
                borderColor: '#74b9ff',
                backgroundColor: 'rgba(116, 185, 255, 0.1)',
                tension: 0.4
              },
              {
                label: '下载',
                data: data.download_data,
                borderColor: '#00b894',
                backgroundColor: 'rgba(0, 184, 148, 0.1)',
                tension: 0.4
              }
            ]
          },
          options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
              legend: {
                display: false
              }
            },
            scales: {
              x: {
                display: true,
                ticks: {
                  color: '#2c3e50',
                  font: {
                    size: 10
                  }
                },
                grid: {
                  color: 'rgba(44, 62, 80, 0.1)'
                }
              },
              y: {
                display: true,
                ticks: {
                  color: '#2c3e50',
                  font: {
                    size: 10
                  },
                  callback: function(value, index, values) {
                    if (value >= 1024 * 1024 * 1024) {
                      return (value / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
                    } else if (value >= 1024 * 1024) {
                      return (value / (1024 * 1024)).toFixed(1) + ' MB';
                    } else if (value >= 1024) {
                      return (value / 1024).toFixed(1) + ' KB';
                    } else {
                      return value + ' B';
                    }
                  }
                },
                grid: {
                  color: 'rgba(44, 62, 80, 0.1)'
                }
              }
            },
            elements: {
              point: {
                radius: 0
              }
            }
          }
        })
      }
    }
  } catch (error) {
    console.error('创建图表失败:', error)
  }
}

onMounted(() => {
  setTimeout(createChart, 100)
})

onUnmounted(() => {
  if (chart) {
    chart.destroy()
  }
})
</script>

<style scoped>
.name-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.display-name {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 600;
  color: #2c3e50;
}

.edit-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  padding: 2px;
  border-radius: 3px;
  transition: all 0.2s ease;
}

.edit-icon:hover {
  background: rgba(52, 152, 219, 0.1);
}
</style> 