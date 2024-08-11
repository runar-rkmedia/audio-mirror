<script>
	import { base } from '$app/paths'
	import { playerState } from '$lib/userSettings.svelte'
	import '../app.css'
	import Player from './Player.svelte'
	export const ssr = false
</script>

<header>
	<div class="navbar bg-base-100">
		<div class="flex-1">
			<a href="/" class="btn btn-ghost text-xl">Audio-Mirror</a>
		</div>
		<div class="flex-none">
			<Player />
			<div class="dropdown dropdown-end">
				<button class="btn btn-square btn-ghost">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						class="inline-block h-5 w-5 stroke-current"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M5 12h.01M12 12h.01M19 12h.01M6 12a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0zm7 0a1 1 0 11-2 0 1 1 0 012 0z"
						></path>
					</svg>
				</button>

				<ul
					class=" menu menu-sm dropdown-content bg-base-100 rounded-box z-[1] mt-3 w-52 p-2 shadow"
				>
					<li><a href="{base}/settings">Settings</a></li>
					<li><a href="{base}/about">About</a></li>
					<li>
						<a target="_blank" href="https://github.com/runar-rkmedia/audio-mirror">Github</a>
					</li>
				</ul>
			</div>
		</div>
	</div>
</header>
<main class="mx-2 lg:mx-4">
	<slot />
</main>
