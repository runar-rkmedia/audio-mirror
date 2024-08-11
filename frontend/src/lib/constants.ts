import type { ServerLoadEvent } from '@sveltejs/kit'

export const rootUrl = (p: Pick<ServerLoadEvent, 'url'>, path: string = '') =>
  p.url.origin.replace('http://', 'https://') + path
