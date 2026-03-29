<template>
  <div>
    <h1 class="page-title">菜单管理</h1>
    
    <!-- Categories -->
    <div class="section">
      <div class="section-header">
        <h2>菜品分类</h2>
        <button @click="showAddCat = true" class="add-btn">+ 添加分类</button>
      </div>
      <div class="cat-list">
        <div v-for="cat in categories" :key="cat.id" class="cat-row">
          <span class="cat-icon">{{ cat.icon }}</span>
          <span class="cat-name">{{ cat.name }}</span>
          <span class="cat-order">排序：{{ cat.sort_order }}</span>
          <div class="row-actions">
            <button @click="editCat(cat)" class="btn-edit">编辑</button>
            <button @click="deleteCat(cat.id)" class="btn-delete">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Dishes -->
    <div class="section">
      <div class="section-header">
        <h2>菜品列表</h2>
        <button @click="showAddDish = true" class="add-btn">+ 添加菜品</button>
      </div>
      <div class="filter-bar">
        <select v-model="filterCatId" @change="loadDishes">
          <option value="">全部分类</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </div>
      <table class="dishes-table">
        <thead>
          <tr><th>名称</th><th>分类</th><th>价格</th><th>描述</th><th>状态</th><th>操作</th></tr>
        </thead>
        <tbody>
          <tr v-for="dish in dishes" :key="dish.id">
            <td>{{ dish.name }}</td>
            <td>{{ dish.category?.name }}</td>
            <td>¥{{ dish.price.toFixed(2) }}</td>
            <td class="desc-cell">{{ dish.description }}</td>
            <td><span :class="['avail-badge', dish.available ? 'on' : 'off']">{{ dish.available ? '在售' : '下架' }}</span></td>
            <td class="actions">
              <button @click="editDish(dish)" class="btn-edit">编辑</button>
              <button @click="deleteDish(dish.id)" class="btn-delete">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add/Edit Category Modal -->
    <div v-if="showAddCat || editingCat" class="modal-overlay" @click.self="closeCatModal">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ editingCat ? '编辑分类' : '添加分类' }}</h2>
          <button @click="closeCatModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group"><label>名称</label><input v-model="catForm.name" placeholder="分类名称" /></div>
          <div class="form-group"><label>图标</label><input v-model="catForm.icon" placeholder="emoji图标" /></div>
          <div class="form-group"><label>排序</label><input v-model.number="catForm.sort_order" type="number" /></div>
          <button @click="saveCat" class="submit-btn">保存</button>
        </div>
      </div>
    </div>

    <!-- Add/Edit Dish Modal -->
    <div v-if="showAddDish || editingDish" class="modal-overlay" @click.self="closeDishModal">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ editingDish ? '编辑菜品' : '添加菜品' }}</h2>
          <button @click="closeDishModal" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group"><label>名称</label><input v-model="dishForm.name" /></div>
          <div class="form-group"><label>分类</label>
            <select v-model.number="dishForm.category_id">
              <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <div class="form-group"><label>价格</label><input v-model.number="dishForm.price" type="number" step="0.01" /></div>
          <div class="form-group"><label>描述</label><textarea v-model="dishForm.description" rows="3"></textarea></div>
          <div class="form-group">
            <label>
              <input type="checkbox" v-model="dishForm.available" /> 在售
            </label>
          </div>
          <button @click="saveDish" class="submit-btn">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const categories = ref([])
const dishes = ref([])
const filterCatId = ref('')
const showAddCat = ref(false)
const showAddDish = ref(false)
const editingCat = ref(null)
const editingDish = ref(null)
const catForm = ref({ name: '', icon: '', sort_order: 1 })
const dishForm = ref({ name: '', category_id: null, price: 0, description: '', available: true, sort_order: 1 })

async function loadCategories() {
  const res = await axios.get('/api/categories')
  categories.value = res.data
}

async function loadDishes() {
  const url = filterCatId.value ? `/api/dishes?category_id=${filterCatId.value}` : '/api/dishes'
  const res = await axios.get(url)
  dishes.value = res.data
}

