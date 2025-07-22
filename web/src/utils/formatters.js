export const formatBytes = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export const formatDate = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN')
}

// 智能时间格式化：根据时间差自动选择“刚刚”、“xx分钟前”、“xx小时前”、“xx天前”、“xx个月前”、“xx年前”或 yyyy-MM-dd HH:mm
export function formatSmartTime(dateStr) {
  if (!dateStr) return ''
  let date
  if (typeof dateStr === 'string') {
    if (dateStr.includes('T') || dateStr.includes('Z')) {
      date = new Date(dateStr)
    } else {
      date = new Date(dateStr.replace(' ', 'T'))
    }
  } else {
    date = new Date(dateStr)
  }
  const now = new Date()
  const diffMs = now - date
  const diffMin = Math.floor(diffMs / (1000 * 60))
  const diffHour = Math.floor(diffMs / (1000 * 60 * 60))
  const diffDay = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin}分钟前`
  if (diffHour < 24) return `${diffHour}小时前`
  if (diffDay < 30) return `${diffDay}天前`
  if (diffDay < 365) return `${Math.floor(diffDay / 30)}个月前`
  if (diffDay >= 365) return `${Math.floor(diffDay / 365)}年前`
  // 超过24小时，显示 yyyy-MM-dd HH:mm
  const y = date.getFullYear()
  const m = (date.getMonth() + 1).toString().padStart(2, '0')
  const d = date.getDate().toString().padStart(2, '0')
  const hh = date.getHours().toString().padStart(2, '0')
  const mm = date.getMinutes().toString().padStart(2, '0')
  return `${y}-${m}-${d} ${hh}:${mm}`
} 