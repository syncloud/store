import { createServer, Response } from 'miragejs'

const STUB_ICONS = [
  '/stub-icons/mastodon.png',
  '/stub-icons/bitwarden.png',
  '/stub-icons/matrix.png',
  '/stub-icons/mattermost.png'
]

const rawApps = [
  {
    snapId: 'mastodon-id',
    name: 'Mastodon',
    summary: 'Decentralized social network — own your timeline.',
    version: '4.3.1'
  },
  {
    snapId: 'bitwarden-id',
    name: 'Bitwarden',
    summary: 'Open-source password manager. Sync your secrets across devices.',
    version: '1.32.7'
  },
  {
    snapId: 'matrix-id',
    name: 'Matrix',
    summary: 'Decentralized chat server (Synapse). Run your own messaging network.',
    version: '1.118.0'
  },
  {
    snapId: 'mattermost-id',
    name: 'Mattermost',
    summary: 'Secure team collaboration. Self-hosted Slack alternative.',
    version: '9.11.3'
  },
  {
    snapId: 'nextcloud-id',
    name: 'Nextcloud',
    summary: 'Self-hosted productivity platform: files, calendar, contacts and more.',
    version: '30.0.4'
  },
  {
    snapId: 'jellyfin-id',
    name: 'Jellyfin',
    summary: 'Free media server. Stream your movies, music and shows from home.',
    version: '10.9.11'
  },
  {
    snapId: 'gogs-id',
    name: 'Gogs',
    summary: 'Self-hosted Git service. A lightweight alternative to GitHub.',
    version: '0.13.0'
  },
  {
    snapId: 'paperless-id',
    name: 'Paperless',
    summary: 'Index and archive your scanned documents with OCR and tags.',
    version: '2.13.5'
  },
  {
    snapId: 'home-assistant-id',
    name: 'Home Assistant',
    summary: 'Open-source home automation. Control your smart devices privately.',
    version: '2025.4.1'
  },
  {
    snapId: 'collabora-id',
    name: 'Collabora',
    summary: 'Online office suite for collaborative document editing.',
    version: '24.04'
  },
  {
    snapId: 'syncthing-id',
    name: 'Syncthing',
    summary: 'Continuous file synchronisation between your devices.',
    version: '1.27.10'
  },
  {
    snapId: 'calibre-id',
    name: 'Calibre',
    summary: 'Manage your e-book library and read from any device.',
    version: '7.21.0'
  },
  {
    snapId: 'grocy-id',
    name: 'Grocy',
    summary: 'ERP for your fridge: groceries, chores and recipes.',
    version: '4.4.0'
  }
]

function pickIcon (snapId) {
  let h = 0
  for (let i = 0; i < snapId.length; i++) h = (h * 31 + snapId.charCodeAt(i)) >>> 0
  return STUB_ICONS[h % STUB_ICONS.length]
}

const apps = rawApps.map(a => ({ ...a, iconUrl: pickIcon(a.snapId) }))

export function mock () {
  createServer({
    routes () {
      this.get('/api/ui/v1/apps', () => {
        return new Response(200, { 'Content-Type': 'application/json' }, apps)
      })

      this.get('/api/ui/v1/version', () => {
        return new Response(200, { 'Content-Type': 'application/json' }, {
          gitSha: 'devstub00000000000000000000000000000000',
          buildNumber: 'dev',
          buildTime: new Date().toISOString()
        })
      })

      this.passthrough()
    }
  })
}
