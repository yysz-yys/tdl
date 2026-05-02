const API_BASE_URL = 'http://localhost:8080/api/v1';

export interface Task {
  id: string;
  name: string;
  status: 'pending' | 'downloading' | 'completed' | 'error';
  progress: number; // 0 to 100
  size: number;
  downloaded: number;
}

export const apiService = {
  async getTasks(): Promise<Task[]> {
    try {
      const response = await fetch(`${API_BASE_URL}/tasks`);
      if (!response.ok) {
        throw new Error('Failed to fetch tasks');
      }
      return await response.json();
    } catch (error) {
      console.error('API Error:', error);
      // Return mock data if API is unreachable for demonstration
      return [
        {
          id: '1',
          name: 'ubuntu-22.04-desktop-amd64.iso',
          status: 'downloading',
          progress: 45.5,
          size: 4900000000,
          downloaded: 2229500000,
        },
        {
          id: '2',
          name: 'project-backup-2023.zip',
          status: 'completed',
          progress: 100,
          size: 1024000000,
          downloaded: 1024000000,
        },
        {
          id: '3',
          name: 'data-dump.sql',
          status: 'pending',
          progress: 0,
          size: 500000000,
          downloaded: 0,
        }
      ];
    }
  },

  async startTask(id: string): Promise<void> {
    await fetch(`${API_BASE_URL}/tasks/${id}/start`, { method: 'POST' });
  },

  async pauseTask(id: string): Promise<void> {
    await fetch(`${API_BASE_URL}/tasks/${id}/pause`, { method: 'POST' });
  },

  async deleteTask(id: string): Promise<void> {
    await fetch(`${API_BASE_URL}/tasks/${id}`, { method: 'DELETE' });
  }
};
