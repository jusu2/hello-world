package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm/schema"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Message struct {
	ID        uint      `gorm:"primaryKey"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type DbConnectionConfig struct {
	Hostname string `yaml:"hostname"`
	Database string `yaml:"database"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	MaxConn  int    `yaml:"max-conn"`
	MaxOpen  int    `yaml:"max_open"`
	LogLevel int    `yaml:"log_level"`
	Schema   string `yaml:"schema"`
	Timeout  int    `yaml:"timeout"`
}

var DB *gorm.DB

func CreateDBConnPool(dbconfig *DbConnectionConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(BuildDSN(dbconfig.Hostname, dbconfig.Database, dbconfig.User, dbconfig.Password,
		dbconfig.Port, dbconfig.Timeout)), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.LogLevel(dbconfig.LogLevel)),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   dbconfig.Schema + ".",
			SingularTable: false,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("DB open failed, %v", err)
	}
	err = db.AutoMigrate(&Message{})
	if err != nil {
		return nil, fmt.Errorf("auto migration failed, %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get wrapped sql db failed, %v", err)
	}
	sqlDB.SetMaxIdleConns(dbconfig.MaxConn)
	sqlDB.SetMaxOpenConns(dbconfig.MaxOpen)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func BuildDSN(host string, database string, user, password string, port int, timeout int) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s connect_timeout=%d", host, port,
		user, database, password, timeout)
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func Init() error {
	cfg := &DbConnectionConfig{
		Hostname: os.Getenv("DB_HOST"),
		Database: os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Port:     atoi(os.Getenv("DB_PORT")),
		Schema:   os.Getenv("DB_SCHEMA"),
		MaxConn:  atoi(os.Getenv("DB_MAX_IDLE")),
		MaxOpen:  atoi(os.Getenv("DB_MAX_OPEN")),
		Timeout:  atoi(os.Getenv("DB_TIMEOUT")),
		LogLevel: atoi(os.Getenv("DB_LOG_LEVEL")),
	}

	var err error
	DB, err = CreateDBConnPool(cfg)
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}
	return nil
}

func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func SaveMessage(content string) error {
	msg := Message{Content: content}
	result := DB.Create(&msg)
	return result.Error
}