function editCat(cat) {
  editingCat.value = cat
  catForm.value = { ...cat }
}

function closeCatModal() {
  showAddCat.value = false
  editingCat.value = null
  catForm.value = { name: '', icon: '', sort_order: 1 }
}

async function saveCat() {
  if (editingCat.value) {
    await axios.put(`/api/categories/${editingCat.value.id}`, catForm.value)
  } else {
    await axios.post('/api/categories', catForm.value)
  }
  closeCatModal()
  loadCategories()
}

async function deleteCat(id) {
  if (confirm('确定删除此分类？')) {
    await axios.delete(`/api/categories/${id}`)
    loadCategories()
  }
}

function editDish(dish) {
  editingDish.value = dish
  dishForm.value = { ...dish }
}

function closeDishModal() {
  showAddDish.value = false
  editingDish.value = null
  dishForm.value = { name: '', category_id: null, price: 0, description: '', available: true, sort_order: 1 }
}

async function saveDish() {
  if (editingDish.value) {
    await axios.put(`/api/dishes/${editingDish.value.id}`, dishForm.value)
  } else {
    await axios.post('/api/dishes', dishForm.value)
  }
  closeDishModal()
  loadDishes()
}

async function deleteDish(id) {
  if (confirm('确定删除此菜品？')) {
    await axios.delete(`/api/dishes/${id}`)
    loadDishes()
  }
}

onMounted(() => {
  loadCategories()
  loadDishes()
})
</script>

<style scoped>
.page-title { font-size: 24px; font-weight: bold; margin-bottom: 24px; }
.section { background: white; border-radius: 12px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.section-header h2 { font-size: 18px; }
.add-btn { background: #e74c3c; color: white; border: none; padding: 8px 16px; border-radius: 8px; }
.cat-list { display: flex; flex-direction: column; gap: 8px; }
.cat-row { display: flex; align-items: center; gap: 12px; padding: 10px; background: #f8f9fa; border-radius: 8px; }
.cat-icon { font-size: 20px; }
.cat-name { font-weight: 600; flex: 1; }
.cat-order { color: #666; font-size: 13px; }
.row-actions { display: flex; gap: 8px; }
.btn-edit { background: #3498db; color: white; border: none; padding: 6px 12px; border-radius: 6px; font-size: 13px; }
.btn-delete { background: #e74c3c; color: white; border: none; padding: 6px 12px; border-radius: 6px; font-size: 13px; }
.filter-bar { margin-bottom: 12px; }
.filter-bar select { padding: 8px 12px; border: 1px solid #ddd; border-radius: 8px; }
.dishes-table { width: 100%; border-collapse: collapse; }
.dishes-table th, .dishes-table td { padding: 12px; text-align: left; border-bottom: 1px solid #f0f0f0; }
.dishes-table th { background: #fafafa; font-weight: 600; color: #555; }
.desc-cell { max-width: 200px; color: #888; font-size: 13px; }
.avail-badge { padding: 3px 8px; border-radius: 10px; font-size: 12px; }
.avail-badge.on { background: #d4edda; color: #155724; }
.avail-badge.off { background: #f8d7da; color: #721c24; }
.actions { display: flex; gap: 8px; }
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; }
.modal { background: white; border-radius: 16px; width: 90%; max-width: 480px; }
.modal-header { padding: 20px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; }
.modal-header h2 { font-size: 18px; }
.close-btn { background: none; border: none; font-size: 20px; color: #666; }
.modal-body { padding: 20px; display: flex; flex-direction: column; gap: 16px; }
.form-group { display: flex; flex-direction: column; gap: 4px; }
.form-group label { font-size: 14px; color: #555; }
.form-group input, .form-group select, .form-group textarea {
  padding: 8px 12px; border: 1px solid #ddd; border-radius: 8px; font-size: 14px;
}
.submit-btn { background: #e74c3c; color: white; border: none; padding: 12px; border-radius: 8px; font-size: 16px; }
</style>
