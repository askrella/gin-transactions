package gintx

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
)

const contextKeyGormTransaction = "gorm_tx"

func GetGormTransaction(ctx *gin.Context) *gorm.DB {
	tx, ok := ctx.Get(contextKeyGormTransaction)
	if !ok {
		return nil
	}

	gormTx := tx.(*gorm.DB)
	return gormTx
}

func SetGormTransaction(ctx *gin.Context, tx *gorm.DB) {
	ctx.Set(contextKeyGormTransaction, tx)
}

func BuildGormTransactionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tx := db.Begin()
		if tx.Error != nil {
			logrus.WithField("error", tx.Error.Error()).Error("Cannot begin Gorm transaction.")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		SetGormTransaction(ctx, tx)

		defer func() {
			if err := recover(); err != nil {
				encodedError, marshalError := json.Marshal(err)
				if marshalError != nil {
					encodedError = []byte("MARSHAL_ERROR")
				}

				logrus.WithField("error", string(encodedError)).Info("Starting Gorm recovery process.")

				err := tx.Rollback().Error
				if err != nil {
					logrus.WithField("error", err.Error()).Error("Cannot rollback Gorm transaction.")
					return
				}
			} else if ctx.Writer.Status() >= http.StatusInternalServerError {
				logrus.Debug("Starting Gorm rollback due to internal server error status code.")

				err := tx.Rollback().Error
				if err != nil {
					logrus.WithField("error", err.Error()).Error("Cannot rollback Gorm transaction.")
					return
				}
			} else {
				logrus.Debug("Starting Gorm commit.")

				err := tx.Commit().Error
				if err != nil {
					logrus.WithField("error", err.Error()).Error("Cannot commit Gorm transaction.")
					return
				}
			}
		}()

		ctx.Next()
	}
}
