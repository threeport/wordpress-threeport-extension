// generated by 'threeport-sdk gen' - do not edit

package database

import (
	"context"
	"fmt"
	tp_database "github.com/threeport/threeport/pkg/api-server/v0/database"
	log "github.com/threeport/threeport/pkg/log/v0"
	util "github.com/threeport/threeport/pkg/util/v0"
	zap "go.uber.org/zap"
	postgres "gorm.io/driver/postgres"
	gorm "gorm.io/gorm"
	logger "gorm.io/gorm/logger"
	"os"
	"reflect"
	"strings"
	"time"
)

// ZapLogger is a custom GORM logger that forwards log messages to a Zap logger.
type ZapLogger struct {
	Logger *zap.Logger
}

// Init initializes the API database.
func Init(autoMigrate bool, logger *zap.Logger) (*gorm.DB, error) {
	dsn, err := GetDsn(false)
	if err != nil {
		return nil, fmt.Errorf("failed to populate DB DSN from environment: %w", err)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: &ZapLogger{Logger: logger},
		NowFunc: func() time.Time {
			utc, _ := time.LoadLocation("UTC")
			return time.Now().In(utc).Truncate(time.Microsecond)
		},
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// GetDsn returns the data source name string or an error if one of the required
// env vars is not set.  If the root user is requested the DSN will include the
// root user and reference the root user's SSL cert creds.
func GetDsn(rootDbUser bool) (string, error) {
	var dbEnvErrors util.MultiError

	requiredDbEnvVars := map[string]string{
		"DB_HOST":     "",
		"DB_NAME":     "",
		"DB_PORT":     "",
		"DB_SSL_MODE": "",
		"DB_USER":     "",
	}

	for env, _ := range requiredDbEnvVars {
		val, ok := os.LookupEnv(env)
		if !ok {
			dbEnvErrors.AppendError(fmt.Errorf("missing required environment variable: %s", env))
			requiredDbEnvVars[env] = ""
		} else {
			requiredDbEnvVars[env] = val
		}
	}

	dbUser := requiredDbEnvVars["DB_USER"]
	if rootDbUser {
		dbUser = "root"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s sslrootcert=%s/ca.crt sslcert=%[6]s/client.%[3]s.crt sslkey=%[6]s/client.%[3]s.key TimeZone=UTC",
		requiredDbEnvVars["DB_HOST"],
		requiredDbEnvVars["DB_PORT"],
		dbUser,
		requiredDbEnvVars["DB_NAME"],
		requiredDbEnvVars["DB_SSL_MODE"],
		tp_database.ThreeportApiDbCertsDir,
	)

	return dsn, dbEnvErrors.Error()
}

// LogMode overrides the standard GORM logger's LogMode method to set the logger mode.
func (zl *ZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	return zl
}

// Info overrides the standard GORM logger's Info method to forward log messages
// to the zap logger.
func (zl *ZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	fields := make([]zap.Field, 0, len(data)/2) // len(data)/2 because pairs of key-value

	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			// if there's no matching pair, log a warning or handle the error appropriately
			zl.Logger.Warn("Odd number of arguments passed to Info method", zap.Any("data", data))
			break
		}

		if reflect.TypeOf(data[i]).Kind() == reflect.Ptr {
			data[i] = fmt.Sprintf("%+v", data[i])
		}

		key, ok := data[i].(string)
		if !ok {
			zl.Logger.Warn("Key is not a string", zap.Any("key", data[i]))
			continue
		}

		fields = append(fields, zap.Any(key, data[i+1]))
	}
	zl.Logger.Info(msg, fields...)
}

// Warn overrides the standard GORM logger's Warn method to forward log messages
// to the zap logger.
func (zl *ZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	fields := make([]zap.Field, 0, len(data)/2) // len(data)/2 because pairs of key-value

	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			// if there's no matching pair, log a warning or handle the error appropriately
			zl.Logger.Warn("Odd number of arguments passed to Warn method", zap.Any("data", data))
			break
		}

		if reflect.TypeOf(data[i]).Kind() == reflect.Ptr {
			data[i] = fmt.Sprintf("%+v", data[i])
		}

		key, ok := data[i].(string)
		if !ok {
			zl.Logger.Warn("Key is not a string", zap.Any("key", data[i]))
			continue
		}

		fields = append(fields, zap.Any(key, data[i+1]))
	}
	zl.Logger.Warn(msg, fields...)
}

// Error overrides the standard GORM logger's Error method to forward log messages
// to the zap logger.
func (zl *ZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	fields := make([]zap.Field, 0, len(data)/2) // len(data)/2 because pairs of key-value

	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			// if there's no matching pair, log a warning or handle the error appropriately
			zl.Logger.Warn("Odd number of arguments passed to Error method", zap.Any("data", data))
			break
		}

		if reflect.TypeOf(data[i]).Kind() == reflect.Ptr {
			data[i] = fmt.Sprintf("%+v", data[i])
		}

		key, ok := data[i].(string)
		if !ok {
			zl.Logger.Warn("Key is not a string", zap.Any("key", data[i]))
			continue
		}

		fields = append(fields, zap.Any(key, data[i+1]))
	}
	zl.Logger.Error(msg, fields...)
}

// Trace overrides the standard GORM logger's Trace method to forward log messages
// to the zap logger.
func (zl *ZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	// use the fc function to get the SQL statement and execution time
	sql, rows := fc()

	// create a new logger with some additional fields
	logger := zl.Logger.With(
		zap.String("type", "sql"),
		zap.String("sql", suppressSensitive(sql)),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", time.Since(begin)),
	)

	// if an error occurred, add it as a field to the logger
	if err != nil {
		logger = logger.With(zap.Error(err))
	}

	// log the message using the logger
	logger.Debug("gorm query")
}

// suppressSensitive supresses messages containing sesitive strings.
func suppressSensitive(msg string) string {
	for _, str := range log.SensitiveStrings() {
		if strings.Contains(msg, str) {
			return fmt.Sprintf("[log message containing %s supporessed]", str)
		}
	}

	return msg
}
