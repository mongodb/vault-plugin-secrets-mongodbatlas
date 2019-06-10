module github.com/mongodb-partners/vault-plugin-secrets-mongodb-atlas

go 1.12

require (
	github.com/hashicorp/go-hclog v0.9.2
	github.com/hashicorp/vault/api v1.0.2
	github.com/hashicorp/vault/sdk v0.1.11
	github.com/mongodb-partners/go-client-mongodb-atlas v0.0.0
)

replace github.com/mongodb-partners/go-client-mongodb-atlas v0.0.0 => ../go-client-mongodb-atlas/
