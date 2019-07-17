---
layout: "docs"
page_title: "Atlas MongoDB - Secrets Engines"
sidebar_title: "Atlas MongoDB"
sidebar_current: "docs-secrets-atlasmongodb"
description: |-
  The Atlas MongoDB secrets engine for Vault generates database user keys dynamically based and Programmatic API Keys.
---

# Atlas MongoDB Secrets Engine

The Atlas MongoDB secrets engine generates database user acces keys and Programmatic
API keys based on their respective roles.  This generally makes working with Atlas 
MongoDB, since it does not involve clicking in the web UI. Additionally, the process 
is codified and mapped to internal auth methods (such as LDAP). The Atlas MongoDB credentials 
are time-based and are automatically revoked when the Vault lease expires.

Vault supports two different types of credentials to retrieve from Atlas MongoDB:

1. `database_user`: Vault will create an database user for each lease, each user has a set of roles 
   that provide access to the project’s databases. And then return the username and password to the caller. 
   [databaseUsers](https://docs.atlas.mongodb.com/reference/api/database-users/)
2. `programmatic_api_key`: Vault will call
   [apiKeys](https://docs.atlas.mongodb.com/reference/api/apiKeys-orgs-create-one/)
   and return the public key, secret key.

## Setup

Most secrets engines must be configured in advance before they can perform their
functions. These steps are usually completed by an operator or configuration
management tool.

1. Enable the Atlas MongoDB secrets engine:

    ```text
    $ vault secrets enable mongodbatlas
    Success! Enabled the aws secrets engine at: mongodbatlas/
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

    Internally, Vault will connect to Atlas MongoDB using these credentials. As such,
    these credentials must be a superset of any policies which might be granted
    on API Keys. Since Vault uses the official Atlas MongoDB Client, it will use the
    specified credentials. 

    ~> **Notice:** Even though the path above is `mongodbatlas/config/root`, do not use
    your Atlas MongoDB root account credentials. Instead generate a dedicated user or
    role.

1. Configure a Vault role that maps to a set of permissions in AWS as well as an
   AWS credential type. When users generate credentials, they are generated
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
    this role, Vault will create an database user and attach the specified roles
    database user. Vault will then create an username and password for the database 
    user and return these credentials. 

    For more information on database user roles, please see the
    [Atlas MongoDB documentation](https://docs.atlas.mongodb.com/reference/api/database-users-create-a-user/).

