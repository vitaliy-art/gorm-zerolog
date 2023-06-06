package main

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	gormzerolog "github.com/vitaliy-art/gorm-zerolog"
	"gorm.io/gorm"
)

type User struct {
	*gorm.Model
	Name string
}

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormzerolog.NewGormLogger()})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&User{})
	user1 := &User{Name: "user1"}
	db.Create(user1)
	fmt.Printf("%s: %d\n\n", user1.Name, user1.ID)

	writer := zerolog.NewConsoleWriter()
	writer.TimeFormat = time.DateTime
	zeroLogger := zerolog.New(writer).With().Timestamp().Logger()

	logger := gormzerolog.NewGormLogger().WithInfo(func() gormzerolog.Event {
		return &gormzerolog.GormLoggerEvent{Event: zeroLogger.Info()}
	})

	db.Config.Logger = logger
	user2 := &User{Name: "user2"}
	db.Create(user2)
	fmt.Printf("%s: %d\n\n", user2.Name, user2.ID)
}
