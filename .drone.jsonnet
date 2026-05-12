local name = "syncloud-store";
local go = "1.23";
local playwright = "v1.48.2-jammy";
local docker_image = "syncloud/store";
local debian = "bookworm-slim";
local platform = "26.04.10";
local version = "${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}";
local image_tag = docker_image + ":" + version;

local grafanaDatasources = |||
    apiVersion: 1
    datasources:
      - name: vm
        type: prometheus
        access: proxy
        url: http://vm:8428
        isDefault: true
        editable: false
|||;
local grafanaDashboards = |||
    apiVersion: 1
    providers:
      - name: store
        type: file
        disableDeletion: true
        updateIntervalSeconds: 10
        options:
          path: /var/lib/grafana/dashboards
          foldersFromFilesStructure: false
|||;
local grafanaPopularity = |||
    {
      "uid": "popularity",
      "title": "Store Popularity",
      "schemaVersion": 39,
      "version": 1,
      "refresh": "5s",
      "time": { "from": "now-5m", "to": "now" },
      "timepicker": {},
      "panels": [
        {
          "id": 1,
          "type": "stat",
          "title": "Unique devices",
          "gridPos": { "x": 0, "y": 0, "w": 6, "h": 5 },
          "datasource": { "type": "prometheus", "uid": "vm" },
          "targets": [
            { "expr": "store_popularity_devices_unique", "refId": "A", "instant": true }
          ],
          "fieldConfig": {
            "defaults": {
              "color": { "mode": "thresholds" },
              "thresholds": {
                "mode": "absolute",
                "steps": [
                  { "color": "blue", "value": null },
                  { "color": "green", "value": 1 }
                ]
              }
            }
          },
          "options": { "reduceOptions": { "calcs": ["lastNotNull"] }, "colorMode": "value", "graphMode": "none" }
        },
        {
          "id": 2,
          "type": "bargauge",
          "title": "Active devices per snap",
          "gridPos": { "x": 6, "y": 0, "w": 18, "h": 5 },
          "datasource": { "type": "prometheus", "uid": "vm" },
          "targets": [
            { "expr": "store_popularity_devices_active", "refId": "A", "legendFormat": "{{snap}}", "instant": true }
          ],
          "options": { "orientation": "horizontal", "displayMode": "gradient", "showUnfilled": true },
          "fieldConfig": {
            "defaults": {
              "color": { "mode": "continuous-GrYlRd" },
              "thresholds": { "mode": "absolute", "steps": [ { "color": "blue", "value": null } ] }
            }
          }
        },
        {
          "id": 3,
          "type": "timeseries",
          "title": "Record events rate (per minute, by snap)",
          "gridPos": { "x": 0, "y": 5, "w": 24, "h": 10 },
          "datasource": { "type": "prometheus", "uid": "vm" },
          "targets": [
            { "expr": "sum by (snap) (rate(store_popularity_record_total[1m])) * 60", "refId": "A", "legendFormat": "{{snap}}" }
          ],
          "fieldConfig": {
            "defaults": {
              "custom": { "drawStyle": "line", "lineInterpolation": "linear", "lineWidth": 2, "fillOpacity": 10, "pointSize": 5, "showPoints": "auto" },
              "color": { "mode": "palette-classic" },
              "unit": "short"
            }
          },
          "options": { "legend": { "displayMode": "list", "placement": "bottom" }, "tooltip": { "mode": "multi" } }
        },
        {
          "id": 4,
          "type": "bargauge",
          "title": "Total record events per snap (cumulative)",
          "gridPos": { "x": 0, "y": 15, "w": 24, "h": 8 },
          "datasource": { "type": "prometheus", "uid": "vm" },
          "targets": [
            { "expr": "store_popularity_record_total", "refId": "A", "legendFormat": "{{snap}}", "instant": true }
          ],
          "options": { "orientation": "horizontal", "displayMode": "gradient", "showUnfilled": true },
          "fieldConfig": {
            "defaults": {
              "color": { "mode": "continuous-BlPu" },
              "thresholds": { "mode": "absolute", "steps": [ { "color": "blue", "value": null } ] }
            }
          }
        }
      ]
    }
