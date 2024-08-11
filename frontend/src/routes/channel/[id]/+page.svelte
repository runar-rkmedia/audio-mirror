<script lang="ts">
	import { playerState } from '$lib/userSettings.svelte'
	import { fade } from 'svelte/transition'
	import PodcastUrls from '../../PodcastUrls.svelte'

	const { data } = $props()
	const { channel, episodes } = data
</script>

{#if !channel}
	<h1>Oh no</h1>
	Could not find what you were looking for.
{:else}
	<div
		class="flex gap-8 justify-center bg-base-300 rounded-4xl p-8"
		transition:fade={{ delay: 2500, duration: 3000 }}
	>
		<div>
			<h1 class="text-4xl mb-8">{channel.title}</h1>
			<p class="max-w-[44rem]">{channel.description}</p>
			<div class="stats shadow">
				<div class="stat">
					<div class="stat-title">Episodes</div>
					<div class="stat-value">{channel.episodeCount || episodes.length}</div>
				</div>
				<div class="stat">
					<div class="stat-title">Last episode date:</div>
					<div class="stat-value">TODO</div>
				</div>
			</div>
		</div>
		<div class="">
			<img
				class="min-w-96 w-96 h-96 ml-auto glass"
				src={channel.imageUrl}
				alt="Poster for {channel.title}"
			/>
			<div class="join join-vertical flex">
				<PodcastUrls feedUrl={channel.feedUrl} buttonClass="join-item" />
			</div>
		</div>
	</div>
	{#if episodes}
		<div class="overflow-x-auto">
			<table class="table">
				<!-- head -->
				<thead>
					<tr>
						<th>Name</th>
					</tr>
				</thead>
				<tbody>
					{#each episodes as epi}
						<!-- row 1 -->
						<tr>
							<td>
								<div class="flex items-center gap-3">
									<div class="avatar">
										<div class="mask mask-squircle h-24 w-24">
											<img src={epi.imageUrl || channel.imageUrl} alt="Poster for {epi.title}" />
										</div>
									</div>
									<div>
										<div class="font-bold">{epi.title}</div>
										<div class="text-sm opacity-50">{epi.description}</div>
										<button
											onclick={() => {
												if (playerState.episode?.soundUrl !== epi.soundUrl) {
													playerState.currentTime = 0
												}
												playerState.episode = epi
												playerState.channel = channel
												playerState.open = true
											}}>Play</button
										>
									</div>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
{/if}
