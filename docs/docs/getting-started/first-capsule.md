# Deploy your first capsule

Now that you have a configured Rig platform up and running, and you have installed the CLI, you are ready to deploy your first capsule. This will be a short introduction to capsules, and how to deploy a containerized application to Rig using the CLI.

## What is a capsule?

In Rig a capsule encapsulates (😉) a collection of resources on Rig, that is used to manage and run an application.
Capsules contain builds, which hold information on how to deploy the application: which container image to use, git repository, etc.
Capsules then contain rollouts, which are deployments of a build to a specific environment and network network configuration. Rollouts are immutable and are used to manage the lifecycle of the application. This means that when the deployed build is updated, the number of replicas is changed or when a different environment or network configuration is used, a new rollout is automatically created.

## Deploy Nginx in a Rig Capsule Using the CLI

This guide will take you through the process of deploying an Nginx Server to Rig using the CLI.

### Create a capsule

```bash
rig capsule create -c nginx-capsule
```

This will create an empty capsule called `nginx-capsule` with no additional resources. We can verify that the capsule was created by running:

```bash
rig capsule get
```

This will list all capsules in the current project, where you should see the capsule you just created.

### Create a build

Next, we need to create a build with the Nginx image for the capsule. This is done using the following command:

```bash
rig capsule -c nginx-capsule build create --image nginx:latest
```
Note that if you are on an arm chip, you should the arm version of the image, 'arm64v8/nginx:latest'.

From the command, we should see an output similar to: `Created new build: <build-id>`

We can verify that the build was created by running:

```bash
rig capsule -c nginx-capsule build get
```

This will list all builds for the capsule, where you should see the build you just created.

### Deploy the build

Now that we have a build, we can deploy it in the nginx-capsule. This is done using the command:

```bash
rig capsule -c nginx-capsule deploy --build-id <build-id>
```

Where `<build-id>` is the id of the build you just created. This will create a rollout for the build, and deploy it with the default configuration. We can verify that the rollout was create by running:

```bash
rig capsule -c nginx-capsule rollout get
```

This will list all rollouts for the capsule, where you should see one for the deployment you just initiated.

if you run Rig in Kubernetes, you can run

```bash
kubectl get pods -n <project-id>
```

to see the Nginx pod running.

Alternatively if you run Rig in Docker, you can run

```bash
docker ps
```

to see the Nginx container running.

We can also shortcut the creation of build and deployment by simply supplying the deploy command with the same image. This wil automatically create a corresponding build and deploy it. This can be done by running:

```bash
rig capsule -c nginx-capsule deploy -i nginx:latest
```

### Scale the capsule

Now that we have a running deployment, we can scale the number of replicas. This is done using the command:

```bash
rig capsule -c nginx-capsule scale horizontal --replicas 3
```

This will create a new rollout with the updated number of replicas. In order to verify this, you can run the previous commands to see the new changes reflected in the rollout and the pods/containers.

### Set Static Content
Instead of the default content, we can mount a config file to the container. This is done by creating an `index.html` file with some content, for example:
  
```html
<html>
  <body>
    <h1>Hello World!</h1>
  </body>
</html>
```

and then running the following command:
  
```bash
rig capsule -c nginx-capsule mount set --src index.html --dst /usr/share/nginx/html/index.html
```

### Configure the network

In order to expose the Nginx server to the public internet, we need to configure the network. This is done by creating a `network.yaml` file, for example with the following content:

```yaml
interfaces:
  - name: http
    port: 80
    public:
      enabled: true
      method:
        load_balancer:
          port: 8081
```

and then running the following command:

```bash
rig capsule -c nginx-capsule network configure network.yaml
```

This will create a new rollout with the updated network configuration. Now open your favorite browser and navigate to [http://localhost:8081](http://localhost:8081) to see the Nginx server running. Well done, you have created and deployed and exposed your first capsule to Rig! 🎉


### Shortcut

The above guide executed every step of the process in each command explicitly. However, the CLI provides a shortcut for created, deploying and configuring the capsule in one command. This above could be achieved by running the following command:

```bash
rig capsule create -c nginx-capsule --interactive
```

Which will take you through the process of creating a capsule, adding an image, configuring the network, setting the number of replicas and then deploying the capsule.

## Deploy Nginx in a Rig Capsule using the Dashboard