|||;

local build(arch) = {
    kind: "pipeline",
    name: arch,

    trigger: {
        event: ["push", "tag"],
    },

    platform: {
        os: "linux",
        arch: arch
    },
    steps: [
        {
            name: "version",
            image: "debian:" + debian,
            commands: [
                "echo $DRONE_BUILD_NUMBER > version"
            ]
        },
    ] + (if arch == "amd64" then [
        {
            name: "web build",
            image: "node:20-bookworm-slim",
            commands: [
              "bash web/build.sh",
            ]
        },
    ] else []) + [
        {
            name: "unit test",
            image: "golang:" + go,
            commands: [
                "./unit-test.sh",
            ]
        },
        {
            name: "build store",
            image: "golang:" + go,
            commands: [
                "VERSION=$(cat version)",
                "./build.sh $VERSION " + arch
            ]
        },
        {
            name: "build apps",
            image: "debian:" + debian,
            commands: [
              "apt update && apt install -y squashfs-tools",
              "./test/build-apps.sh",
              "./test/publish.sh " + arch
            ]
        },
        {
            name: "build test",
            image: "golang:" + go,
            commands: [
              "./test/build-tests.sh",
            ]
        },
        {
            name: "test",
            image: "debian:" + debian,
            commands: [
              "VERSION=$(cat version)",
              "./test/test.sh"
            ]
        },
        {
            name: "metrics",
            image: "debian:" + debian,
            commands: [
              "./ci/metrics-verify.sh",
            ],
        },
    ] + (if arch == "amd64" then [
        {
            name: "docker",
            image: "plugins/docker:20.18",
            settings: {
                repo: docker_image,
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                tags: [
                    version,
                    "${DRONE_BRANCH}",
                ],
            },
            when: {
                event: ["push", "tag"],
            },
        },
        {
            name: "docker latest",
            image: "plugins/docker:20.18",
            settings: {
                repo: docker_image,
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                tags: ["latest"],
            },
            when: {
                event: ["push"],
                branch: ["stable"],
            },
        },
        {
            name: "deploy test",
            image: "debian:" + debian,
            environment: {
                DEPLOY_HOST: "api.store.test",
                DEPLOY_USER: "root",
                DEPLOY_URL: "http://api.store.test",
            },
            commands: [
                "./ci/test-init.sh",
                "./ci/deploy-prepare.sh test",
                "./ci/deploy-run.sh test " + image_tag,
                "./ci/deploy-verify.sh test",
            ],
            when: {
                event: ["push", "tag"],
            },
        },
        {
            name: "web e2e",
            image: "mcr.microsoft.com/playwright:" + playwright,
            environment: {
                PLAYWRIGHT_BASE_URL: "http://api.store.test",
            },
            commands: [
                "bash web/e2e/run.sh",
            ],
            when: {
                event: ["push", "tag"],
            },
        },
        {
            name: "deploy uat",
            image: "debian:" + debian,
            environment: {
                DEPLOY_HOST: { from_secret: "uat_deploy_host" },
                DEPLOY_USER: { from_secret: "uat_deploy_user" },
                DEPLOY_KEY: { from_secret: "uat_deploy_key" },
                DEPLOY_URL: { from_secret: "uat_deploy_url" },
            },
            commands: [
                "./ci/deploy-prepare.sh uat",
                "./ci/deploy-run.sh uat " + image_tag,
                "./ci/deploy-verify.sh uat",
            ],
            when: { event: ["push"] },
        },
        {
            name: "deploy prod",
            image: "debian:" + debian,
            environment: {
                DEPLOY_HOST: { from_secret: "prod_deploy_host" },
                DEPLOY_USER: { from_secret: "prod_deploy_user" },
                DEPLOY_KEY: { from_secret: "prod_deploy_key" },
                DEPLOY_URL: { from_secret: "prod_deploy_url" },
            },
            commands: [
                "./ci/deploy-prepare.sh prod",
                "./ci/deploy-run.sh prod " + image_tag,
                "./ci/deploy-verify.sh prod",
            ],
            when: { event: ["push"], branch: ["stable"] },
        },
    ] else []) + [
        {
            name: "artifact",
            image: "appleboy/drone-scp:1.6.4",
            settings: {
                host: {
                    from_secret: "artifact_host"
                },
                username: "artifact",
                key: {
                    from_secret: "artifact_key"
                },
                timeout: "2m",
                command_timeout: "2m",
                target: "/home/artifact/repo/" + name + "/${DRONE_BUILD_NUMBER}-" + arch,
                source: [
                    "test/*.snap",
                    "out/*",
                    "test/artifacts/*",
                    "artifact/*",
                ]
            },
            when: {
              status: [ "failure", "success" ]
            }
        },
        {
            name: "publish to github",
            image: "plugins/github-release:1.0.0",
            settings: {
                api_key: {
                    from_secret: "github_token"
                },
                files: "out/*",
                overwrite: true,
                file_exists: "overwrite"
            },
            when: {
                event: [ "tag" ]
            }
        },
    ],
    services:
    [
        {
            name: "device",
            image: "syncloud/bootstrap-bookworm-" + arch + ":" + platform,
            privileged: true,
            volumes: [
                {
                    name: "dbus",
                    path: "/var/run/dbus"
                },
                {
                    name: "dev",
                    path: "/dev"
                }
            ]
        },
        {
            name: "api.store.test",
            image: "syncloud/bootstrap-bookworm-" + arch + ":" + platform,
            privileged: true,
            volumes: [
                {
                    name: "dbus",
                    path: "/var/run/dbus"
                },
                {
                    name: "dev",
                    path: "/dev"
                }
            ]
        },
        {
            name: "apps.syncloud.org",
            image: "syncloud/bootstrap-bookworm-" + arch + ":" + platform,
            privileged: true,
            volumes: [
                {
                    name: "dbus",
                    path: "/var/run/dbus"
                },
                {
                    name: "dev",
                    path: "/dev"
                }
            ]
        },
        {
            name: "vm",
            image: "victoriametrics/victoria-metrics:v1.110.0",
            command: ["-search.latencyOffset=0s"],
        },
        {
            name: "grafana",
            image: "grafana/grafana:11.3.0",
            environment: {
                GF_AUTH_ANONYMOUS_ENABLED: "true",
                GF_AUTH_ANONYMOUS_ORG_ROLE: "Admin",
                GF_AUTH_DISABLE_LOGIN_FORM: "true",
                GF_SECURITY_ADMIN_PASSWORD: "admin",
                DS_YML: grafanaDatasources,
                DB_YML: grafanaDashboards,
                DASH_JSON: grafanaPopularity,
            },
            entrypoint: ["/bin/sh", "-c"],
            command: [
                "mkdir -p /var/lib/grafana/dashboards && "
                + "printf '%s' \"$DS_YML\" > /etc/grafana/provisioning/datasources/vm.yml && "
                + "printf '%s' \"$DB_YML\" > /etc/grafana/provisioning/dashboards/store.yml && "
                + "printf '%s' \"$DASH_JSON\" > /var/lib/grafana/dashboards/popularity.json && "
                + "exec /run.sh",
            ],
        }
    ],
    volumes: [
        {
            name: "dbus",
            host: {
                path: "/var/run/dbus"
            }
        },
        {
            name: "dev",
            host: {
                path: "/dev"
            }
        },
        {
            name: "shm",
            temp: {}
        }
    ]
};


[
    build("amd64"),
    build("arm64"),
    build("arm"),
]
