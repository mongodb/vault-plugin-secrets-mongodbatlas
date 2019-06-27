module github.com/mongodb-partners/vault-plugin-secrets-mongodb-atlas

go 1.12

require (
	github.com/Sectorbob/mlab-ns2 v0.0.0-20171030222938-d3aa0c295a8a
	github.com/hashicorp/errwrap v1.0.0
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/vault/api v1.0.2
	github.com/hashicorp/vault/sdk v0.1.11
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mongodb-partners/go-client-mongodb-atlas v0.0.0
	github.com/sethvargo/go-password v0.1.2
)

replace github.com/mongodb-partners/go-client-mongodb-atlas v0.0.0 => ../go-client-mongodb-atlas/
