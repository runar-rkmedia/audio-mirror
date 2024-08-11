import apiClient from '$lib/apiClient'
import type { PageServerLoad } from './$types'

export const load: PageServerLoad = async (p) => {
  const channel = await apiClient.getChannel({
    id: p.params.id,
  })
  const out = channel.toJson() as typeof channel
  return {
    channel: out.channel!,
    episodes: out.episodes,
  }
}
