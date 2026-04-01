import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

api.interceptors.response.use(
  (res) => res.data,
  (err) => Promise.reject(err)
)

export default api

export const getCategories = () => api.get('/categories')
export const getProducts = (categoryId) => api.get('/products', { params: { category_id: categoryId } })
export const getProduct = (id) => api.get(`/products/${id}`)
export const getTable = (id) => api.get(`/tables/${id}`)
export const createOrder = (data) => api.post('/orders', data)
export const getOrdersByTable = (tableId) => api.get(`/tables/${tableId}/orders`)
