<template>
  <div class="container">
    <button class="back-button" @click="backToHome">
      ← 返回主页
    </button>
    <div class="header">
      <h1>HY2设置</h1>
      <p>可配置多组HY2流量同步参数，每组独立同步</p>
    </div>
    <div class="content-blocks">
      <!-- XTrafficDash地址设置卡片 -->
      <div class="card block-card">
        <div class="block-title">XTrafficDash地址</div>
        <form class="target-form" @submit.prevent="saveAll">
          <input v-model="targetApiUrl" type="text" class="target-input" placeholder="http://127.0.0.1:37022/api/traffic" required />
          <p class="hint">所有HY2流量数据将发送到此地址</p>
        </form>
      </div>
      <!-- HY2配置列表卡片 -->
      <div class="card block-card">
        <div class="block-title">HY2配置列表</div>
        <form class="hy2-form" @submit.prevent="saveAll">
          <div class="hy2-table-card">
            <table class="hy2-table">
              <thead>
                <tr>
                  <th>IP</th>
                  <th>端口</th>
                  <th>密码</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, idx) in configs" :key="row.id || idx">
                  <td><input v-model="row.source_api_host" type="text" required /></td>
                  <td><input v-model="row.source_api_port" type="text" required /></td>
                  <td><input v-model="row.source_api_password" type="text" required /></td>
                  <td>
                    <button type="button" class="del-btn" @click="removeRow(idx)">删除</button>
                  </td>
                </tr>
              </tbody>
            </table>
            <div class="actions">
              <button type="button" class="add-btn" @click="addRow">添加配置</button>
              <button class="save-button" type="submit" :disabled="loading">{{ loading ? '保存中...' : '保存全部' }}</button>
            </div>
            <div v-if="msg" class="msg">{{ msg }}</div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const router = useRouter()
const configs = ref([])
const targetApiUrl = ref('http://127.0.0.1:37022/api/traffic')
const loading = ref(false)
const msg = ref('')

const backToHome = () => {
  router.push('/home')
}

const loadConfigs = async () => {
  loading.value = true
  msg.value = ''
  try {
    const res = await axios.get('/api/hy2-configs')
    if (res.data.success) {
      configs.value = Array.isArray(res.data.data) ? res.data.data : []
      // 从第一个配置中获取目标地址
      if (configs.value.length > 0) {
        targetApiUrl.value = configs.value[0].target_api_url || 'http://127.0.0.1:37022/api/traffic'
      }
    } else {
      msg.value = res.data.error || '加载失败'
    }
  } catch (e) {
    msg.value = '网络错误，无法加载配置'
  } finally {
    loading.value = false
  }
}

const saveAll = async () => {
  loading.value = true
  msg.value = ''
  // 过滤掉空行，并为每个配置设置相同的目标地址
  const arr = Array.isArray(configs.value) ? configs.value : []
  const toSave = arr
    .filter(row => row.source_api_host && row.source_api_port && row.source_api_password)
    .map(row => ({
      ...row,
      target_api_url: targetApiUrl.value
    }))
  // 允许全部删除后保存（即 toSave 可以为空数组）
  try {
    const res = await axios.post('/api/hy2-configs', toSave)
    if (res.data.success) {
      msg.value = '保存成功！'
      await loadConfigs()
    } else {
      msg.value = res.data.error || '保存失败'
    }
  } catch (e) {
    if (e.response && e.response.data && e.response.data.error) {
      msg.value = e.response.data.error
    } else {
      msg.value = '网络错误，保存失败'
    }
  } finally {
    loading.value = false
  }
}

const addRow = () => {
  configs.value.push(emptyRow())
}

const removeRow = (idx) => {
  configs.value.splice(idx, 1)
}

function emptyRow() {
  return {
    source_api_password: '',
    source_api_host: '',
    source_api_port: ''
  }
}

onMounted(loadConfigs)
</script>

<style scoped>
.back-button {
  position: fixed;
  top: 24px;
  left: 24px;
  background: #fff;
  border: none;
  padding: 10px 22px;
  border-radius: 25px;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 500;
  box-shadow: 0 4px 15px rgba(0,0,0,0.13);
  transition: all 0.3s ease;
  z-index: 1000;
  color: #2c3e50;
}
.back-button:hover {
  background: #f4f6fa;
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0,0,0,0.18);
}



