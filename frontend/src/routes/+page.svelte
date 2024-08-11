<script lang="ts">
	import { browser } from '$app/environment'
	import { fly } from 'svelte/transition'
	import PodcastUrls from './PodcastUrls.svelte'
	import { base } from '$app/paths'

	let { data } = $props()
	let searchInput = $state('')
	let filteredChannesl = $derived(
		data.channels.filter((ch) =>
			searchInput == '' ? true : ch.title.toLowerCase().includes(searchInput.toLowerCase()),
		),
	)
	// let view = $state('side' as 'side' | 'grid')
</script>

<hgroup>
	<h1 class="text-4xl text-blue-400">Welcome to Audio-Mirror</h1>
	<p>Where any podcast is welcome, whether they like it or not!</p>
</hgroup>
{#if data.channels.length}
	<div>
		<label class="input input-bordered flex items-center gap-2">
			<input type="text" class="grow" placeholder="Search" bind:value={searchInput} />
			<svg
				xmlns="http://www.w3.org/2000/svg"
				viewBox="0 0 16 16"
				fill="currentColor"
				class="h-4 w-4 opacity-70"
			>
				<path
					fill-rule="evenodd"
					d="M9.965 11.026a5 5 0 1 1 1.06-1.06l2.755 2.754a.75.75 0 1 1-1.06 1.06l-2.755-2.754ZM10.5 7a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0Z"
					clip-rule="evenodd"
				/>
			</svg>
		</label>
		{#if filteredChannesl.length < data.channels.length}
			{filteredChannesl.length}/{data.channels.length} channels found matching {searchInput}
		{:else}
			{data.channels.length} channels listed
		{/if}
	</div>
	{#if filteredChannesl.length}
		<div class="grid gap-8 grid-cols-1 lg:grid-cols-1 md:grid-cols-2">
			{#each filteredChannesl as ch (ch.id)}
				{@const feedUrl = `${data.feedUrlPrefix}/${ch.id}`}
				<article class="group card glass lg:card-side lg:h-56" in:fly out:fly>
					<a href="{base}/channel/{ch.id}">
						<figure class="lg:w-56 lg:min-w-56">
							<img
								src={ch.imageUrl}
								alt={ch.title}
								class="md:w-40 md:h-40 w-96 h-96 lg:h-56 lg:w-56 object-contain"
							/>
						</figure>
					</a>
					<div class="card-body">
						<h2 class="card-title">
							<a href="{base}/channel/{ch.id}">
								{ch.title}
							</a>
						</h2>
						<p title={ch.description} class="max-h-24 overflow-y-auto">
							{ch.description}
						</p>
						<div class="card-actions justify-end">
							<PodcastUrls {feedUrl} />
						</div>
					</div>
				</article>
			{/each}
		</div>
	{/if}
{/if}
