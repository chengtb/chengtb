import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || 'http://localhost:8080/api/chef',
})

api.interceptors.request.use(config => {
  const chefId = localStorage.getItem('chef_id')
  if (chefId) config.headers['X-Chef-ID'] = chefId
  return config
})

export default api
