package gintx

import (
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/sirupsen/logrus"
	"net/http"
)

const contextKeyNeo4JTransaction = "neo4j_tx"

func GetNeo4JTransaction(ctx *gin.Context) *neo4j.ExplicitTransaction {
	tx, ok := ctx.Get(contextKeyNeo4JTransaction)
	if !ok {
		return nil
	}

	neoTx := tx.(neo4j.ExplicitTransaction)
	return &neoTx
}

func SetNeo4JTransaction(ctx *gin.Context, tx neo4j.ExplicitTransaction) {
	ctx.Set(contextKeyNeo4JTransaction, tx)
}

func BuildNeo4JTransactionMiddleware(sessionConfig neo4j.SessionConfig, driver neo4j.DriverWithContext) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := driver.NewSession(ctx, sessionConfig)
		tx, err := session.BeginTransaction(ctx)
		if err != nil {
			logrus.WithField("error", err.Error()).Error("Cannot begin Neo4J transaction.")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		SetNeo4JTransaction(ctx, tx)

		defer func() {
			if err := recover(); err != nil {
				logrus.Info("Starting neo4j recovery process.")

				err := tx.Rollback(ctx)
				if err != nil {
					logrus.WithField("error", err.Error()).Error("Cannot rollback neo4j transaction.")
					return
				}
			} else if ctx.Writer.Status() >= http.StatusInternalServerError {
				logrus.Debug("Starting neo4j rollback due to internal server error status code.")

				err := tx.Rollback(ctx)
				if err != nil {
					logrus.WithField("error", err.Error()).Error("Cannot rollback neo4j transaction.")
					return
				}
			} else {
				logrus.Debug("Starting neo4j commit.")

				err := tx.Commit(ctx)
				if err != nil {
					logrus.WithField("error", err.Error()).Error("Cannot commit neo4j transaction.")
					return
				}
			}

			err := session.Close(ctx)
			if err != nil {
				logrus.WithField("error", err.Error()).Error("Cannot close neo4j session.")
				return
			}
		}()

		ctx.Next()
	}
}
