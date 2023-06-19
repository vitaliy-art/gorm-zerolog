# Zerolog GORM logger
Simple zerolog implementation of GORM logger

# Usage

```go
import (
	"github.com/glebarez/sqlite"
	gormzerolog "github.com/vitaliy-art/gorm-zerolog"
	"gorm.io/gorm"
)

db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
    Logger: gormzerolog.NewGormLogger()
})
```

# Example with logger customization

```go
import (
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	gormzerolog "github.com/vitaliy-art/gorm-zerolog"
	"gorm.io/gorm"
)

writer := zerolog.NewConsoleWriter()
writer.TimeFormat = time.DateTime
zeroLogger := zerolog.New(writer).With().Timestamp().Logger()

logger := gormzerolog.NewGormLogger().WithInfo(func() gormzerolog.Event {
    return &gormzerolog.GormLoggerEvent{Event: zeroLogger.Info()}
})

db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
    Logger: logger,
})
```
