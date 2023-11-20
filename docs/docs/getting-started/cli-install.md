# CLI Installation

import {RIG_VERSION} from "../../src/constants/versions"

## Install

The following are the different installation options for how to install the CLI

### Homebrew

Add the rig homebrew tap and install the CLI.

```bash
brew tap rigdev/tap
brew install rigdev/tap/rig
```

### Binaries

Rig can be installed manually by downloading a precompiled binary and adding
it to your `$PATH`

Every GitHub release has prebuilt binaries for common platforms and
architectures. Go to [the releases
page](https://github.com/rigdev/rig/releases/latest) to find yours.

### From source

Installation from source requires the go toolchain to be installed.

<pre><code className="language-bash">go install github.com/rigdev/rig/cmd/rig@{RIG_VERSION}</code></pre>

## Next step

Use the [Getting Started guide](/getting-started) to continue installing the Rig Platform.
