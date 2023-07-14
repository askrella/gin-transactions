# gin-transactions

The "gin-transactions" library, developed by [Askrella](https://askrella.de), provides a middleware that seamlessly integrates database transactions into the gin context. Our middleware injects the database transaction into the gin context, enabling developers to effortlessly retrieve and utilize the transaction within their routes.

# Database support

- Neo4j
- Gorm

# Installation
```
go get -u github.com/askrella/gin-transactions
```

# Register gin middleware
```
gintx.BuildNeo4JTransactionMiddleware
gintx.BuildGormTransactionMiddleware
```

# Retrieve database transaction
```
gintx.GetNeo4JTransaction
gintx.GetGormTransaction
```
