---
layout: "docs"
page_title: "MongoDB Atlas - Secrets Engines"
sidebar_title: "MongoDB Atlas"
sidebar_current: "docs-secrets-atlasmongodb"
description: |-
  The MongoDB Atlas Secrets Engine for Vault generates MongoDB Database User Credentials and Programmatic API Keys dynamically.
---

# MongoDB Atlas Secrets Engine

The MongoDB Atlas Secrets Engine generates Database User credentials and Programmatic API keys. 
This allows one to manage the lifecycle of these MongoDB Atlas secrets programmatically. The 
created MongoDB Atlas secrets are time-based and are automatically revoked when the Vault lease 
expires, unless renewed.

This Secrets Engine supports two different types of MongoDB Atlas Secrets:

1. `database_user`: Vault will create a database user for each lease, each database user has defined role(s) that provide appropriate access to the project’s databases/collections. The username and passwordis returned to the caller. To see more about database users visit the [MongoDB Atlas Database Users Documentation](https://docs.atlas.mongodb.com/reference/api/database-users/).
2. `programmatic_api_key`: Vault will create a Programmatic API key for each lease, each key has defined role(s) that provide appropriate access to the defined MongoDB Atlas project or organization. The public and private key is returned to the caller. To see more about Programmatic API Keys visit the [Programmatic API Keys Doc](https://docs.atlas.mongodb.com/reference/api/database-users/).

## Setup

Most Secrets Engines must be configured in advance before they can perform their
functions. These steps are usually completed by an operator or configuration
management tool.

  ~> **Notice:** The following will be accurate after review and approval by Hashicorp, which is in 
    progress. Until then follow the instructions in the [README developing section](./../../../../../README.md).


1. Enable the MongoDB Atlas Secrets Engine:

    ```bash
    $ vault secrets enable mongodbatlas
    Success! Enabled the mongodbatlas Secrets Engine at: mongodbatlas/
    ```

    By default, the Secrets Engine will mount at the name of the engine. To
    enable the Secrets Engine at a different path, use the `-path` argument.

1. It's necessary to generate and configure a MongoDB Atlas Programmatic API Key for your organization that has sufficient permissions to allow Vault to create other Programmatic API Keys.

    In order to grant Vault programmatic access to an organization or project using only the [API](https://docs.atlas.mongodb.com/api/) you need to create a MongoDB Atlas Programmatic API Key with the appropriate roles if you have not already done so. A Programmatic API Key consists of a public and private key so ensure you have both. Regarding roles, the Organization Owner and Project Owner roles should be sufficient for most needs, however be sure to check what each roles grants in the [MongoDB Atlas Programmatic API Key User Roles documentation](https://docs.atlas.mongodb.com/reference/user-roles/). Also ensure you set an IP Whitelist when creating the key.

    For more detailed instructions on how to create a Programmatic API Key in the Atlas UI, including available roles, visit the [Programmatic API Key documenation](https://docs.atlas.mongodb.com/configure-api-access/#programmatic-api-keys).

1. Once you have a MongoDB Atlas Programmatic Key pair, as created in the previous step, Vault can now be configured to use it with MongoDB Atlas:

    ```bash
    $ vault write mongodbatlas/config/root \
        public_key=yhltsvan \
        private_key=2c130c23-e6b6-4da8-a93f-a8bf33218830
    ```

    Internally, Vault will connect to MongoDB Atlas using these credentials. As such,
    these credentials must be a superset of any policies which might be granted
    on API Keys. Since Vault uses the official [MongoDB Atlas Client](https://github.com/mongodb/go-client-mongodb-atlas), it will use the specified credentials. 

    <!-- ~> **Notice:** Even though the path above is `mongodbatlas/config/root`, do not use
    your MongoDB Atlas root account credentials. Instead generate a dedicated Programmatic API key with appropriate roles. -->

## Database Users
The Database User, `database_user`, credential type creates a Vault role that maps a set of MongoDB Atlas database roles to a database/collection in the specified MongoDB Atlas project. An example:

```bash
$ vault write mongodbatlas/roles/testDB \
    credential_type=database_user \
    project_id=5cf5a481ok7f6400e60981b6 \
    database_name=admin \
    roles=-<<EOF
      [{
        "databaseName": "admin",
        "roleName": "atlasAdmin"
      }]
    EOF
```
~> **Notice:** Each user has a set of database roles that provide access to the Atlas project’s databases/collections. A user’s database roles apply to all the clusters in the project: if two clusters have a products database and a user has a database role granting read access on the products database, the user has that access on both clusters.

This write creates a Vault role named "testDB" in the project specified by the project_id with an authentication database of admin. When users generate credentials against "testDB", Vault will create a database user in MongoDB Atlas project with the database role specified to the
appropriate database/collection, in this case the database is testDB and the role granted is atlasAdmin. Vault will then return the username and password for the database user.

Once the Vault role test is created it can be called to generate credentials. Note that the following read command uses the Vault default lease settings since they were not specified for the Vault role:

```bash
$ vault read mongodbatlas/creds/testDB
    Key                Value
    ---                -----
    lease_id           mongodbatlas/creds/testDB/sZz3qvggwcULgERDqe9r151h
    lease_duration     20s
    lease_renewable    true
    password           A3mOyxvSnBpKG5sdID2iNR
    username           vault-testDB-1563475091-2081
```

## Programmatic API Keys

 The Programmatic API Key credential types, org_programmatic_api_key and project_programmatic_api_key, create a Vault role to generate a Programmatic API Key at either the MongoDB Atlas Organization or Project level with the appropriate role(s) for programmatic access.

  Programmatic API Keys:
  - Has two parts, a public key and a private key
  - Cannot be used to log into Atlas through the user interface
  - Must be granted appropriate roles to complete required tasks
  - Must belong to one organization, but may be granted access to any number of projects in that organization.
  - May have an IP whitelist configured and some capabilities may require a whitelist to be configured (these are noted in the MongoDB Atlas API documentation).


1. Create a Vault role for a MongoDB Atlas Programmatic API Key by mapping appropriate Key role(s) to the organization or project designated.

    - org_programmatic_api_key: Set organization_id to the MongoDB Atlas Organization Id with the appropriate [Organization Level Roles](https://docs.atlas.mongodb.com/reference/user-roles/#organization-roles).

    - project_programmatic_api_key: Set project_id to the MongoDB Atlas Project Id with the appropriate [Project Level Roles](https://docs.atlas.mongodb.com/reference/user-roles/#project-roles).

~> **Notice:** Programmatic API keys can belong to only one Organization but can belong to one or more Projects.

Examples:

```bash
$ vault write mongodbatlas/roles/test \
    credential_type=org_programmatic_api_key \
    organization_id=5b23ff2f96e82130d0aaec13 \
    programmatic_key_roles=ORG_MEMBER
```
```bash 
$ vault write mongodbatlas/roles/test \
    credential_type=project_programmatic_api_key \
    project_id=5cf5a45a9ccf6400e60981b6 \
    programmatic_key_roles=GROUP_DATA_ACCESS_READ_ONLY
```

In both of these examples Vault returns the public/private key pair generated.

## Programmatic API Key Whitelist

Programmatic API Key access can and should be limited with a IP Whitelist. In the following example both a CIDR block and IP address are added to the IP whitelist for Keys generated with this Vault role:
  
```bash 
  $ vault write atlas/roles/test \
      credential_type=project_programmatic_api_key \
      project_id=5cf5a45a9ccf6400e60981b6 \
      programmatic_key_roles=GROUP_CLUSTER_MANAGER \
      cidr_blocks=192.168.1.3/32 \
      ip_addresses=192.168.1.3
```

Verify the created Programmatic API Key Vault role has the added CIDR block and IP address by running:

```bash 
  $ vault read atlas/roles/test
  
    Key                       Value
    ---                       -----
    cidr_blocks               [192.168.1.3/32]
    credential_type           project_programmatic_api_key
    database_name             n/a
    ip_addresses              [192.168.1.3]
    max_ttl                   0s
    organization_id           n/a
    programmatic_key_roles    [GROUP_CLUSTER_MANAGER]
    project_id                5cf5a45a9ccf6400e60981b6
    roles                     n/a
    ttl                       0s
```

  ```bash 
    $ vault read mongodbatlas/creds/test

    Key                Value
    ---                -----
    lease_id           mongodbatlas/creds/test/0fLBv1c2YDzPlJB1PwsRRKHR
    lease_duration     20s
    lease_renewable    true
    description        vault-test-1563980947-1318
    private_key        905ae89e-6ee8-40rd-ab12-613t8e3fe836
    public_key         klpruxce
  ```

## TTL and Max TTL

Database User and Programmatic API Keys Vault role can also have a time-to-live (TTL) and maximum time-to-live (Max TTL). When a credential expires and it's not renewed, it's automatically revoked. You can set the TTL and Max TTL for each role or globally using config/lease.

The following creates a Vault role "test" for a Project level Programmatic API key with a 2 hours time-to-live and a max time-to-live of 5 hours.

```bash 
$ vault write mongodbatlas/roles/test \
    credential_type=project_programmatic_api_key \
    project_id=5cf5a45a9ccf6400e60981b6 \
    programmatic_key_roles=GROUP_DATA_ACCESS_READ_ONLY \
    ttl=2h \
    max_ttl=5h
```

This then creates a credential with the lease time-to-live values:

```bash
$ vault read mongodbatlas/creds/test

    Key                Value
    ---                -----
    lease_id           mongodbatlas/creds/test/0fLBv1c2YDzPlJB1PwsRRKHR
    lease_duration     2h
    lease_renewable    true
    description        vault-test-1563980947-1318
    private_key        905ae89e-6ee8-40rd-ab12-613t8e3fe836
    public_key         klpruxce
```

You can verify the role that you have created with:

```bash
$ vault read mongodbatlas/roles/test   

    Key                       Value
    ---                       -----
    credential_type           org_programmatic_api_key
    database_name             n/a
    max_ttl                   5h0m0s
    organization_id           5b71ff2f96e82120d0aaec14
    programmatic_key_roles    [GROUP_DATA_ACCESS_READ_ONLY]
    project_id                5cf5a45a9ccf6400e60981b6
    roles                     n/a
    ttl                       2h0m0s
```

 ~> **Notice:** If you don't set the TTL and Max TTL when you are creating a role the default lease will be used if it was previously configured in the `mongodbatlas/config/lease` path. If a default was not created for MongoDB Atlas then Vault's default will be used.
