import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1', timeout: 10000 })
api.interceptors.response.use(res => res.data, err => Promise.reject(err))

export default api

export const listTasks = (params) => api.get('/kitchen/tasks', { params })
export const completeTask = (id) => api.put(`/kitchen/tasks/${id}/complete`)
export const reassignTask = (id, chefId) => api.put(`/kitchen/tasks/${id}/reassign`, { chef_id: chefId })
export const splitOrder = (orderId) => api.post(`/orders/${orderId}/split`)
