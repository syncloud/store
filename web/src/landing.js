import LANDINGS from './landings.json'

const BASE = 'https://syncloud.org/en/'

export function landingUrl (appId) {
  return LANDINGS.includes(appId) ? BASE + appId : null
}
