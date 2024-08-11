<script lang="ts">
	import { base } from '$app/paths'
	import { playerState, savePlayerState } from '$lib/userSettings.svelte'

	let audioTag = $state<HTMLAudioElement>()
	let paused = $state(false)
	let compact = $state(true)
	let duration = $state(0)
	$effect(() => {
		if (!audioTag) {
			return
		}
		paused = audioTag.paused
		if (audioTag.currentTime === 0) {
			audioTag.currentTime = playerState.currentTime || 0
		}
		audioTag.volume = playerState.volume || 1
		audioTag.playbackRate = playerState.playbackRate || 1
	})
	const formatDuration = (n: number | undefined | null) => {
		if (n === null || n === undefined) {
			return '-:-'
		}
		if (isNaN(n)) {
			return '-:-'
		}
		n = Math.floor(n)
		const m = Math.floor(n / 60)
		const s = n % 60
		return [m, s.toString().padStart(2, '0')].join(':')
	}
</script>

{#snippet audio()}
	{#if playerState.episode?.soundUrl}
		<audio
			src={playerState.episode.soundUrl}
			autoplay
			bind:this={audioTag}
			controls={playerState.nativeControls}
			onvolumechange={(e) => {
				const { volume } = e.currentTarget
				playerState.volume = volume
			}}
			onratechange={(e) => {
				const { playbackRate } = e.currentTarget

				playerState.playbackRate = playbackRate
			}}
			onpause={() => (paused = true)}
			onplay={() => (paused = false)}
			ontimeupdate={(e) => {
				const { currentTime } = e.currentTarget
				duration = e.currentTarget.duration
				if (currentTime === 0) {
					e.currentTarget.currentTime = playerState.currentTime || 0
				}
				if (playerState.currentTime === currentTime) {
					return
				}
				playerState.currentTime = currentTime
				savePlayerState()
			}}
		></audio>
	{/if}
{/snippet}
{#snippet controls(className: string)}
	{#if audioTag}
		<div class="join mx-auto">
			<button
				class="btn join-item {className}"
				onclick={() => {
					if (!audioTag) {
						return
					}
					audioTag.currentTime -= 15
				}}
				>&lt;
			</button>
			<button
				class="btn join-item {className}"
				onclick={() => {
					if (!audioTag) {
						return
					}
					if (audioTag.paused) {
						audioTag.play()
					} else {
						audioTag.pause()
					}
				}}
			>
				{#if paused}
					▶️
				{:else}
					⏸
				{/if}
			</button>
			<button
				class="btn join-item {className}"
				onclick={() => {
					if (!audioTag) {
						return
					}
					audioTag.currentTime += 15
				}}
				>&gt;
			</button>
			<div class="flex items-center">
				{formatDuration(playerState.currentTime)}
				{formatDuration(duration)}
			</div>
		</div>
	{/if}
{/snippet}

<div class:fixed={!compact} class="bottom-0 right-0 z-10">
	{@render audio()}
	{#if playerState?.episode?.soundUrl}
		{#if compact}
			<div class="flex">
				<button onclick={() => (compact = !compact)}>
					<figure class="w-12 h-12">
						{#if playerState.episode.imageUrl}
							<!-- content here -->
							<img
								src={playerState.episode.imageUrl}
								alt="Poster for episode {playerState.episode.title}"
							/>
						{:else if playerState.channel?.imageUrl}
							<img
								src={playerState.channel.imageUrl}
								alt="Poster for channel {playerState.channel.title}"
							/>
						{/if}
					</figure>
				</button>
				<div>
					{@render controls('btn-sm')}
					<div
						title={[
							playerState.episode.title,
							playerState.episode.description,
							'---',
							playerState.channel?.title,
							playerState.channel?.description,
						]
							.filter(Boolean)
							.join('\n\n')}
						class="px-4 text-xs"
					>
						{playerState.episode.title}
					</div>
				</div>
			</div>
		{:else}
			<div
				class="card bg-base-100 h-96 w-72 shadow-xl relative outline-3 outline outline-black/70 rounded-bl-none rounded-tr-none"
			>
				<figure class="relative">
					{#if playerState.episode.imageUrl}
						<!-- content here -->
						<img
							src={playerState.episode.imageUrl}
							alt="Poster for episode {playerState.episode.title}"
						/>
					{:else if playerState.channel?.imageUrl}
						<img
							src={playerState.channel.imageUrl}
							alt="Poster for channel {playerState.channel.title}"
						/>
					{/if}
					<button
						class="w-full absolute -bottom-2 -mt-8 pt-8 group"
						onclick={(e) => {
							const rect = e.currentTarget.getBoundingClientRect()
							const x = e.clientX - rect.left //x position within the element.
							const percent = x / rect.width
							if (!audioTag) {
								return
							}
							audioTag.currentTime = percent * audioTag.duration
							playerState.currentTime = percent * audioTag.duration
						}}
					>
						<progress
							class="w-full progress progress-info h-2 group-hover:h-5 transition-all"
							value={playerState.currentTime}
							max={duration}
						></progress>
					</button>
				</figure>
				<button
					class="btn btn-circle btn-outline glass absolute -left-4 -top-4 text-black"
					onclick={() => (compact = true)}
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						class="h-6 w-6"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</button>
				<div class="card-body">
					<div class="card-title">
						{#if playerState.channel}
							<a href="{base}/channel/{playerState.channel.id}">
								{playerState.channel?.title}
							</a>
						{/if}
					</div>
					<div title={playerState.episode.description}>{playerState.episode.title}</div>
					{@render controls()}
				</div>
			</div>
		{/if}
	{/if}
</div>
