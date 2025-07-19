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
            <div 
              class="table-label clickable" 
              @click="viewPortDetail(selectedService.id, inbound.tag)"
            >
              {{ inbound.tag }}
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
            <div 
              class="table-label clickable" 
              @click="viewUserDetail(selectedService.id, client.email)"
            >
              {{ client.email }}
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
import { onMounted, onUnmounted, computed, ref } from 'vue'
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