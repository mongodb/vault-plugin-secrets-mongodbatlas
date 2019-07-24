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

1. `database_user`: Vault will create an database user for each lease, each user has a set of roles
   that provide access to the project’s databases. And then return the username and password to the
   caller. 
   [databaseUsers](https://docs.atlas.mongodb.com/reference/api/database-users/)
2. `programmatic_api_key`: Vault will call
   [apiKeys](https://docs.atlas.mongodb.com/reference/api/apiKeys-orgs-create-one/)
   and return the public key, secret key.

## Setup

Most secrets engines must be configured in advance before they can perform their
functions. These steps are usually completed by an operator or configuration
management tool.

1. Enable the MongoDB Atlas Secrets Engine:

    ```text
    $ vault secrets enable mongodbatlas
    Success! Enabled the mongodbatlas secrets engine at: mongodbatlas/
    ```

    By default, the secrets engine will mount at the name of the engine. To
    enable the secrets engine at a different path, use the `-path` argument.

1. Configure the credentials that Vault uses to communicate with AWS to generate
the IAM credentials:

    ```text
    $ vault write mongodbatlas/config/root \
        public_key=yhltsvan \
        private_key=2c130c23-e6b6-4da8-a93f-a8bf33218830
    ```

    Internally, Vault will connect to MongoDB Atlas using these credentials. As such,
    these credentials must be a superset of any policies which might be granted
    on API Keys. Since Vault uses the official [MongoDB Atlas Client](https://github.com/mongodb/go-client-mongodb-atlas), it will use the specified credentials. 

    ~> **Notice:** Even though the path above is `mongodbatlas/config/root`, do not use
    your MongoDB Atlas root account credentials. Instead generate a dedicated dedicated  Programmatic API key with appropriate roles.

1. Configure a Vault role that maps to a set of permissions in MongoDB Atlas as well as 
   a MongoDB Atlas credentials/keys. When users generate credentials, they are generated
   against this role. An example:

    ```text
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
    an username and password for the database user and return these credentials.

    ```text
    $ vault read mongodbatlas/creds/test
        Key                Value
        ---                -----
        lease_id           mongodbatlas/creds/test/sZz3qvggwcULgERDqe9r151h
        lease_duration     20s
        lease_renewable    true
        password           A3mOyxvSnBpKG5sdID2iNR
        username           vault-test-1563475091-2081
    ```

    For more information on database user roles, please see the
    [MongoDB Atlas documentation](https://docs.atlas.mongodb.com/reference/api/database-users-create-a-user/).

## Programmatic API Keys
  To grant programmatic access to an organization or project using only the API, you can create an API key. API keys:

  Have two parts: a Public Key and a Private Key.
  Cannot be used to log into Atlas through the user interface.
  Must be granted roles as you would Users to make sure the API Keys can call API endpoints without errors.
  Can belong to one organization, but may be granted access to any number of projects in that organization.


  Most secrets engines must be configured in advance before they can perform their
  functions. These steps are usually completed by an operator or configuration
  management tool.


  ~> **Notice:** Even though the path above is `atlas/config/root`, do not use
  your MongoDB Atlas root account credentials. Instead generate a dedicated user or
  role.


1. Configure a Vault role that maps to a set of permissions in MongoDB Atlas as well as an
   MongoDB Atlas credential type. When users generate credentials, they are generated
   against this role and can use `organization_id` or `project_id` (if you will want to bind the Programmatic API key to a project, use `project_id` instead of `organization_id` and `credential_type` should be project_programmatic_api_key). An example:

    ```bash
    $ vault write atlas/roles/test \
        credential_type=org_programmatic_api_key \
        organization_id=5b23ff2f96e82130d0aaec13 \
        programmatic_key_roles=ORG_MEMBER
    ```

    or

     ```bash
    $ vault write atlas/roles/test \
        credential_type=project_programmatic_api_key \
        project_id=5b23ff2f96e82130d0aaec13 \
        programmatic_key_roles=ORG_MEMBER
    ```

  ~> **Notice:**  Vault will create an `org_programmatic_api_key` or `project_programmatic_api_key` depending on the configuration as in the previous example.

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

