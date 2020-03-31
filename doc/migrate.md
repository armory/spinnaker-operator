# Migrating from halyard to Operator

If you have a current spinnaker instance installed via halyard, you can use this guide to take the existing 
configuration files and start using operator.

## Moving from halyard to Operator

The migration process from halyard to Operator can be completed in n steps:

1. Install Operator.
2. Export Spinnaker configuration.
3. Export Spinnaker profiles.
4. Export Spinnaker settings.
4. Export Spinnaker files.
n. Prevalidate your spinnaker configuration (only cluster mode).

### 1. Install Operator

To get started, install [Spinnaker Operator](https://github.com/armory/spinnaker-operator/blob/master/README.md#L69). 

### 2. Export Spinnaker configuration 

From the `config` file copy content of desired profile 

Lets say you want to migrate `default` hal profile, then you will have the following structure:

```yaml
currentDeployment: default
deploymentConfigurations:
- name: default
  <<CONTENT>> 
```

Where `<<CONTENT>>` needs to be added under `spec.spinnakerConfig.config` on `SpinnakerService` manifest as follows:

```yaml
spec:
  spinnakerConfig:
    config:
      <<CONTENT>> 
```

Note: `config` is under `~/.hal`

### 3. Export Spinnaker profiles

If we have configured spinnaker profiles, we will need to migrate these profiles to `SpinnakerService` manifest.

Let's identify the current profiles under  `~/.hal/default/profiles`

For each file let's create an entry under `spec.spinnakerConfig.profiles`

Lets say we have the following profiles 

```bash
$ ls -a ~/.hal/default/profiles | sort
echo-local.yml
```

We need to create new entry with the name of the file without `-local.yaml` as follows:

```yaml
spec:
  spinnakerConfig:
    profiles:
      echo: 
        <<CONTENT>>
```

### 4. Export Spinnaker settings

If we have configured spinnaker settings, we will need to migrate these profiles to `SpinnakerService` manifest.

Let's identify the current settings under  `~/.hal/default/service-settings`

For each file let's create an entry under `spec.spinnakerConfig.service-settings`

Lets say we have the following settings 

```bash
$ ls -a ~/.hal/default/service-settings | sort
echo.yml
```

We need to create new entry with the name of the file without `.yaml` as follows:

```yaml
spec:
  spinnakerConfig:
    service-settings: 
      echo:
        <<CONTENT>>
```


### n. Prevalidate your spinnaker configuration (only cluster mode)

```bash
$ kubectl -n <namespace> apply -f <spinnaker service> --server-dry-run
```

