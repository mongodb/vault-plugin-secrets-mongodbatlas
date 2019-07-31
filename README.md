# Vault Plugin: MongoDB Atlas Secrets Engine

This is a Secrets engine to be used with [Hashicorp Vault](https://www.github.com/hashicorp/vault).
This plugin generates unique, ephemeral datababase user credentials and programmatic API keys.

**Please note**: Hashicorp takes Vault's security and their users' trust very seriously, as does MongoDB.

If you believe you have found a security issue in Vault or with this plugin, 
_please responsibly disclose_ by 
contacting us at [security@hashicorp.com](mailto:security@hashicorp.com) and contact MongoDB
directly via [security@mongodb.com](mailto:security@mongodb.com) or [open a ticket](https://jira.mongodb.org/plugins/servlet/samlsso?redirectTo=%2Fbrowse%2FSECURITY) (link is external).

## Quick Links
- [Vault Website](https://www.vaultproject.io)
- [MongoDB Atlas Website](https://www.mongodb.com/cloud/atlas)
- [Atlas MongoDB Secrets Docs](https://www.vaultproject.io/docs/secrets/atlasmongodb/index.html)
- [Vault Github](https://www.github.com/hashicorp/vault)
- [General Announcement List](https://groups.google.com/forum/#!forum/hashicorp-announce)
- [Discussion List](https://groups.google.com/forum/#!forum/vault-tool)


## Usage

**The following will be accurate after review and approval by Hashicorp, which is in progress. Until then follow the instructions in the developing section that follows:**

This is a [Vault plugin](https://www.vaultproject.io/docs/internals/plugins.html)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works. Otherwise, first read this guide on 
how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

If you are using Vault 11.0.1 or above, this plugin is packaged with Vault
and by default can be enabled by running:
 ```sh
 
 $ vault secrets enable mongodbatlas
 
 Success! Enabled the mongodbatlas secrets engine at: mongodbatlas/
 
 ```
 
 If you are testing this plugin in an earlier version of Vault or 
 want to develop, see the next section. 

## Developing

If you wish to work on this plugin, you'll first need [Go](https://www.golang.org) 
installed on your machine (whichever version is required by Vault).

Make sure Go is properly installed, including setting up a [GOPATH](https://golang.org/doc/code.html#GOPATH).

### Get Plugin 
Clone this repository: 

```

mkdir $GOPATH/src/github.com/hashicorp/vault-plugin-secrets-alicloud`
cd $GOPATH/src/github.com/hashicorp/
git clone git@github.com:mongodb/vault-plugin-secrets-mongodbatlas.git

```
(or use `go get github.com/mongodb/vault-plugin-secrets-mongodbatlas` ).

You can then download any required build tools by bootstrapping your environment:

```sh
$ make bootstrap
```

To compile a development version of this plugin, run `make` or `make dev`.
This will put the plugin binary in the `bin` and `$GOPATH/bin` folders. `dev`
mode will only generate the binary for your platform and is faster:

```sh
$ make
$ make dev
```

### Install Plugin in Vault

Put the plugin binary into a location of your choice. This directory
will be specified as the [`plugin_directory`](https://www.vaultproject.io/docs/configuration/index.html#plugin_directory)
in the Vault config used to start the server.

```hcl

plugin_directory = "path/to/plugin/directory"

```

Start a Vault server with this config file:
```sh
$ vault server -config=path/to/config.json ...
```

Once the server is started, register the plugin in the Vault server's [plugin catalog](https://www.vaultproject.io/docs/internals/plugins.html#plugin-catalog):

```sh
$ vault write sys/plugins/catalog/alicloudsecrets \
        sha_256="$(shasum -a 256 path/to/plugin/directory/vault-plugin-secrets-mongodbatlas | cut -d " " -f1)" \
        command="vault-plugin-secrets-mongodbatlas"
```

Any name can be substituted for the plugin name "mongodbatlas". This
name will be referenced in the next step, where we enable the secrets
plugin backend using the AliCloud secrets plugin:

```sh
$ vault secrets enable --plugin-name='mongodbatlas' --path="mongodbatlas" plugin

```

### Tests

This plugin has both integration tests, and acceptance tests. 

The integration tests are run by `$ make test` and rather than firing real
API calls, they fire API calls at a local test server that returns expected
responses.

The acceptance tests fire real API calls, and are located in `acceptance_test.go`.
These should be run once as a final step before placing a PR. Please see 
`acceptance_test.go` to learn the environment variables that will need to be set.

**Warning:** The acceptance tests create/destroy/modify *real resources*,
which may incur real costs in some cases. In the presence of a bug,
it is technically possible that broken backends could leave dangling
data behind. Therefore, please run the acceptance tests at your own risk.
At the very least, we recommend running them in their own private
account for whatever backend you're testing.

To run the acceptance tests, after exporting the necessary environment variables, 
from the home directory run `go test`:

```sh
$ go test
```

## Other Docs

**The following will be accurate after review and approval by Hashicorp, which is in progress. Until then read the docs within this repo for more information.**

See up-to-date [docs](https://www.vaultproject.io/docs/secrets/mongodbatlas/index.html)
and general [API docs](https://www.vaultproject.io/api/secret/mongodbatlas/index.html).