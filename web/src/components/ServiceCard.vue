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
        {{ service.ip_address }}
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
import { onMounted, onUnmounted } from 'vue'
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