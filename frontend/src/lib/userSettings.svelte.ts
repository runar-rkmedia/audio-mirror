import { browser } from '$app/environment'
import { objectKeys } from 'simplytyped'
import type { Channel, Episode } from '../gen/api/v1/pods_pb'

type PreferredPodcastPlayer = 'overcast' | 'apple-podcasts' | 'custom' | ''

export type UserSettings = {
  customPodcastPlayerUrl: string
  preferredPodcastPlayer: PreferredPodcastPlayer
}

const defaultUserSettings = (): UserSettings => {
  return {
    preferredPodcastPlayer: '',
    customPodcastPlayerUrl: '',
  }
}
const loadSettings = (prevSettings?: UserSettings): UserSettings => {
  if (!browser) {
    return prevSettings || defaultUserSettings()
  }
  const raw = window.localStorage.getItem('aum_settings')
  if (!raw) {
    return prevSettings || defaultUserSettings()
  }
  try {
    const defaults = defaultUserSettings()
    const parsed: UserSettings = {
      ...defaults,
      ...JSON.parse(raw),
    }
    const errors = validateSettings(parsed)
    if (!errors) {
      return parsed
    }

    for (const key of objectKeys(errors)) {
      if (!errors[key].length) {
        continue
      }
      delete parsed[key]
    }
    return {
      ...defaults,
      ...parsed,
    }
  } catch (err) {
    console.error(err)
    return prevSettings || defaultUserSettings()
  }
}

export type Error = { message: string; shadow?: boolean }
export type ErrorMap<T> = Record<keyof T, Error[]>
const validateSettings = (settings = userSettings) => {
  const errors: ErrorMap<UserSettings> = {
    preferredPodcastPlayer: [],
    customPodcastPlayerUrl: [],
  }
  for (const k of objectKeys(settings)) {
    switch (k) {
      case 'customPodcastPlayerUrl':
        break
      case 'preferredPodcastPlayer':
        switch (settings.preferredPodcastPlayer) {
          case 'apple-podcasts':
          case 'overcast':
            break
          case 'custom':
            if (!settings.customPodcastPlayerUrl) {
              errors.customPodcastPlayerUrl.push({
                message: 'must be set when preferredPodcastPlayer is set to custom',
              })
              errors.preferredPodcastPlayer.push({
                shadow: true,
                message:
                  'customPodcastPlayerUrl must be set when preferredPodcastPlayer is set to custom',
              })
            }
            break
          default:
            errors.preferredPodcastPlayer.push({
              message: `invalid value: ${settings.preferredPodcastPlayer}. Must be one of: ['apple-podcasts', 'overcast', 'custom']`,
            })
        }
        break
      default:
        delete settings[k]
    }
  }
  const allErrors = Object.values(errors).flat()
  if (!allErrors.length) {
    return null
  }
  return errors
}
const userSettings = $state<UserSettings>(loadSettings())
type PlayerState = Partial<{
  episode: Episode
  channel: Channel
  currentTime: number
  volume: number
  playbackRate: number
  open: boolean
  nativeControls: boolean
}>

const loadPlayerState = (): PlayerState => {
  if (!browser) {
    return {}
  }
  return JSON.parse(window.localStorage.getItem('aum_player') || '{}') as PlayerState
}
export const playerState = $state<PlayerState>(loadPlayerState())
export const savePlayerState = () => {
  const json = JSON.stringify(playerState)
  window.localStorage.setItem('aum_player', json)
}

export const saveSettings = async (settings = userSettings) => {
  const errors = validateSettings(settings)
  if (errors) {
    return [false, errors] as const
  }
  const json = JSON.stringify(settings)
  localStorage.setItem('aum_settings', json)
  return [true, null] as const
}

export default userSettings
