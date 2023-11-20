import {
RIG_PLATFORM_CHART_VERSION,
RIG_OPERATOR_CHART_VERSION
} from "../../src/constants/versions"

# Platform Installation

## Local Installation

Rig can run locally in both Kubernetes (KIND) and in a local Docker environment.

:::info Prerequisites
Make sure that you have the [CLI Installed](/getting-started/cli-install).
:::

### Option 1: Docker

To create a Rig setup on your local machine within Docker, simply run the following command:

```bash
rig dev docker create
```

The above command will guide you through the installation. If anything goes wrong, you can always run the command again.

### Option 2: Kubernetes (KIND)

To easily create a Rig Kubernetes setup on your local machine, Rig comes with support for starting up a KIND cluster on your local machine. Run the following command:

```bash
rig dev kind create
```

## Kubernetes Cluster Installation

See [here](/operator-manual/setup-guide) for how to setup Rig on an already existing Kubernetes cluster.

### Next step

And that's it, you're now ready to login on the dashboard at [http://localhost:4747](http://localhost:4747).

### Setup Rig

The next step is to do some simple setup of Rig. This amounts to creating yourself a new Admin user and create a proejct. The Rig docker image comes with a `rig-admin` tool, that can be used for exactly this:

```bash
kubectl exec -it --namespace rig-system deploy/rig-platform \
  -- rig-admin init
```

## Configuration

For more information about how to configure Rig, see the [configuration](/configuration) section.
