local name = "syncloud-store";
local go = "1.23";
local playwright = "v1.48.2-jammy";
local docker_image = "syncloud/store";
local publisher_image = "syncloud/store-publisher";
local debian = "bookworm-slim";
local platform = "26.04.10";
local version = "${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}";
local image_tag = docker_image + ":" + version;


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
        {
            name: "vm",
            image: "victoriametrics/victoria-metrics:v1.110.0",
            detach: true,
            command: [
                "-storageDataPath=/storage",
                "-promscrape.config=/drone/src/ci/vm/prometheus.yml",
                "-httpListenAddr=:8428",
                "-search.latencyOffset=0s",
            ],
        },
        {
            name: "s3",
            image: "dxflrs/garage:v1.0.1",
            detach: true,
            environment: {
                GARAGE_CONFIG_FILE: "/drone/src/ci/garage.toml",
            },
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
            name: "s3 init",
            image: "alpine:3.20",
            environment: {
                GARAGE_RPC_SECRET: "1799ff75e85715cd0bd91e09f2a9d70b1799ff75e85715cd0bd91e09f2a9d70b",
                GARAGE_TRIPLE:
                    if arch == "amd64" then "x86_64-unknown-linux-musl"
                    else if arch == "arm64" then "aarch64-unknown-linux-musl"
                    else "armv6l-unknown-linux-musleabihf",
            },
            commands: [
                "apk add --no-cache curl jq",
                "wget -q -O /usr/local/bin/garage https://garagehq.deuxfleurs.fr/_releases/v1.0.1/$GARAGE_TRIPLE/garage",
                "chmod +x /usr/local/bin/garage",
                "for i in $(seq 120); do curl -fsS http://s3:3903/health >/dev/null 2>&1 && break; sleep 1; done",
                "NODE=$(curl -fsS -H 'Authorization: Bearer test-admin-token' http://s3:3903/v1/status | jq -r .node)",
                "export GARAGE_RPC_HOST=$NODE@s3:3901",
                "garage layout assign -z dc1 -c 1G $NODE",
                "garage layout apply --version 1",
                "garage bucket create test",
                "garage key import --yes -n test GK31c4cef60f8f78b1bf12cd71 testtest",
                "garage bucket allow --read --write --owner test --key test",
                "garage bucket website --allow test",
                "garage status",
            ],
        },
        {
            name: "seed s3",
            image: "debian:" + debian,
            commands: [
              "./test/seed",
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
            name: "grafana provision",
            image: "debian:" + debian,
            commands: [
              "./ci/grafana-provision.sh",
            ],
        },
        {
            name: "docker push publisher",
            image: "plugins/docker:20.18",
            settings: {
                repo: publisher_image,
                dockerfile: "Dockerfile.store-publisher",
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                tags: [version + "-" + arch],
            },
            when: { event: ["push", "tag"] },
        },
        {
            name: "e2e publish image",
            image: "docker:24-cli",
            environment: { SYNCLOUD_TOKEN: "test" },
            volumes: [{ name: "docker-sock", path: "/var/run/docker.sock" }],
            commands: [
                "NET=$(docker inspect $(hostname) --format '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}\\n{{end}}' | grep -m1 '^drone-')",
                "echo using network=$NET PWD=$PWD",
                "docker pull " + publisher_image + ":" + version + "-" + arch,
                "docker run --rm --network \"$NET\" --volumes-from $(hostname) -e SYNCLOUD_TOKEN -w $PWD " +
                  publisher_image + ":" + version + "-" + arch + " snap -f test/testapp1_3_" + arch + ".snap -c stable -s http://api.store.test -y test/testapp1/meta/snap.yaml -i test/images/testapp1.png",
                "docker run --rm --network \"$NET\" curlimages/curl:8.10.1 -fsS 'http://api.store.test/api/ui/v1/apps?channel=stable' | grep -q testapp1",
            ],
            when: { event: ["push", "tag"] },
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
                AWS_ACCESS_KEY_ID: "GK31c4cef60f8f78b1bf12cd71",
                AWS_SECRET_ACCESS_KEY: "testtest",
                AWS_S3_ENDPOINT: "http://s3",
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
                AWS_ACCESS_KEY_ID: { from_secret: "aws_access_key_id" },
                AWS_SECRET_ACCESS_KEY: { from_secret: "aws_secret_access_key" },
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
                AWS_ACCESS_KEY_ID: { from_secret: "aws_access_key_id" },
                AWS_SECRET_ACCESS_KEY: { from_secret: "aws_secret_access_key" },
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
            name: "grafana",
            image: "grafana/grafana:11.3.0",
            environment: {
                GF_AUTH_ANONYMOUS_ENABLED: "true",
                GF_AUTH_ANONYMOUS_ORG_ROLE: "Viewer",
                GF_SECURITY_ADMIN_PASSWORD: "admin",
            },
        },
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
            name: "docker-sock",
            host: {
                path: "/var/run/docker.sock"
            }
        },
        {
            name: "shm",
            temp: {}
        }
    ]
};


local publisherManifest = {
    kind: "pipeline",
    name: "publisher manifest",
    depends_on: ["amd64", "arm64", "arm"],
    trigger: { event: ["push", "tag"] },
    steps: [
        {
            name: "manifest version",
            image: "plugins/manifest:1.4",
            settings: {
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                target: publisher_image + ":${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}",
                template: publisher_image + ":${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}-ARCH",
                platforms: ["linux/amd64", "linux/arm64", "linux/arm"],
            },
        },
        {
            name: "manifest branch",
            image: "plugins/manifest:1.4",
            settings: {
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                target: publisher_image + ":${DRONE_BRANCH}",
                template: publisher_image + ":" + version + "-ARCH",
                platforms: ["linux/amd64", "linux/arm64", "linux/arm"],
            },
        },
    ],
};

[
    build("amd64"),
    build("arm64"),
    build("arm"),
    publisherManifest,
]
