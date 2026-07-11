package helpers

import (
	"clean-template/migrate"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetupMySQL() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", GetEnv("DB_USER", ""), GetEnv("DB_PASSWORD", ""), GetEnv("DB_HOST", "127.0.0.1"), GetEnv("DB_PORT", "3306"), GetEnv("DB_NAME", ""))

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	logrus.Info("Successfully connect to database..")

	DB.AutoMigrate(migrate.MigrateList()...)
}
