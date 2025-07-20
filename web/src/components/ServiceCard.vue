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
          <span v-if="!isEditingName" class="display-name">
            {{ displayName }}
            <button 
              class="edit-icon" 
              @click.stop="startEditName"
              title="编辑名称"
            >
              ✏️
            </button>
          </span>
          <div v-else class="edit-container">
            <input 
              v-model="editingName" 
              @keyup.enter="saveName"
              @blur="handleBlur"
              class="edit-input"
              ref="nameInput"
            />
            <button 
              class="save-button" 
              @click.stop="saveName"
              title="确认"
            >
              确认
            </button>
            <button 
              class="cancel-button" 
              @click.stop="cancelEditName"
              title="取消"
            >
              取消
            </button>
          </div>
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
</template>

<script setup>
import { onMounted, onUnmounted, ref, computed, nextTick } from 'vue'
import { formatBytes } from '../utils/formatters'
import { servicesAPI } from '../utils/api'
import Chart from 'chart.js/auto'

const props = defineProps({
  service: {
    type: Object,
    required: true
  }
})

defineEmits(['select', 'delete'])

let chart = null

// 编辑名称相关状态
const isEditingName = ref(false)
const editingName = ref('')
const nameInput = ref(null)

// 计算显示名称
const displayName = computed(() => {
  return props.service.custom_name || props.service.ip_address
})

// 开始编辑名称
const startEditName = () => {
  editingName.value = props.service.custom_name || props.service.ip_address
  isEditingName.value = true
  nextTick(() => {
    nameInput.value?.focus()
  })
}

// 保存名称
const saveName = async () => {
  try {
    const response = await servicesAPI.updateServiceCustomName(props.service.id, editingName.value)
    if (response.data.success) {
      // 更新本地数据
      props.service.custom_name = editingName.value
      isEditingName.value = false
    } else {
      alert('保存失败: ' + response.data.error)
    }
  } catch (error) {
    console.error('保存名称失败:', error)
    alert('保存失败: ' + error.message)
  }
}

// 取消编辑
const cancelEditName = () => {
  isEditingName.value = false
  editingName.value = ''
}

// 处理失去焦点事件
const handleBlur = (event) => {
  // 延迟执行，避免与按钮点击冲突
  setTimeout(() => {
    // 检查是否点击了保存或取消按钮
    const relatedTarget = event.relatedTarget
    if (relatedTarget && (relatedTarget.classList.contains('save-button') || relatedTarget.classList.contains('cancel-button'))) {
      return
    }
    // 如果没有点击按钮，则取消编辑
    cancelEditName()
  }, 100)
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

.save-button, .cancel-button {
  background: none;
  border: 1px solid #ddd;
  cursor: pointer;
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 3px;
  transition: all 0.2s ease;
  color: #2c3e50;
}

.save-button {
  background: #27ae60;
  color: white;
  border-color: #27ae60;
}

.save-button:hover {
  background: #229954;
  border-color: #229954;
}

.cancel-button {
  background: #ecf0f1;
  border-color: #bdc3c7;
}

.cancel-button:hover {
  background: #d5dbdb;
  border-color: #95a5a6;
}

.edit-container {
  display: flex;
  align-items: center;
  gap: 4px;
}

.edit-input {
  border: 1px solid #3498db;
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 14px;
  background: white;
  color: #2c3e50;
  min-width: 120px;
}

.edit-input:focus {
  outline: none;
  border-color: #2980b9;
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}
</style> 