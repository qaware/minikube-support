# Minikube Support Tools

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fqaware%2Fminikube-support.svg?type=small)](https://app.fossa.com/projects/git%2Bgithub.com%2Fqaware%2Fminikube-support?ref=badge_small)
[![Go-Build](https://github.com/qaware/minikube-support/workflows/Go-Build/badge.svg?branch=master)](https://github.com/qaware/minikube-support/actions?query=workflow%3AGo-Build)

The minikube-support tools are intended to automate and simplifies
common tasks which you need to do every time you set up a new minikube
cluster. For this they combine some cluster internal and external tools
to provide a better interaction between minikube and the developers
local os. The main entry point to run everything on the current minikube
instance, is the `minikube-support` command. For this it installs,
configures and provides the following:

- A local [CoreDNS](https://coredns.io/) to access the services and
  ingresses using a domain name `*.minikube`.
- [mkcert](https://github.com/FiloSottile/mkcert) to set up an own CA
  for generating certificates for the domain names served by CoreDNS.
- The [Cert-Manager](https://github.com/jetstack/cert-manager) to manage
  that certificate generation within the cluster.
- A
  [Nginx Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
  to provide access to the ingresses deployed in minikube.
- A dashboard that shows the status of Ingresses,
  `LoadBalancer`-Services, served DNS entries and the `minikube tunnel`
  status.

[TOC]: # "## Table of Contents"

## Table of Contents
- [Quickstart](#quickstart)
- [Installing](#installing)
  - [Requirements](#requirements)
  - [Homebrew Tap](#homebrew-tap)
  - [Building from Source](#building-from-source)
  - [Using prebuilt images](#using-prebuilt-images)
  - [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)
- [Maintainer](#maintainer)


## Quickstart

1. Install `minikube`:
   - Download the latest release and place it somewhere into your
     `$PATH`.
   - Or using Brew:

     ```shell script
     brew install qaware/minikube-support/minikube-support  
     ```
2. Start your cluster `minikube start`
3. Install internal and external components with `minikube-support
   install -l`
   - If you don't want to install the external components (mkcert,
     CoreDNS) just run `minikube-support install`.
4. Run the dashboard: `minikube-support run`
   ![Dashboard after start of `minikube-support run`](docs/run.png)

For more information about using please take a look into the
documentation:
- [How to start](docs/how-to-start.md)

## Installing

### Requirements

The `minikube-support`-Tools requires a supported and preinstalled
package manager on the system. This allows the tools to install the
additional helper tools directly using the system's package manger.

Currently, the following package managers are supported:

- [**HomeBrew**](https://brew.sh/) (macOS and Linux using
  [Linuxbrew](https://docs.brew.sh/Homebrew-on-Linux))

### Homebrew Tap

```shell script
brew install qaware/minikube-support/minikube-support
```

### Building from Source

```shell script
git clone https://github.com/qaware/minikube-support.git
make build
```

Then the built binary is located under [`bin`](bin).

### Using prebuilt images

Take a look under
[Releases](https://github.com/qaware/minikube-support/releases) and
download the prebuilt image for your operating system. To use it just
place the `minikube-support` binary into your `$PATH`. For example into
`/usr/local/bin`.

### Configuration

The `minikube-support` tools do not have own configuration properties,
but requires to configure the dns resolver of the local os. For macOS
the minikube-support tools will do this automatically. Just ensure that
all your Ingresses uses the top level domain `.minikube`.

For Windows and Linux systems there are no known interfaces or
configuration properties to configure conditional forwarding for DNS
requests to `.minikube` programmatically. In this case you have to
configure your DNS resolving by your own.

#### macOS: Add additional TLD

It is possible to add additional top level domains to serve them using your minikube cluster. The following shows how to add the domain `mk.local` to serve them using minikube-support tools.

1. Create a file `mk.local` under `/etc/resolver` with the following content:
   ```
   nameserver ::1
   ```
2. Adjust the CoreDNS configuration in `/opt/mks/coredns/etc/corefile` and after `grpc minikube 127.0.0.1:8053` the following line:
   ```
       grpc mk.local 127.0.0.1:8053
   ```
3. Restart CoreDNS with the following command:
   ```shell script
   launchctl unload /Library/LaunchDaemons/de.chrfritz.minikube-support.coredns.plist
   launchctl load /Library/LaunchDaemons/de.chrfritz.minikube-support.coredns.plist
   ```

## Contributing

1. Fork it
2. Download your fork to your PC (`git clone
   https://github.com/your_username/minikube-support && cd
   minikube-support`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request

## License

Minikube-Support is released under the MIT license. See
[LICENSE](https://github.com/qaware/minikube-support/blob/master/LICENSE)

## Maintainer

Christian Fritz (@chrfritz)
