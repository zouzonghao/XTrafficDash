<template>
  <div class="container">
    <button class="back-button" @click="backToHome">
      ← 返回主页
    </button>

    <div class="header">
      <h1>{{ selectedService?.ip_address }}</h1>
      <p>服务今日流量详情</p>
    </div>

    <div class="detail-container" v-if="selectedService">
      <div class="detail-header">
        <div class="detail-title">服务信息</div>
        <button class="refresh-button" @click="refreshDetail" :disabled="isRefreshing">
          {{ isRefreshing ? '刷新中...' : '刷新数据' }}
        </button>
      </div>

      <div class="traffic-tables">
        <div class="traffic-table">
          <div class="table-title">入站今日流量</div>
          <div 
            v-for="inbound in selectedService.inbound_traffics" 
            :key="inbound.id" 
            class="table-row"
          >
            <div class="table-label-container">
              <div 
                class="table-label clickable" 
                @click="viewPortDetail(selectedService.id, inbound.tag)"
              >
                <span v-if="!inbound.isEditing" class="display-name">
                  {{ inbound.custom_name || inbound.tag }}
                  <button 
                    class="edit-icon" 
                    @click.stop="startEditInbound(inbound)"
                    title="编辑端口名称"
                  >
                    ✏️
                  </button>
                </span>
                <div v-else class="edit-container">
                  <input 
                    v-model="inbound.editingName" 
                    @keyup.enter="saveInboundName(inbound)"
                    @blur="cancelEditInbound(inbound)"
                    class="edit-input"
                    ref="inboundInput"
                  />
                  <button 
                    class="save-icon" 
                    @click.stop="saveInboundName(inbound)"
                    title="保存"
                  >
                    ✅
                  </button>
                </div>
              </div>
            </div>
            <div class="table-value">
              <span class="upload-traffic">
                <span class="traffic-icon">↑</span>
                {{ formatBytes(inbound.up) }}
              </span>
              <span class="download-traffic">
                <span class="traffic-icon">↓</span>
                {{ formatBytes(inbound.down) }}
              </span>
            </div>
          </div>
        </div>

        <div class="traffic-table">
          <div class="table-title">用户今日流量</div>
          <div 
            v-for="client in selectedService.client_traffics" 
            :key="client.id" 
            class="table-row"
          >
            <div class="table-label-container">
              <div 
                class="table-label clickable" 
                @click="viewUserDetail(selectedService.id, client.email)"
              >
                <span v-if="!client.isEditing" class="display-name">
                  {{ client.custom_name || client.email }}
                  <button 
                    class="edit-icon" 
                    @click.stop="startEditClient(client)"
                    title="编辑用户名称"
                  >
                    ✏️
                  </button>
                </span>
                <div v-else class="edit-container">
                  <input 
                    v-model="client.editingName" 
                    @keyup.enter="saveClientName(client)"
                    @blur="cancelEditClient(client)"
                    class="edit-input"
                    ref="clientInput"
                  />
                  <button 
                    class="save-icon" 
                    @click.stop="saveClientName(client)"
                    title="保存"
                  >
                    ✅
                  </button>
                </div>
              </div>
            </div>
            <div class="table-value">
              <span class="upload-traffic">
                <span class="traffic-icon">↑</span>
                {{ formatBytes(client.up) }}
              </span>
              <span class="download-traffic">
                <span class="traffic-icon">↓</span>
                {{ formatBytes(client.down) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="chart-container" style="height: 400px;">
        <canvas id="detail-chart"></canvas>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, computed, ref, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServicesStore } from '../stores/services'
import { formatBytes } from '../utils/formatters'
import { servicesAPI } from '../utils/api'
import Chart from 'chart.js/auto'

const route = useRoute()
const router = useRouter()
const servicesStore = useServicesStore()

const selectedService = computed(() => servicesStore.selectedService)

let detailChart = null
let refreshInterval = null
const isRefreshing = ref(false)

// 编辑相关函数
const startEditInbound = (inbound) => {
  inbound.isEditing = true
  inbound.editingName = inbound.custom_name || inbound.tag
  nextTick(() => {
    // 找到对应的输入框并聚焦
    const inputs = document.querySelectorAll('.edit-input')
    inputs.forEach(input => {
      if (input.value === inbound.editingName) {
        input.focus()
      }
    })
  })
}

const saveInboundName = async (inbound) => {
  try {
    const response = await servicesAPI.updateInboundCustomName(
      selectedService.value.id, 
      inbound.tag, 
      inbound.editingName
    )
    if (response.data.success) {
      inbound.custom_name = inbound.editingName
      inbound.isEditing = false
    } else {
      alert('保存失败: ' + response.data.error)
    }
  } catch (error) {
    console.error('保存端口名称失败:', error)
    alert('保存失败: ' + error.message)
  }
}

const cancelEditInbound = (inbound) => {
  inbound.isEditing = false
  inbound.editingName = ''
}

const startEditClient = (client) => {
  client.isEditing = true
  client.editingName = client.custom_name || client.email
  nextTick(() => {
    // 找到对应的输入框并聚焦
    const inputs = document.querySelectorAll('.edit-input')
    inputs.forEach(input => {
      if (input.value === client.editingName) {
        input.focus()
      }
    })
  })
}

const saveClientName = async (client) => {
  try {
    const response = await servicesAPI.updateClientCustomName(
      selectedService.value.id, 
      client.email, 
      client.editingName
    )
    if (response.data.success) {
      client.custom_name = client.editingName
      client.isEditing = false
    } else {
      alert('保存失败: ' + response.data.error)
    }
  } catch (error) {
    console.error('保存用户名称失败:', error)
    alert('保存失败: ' + error.message)
  }
}

const cancelEditClient = (client) => {
  client.isEditing = false
  client.editingName = ''
}

const createDetailChart = async () => {
  try {
    const response = await servicesAPI.getWeeklyTraffic(selectedService.value.id)
    if (response.data.success) {
      const data = response.data.data
      const ctx = document.getElementById('detail-chart')
      
      if (ctx) {
        // 销毁现有图表
        if (detailChart) {
          detailChart.destroy()
        }
        
        detailChart = new Chart(ctx, {
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
                display: true,
                position: 'top',
                labels: {
                  color: '#2c3e50',
                  font: {
                    size: 14,
                    weight: 'bold'
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
            }
          }
        })
      }
    }
  } catch (error) {
    console.error('创建详情图表失败:', error)
  }
}

const refreshDetail = async () => {
  if (selectedService.value && !isRefreshing.value) {
    isRefreshing.value = true
    try {
      await servicesStore.loadServiceDetail(selectedService.value.id)
      await createDetailChart()
    } catch (error) {
      console.error('刷新数据失败:', error)
    } finally {
      isRefreshing.value = false
    }
  }
}

// 开始自动刷新
const startAutoRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
  // 每60秒刷新一次，确保图表数据实时更新
  refreshInterval = setInterval(async () => {
    if (selectedService.value) {
      await servicesStore.loadServiceDetail(selectedService.value.id)
      await createDetailChart()
    }
  }, 60000)
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

const backToHome = () => {
  router.push('/home')
}

const viewPortDetail = (serviceId, tag) => {
  router.push(`/port/${serviceId}/${tag}`)
}

const viewUserDetail = (serviceId, email) => {
  router.push(`/user/${serviceId}/${email}`)
}

onMounted(async () => {
  const serviceId = parseInt(route.params.serviceId)
  
  // 如果没有选中的服务，从主页获取
  if (!selectedService.value || selectedService.value.id !== serviceId) {
    // 这里需要从主页获取服务列表，然后选择对应的服务
    await servicesStore.loadServices()
    const service = servicesStore.services.find(s => s.id === serviceId)
    if (service) {
      servicesStore.selectService(service)
    }
  }
  
  // 加载服务详情
  if (selectedService.value) {
    await servicesStore.loadServiceDetail(serviceId)
    await createDetailChart()
    startAutoRefresh()
  }
})

onUnmounted(() => {
  if (detailChart) {
    detailChart.destroy()
  }
  stopAutoRefresh()
})
</script>

<style scoped>
.table-label-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.display-name {
  display: flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
  color: #2c3e50;
}

.edit-icon, .save-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 12px;
  padding: 2px;
  border-radius: 3px;
  transition: all 0.2s ease;
}

.edit-icon:hover {
  background: rgba(52, 152, 219, 0.1);
}

.save-icon:hover {
  background: rgba(46, 204, 113, 0.1);
}

.edit-container {
  display: flex;
  align-items: center;
  gap: 4px;
}

.edit-input {
  border: 1px solid #3498db;
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 12px;
  background: white;
  color: #2c3e50;
  min-width: 100px;
}

.edit-input:focus {
  outline: none;
  border-color: #2980b9;
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}
</style> 