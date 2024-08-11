import apiClient from '$lib/apiClient'
import { rootUrl } from '$lib/constants'
import type { PageServerLoad } from './$types'

export const load: PageServerLoad = async (p) => {
  const channels = await apiClient.getChannels({})
  return {
    channels: (channels.toJson() as any as typeof channels).channels,
    feedUrlPrefix: rootUrl(p, '/feed'),
  }
}
