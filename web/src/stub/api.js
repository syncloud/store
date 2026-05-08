import { createServer, Response } from 'miragejs'

const apps = [
  {
    'snap-id': 'nextcloud-id',
    name: 'Nextcloud',
    summary: 'Self-hosted productivity platform: files, calendar, contacts and more.',
    version: '30.0.4',
    icon: 'https://apps.syncloud.org/apps/nextcloud/icon.png'
  },
  {
    'snap-id': 'jellyfin-id',
    name: 'Jellyfin',
    summary: 'Free media server. Stream your movies, music and shows from home.',
    version: '10.9.11',
    icon: 'https://apps.syncloud.org/apps/jellyfin/icon.png'
  },
  {
    'snap-id': 'bitwarden-id',
    name: 'Bitwarden',
    summary: 'Open-source password manager. Sync your secrets across devices.',
    version: '1.32.7',
    icon: 'https://apps.syncloud.org/apps/bitwarden/icon.png'
  },
  {
    'snap-id': 'matrix-id',
    name: 'Matrix',
    summary: 'Decentralized chat server (Synapse). Run your own messaging network.',
    version: '1.118.0',
    icon: 'https://apps.syncloud.org/apps/matrix/icon.png'
  },
  {
    'snap-id': 'gogs-id',
    name: 'Gogs',
    summary: 'Self-hosted Git service. A lightweight alternative to GitHub.',
    version: '0.13.0',
    icon: 'https://apps.syncloud.org/apps/gogs/icon.png'
  },
  {
    'snap-id': 'paperless-id',
    name: 'Paperless',
    summary: 'Index and archive your scanned documents with OCR and tags.',
    version: '2.13.5',
    icon: 'https://apps.syncloud.org/apps/paperless/icon.png'
  },
  {
    'snap-id': 'home-assistant-id',
    name: 'Home Assistant',
    summary: 'Open-source home automation. Control your smart devices privately.',
    version: '2025.4.1',
    icon: 'https://apps.syncloud.org/apps/home-assistant/icon.png'
  },
  {
    'snap-id': 'mattermost-id',
    name: 'Mattermost',
    summary: 'Secure team collaboration. Self-hosted Slack alternative.',
    version: '9.11.3',
    icon: 'https://apps.syncloud.org/apps/mattermost/icon.png'
  },
  {
    'snap-id': 'collabora-id',
    name: 'Collabora',
    summary: 'Online office suite for collaborative document editing.',
    version: '24.04',
    icon: 'https://apps.syncloud.org/apps/collabora/icon.png'
  },
  {
    'snap-id': 'syncthing-id',
    name: 'Syncthing',
    summary: 'Continuous file synchronisation between your devices.',
    version: '1.27.10',
    icon: 'https://apps.syncloud.org/apps/syncthing/icon.png'
  },
  {
    'snap-id': 'calibre-id',
    name: 'Calibre',
    summary: 'Manage your e-book library and read from any device.',
    version: '7.21.0',
    icon: 'https://apps.syncloud.org/apps/calibre/icon.png'
  },
  {
    'snap-id': 'grocy-id',
    name: 'Grocy',
    summary: 'ERP for your fridge: groceries, chores and recipes.',
    version: '4.4.0',
    icon: 'https://apps.syncloud.org/apps/grocy/icon.png'
  }
]

function toResult (a) {
  return {
    'snap-id': a['snap-id'],
    name: a.name,
    revision: { revision: 1 },
    snap: {
      'snap-id': a['snap-id'],
      name: a.name,
      summary: a.summary,
      version: a.version,
      type: 'app',
      architectures: ['amd64', 'arm64', 'arm'],
      revision: 1,
      media: a.icon ? [{ type: 'icon', url: a.icon, width: 256, height: 256 }] : []
    }
  }
}

export function mock () {
  createServer({
    routes () {
      this.get('/v2/snaps/find', (_schema, request) => {
        const q = (request.queryParams.q || '').toLowerCase()
        const results = apps
          .filter(a => !q || a.name.toLowerCase().includes(q) || a.summary.toLowerCase().includes(q))
          .map(toResult)
        return new Response(200, { 'Content-Type': 'application/json' }, {
          results,
          'error-list': []
        })
      })

      this.get('/v2/snaps/info/:name', (_schema, request) => {
        const a = apps.find(x => x.name.toLowerCase() === request.params.name.toLowerCase())
        if (!a) return new Response(404, {}, 'not found')
        return new Response(200, { 'Content-Type': 'application/json' }, {
          'snap-id': a['snap-id'],
          name: a.name,
          snap: toResult(a).snap,
          'channel-map': []
        })
      })

      this.passthrough()
    }
  })
}
