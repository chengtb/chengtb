import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/waiter',
})

api.interceptors.request.use(config => {
  const waiterId = localStorage.getItem('waiter_id')
  if (waiterId) config.headers['X-Waiter-ID'] = waiterId
  return config
})

export default api
