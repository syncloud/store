local name = "syncloud-store";
local go = "1.20";
local playwright = "v1.48.2-jammy";
local docker_image = "syncloud/store";

local deployStep(env, hostSecret) = {
    name: "deploy " + env,
    image: "appleboy/drone-ssh:1.7.0",
    settings: {
        host: { from_secret: hostSecret },
        username: { from_secret: env + "_deploy_user" },
        key: { from_secret: env + "_deploy_key" },
        command_timeout: "10m",
        script: [
            "set -ex",
            "TAG=" + docker_image + ":${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}",
            "command -v docker >/dev/null 2>&1 || { apt-get update && apt-get install -y docker.io; }",
            "systemctl is-active --quiet syncloud-store.service && systemctl stop syncloud-store.service || true",
            "systemctl is-enabled --quiet syncloud-store.service 2>/dev/null && systemctl disable syncloud-store.service || true",
            "STORE_UID=$(id -u store)",
            "STORE_GID=$(id -g store)",
            "mkdir -p /var/www/store",
            "chown $STORE_UID:$STORE_GID /var/www/store",
            "docker pull $TAG",
            "docker rm -f syncloud-store 2>/dev/null || true",
            "docker run -d --name syncloud-store --restart=unless-stopped --user $STORE_UID:$STORE_GID -v /var/www/store:/var/www/store $TAG",
            "docker image prune -f",
        ],
    },
};

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
            image: "debian:buster-slim",
            commands: [
                "echo $DRONE_BUILD_NUMBER > version"
            ]
        },
    ] + (if arch == "amd64" then [
        {
            name: "web build",
            image: "mcr.microsoft.com/playwright:" + playwright,
            commands: [
              "cd web",
              "npm ci --prefer-offline --no-audit --no-fund",
              "npm run build",
              "npm run build:stub",
            ]
        },
        {
            name: "web e2e",
            image: "mcr.microsoft.com/playwright:" + playwright,
            environment: {
              PLAYWRIGHT_ARTIFACT_DIR: "/drone/src/artifact",
            },
            commands: [
              "(cd web && npx vite preview --host 127.0.0.1 --port 4173) &",
              "for i in $(seq 1 30); do curl -fsS http://127.0.0.1:4173 >/dev/null && break || sleep 1; done",
              "cd web/e2e",
              "npm ci --no-audit --no-fund",
              "PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 npx playwright test --project=desktop",
              "PLAYWRIGHT_BASE_URL=http://127.0.0.1:4173 npx playwright test --project=mobile",
            ]
        },
    ] else []) + [
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
            image: "debian:buster-slim",
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
            image: "debian:buster-slim",
            commands: [
              "VERSION=$(cat version)",
              "./test/test.sh"
            ]
        },
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
    ] + (if arch == "amd64" then [
        {
            name: "docker",
            image: "plugins/docker:20.18",
            settings: {
                repo: docker_image,
                username: { from_secret: "docker_username" },
                password: { from_secret: "docker_password" },
                tags: [
                    "${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}",
                    "${DRONE_BRANCH}",
                ],
            },
            when: {
                event: ["push", "tag"],
            },
        },
        deployStep("uat", "uat_deploy_host") + {
            when: {
                event: ["push"],
            },
        },
        deployStep("prod", "prod_deploy_host") + {
            when: {
                event: ["push"],
                branch: ["stable"],
            },
        },
    ] else []),
    services:
    [
        {
            name: "device",
            image: "syncloud/bootstrap-buster-" + arch,
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
            image: "syncloud/bootstrap-buster-" + arch,
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
            image: "syncloud/bootstrap-buster-" + arch,
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
