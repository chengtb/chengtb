import { createApp } from 'vue'
import App from './App.vue'
import router from './router/index.js'

// Import Vant components
import {
  Button, NavBar, Sidebar, SidebarItem, Image as VanImage,
  Stepper, Badge, Icon, Popup, PullRefresh, List, Tag,
  Empty, RadioGroup, Radio, Toast, showToast, showSuccessToast
} from 'vant'
import 'vant/lib/index.css'

const app = createApp(App)

app.use(router)
app.use(Button)
app.use(NavBar)
app.use(Sidebar)
app.use(SidebarItem)
app.use(VanImage)
app.use(Stepper)
app.use(Badge)
app.use(Icon)
app.use(Popup)
app.use(PullRefresh)
app.use(List)
app.use(Tag)
app.use(Empty)
app.use(RadioGroup)
app.use(Radio)

app.mount('#app')
