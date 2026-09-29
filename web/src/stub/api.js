import { createServer, Response } from 'miragejs'

const STUB_ICONS = [
  '/stub-icons/mastodon.png',
  '/stub-icons/bitwarden.png',
  '/stub-icons/matrix.png',
  '/stub-icons/mattermost.png'
]

const rawApps = [
  {
    snapId: 'nextcloud.101',
    id: 'nextcloud',
    name: 'Nextcloud',
    summary: 'Self-hosted productivity platform: files, calendar, contacts and more.',
    version: '30.0.4',
    popularity: 3128
  },
  {
    snapId: 'jellyfin.102',
    id: 'jellyfin',
    name: 'Jellyfin',
    summary: 'Free media server. Stream your movies, music and shows from home.',
    version: '10.9.11',
    popularity: 2104
  },
  {
    snapId: 'actual-budget.115',
    id: 'actual-budget',
    name: 'Actual Budget',
    summary: 'Local-first personal finance and budgeting',
    version: '25.9.0',
    popularity: 118
  },
  {
    snapId: 'games.116',
    id: 'games',
    name: 'Game Hub',
    summary: 'Dedicated game server panel for SteamCMD and Pelican-egg games',
    version: '0.4.1',
    popularity: 96
  },
  {
    snapId: 'bitwarden.103',
    id: 'bitwarden',
    name: 'Bitwarden',
    summary: 'Open-source password manager. Sync your secrets across devices.',
    version: '1.32.7',
    popularity: 1487
  },
  {
    snapId: 'syncthing.104',
    id: 'syncthing',
    name: 'Syncthing',
    summary: 'Continuous file synchronisation between your devices.',
    version: '1.27.10',
    popularity: 1142
  },
  {
    snapId: 'home-assistant.105',
    id: 'home-assistant',
    name: 'Home Assistant',
    summary: 'Open-source home automation. Control your smart devices privately.',
    version: '2025.4.1',
    popularity: 980
  },
  {
    snapId: 'matrix.106',
    id: 'matrix',
    name: 'Matrix',
    summary: 'Decentralized chat server (Synapse). Run your own messaging network.',
    version: '1.118.0',
    popularity: 612
  },
  {
    snapId: 'paperless.107',
    id: 'paperless',
    name: 'Paperless',
    summary: 'Index and archive your scanned documents with OCR and tags.',
    version: '2.13.5',
    popularity: 504
  },
  {
    snapId: 'mastodon.108',
    id: 'mastodon',
    name: 'Mastodon',
    summary: 'Decentralized social network — own your timeline.',
    version: '4.3.1',
    popularity: 388
  },
  {
    snapId: 'collabora.109',
    id: 'collabora',
    name: 'Collabora',
    summary: 'Online office suite for collaborative document editing.',
    version: '24.04',
    popularity: 271
  },
  {
    snapId: 'mattermost.110',
    id: 'mattermost',
    name: 'Mattermost',
    summary: 'Secure team collaboration. Self-hosted Slack alternative.',
    version: '9.11.3',
    popularity: 219
  },
  {
    snapId: 'gogs.111',
    id: 'gogs',
    name: 'Gogs',
    summary: 'Self-hosted Git service. A lightweight alternative to GitHub.',
    version: '0.13.0',
    popularity: 156
  },
  {
    snapId: 'calibre.112',
    id: 'calibre',
    name: 'Calibre',
    summary: 'Manage your e-book library and read from any device.',
    version: '7.21.0',
    popularity: 92
  },
  {
    snapId: 'grocy.113',
    id: 'grocy',
    name: 'Grocy',
    summary: 'ERP for your fridge: groceries, chores and recipes.',
    version: '4.4.0',
    popularity: 41
  },
  {
    snapId: 'newcomer.114',
    id: 'newcomer',
    name: 'Newcomer',
    summary: 'Freshly published app — no devices have checked in yet.',
    version: '0.1.0',
    popularity: 0
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
