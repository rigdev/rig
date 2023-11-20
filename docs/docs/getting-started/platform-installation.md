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

### Next steps

And that's it, you're now ready to login on the dashboard at [http://localhost:4747](http://localhost:4747).

You can now either explore the [Your First Capsule](first-capsule) walkthrough or check out the [Next steps](next-steps) guide.
