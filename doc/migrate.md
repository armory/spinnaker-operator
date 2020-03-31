# Migrating from halyard to Operator

If you have a current spinnaker instance installed via halyard, you can use this guide to take the existing 
configuration files and start using operator.

## Moving from halyard to Operator

The migration process from halyard to Operator can be completed in n steps:

1. Install Operator.
2. Export Spinnaker configuration.
3. Export Spinnaker profiles.
4. Export Spinnaker settings.
5. Export Spinnaker files. 
6. Prevalidate your spinnaker configuration (only cluster mode).
7. Apply your SpinnakerService

### 1. Install Operator

To get started, [install](../README.md) Spinnaker Operator. 

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

More details on [SpinnakerService Options](options.md) on `.spec.spinnakerConfig.config` section

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

More details on [SpinnakerService Options](options.md) on `.spec.spinnakerConfig.profiles` section

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

### 5. Export Spinnaker files

If we have configured spinnaker files, we will need to migrate these files to `SpinnakerService` manifest.

Let's identify the current files under  `~/.hal/default/profiles`

For each file let's create an entry under `spec.spinnakerConfig.files`

Lets say we have the following files 

```bash
$ tree -v ~/.hal/default/profiles
├── echo.yml
└── rosco
    └── packer
        └── example-packer-config.json

2 directories, 2 files
```

We need to create new entry with the name of the file following  this instructions:
 
- For each folder put the folder name followed by double underscores (__) and at the very end the name of the file.

```yaml
spec:
  spinnakerConfig:
    files: 
      profiles__rosco__packer__example-packer-config.json:
        <<CONTENT>>
```

### 6. Prevalidate your spinnaker configuration (only cluster mode)

```bash
$ kubectl -n <namespace> apply -f <spinnaker service> --server-dry-run
```

If something is wrong with your manifest validation service will throw an error.

### 7. Apply your SpinnakerService

```bash
$ kubectl -n <namespace> apply -f <spinnaker service>
```
