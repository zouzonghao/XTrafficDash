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

export const formatDateTime = (dateString) => {
  // 处理时区问题：确保时间字符串被正确解析为本地时间
  let date
  if (typeof dateString === 'string') {
    // 如果时间字符串包含 'T' 或 'Z'，说明是ISO格式，需要特殊处理
    if (dateString.includes('T') || dateString.includes('Z')) {
      date = new Date(dateString)
    } else {
      // 对于SQLite的TIMESTAMP格式，直接解析
      date = new Date(dateString.replace(' ', 'T'))
    }
  } else {
    date = new Date(dateString)
  }
  
  const now = new Date()
  const diffInMs = now - date
  const diffInMinutes = Math.floor(diffInMs / (1000 * 60))
  const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60))
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24))

  // 如果是今天
  if (diffInDays === 0) {
    if (diffInMinutes < 1) {
      return '刚刚'
    } else if (diffInMinutes < 60) {
      return `${diffInMinutes}分钟前`
    } else {
      return `${diffInHours}小时前`
    }
  }
  // 如果是昨天
  else if (diffInDays === 1) {
    return '昨天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  // 如果是前天
  else if (diffInDays === 2) {
    return '前天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  // 如果是一周内
  else if (diffInDays < 7) {
    return `${diffInDays}天前`
  }
  // 如果是一年内
  else if (diffInDays < 365) {
    return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
  }
  // 超过一年
  else {
    return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
  }
}

export const formatTimeAgo = (dateString) => {
  // 处理时区问题：确保时间字符串被正确解析为本地时间
  let date
  if (typeof dateString === 'string') {
    // 如果时间字符串包含 'T' 或 'Z'，说明是ISO格式，需要特殊处理
    if (dateString.includes('T') || dateString.includes('Z')) {
      date = new Date(dateString)
    } else {
      // 对于SQLite的TIMESTAMP格式，直接解析
      date = new Date(dateString.replace(' ', 'T'))
    }
  } else {
    date = new Date(dateString)
  }
  
  const now = new Date()
  const diffInMs = now - date
  const diffInMinutes = Math.floor(diffInMs / (1000 * 60))
  const diffInHours = Math.floor(diffInMs / (1000 * 60 * 60))
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24))

  if (diffInMinutes < 1) {
    return '刚刚'
  } else if (diffInMinutes < 60) {
    return `${diffInMinutes}分钟前`
  } else if (diffInHours < 24) {
    return `${diffInHours}小时前`
  } else if (diffInDays < 30) {
    return `${diffInDays}天前`
  } else if (diffInDays < 365) {
    const months = Math.floor(diffInDays / 30)
    return `${months}个月前`
  } else {
    const years = Math.floor(diffInDays / 365)
    return `${years}年前`
  }
} 