<template>
  <div class="service-card" @click="$emit('select', service)">
    <button 
      class="delete-button" 
      @click.stop="$emit('delete', service)"
      title="删除服务"
    >
      X
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
        <div class="stat-value">{{ formatBytes(service.today_inbound_up) }}</div>
        <div class="stat-label">今日上传</div>
      </div>
      <div class="stat-item">
        <div class="stat-value">{{ formatBytes(service.today_inbound_down) }}</div>
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
import { watch } from 'vue'

const props = defineProps({
  service: {
    type: Object,
    required: true
  },
  trafficData: {
    type: Object,
    required: false
  }
})

defineEmits(['select', 'delete'])

let chart = null
const servicesStore = useServicesStore() // 获取 store 实例

// 弹窗相关状态
const showEditModal = ref(false)
const currentEditingValue = ref('')

// 计算显示名称
const displayName = computed(() => {
  return props.service.custom_name || props.service.ip
})

// 打开编辑弹窗
const openEditModal = () => {
  currentEditingValue.value = props.service.custom_name || props.service.ip
  showEditModal.value = true
}

// 保存名称
const saveName = async (newName, done) => {
  try {
    const response = await servicesAPI.updateServiceCustomName(props.service.id, newName)
    if (response.data.success) {
      await servicesStore.forceRefreshAllData()
      done(true); // 成功
    } else {
      alert('保存失败: ' + response.data.error)
      done(false); // 失败
    }
  } catch (error) {
    console.error('保存名称失败:', error)
    alert('保存失败: ' + error.message)
    done(false); // 失败
  }
}

// 关闭弹窗
const closeModal = () => {
  showEditModal.value = false
}

const createChart = () => {
  if (!props.trafficData) return
  const ctx = document.getElementById(`chart-${props.service.id}`)
  if (ctx) {
    if (chart) {
      chart.destroy()
    }
    chart = new Chart(ctx, {
      type: 'line',
      data: {
        labels: props.trafficData.dates,
        datasets: [
          {
            label: '上传',
            data: props.trafficData.upload_data,
            borderColor: '#74b9ff',
            backgroundColor: 'rgba(116, 185, 255, 0.1)',
            tension: 0.4,
            fill: true
          },
          {
            label: '下载',
            data: props.trafficData.download_data,
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

watch(() => props.trafficData, () => {
  createChart()
})

onMounted(() => {
  createChart()
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

.delete-button {
  position: absolute;
  top: 15px;
  right: 15px;
  background: #FF6B81;
  color: #fff;
  border: none;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  cursor: pointer;
  font-size: 14px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s, color 0.2s, box-shadow 0.2s, transform 0.18s;
  z-index: 10;
  box-shadow: 0 2px 8px rgba(255,107,129,0.10);
}

.delete-button:hover {
  background: #FF4757;
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(255,107,129,0.18);
}
</style> 