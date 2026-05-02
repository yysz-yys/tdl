<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { apiService, type Task } from '../services/api'
import { wsManager } from '../services/ws'

const tasks = ref<Task[]>([])
const loading = ref(true)

const fetchTasks = async () => {
  loading.value = true
  try {
    tasks.value = await apiService.getTasks()
  } catch (error) {
    console.error('Failed to load tasks', error)
  } finally {
    loading.value = false
  }
}

const handleTaskUpdate = (updatedTask: Task) => {
  const index = tasks.value.findIndex(t => t.id === updatedTask.id)
  if (index !== -1) {
    tasks.value[index] = { ...tasks.value[index], ...updatedTask }
  } else {
    tasks.value.push(updatedTask)
  }
}

onMounted(() => {
  fetchTasks()
  
  // Connect to WebSocket
  wsManager.connect()
  
  // Listen for task updates
  wsManager.on('task_update', handleTaskUpdate)
  
  // Mock WS events for demonstration if backend is not available
  // setTimeout(() => {
  //   handleTaskUpdate({
  //     id: '1',
  //     name: 'ubuntu-22.04-desktop-amd64.iso',
  //     status: 'downloading',
  //     progress: 55.0,
  //     size: 4900000000,
  //     downloaded: 2695000000,
  //   })
  // }, 3000)
})

onUnmounted(() => {
  wsManager.off('task_update', handleTaskUpdate)
  wsManager.disconnect()
})

const formatBytes = (bytes: number, decimals = 2) => {
  if (!+bytes) return '0 Bytes'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}

const getStatusColor = (status: Task['status']) => {
  switch (status) {
    case 'downloading': return 'text-blue-600 bg-blue-100'
    case 'completed': return 'text-green-600 bg-green-100'
    case 'error': return 'text-red-600 bg-red-100'
    case 'pending': default: return 'text-gray-600 bg-gray-100'
  }
}

const getProgressBarColor = (status: Task['status']) => {
  switch (status) {
    case 'downloading': return 'bg-blue-600'
    case 'completed': return 'bg-green-600'
    case 'error': return 'bg-red-600'
    case 'pending': default: return 'bg-gray-400'
  }
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
    <div class="max-w-4xl mx-auto">
      <div class="flex items-center justify-between mb-8">
        <h1 class="text-3xl font-bold text-gray-900">TDL Dashboard</h1>
        <button 
          @click="fetchTasks" 
          class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition-colors shadow-sm text-sm font-medium"
        >
          Refresh Tasks
        </button>
      </div>

      <div v-if="loading" class="flex justify-center py-12">
        <div class="animate-spin rounded-full h-10 w-10 border-b-2 border-indigo-600"></div>
      </div>

      <div v-else-if="tasks.length === 0" class="bg-white rounded-lg shadow p-8 text-center text-gray-500">
        No active tasks found.
      </div>

      <div v-else class="space-y-4">
        <div 
          v-for="task in tasks" 
          :key="task.id" 
          class="bg-white rounded-lg shadow overflow-hidden p-6 transition-all hover:shadow-md"
        >
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-medium text-gray-900 truncate pr-4" :title="task.name">
              {{ task.name }}
            </h3>
            <span 
              class="px-3 py-1 inline-flex text-xs leading-5 font-semibold rounded-full capitalize"
              :class="getStatusColor(task.status)"
            >
              {{ task.status }}
            </span>
          </div>
          
          <div class="mb-2 flex justify-between text-sm text-gray-600">
            <span>{{ formatBytes(task.downloaded) }} / {{ formatBytes(task.size) }}</span>
            <span class="font-medium">{{ task.progress.toFixed(1) }}%</span>
          </div>
          
          <div class="w-full bg-gray-200 rounded-full h-2.5">
            <div 
              class="h-2.5 rounded-full transition-all duration-500 ease-in-out" 
              :class="getProgressBarColor(task.status)"
              :style="{ width: `${task.progress}%` }"
            ></div>
          </div>
          
          <div class="mt-4 flex justify-end space-x-3">
            <button 
              v-if="task.status === 'downloading'"
              @click="apiService.pauseTask(task.id)"
              class="text-sm text-gray-600 hover:text-gray-900 font-medium px-3 py-1.5 rounded hover:bg-gray-100 transition-colors"
            >
              Pause
            </button>
            <button 
              v-if="task.status === 'pending' || task.status === 'error'"
              @click="apiService.startTask(task.id)"
              class="text-sm text-blue-600 hover:text-blue-900 font-medium px-3 py-1.5 rounded hover:bg-blue-50 transition-colors"
            >
              Start
            </button>
            <button 
              @click="apiService.deleteTask(task.id)"
              class="text-sm text-red-600 hover:text-red-900 font-medium px-3 py-1.5 rounded hover:bg-red-50 transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
