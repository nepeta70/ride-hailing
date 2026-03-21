package mongodb

import (
	"net/url"
	"strconv"
	"time"

	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
)

type MongoConfig struct {
	User         string        `json:"user" env:"MONGO_USER"`
	Password     string        `json:"password" env:"MONGO_PASSWORD"`
	Database     string        `json:"database" env:"MONGO_DB"`
	Host         string        `json:"host" env:"MONGO_HOST"`
	Port         int           `json:"port" env:"MONGO_PORT"`
	PingSeconds  int           `json:"ping_timeout_seconds" env:"MONGO_PING_TIMEOUT_SECONDS"`
	PingTimeout  time.Duration `json:"-"`
	QuerySeconds int           `json:"query_timeout_seconds" env:"MONGO_QUERY_TIMEOUT_SECONDS"`
	QueryTimeout time.Duration `json:"-"`
}

// mongodb://${MONGO_USER}:${MONGO_PASSWORD}@localhost:27017
func (c *MongoConfig) GetURI() string {
	userInfo := url.UserPassword(c.User, c.Password).String()
	return "mongodb://" + userInfo + "@" + c.Host + ":" + strconv.Itoa(c.Port) + "/" + c.Database + "?authSource=admin"
}

func (c *MongoConfig) Validate() error {
	if c.Host == "" {
		return errors.NewValidationErrorf("invalid host")
	}
	if c.Port == 0 {
		return errors.NewValidationErrorf("invalid port")
	}
	if c.Database == "" {
		return errors.NewValidationErrorf("invalid database")
	}
	if c.User == "" {
		return errors.NewValidationErrorf("invalid username")
	}
	if c.Password == "" {
		return errors.NewValidationErrorf("invalid password")
	}
	return nil
}

func (c *MongoConfig) Init() error {
	c.PingTimeout = time.Duration(c.PingSeconds) * time.Second
	c.QueryTimeout = time.Duration(c.QuerySeconds) * time.Second
	return nil
}

func DefaultMongoConfig() MongoConfig {
	return MongoConfig{
		User:         "",
		Password:     "",
		Database:     "",
		Host:         "localhost",
		Port:         27017,
		PingSeconds:  5,
		QuerySeconds: 5,
	}
}
