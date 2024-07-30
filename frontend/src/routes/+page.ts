import apiClient from '$lib/apiClient'
import type { PageLoad } from './$types'

export const load: PageLoad = async (p) => {
  const channels = await apiClient.getChannels({})

  return {
    channels: channels.toJson(),
  }
}
