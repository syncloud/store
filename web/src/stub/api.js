import { createServer, Response } from 'miragejs'

const apps = [
  {
    snapId: 'nextcloud-id',
    name: 'Nextcloud',
    summary: 'Self-hosted productivity platform: files, calendar, contacts and more.',
    version: '30.0.4',
    iconUrl: ''
  },
  {
    snapId: 'jellyfin-id',
    name: 'Jellyfin',
    summary: 'Free media server. Stream your movies, music and shows from home.',
    version: '10.9.11',
    iconUrl: ''
  },
  {
    snapId: 'bitwarden-id',
    name: 'Bitwarden',
    summary: 'Open-source password manager. Sync your secrets across devices.',
    version: '1.32.7',
    iconUrl: ''
  },
  {
    snapId: 'matrix-id',
    name: 'Matrix',
    summary: 'Decentralized chat server (Synapse). Run your own messaging network.',
    version: '1.118.0',
    iconUrl: ''
  },
  {
    snapId: 'gogs-id',
    name: 'Gogs',
    summary: 'Self-hosted Git service. A lightweight alternative to GitHub.',
    version: '0.13.0',
    iconUrl: ''
  },
  {
    snapId: 'paperless-id',
    name: 'Paperless',
    summary: 'Index and archive your scanned documents with OCR and tags.',
    version: '2.13.5',
    iconUrl: ''
  },
  {
    snapId: 'home-assistant-id',
    name: 'Home Assistant',
    summary: 'Open-source home automation. Control your smart devices privately.',
    version: '2025.4.1',
    iconUrl: ''
  },
  {
    snapId: 'mattermost-id',
    name: 'Mattermost',
    summary: 'Secure team collaboration. Self-hosted Slack alternative.',
    version: '9.11.3',
    iconUrl: ''
  },
  {
    snapId: 'collabora-id',
    name: 'Collabora',
    summary: 'Online office suite for collaborative document editing.',
    version: '24.04',
    iconUrl: ''
  },
  {
    snapId: 'syncthing-id',
    name: 'Syncthing',
    summary: 'Continuous file synchronisation between your devices.',
    version: '1.27.10',
    iconUrl: ''
  },
  {
    snapId: 'calibre-id',
    name: 'Calibre',
    summary: 'Manage your e-book library and read from any device.',
    version: '7.21.0',
    iconUrl: ''
  },
  {
    snapId: 'grocy-id',
    name: 'Grocy',
    summary: 'ERP for your fridge: groceries, chores and recipes.',
    version: '4.4.0',
    iconUrl: ''
  }
]

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
