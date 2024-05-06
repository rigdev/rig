<p align="center">
  <a href="https://www.rig.dev">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="https://github.com/rigdev/rig/assets/3807831/2b31efd1-c518-4939-8f2a-411805902d03">
      <img alt="rig" src="https://github.com/rigdev/rig/assets/3807831/ddf2a96b-e9a8-44c5-9b83-a333736bd472" width="230px">
    </picture>
  </a>
</p>

<p align="center"><b><a href="https://docs.rig.dev/">Documentation</a> | <a href="https://rig.dev/">Website</a></b></p>

<p align="center">
  The DevEx & Application-layer for your Internal Developer Platform
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/rigdev/rig">
    <img src="https://pkg.go.dev/badge/github.com/rigdev/rig.svg" alt="Go Reference">
  </a>
  <a href="https://goreportcard.com/badge/github.com/rigdev/rig">
    <img src="https://goreportcard.com/badge/github.com/rigdev/rig" alt="Go Report">
  </a>
  <a href="https://github.com/rigdev/rig/releases/latest">
    <img src="https://img.shields.io/github/release/rigdev/rig.svg" alt="Rig is released under the Apache2 license." />
  </a>
  <a href="https://github.com/rigdev/rig/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-apache2-blue.svg" alt="Rig is released under the Apache2 license." />
  </a>
  <a href="https://join.slack.com/t/rig-community/shared_invite/zt-26104sb0m-lzmGdbR~XvCZU3xiM0MR7g">
    <img src="https://img.shields.io/badge/join-slack-blue.svg?logo=slack" alt="Join Slack" />
  </a>
  <a href="https://twitter.com/intent/follow?screen_name=Rig_dev">
    <img src="https://img.shields.io/badge/follow-%40Rig__dev-blue?logo=x" alt="Follow @Rig_dev" />
  </a>
</p>

## 🌟 What is Rig?

Rig.dev is a complete service-lifecycle platform for Kubernetes. The Platform empower developers with a developer-friendly deployment engine that simplifies the process of rolling out, managing, debugging, and scaling applications.

The Rig platform is self-hosted can be installed in any Kubernetes cluster and will immediately simplify maintaining services in the cluster.

## 📦 Features

The complete stack offers:

- rig - The CLI for interacting with the rig-platform and its resources
- rig-operator - Our open-core abstraction implementation running in Kubernetes
- rig-platform - Our developer-friendly rollout engine and dashboard
- Helm charts for installing rig-operator and rig-platform
- The platform protobuf interfaces (allows for easy API-client generation)
- Plugin framework for easy integrations of the Platform with _any_ infrastructure
- Simple CLI commands for integrating with any CI/CD pipeline

## ⚙️ Plugins

The Rig platform comes with a open Plugin framework, for easy customization.

The default configuration will run with the basic plugins:

- Deployment Plugin - [`rigdev.deployment`](https://github.com/rigdev/rig/tree/main/plugins/capsulesteps/deployment)
- CronJob Plugin - [`rigdev.cronjob`](https://github.com/rigdev/rig/tree/main/plugins/capsulesteps/cron_jobs)
- Service Account Plugin - [`rigdev.service_account`](https://github.com/rigdev/rig/tree/main/plugins/capsulesteps/service_account)
- Ingress Rources Plugin - [`rigdev.ingress_routes`](https://github.com/rigdev/rig/tree/main/plugins/capsulesteps/ingress_routes) (must be configured, see [here](https://docs.rig.dev/operator-manual/setup-guide/networking))

More helper-plugins are available [here](https://docs.rig.dev/operator-manual/plugins/builtin) and used in a few examples described
[here](https://docs.rig.dev/operator-manual/plugins/examples).

To write your own plugins, see our [Custom Plugin guide](https://docs.prod.rig.dev/operator-manual/plugins/thirdparty/).

## 🧑‍💻 Getting Started

Our Setup Guide is available [here](https://docs.rig.dev/operator-manual/setup-guide).

The guide allows you to set up either your local machine or a Kubernetes cluster in production.

## 👯 Community

For support, development, and community questions, we recommend checking out our [Slack channel](https://join.slack.com/t/rig-community/shared_invite/zt-26104sb0m-lzmGdbR~XvCZU3xiM0MR7g).

Furthermore, be sure to check out our [Code of Conduct](https://github.com/rigdev/rig/blob/main/CODE_OF_CONDUCT.md).

## ➕ Contributions

We love additions in all forms, to make Rig even greater.

The easiest steps are to file bug reports, gaps in documentation, etc. If you know how to improve it yourself, we encourage you to fork the relevant repository and create a Pull Request.

## 📖 License

Rig is licensed under the Apache 2.0 License.
