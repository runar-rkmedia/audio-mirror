<script lang="ts">
	import userSettings, {
		saveSettings,
		type ErrorMap,
		type UserSettings,
	} from '$lib/userSettings.svelte'

	let errors = $state<ErrorMap<UserSettings> | null>()
</script>

<h1>Settings</h1>

{#snippet err(key: keyof UserSettings)}
	{#if errors?.[key]}
		<!-- content here -->
		{#each errors[key] as err}
			<p>{err.message}</p>
		{/each}
	{/if}
{/snippet}

<form
	onsubmit={async (e) => {
		e.preventDefault()
		const [ok, errs] = await saveSettings()
		errors = errs
		if (ok) {
			alert('saved')
		}
	}}
>
	<input type="text" bind:value={userSettings.preferredPodcastPlayer} />
	{@render err('preferredPodcastPlayer')}
	<button class="btn btn-primary">Save</button>
</form>
