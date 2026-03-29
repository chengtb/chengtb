import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export const useTableStore = defineStore('table', () => {
  const tableId = ref(localStorage.getItem('tableId') || null)
  const tableInfo = ref(null)

  async function setTable(id) {
    tableId.value = id
    localStorage.setItem('tableId', id)
    try {
      const res = await axios.get(`/api/tables/${id}`)
      tableInfo.value = res.data
    } catch (e) {
      console.error('Failed to load table info', e)
    }
  }

  return { tableId, tableInfo, setTable }
})
