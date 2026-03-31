import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1', timeout: 10000 })
api.interceptors.response.use(res => res.data, err => Promise.reject(err))

export default api

export const listOrders = (params) => api.get('/orders', { params })
export const getOrder = (id) => api.get(`/orders/${id}`)
export const confirmOrder = (id, waiterID) => api.put(`/orders/${id}/confirm`, { waiter_id: waiterID })
export const updateOrderItems = (id, data) => api.put(`/orders/${id}/items`, data)
export const listRegions = () => api.get('/regions')
export const createRefund = (data) => api.post('/refunds', data)
export const createGift = (data) => api.post('/gifts', data)
