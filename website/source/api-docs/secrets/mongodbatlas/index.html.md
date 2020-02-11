---
layout: "api"
page_title: "MongoDB Atlas - Secrets Engines - HTTP API"
sidebar_title: "MongoDB Atlas"
sidebar_current: "docs-secrets-engines-mongodbatlas"
description: |-
  The MongoDB Atlas Secrets Engine for Vault generates MongoDB Atlas Programmatic API Keys dynamically.
---

# MongoDB Atlas Secrets Engine

The MongoDB Atlas Secrets Engine generates Programmatic API keys for MongoDB Atlas. This allows one to manage the lifecycle of these MongoDB Atlas secrets programmatically. The created MongoDB Atlas secrets are
time-based and are automatically revoked when the Vault lease expires, unless renewed. Vault will create a Programmatic API key for each lease scoped to the MongoDB Atlas project or organization denoted with the included role(s). An IP Whitelist may also be configured for the Programmatic API key with desired IPs and/or CIDR blocks.

The MongoDB Atlas Programmatic API Key Public and
Private Key is returned to the caller. To learn more about Programmatic API Keys visit the [Programmatic API Keys Doc](https://docs.atlas.mongodb.com/reference/api/apiKeys/).

## Configure Connection

In addition to the parameters defined by the Secrets Engines Backend, this plugin has a number of parameters to further configure a connection.

| Method   | Path                         |
| :--------------------------- | :--------------------- |
| `POST`   | `/mongodbatlas/config`     |


## Parameters

- `public_key` `(string: <required>)` – The Public Programmatic API Key used to authenticate with the MongoDB Atlas API.
- `private_key` `(string: <required>)` - The Private Programmatic API Key used to connect with MongoDB Atlas API.

### Sample Payload

```json
{
  "public_key": "aPublicKey",
  "private_key": "aPrivateKey",
}
```

### Sample Request
```bash
$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
    --data @payload.json \
    http://127.0.0.1:8200/mongodbatlas/config`
```

## Programmatic API Keys
Programmatic API Key credential types create a Vault role to generate a Programmatic API Key at
either the MongoDB Atlas Organization or Project level with the designated role(s) for programmatic access.

| Method   | Path                         |
| :--------------------------- | :--------------------- |
| `POST`   | `/roles/{name}`     |


## Parameters

`project_id` `(string <required>)` - Unique identifier for the organization to which the target API Key belongs. Use the /orgs endpoint to retrieve all organizations to which the authenticated user has access.
`roles` `(list [string] <required>)` - List of roles that the API Key needs to have. If the roles array is provided
`ip_addresses` `(list [string] <Optional>)` - IP address to be added to the whitelist for the API key. This field is mutually exclusive with the cidrBlock field.
`cidr_blocks` `(list [string] <Optional>)` - Whitelist entry in CIDR notation to be added for the API key. This field is mutually exclusive with the ipAddress field.

### Sample Payload

```json
{
  "project_id": "5cf5a45a9ccf6400e60981b6",
  "roles": ["GROUP_CLUSTER_MANAGER"],
  "cidr_blocks": ["192.168.1.3/32"],
  "ip_addresses": ["192.168.1.3", "192.168.1.3"]
}
```

```bash
$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
    --data @payload.json \
    http://127.0.0.1:8200/mongodbatlas/roles/test-programmatic-key
```

### Sample Response
```json
{
  "project_id": "5cf5a45a9ccf6400e60981b6",
  "roles": ["GROUP_CLUSTER_MANAGER"],
  "cidr_blocks": ["192.168.1.3/32"],
  "ip_addresses": ["192.168.1.3", "192.168.1.3"],
  "organization_id": "7cf5a45a9ccf6400e60981b7",
  "ttl": "0s",
  "max_ttl": "0s"
}
```

## Read Credential

### Sample Request

| Method   | Path                         |
| :--------------------------- | :--------------------- |
| `GET`   | `/creds/{name}`     |

```bash
$ curl \
    --header "X-Vault-Token: ..." \
    http://127.0.0.1:8200/mongodbatlas/creds/0fLBv1c2YDzPlJB1PwsRRKHR
```

### Sample Response
```json
{
  "lease_duration": "20s",
  "lease_renewable": true,
  "description": "vault-test-1563980947-1318",
  "private_key": "905ae89e-6ee8-40rd-ab12-613t8e3fe836",
  "public_key": "klpruxce"
}
```