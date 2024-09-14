# Zookeeper framework for high level recipes

High level recipes:

- framework reconnection
- simplified node operations
- watchers
- cache
- locks
- config set and encryption
- group and leadership
- ...

### TODO

- Consistent/pluggable logger

## module `framework`

Baseline connection manager with reconnection capability

### TODO

More connection options, in particular:

- Confidential connection using TLS
- Authenticated connection
- Create framework with context

## module `operation`

Baseline CRUD operations on nodes

### TODO

- On get/exists: stats

## module `watchers`

Monitor and notify node changes
