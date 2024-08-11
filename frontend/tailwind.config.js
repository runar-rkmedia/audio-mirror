import daisyui from 'daisyui'
import lineclamp from '@tailwindcss/line-clamp'
/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {},
	},
	plugins: [daisyui, lineclamp],
	daisyui: {
		themes: ['night'],
		logs: true,
	},
}
