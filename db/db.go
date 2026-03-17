package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	Config DBConfig `yaml:",inline"`
	Conn   *sqlx.DB
}

type DBConfig struct {
	Host     string `yaml:"host" validate:"required"`
	Port     int    `yaml:"port" validate:"required"`
	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Name     string `yaml:"name" validate:"required"`
	SSLMode  string `yaml:"sslMode" validate:"required"`
}

func (db *DB) Init() {
	db.newDB()
}

func (db *DB) newDB() {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Config.Host,
		db.Config.Port,
		db.Config.User,
		db.Config.Password,
		db.Config.Name,
		db.Config.SSLMode,
	)
	newDB, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	err = newDB.Ping()
	if err != nil {
		panic(err)
	}

	newDB.SetMaxOpenConns(25)
	newDB.SetMaxIdleConns(25)

	db.Conn = newDB
}
