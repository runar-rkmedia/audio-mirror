<script lang="ts">
	import userSettings from '$lib/userSettings.svelte'

	let { feedUrl = '', buttonClass }: { feedUrl: string; buttonClass?: string } = $props()

	const feedSchemedUrl = feedUrl.replace('https', 'podcast')
	const overcastUrl = `overcast://x-callback-url/add?url=${feedUrl}`
	const preferred = userSettings.preferredPodcastPlayer
</script>

{#if !feedUrl}
	<div class="alert alert-error">Missing feedurl</div>
{:else}
	<button
		class="btn btn-primary {buttonClass}"
		onclick={async () => {
			await navigator.clipboard.writeText(feedUrl)
			alert(
				'Feed-url copied to clipboard.\n\nPaste it into your favourite podcast-app to listen. (For instance OverCast)',
			)
		}}>Copy feed link</button
	>
	<a class="btn btn-secondary {buttonClass}" target="_blank" href={feedUrl}>Direct link</a>

	{#if !preferred || preferred === 'overcast'}
		<a class="btn btn-ghost {buttonClass}" href={overcastUrl}>
			<img src="https://overcast.fm/img/logo.svg?3" alt="Overcast logo" class="w-6 h-6" />
			Add to Overcast
		</a>
	{/if}

	{#if !preferred || preferred === 'apple-podcasts'}
		<a class="btn btn-ghost {buttonClass}" href={feedSchemedUrl}>
			<img
				src="https://marketing.services.apple/api/storage/images/64dbe2a8587a700007115373/en-us-large@2x.png"
				alt="Overcast logo"
				class="w-6 h-6"
			/>
			Add to Apple Podcast</a
		>
	{/if}
{/if}
