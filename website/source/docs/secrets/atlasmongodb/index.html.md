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
created MongoDB Atlas secrets are time-based and are automatically revoked when the Vault lease expires, unless renewed.

This Secrets Engine supports two different types of MongoDB Atlas Secrets:

1. `database_user`: Vault will create a database user for each lease, each user has a defined role(s) that provide appropriate access to the project’s databases. The username and password is returned to the caller. To see more about database users visit: https://docs.atlas.mongodb.com/reference/api/database-users/
2. `programmatic_api_key`: Vault will call
   [apiKeys](https://docs.atlas.mongodb.com/reference/api/apiKeys-orgs-create-one/)
   and return the public key, secret key.

## Setup

Most Secrets Engines must be configured in advance before they can perform their
functions. These steps are usually completed by an operator or configuration
management tool.

  ~> **Notice:** The following will be accurate after review and approval by Hashicorp, which is in progress. Until then follow the instructions in the [README developing section](./../../../../../README.md).


1. Enable the MongoDB Atlas Secrets Engine:

    ```bash
    $ vault secrets enable mongodbatlas
    Success! Enabled the mongodbatlas Secrets Engine at: mongodbatlas/
    ```

    By default, the Secrets Engine will mount at the name of the engine. To
    enable the Secrets Engine at a different path, use the `-path` argument.

1. It's necessary to generate and configure an API key for your organization for the acceptance test to succeed. To grant programmatic access to an organization or project using only the [API](https://docs.atlas.mongodb.com/api/) you need to know:

    1. The programmatic API key has two parts: a Public Key and a Private Key. To see more details on how to create a programmatic API key visit https://docs.atlas.mongodb.com/configure-api-access/#programmatic-api-keys.
    
    1. The programmatic API key must be granted roles sufficient for the acceptance test to succeed. The Organization Owner and Project Owner roles should be sufficient. You can see the available roles at https://docs.atlas.mongodb.com/reference/user-roles.

    1. You must [configure Atlas API Access](https://docs.atlas.mongodb.com/configure-api-access/) for your programmatic API key. You should allow API access for the IP address from which the acceptance test runs.

1. Later you must configure the MongoDB credentials/keys created in the previous step, which Vault will use to communicate with MongoDB Atlas:

    ```bash
    $ vault write mongodbatlas/config/root \
        public_key=yhltsvan \
        private_key=2c130c23-e6b6-4da8-a93f-a8bf33218830
    ```

    Internally, Vault will connect to MongoDB Atlas using these credentials. As such,
    these credentials must be a superset of any policies which might be granted
    on API Keys. Since Vault uses the official [MongoDB Atlas Client](https://github.com/mongodb/go-client-mongodb-atlas), it will use the specified credentials. 

    ~> **Notice:** Even though the path above is `mongodbatlas/config/root`, do not use
    your MongoDB Atlas root account credentials. Instead generate a dedicated Programmatic API key with appropriate roles.

1. Configure a Vault role that maps to a set of permissions in MongoDB Atlas as well as 
   a MongoDB Atlas credentials/keys. When users generate credentials, they are generated
   against this role. An example:

    ```bash
    $ vault write mongodbatlas/roles/test \
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

    This creates a role named "my-role". When users generate credentials against
    this role, Vault will create a database user and attach the specified roles to that
    database user mapped to the appropriate database/collection. Vault will then create 
    a username and password for the database user and return these credentials.

    ```bash
    $ vault read mongodbatlas/creds/test
        Key                Value
        ---                -----
        lease_id           mongodbatlas/creds/test/sZz3qvggwcULgERDqe9r151h
        lease_duration     20s
        lease_renewable    true
        password           A3mOyxvSnBpKG5sdID2iNR
        username           vault-test-1563475091-2081
    ```

    For more information on database user roles, please see
    [step two of the setup section](#Setup).

## Programmatic API Keys


  One may grant programmatic access to MongoDB Atlas by creating a Programmatic API key with access to a organization and project(s).
  Programmatic API Keys:
  - Have two parts, a public key and a private key
  - Cannot be used to log into Atlas through the user interface
  - Must be granted appropriate roles to complete required tasks
  - Must belong to one organization, but may be granted access to any number of projects in that organization.
  - May have an IP whitelist configured and some capabilities may require a whitelist to be configured (these are noted in the MongoDB Atlas API documentation).



  Most Secrets Engines must be configured in advance before they can perform their
  functions. These steps are usually completed by an operator or configuration
  management tool.


  ~> **Notice:** Even though the path above is `atlas/config/root`, do not use
  your MongoDB Atlas root account credentials. Instead generate a dedicated user or
  role.


1. Create a Vault role for a Programmatic API key by mapping a Programmatic API key role(s) to a organization or project in MongoDB Atlas.
- If the key/role is for the MongoDB Atlas Organization level use organization_id with the appropriate Id and roles
- If the key/role is for the MongoDB Atlas Project level use project_id with the appropriate Id and roles

~> **Notice:** Programmatic API keys can belong to only one Organization but can belong to one or more Projects. An examples:

```bash
$ vault write atlas/roles/test \
    credential_type=org_programmatic_api_key \
    organization_id=5b23ff2f96e82130d0aaec13 \
    programmatic_key_roles=ORG_MEMBER
```
```bash 
$ vault write atlas/roles/test \
    credential_type=project_programmatic_api_key \
    project_id=5cf5a45a9ccf6400e60981b6 \
    programmatic_key_roles=GROUP_DATA_ACCESS_READ_ONLY
```

```bash 
$ vault write atlas/roles/test \
    credential_type=project_programmatic_api_key \
    project_id=5cf5a45a9ccf6400e60981b6 \
    programmatic_key_roles=GROUP_CLUSTER_MANAGER
```

  ~> **Notice:**  The above examples creates two roles in Vault for Programmatic API keys. The first one is created at the [Organization](https://docs.atlas.mongodb.com/configure-api-access/) level with a role of ORG_MEMBER. The second example creates a Programmatic API key for the specified Project and grants access only GROUP_DATA_ACCESS_READ_ONLY.

   This creates a set of Programmatic API keys that is attached to an [Organization](https://docs.atlas.mongodb.com/configure-api-access/#view-the-details-of-an-api-key-in-an-organization), if `project_id` is used, is attached to a [Project](https://docs.atlas.mongodb.com/configure-api-access/#manage-programmatic-access-to-a-project).

  ```bash 
    $ vault read atlas/creds/test

    Key                Value
    ---                -----
    lease_id           atlas/creds/test/0fLBv1c2YDzPlJB1PwsRRKHR
    lease_duration     20s
    lease_renewable    true
    description        vault-test-1563980947-1318
    private_key        905ae89e-6ee8-40rd-ab12-613t8e3fe836
    public_key         klpruxce
  ```

## TTL and Max TTL


Every role has a time-to-live (TTL) and maximum time-to-live (Max TTL) which is used to validate the TTL.  When a role expires and it's not renewed, the role is automatically revoked. You can set the TTL and Max TTL for each created role.
```bash 
$ vault write atlas/roles/test \
    credential_type=project_programmatic_api_key \
    project_id=5cf5a45a9ccf6400e60981b6 \
    programmatic_key_roles=GROUP_DATA_ACCESS_READ_ONLY \
    ttl=2h \
    max_ttl=5h
```

This creates a role that is attached to an associated lease:
```bash
$ vault read atlas/creds/test

    Key                Value
    ---                -----
    lease_id           atlas/creds/test/0fLBv1c2YDzPlJB1PwsRRKHR
    lease_duration     2h
    lease_renewable    true
    description        vault-test-1563980947-1318
    private_key        905ae89e-6ee8-40rd-ab12-613t8e3fe836
    public_key         klpruxce
```

You can verify the role that you have created before with:
```bash
$ vault read atlas/roles/test   
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
