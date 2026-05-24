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
    trigger: { event: ["push", "tag"] },
    platform: { os: "linux", arch: arch },
    steps:
        [
            {
                name: "version",
                image: "debian:" + debian,
                commands: ["echo $DRONE_BUILD_NUMBER > version"],
            },
        ] + (if arch == "amd64" then [
            {
                name: "victoria-metrics",
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
                name: "apps.s3",
                image: "dxflrs/garage:v1.0.1",
                detach: true,
                environment: {
                    GARAGE_CONFIG_FILE: "/drone/src/ci/garage.toml",
                },
            },
            {
                name: "web build",
                image: "node:20-bookworm-slim",
                commands: ["bash web/build.sh"],
            },
        ] else []) + [
            {
                name: "unit test",
                image: "golang:" + go,
                commands: ["./unit-test.sh"],
            },
            {
                name: "build store",
                image: "golang:" + go,
                commands: [
                    "VERSION=$(cat version)",
                    "./build.sh $VERSION " + arch,
                ],
            },
            {
                name: "build apps",
                image: "debian:" + debian,
                commands: [
                    "apt update && apt install -y squashfs-tools",
                    "./test/build-apps.sh",
                ],
            },
            {
                name: "build test",
                image: "golang:" + go,
                commands: ["./test/build-tests.sh"],
            },
            {
                name: "docker push publisher",
                image: "plugins/docker:20.13",
                settings: {
                    repo: publisher_image,
                    dockerfile: "Dockerfile.store-publisher",
                    username: { from_secret: "DOCKER_USERNAME" },
                    password: { from_secret: "DOCKER_PASSWORD" },
                    tags: [version + "-" + arch],
                },
                when: { event: ["push", "tag"] },
            },
        ] + (if arch == "amd64" then [
            {
                name: "seed s3",
                image: "alpine:3.20",
                environment: {
                    GARAGE_RPC_SECRET: "1799ff75e85715cd0bd91e09f2a9d70b1799ff75e85715cd0bd91e09f2a9d70b",
                    GARAGE_TRIPLE: "x86_64-unknown-linux-musl",
                },
                commands: ["./ci/seed.sh"],
            },
            {
                name: "docker push store",
                image: "plugins/docker:20.13",
                settings: {
                    repo: docker_image,
                    username: { from_secret: "DOCKER_USERNAME" },
                    password: { from_secret: "DOCKER_PASSWORD" },
                    tags: [version, "${DRONE_BRANCH}"],
                },
                when: { event: ["push", "tag"] },
            },
            {
                name: "deploy test",
                image: "debian:" + debian,
                environment: {
                    DEPLOY_HOST: "api.store",
                    DEPLOY_USER: "root",
                    DEPLOY_URL: "http://api.store",
                    SYNCLOUD_TOKEN: "test",
                },
                commands: [
                    "./ci/test-init.sh",
                    "./ci/deploy-prepare.sh test",
                    "./ci/deploy-run.sh test " + image_tag,
                    "./ci/deploy-verify.sh test",
                ],
                when: { event: ["push", "tag"] },
            },
            {
                name: "publish testapp1",
                image: publisher_image + ":" + version + "-amd64",
                environment: { SYNCLOUD_TOKEN: "test" },
                command: [
                    "snap",
                    "-d", "test/testapp1",
                    "-c", "stable",
                    "-s", "http://api.store",
                ],
                when: { event: ["push", "tag"] },
            },
            {
                name: "publish testapp2",
                image: publisher_image + ":" + version + "-amd64",
                environment: { SYNCLOUD_TOKEN: "test" },
                command: [
                    "snap",
                    "-d", "test/testapp2",
                    "-c", "stable",
                    "-s", "http://api.store",
                ],
                when: { event: ["push", "tag"] },
            },
            {
                name: "test",
                image: "debian:" + debian,
                commands: ["./test/test.sh"],
            },
            {
                name: "grafana provision",
                image: "debian:" + debian,
                commands: ["./ci/grafana-provision.sh"],
            },
            {
                name: "docker push store latest",
                image: "plugins/docker:20.13",
                settings: {
                    repo: docker_image,
                    username: { from_secret: "DOCKER_USERNAME" },
                    password: { from_secret: "DOCKER_PASSWORD" },
                    tags: ["latest"],
                },
                when: { event: ["push"], branch: ["stable"] },
            },
            {
                name: "web e2e",
                image: "mcr.microsoft.com/playwright:" + playwright,
                environment: {
                    PLAYWRIGHT_BASE_URL: "http://api.store",
                },
                commands: ["bash web/e2e/run.sh"],
                when: { event: ["push", "tag"] },
            },
            {
                name: "deploy uat",
                image: "debian:" + debian,
                environment: {
                    DEPLOY_HOST: { from_secret: "uat_deploy_host" },
                    DEPLOY_USER: { from_secret: "uat_deploy_user" },
                    DEPLOY_KEY: { from_secret: "uat_deploy_key" },
                    DEPLOY_URL: { from_secret: "uat_deploy_url" },
                    SYNCLOUD_TOKEN: { from_secret: "uat_token" },
                    AWS_ACCESS_KEY_ID: { from_secret: "AWS_ACCESS_KEY_ID" },
                    AWS_SECRET_ACCESS_KEY: { from_secret: "AWS_SECRET_ACCESS_KEY" },
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
                    SYNCLOUD_TOKEN: { from_secret: "prod_token" },
                    AWS_ACCESS_KEY_ID: { from_secret: "AWS_ACCESS_KEY_ID" },
                    AWS_SECRET_ACCESS_KEY: { from_secret: "AWS_SECRET_ACCESS_KEY" },
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
                    host: { from_secret: "artifact_host" },
                    username: "artifact",
                    key: { from_secret: "artifact_key" },
                    timeout: "2m",
                    command_timeout: "2m",
                    target: "/home/artifact/repo/" + name + "/${DRONE_BUILD_NUMBER}-" + arch,
                    source: [
                        "test/testapp*/*.snap",
                        "out/*",
                        "test/artifacts/*",
                        "artifact/*",
                    ],
                },
                when: { status: ["failure", "success"] },
            },
            {
                name: "publish to github",
                image: "plugins/github-release:1.0.0",
                settings: {
                    api_key: { from_secret: "github_token" },
                    files: "out/*",
                    overwrite: true,
                    file_exists: "overwrite",
                },
                when: { event: ["tag"] },
            },
        ],
    services: if arch == "amd64" then [
        {
            name: "device",
            image: "syncloud/bootstrap-bookworm-amd64:" + platform,
            privileged: true,
            volumes: [
                { name: "dbus", path: "/var/run/dbus" },
                { name: "dev", path: "/dev" },
            ],
        },
        {
            name: "api.store",
            image: "syncloud/bootstrap-bookworm-amd64:" + platform,
            privileged: true,
            volumes: [
                { name: "dbus", path: "/var/run/dbus" },
                { name: "dev", path: "/dev" },
            ],
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
    ] else [],
    volumes: [
        { name: "dbus", host: { path: "/var/run/dbus" } },
        { name: "dev", host: { path: "/dev" } },
        { name: "docker-sock", host: { path: "/var/run/docker.sock" } },
        { name: "shm", temp: {} },
    ],
};


local publisherManifest = {
    kind: "pipeline",
    name: "publisher manifest",
    depends_on: ["amd64", "arm64", "arm"],
    trigger: { event: ["push", "tag"] },
    steps: [
        {
            name: "manifest",
            image: "plugins/manifest:1.4",
            settings: {
                username: { from_secret: "DOCKER_USERNAME" },
                password: { from_secret: "DOCKER_PASSWORD" },
                target: publisher_image + ":${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}",
                template: publisher_image + ":${DRONE_BRANCH}-${DRONE_BUILD_NUMBER}-ARCH",
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
