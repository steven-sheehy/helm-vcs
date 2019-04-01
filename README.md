# Helm VCS

[![CircleCI](https://circleci.com/gh/steven-sheehy/helm-vcs.svg?style=shield)](https://circleci.com/gh/steven-sheehy/helm-vcs)
[![License Apache](https://img.shields.io/badge/license-apache-blue.svg?style=flat)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/steven-sheehy/helm-vcs.svg)](https://github.com/steven-sheehy/helm-vcs/releases)

A [Helm](https://helm.sh) plugin that turns any version control system into a chart repository

## Motivation

Setting up a Helm chart repository has always been more difficult than necessary. Tools like [ChartMuseum](https://chartmuseum.com/)
make it easier, but still require a server and the repository must live separate from the code. [Chart Releaser](https://github.com/helm/chart-releaser)
and the [Helm Github Plugin](https://github.com/technosophos/helm-github) are great, but require integration with the release process
and are specific to git and GitHub.

The goal of Helm VCS is to turn any version control system (VCS) into a chart repository, without requiring any
changes to that repository. It does this by recursively searching the repository for [valid charts](https://helm.sh/docs/developing_charts/)
and generating the chart repository dynamically on the client-side. By default, every tag is checked out and examined for charts.
If needed, the search can also be done for a specific ref (branch, tag or commit). 

Packaging the chart client-side alleviates developers from having to package and store chart binaries and focus purely on the source
code. The chart artifact is generated consistently for different consumers due to it being backed by the VCS. Using the VCS with its 
immutable tags for chart versioning simplifies the release process.

## Installation

[Releases](https://github.com/steven-sheehy/helm-vcs/releases) can be installed directly as a helm plugin. Find
the release for your OS and architecture and pass the URL to `helm plugin install <URL>`. For Example:

```shell
$ helm plugin install https://github.com/steven-sheehy/helm-vcs/releases/download/v0.1.0/helm-vcs_0.1.0_linux_amd64.tar.gz
```

## Usage

Since the VCS repo may not always be a valid URI (e.g., `git@github.com:steven-sheehy/helm-vcs.git`), we can't use the
normal approach of `helm repo add <name> <uri>`. To add a chart repository, instead use the plugin specific init command:

```shell
$ helm vcs init git://github.com/steven-sheehy/helm-vcs-test.git --path charts/ --tags
```

This command will scan the URI for charts, generate an `index.yaml` from that information and add it as a chart repository to helm with the
given name. If needed, this command can be ran multiple times to update the URI or parameters. After this one time setup, you shouldn't
need to interact with the helm-vcs plugin again. Repository updates, chart installs, etc. will be handled by the regular Helm
[commands](https://helm.sh/docs/helm/#see-also).

## Uninstall

```shell
$ helm repo remove helm-vcs-test
$ helm plugin remove vcs
```

