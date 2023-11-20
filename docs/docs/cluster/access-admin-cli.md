---
sidebar_position: 1
hide_table_of_contents: true
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Access the Admin CLI within the Rig Container

The Rig binary has a built-in Admin CLI to perform admin requests and manage your cluster. Below, we will show you how to execute these commands inside of your Docker or Kubernetes environment. To find a list of available commands, run:

<Tabs>
<TabItem value="docker" label="Docker">

- Run the Admin CLI inside of the container by running `docker exec rig rig-admin`. You should now get an output like below:
```
Admin tool for managing a Rig setup

Usage:
  rig-admin [command]

Available Commands:
  capsule     
  cluster     
  completion  Generate the autocompletion script for the specified shell
  database    
  groups      
  help        Help about any command
  projects    
  push-image  
  storage     
  users       

Flags:
  -c, --config string    config file to use
  -h, --help             help for rig-admin
      --project string   project to target (default "rig")

Use "rig-admin [command] --help" for more information about a command.
```

</TabItem>
<TabItem value="k8s" label="Kubernetes">

- Assuming you have access to your Kubernetes cluster, run `kubectl exec deploy/rig -- rig-admin` to view a list of commands. You should now get an output like below:
```
Admin tool for managing a Rig setup

Usage:
  rig-admin [command]

Available Commands:
  capsule     
  cluster     
  completion  Generate the autocompletion script for the specified shell
  database    
  groups      
  help        Help about any command
  projects    
  push-image  
  storage     
  users       

Flags:
  -c, --config string    config file to use
  -h, --help             help for rig-admin
      --project string   project to target (default "rig")

Use "rig-admin [command] --help" for more information about a command.
```
</TabItem>
</Tabs>

Run any of the commands using the Admin CLI by appending your command to your request. Eg. to access admin users, run 
```bash
rig-admin users
```