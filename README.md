# WordPress Threeport Extension

A reference implementation for extending
[Threeport](https://github.com/threeport/threeport).

This project demonstrates how to use the Threeport SDK to build extensions to
Threeport.  See the [Threeport Docs](https://threeport.io/sdk/tutorial/) for a
tutorial that provides step-by-step instructions on how to build this extension.

## Quickstart

If you already have the [Threeport CLI
installed](https://threeport.io/install/install-tptctl/), you can install the
plugin for this WordPress extension as follows.

1. Clone this repo.

1. Build the plugin binary.  Requires [mage](https://magefile.org/).
   ```bash
   mage build:plugin
   ```

1. Install the plugin binary.
   ```bash
   cp bin/wordpress ~/.threeport/plugins/
   ```

1. View the usage info for the plugin.
   ```bash
   tptctl wordpress -h
   ```

If you have a local Threeport control plane running, you can install an instance
of Wordpress as follows.

1. Create a local container registry.
   ```bash
   mage dev:localRegistryUp
   ```

1. Build and push the WordPress extensions control plane components to the local
   registry.
   ```bash
   mage build:allImagesDev
   ```

1. Install this extension's control plane components.
   ```bash
   tptctl wordpress install -r localhost:5001
   ```

1. Install an instance of the WordPress app.
   ```bash
   tptctl wordpress create wordpress -c samples/wordpress.yaml
   ```

Clean up with the following steps.

1. Remove the WordPress app.
   ```bash
   tptctl wordpress delete wordpress -c samples/wordpress.yaml
   ```

1. Remove the local container registry.
   ```bash
   mage dev:localRegistryDown
   ```

