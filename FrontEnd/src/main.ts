import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { AVPlugin } from "vue-audio-visual"; // 引入插件


import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(AVPlugin) // 使用插件

app.mount('#app')