h1 .header {
  font-size: 16px;
  padding: 4px;
  margin-left: 8px;
  vertical-align: middle;
}

.content-blocks {
  display: flex;
  flex-direction: column;
  gap: 32px;
  align-items: center;
}
.card.block-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(44,62,80,0.10);
  padding: 32px 40px 28px 40px;
  width: 100%;
  max-width: 900px;
  margin: 0 auto;
  border: 1px solid #e9ecef;
}
.block-title {
  font-size: 1.18rem;
  font-weight: bold;
  color: #26324b;
  margin-bottom: 22px;
  letter-spacing: 0.5px;
}
.target-form {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
}
.target-input {
  width: 100%;
  padding: 12px 16px;
  border: 1.5px solid #dfe6e9;
  border-radius: 8px;
  font-size: 1.08rem;
  background: #fff;
  transition: border 0.2s;
}
.target-input:focus {
  border: 1.5px solid #0984e3;
  outline: none;
}
.hint {
  color: #6c757d;
  font-size: 0.97rem;
  margin: 0;
}
.hy2-form {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.hy2-table-card {
  width: 100%;
  overflow-x: auto;
}
.hy2-table {
  width: 100%;
  min-width: 0;
  border-collapse: separate;
  border-spacing: 0;
  margin-bottom: 12px;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(44,62,80,0.04);
}
.hy2-table th, .hy2-table td {
  min-width: 0;
  padding: 6px 2px;
  text-align: center;
  font-size: 1.05rem;
  white-space: nowrap;
}
.hy2-table th:last-child, .hy2-table td:last-child {
  min-width: 100px;
}
.hy2-table th {
  background: #f8f9fa;
  font-weight: 600;
  color: #2c3e50;
}
.hy2-table tr:last-child td {
  border-bottom: none;
}
.hy2-table input[type="text"] {
  display: block;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
  text-align: center;
}
.del-btn {
  display: inline-block;
  margin: 0 auto;
}
input[type="text"] {
  padding: 9px 12px;
  border: 1px solid #dfe6e9;
  border-radius: 8px;
  font-size: 1.05rem;
  background: #f9fafb;
  transition: border 0.2s;
}
input[type="text"]:focus {
  border: 1.5px solid #0984e3;
  outline: none;
  background: #fff;
}
.actions {
  display: flex;
  gap: 22px;
  align-items: center;
  margin-bottom: 0;
  margin-top: 12px;
  justify-content: center;
}
.add-btn, .save-button {
  background: linear-gradient(135deg, #74b9ff 0%, #0984e3 100%);
  color: white;
  border: none;
  padding: 10px 26px;
  border-radius: 22px;
  font-size: 1.08rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s, box-shadow 0.2s, transform 0.1s, filter 0.2s;
  box-shadow: 0 2px 8px rgba(9,132,227,0.08);
}
.add-btn {
  background: #27ae60;
  box-shadow: 0 2px 8px rgba(39,174,96,0.08);
}
.add-btn:hover, .save-button:hover {
  filter: brightness(1.08);
  transform: translateY(-2px) scale(1.03);
  box-shadow: 0 4px 16px rgba(44,62,80,0.13);
}
.add-btn:active, .save-button:active {
  filter: brightness(0.96);
  transform: scale(0.98);
  box-shadow: 0 2px 8px rgba(44,62,80,0.08);
}
.del-btn {
  background: #e74c3c;
  color: white;
  border: none;
  padding: 9px 18px;
  border-radius: 18px;
  font-size: 1.03rem;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
  box-shadow: 0 2px 8px rgba(231,76,60,0.08);
}
.del-btn:hover {
  background: #c0392b;
}
.msg {
  margin-top: 18px;
  color: #0984e3;
  text-align: center;
  font-size: 1.08rem;
  font-weight: 500;
}
@media (max-width: 1000px) {
  .card.block-card {
    padding: 18px 8px 18px 8px;
  }
  .hy2-table {
    min-width: 600px;
  }
}
</style> 