# Zookeeper framework for high level recipes

The objective (actually a playground) is to bring Go into Cognitive to replace JVM simple components:

- Push connector
- Presentation servers
- ...

## module `framework`

Baseline connection manager with reconnection capability

### TODO

More connection options, in particular:

- Confidential connection using TLS
- Authenticated connection

## module `operation`

Baseline CRUD operations on nodes

### TODO

- On creation: node types, data, ACL
- On get/exists: stats

## module `watchers`

Monitor and notify node changes

## module `recipes`

High level recipes:

- cache
- locks
- config set and encryption
- group and leadership
- ...
