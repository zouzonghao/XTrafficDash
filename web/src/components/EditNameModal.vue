<template>
  <div v-if="visible" class="modal-overlay" @click="handleOverlayClick">
    <div class="modal-container" @click.stop>
      <div class="modal-header">
        <h3 class="modal-title">{{ title }}</h3>
        <button class="modal-close" @click="handleClose">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
      
      <div class="modal-body">
        <div class="input-group">
          <label class="input-label">{{ label }}</label>
          <input 
            v-model="editingValue" 
            @keyup.enter="handleSave"
            @keyup.esc="handleClose"
            class="modal-input"
            ref="inputRef"
            :placeholder="placeholder"
            type="text"
          />
        </div>
        
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="handleClose">
            取消
          </button>
          <button class="btn btn-primary" @click="handleSave" :disabled="!editingValue.trim()">
            确认
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: '编辑名称'
  },
  label: {
    type: String,
    default: '名称'
  },
  placeholder: {
    type: String,
    default: '请输入名称'
  },
  value: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:visible', 'save', 'close'])

const editingValue = ref('')
const inputRef = ref(null)

// 监听visible变化，当弹窗打开时设置初始值并聚焦
watch(() => props.visible, (newVal) => {
  if (newVal) {
    editingValue.value = props.value
    nextTick(() => {
      inputRef.value?.focus()
      inputRef.value?.select()
    })
  }
})

// 监听value变化
watch(() => props.value, (newVal) => {
  if (props.visible) {
    editingValue.value = newVal
  }
})

const handleSave = () => {
  const trimmedValue = editingValue.value.trim()
  if (trimmedValue) {
    emit('save', trimmedValue)
    emit('update:visible', false)
  }
}

const handleClose = () => {
  emit('close')
  emit('update:visible', false)
}

const handleOverlayClick = () => {
  handleClose()
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  animation: fadeIn 0.3s ease;
}

.modal-container {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  width: 90%;
  max-width: 480px;
  animation: slideUp 0.3s ease;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 24px 0 24px;
  border-bottom: 1px solid #e1e8ed;
  padding-bottom: 20px;
}

.modal-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: #2c3e50;
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  color: #7f8c8d;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-close:hover {
  background: #f8f9fa;
  color: #2c3e50;
}

.modal-body {
  padding: 24px;
}

.input-group {
  margin-bottom: 24px;
}

.input-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 8px;
}

.modal-input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e1e8ed;
  border-radius: 8px;
  font-size: 1rem;
  color: #2c3e50;
  background: white;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

.modal-input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.modal-input::placeholder {
  color: #bdc3c7;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 80px;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.btn-secondary {
  background: #f8f9fa;
  color: #2c3e50;
  border: 1px solid #e1e8ed;
}

.btn-secondary:hover {
  background: #e9ecef;
  border-color: #ced4da;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@media (max-width: 480px) {
  .modal-container {
    width: 95%;
    margin: 20px;
  }
  
  .modal-header {
    padding: 20px 20px 0 20px;
  }
  
  .modal-body {
    padding: 20px;
  }
  
  .modal-actions {
    flex-direction: column;
  }
  
  .btn {
    width: 100%;
  }
}
</style> 