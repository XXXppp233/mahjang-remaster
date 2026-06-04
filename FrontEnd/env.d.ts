/// <reference types="vite/client" />

// 解决 TS 无法识别 *.vue 导致的 “Cannot find module '*.vue'” 报错
// (Vite 官方模板里通常自带, 这里补回来)
declare module '*.vue' {
	import type { DefineComponent } from 'vue'
	const component: DefineComponent<{}, {}, any>
	export default component
}
