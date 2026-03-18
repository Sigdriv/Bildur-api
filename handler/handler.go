package handler

import (
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Sigdriv/Bildur-api/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Handler struct {
	Config    Config `yaml:",inline"`
	DB        db.DB  `yaml:"db" validate:"required"`
	Validator *validator.Validate
}

type Config struct {
	Port string `yaml:"port" validate:"required"`
}

func CreateHandler() (srv Handler) {
	data, err := os.ReadFile("./cfg/cfg.yml")
	if err != nil {
		logrus.Fatalf("Failed reading config file >> %v", err)
	}

	err = yaml.Unmarshal(data, &srv)
	if err != nil {
		logrus.Fatalf("Failed unmarshalling config file >> %v", err)
	}

	srv.resolveFileReferences(&srv)

	srv.Validator = validator.New()
	err = srv.Validator.Struct(&srv.Config)
	if err != nil {
		logrus.Fatalf("Config validation failed >> %v", err)
	}

	srv.DB.Init()

	return
}

func (srv *Handler) CreateGinGroup() {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logrus.WithFields(logrus.Fields{
			"status":  c.Writer.Status(),
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"latency": time.Since(start),
		}).Info("Request handled")
	})

	router.Use(configureCors())

	router.GET("/images", srv.HandleGetImages)
	router.Static("/media/thumb", "./media/thumbnails")

	runner := fmt.Sprintf("localhost:%s", srv.Config.Port)
	router.Run(runner)
}

func configureCors() gin.HandlerFunc {
	local := fmt.Sprintf("http://%s:3000", getLocalIP())

	return cors.New(cors.Config{
		AllowOrigins:     []string{local, "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

func getLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

// resolveFileReferences recursively iterates through all fields in the config
// and replaces string values starting with "file::" with the content of the file
func (srv *Handler) resolveFileReferences(config interface{}) {
	srv.resolveFileReferencesRecursive(reflect.ValueOf(config))
}

func (srv *Handler) resolveFileReferencesRecursive(v reflect.Value) {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	// Only process structs
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			// Process string fields with "file::" prefix
			value := field.String()
			if len(value) > 6 && value[:6] == "file::" {
				filePath := value[6:]
				correctPath := fmt.Sprintf("./cfg/%s", strings.Replace(filePath, "./", "", 1))

				fileData, err := os.ReadFile(correctPath)
				if err != nil {
					logrus.Fatalf("Failed to read file for field %s: %v", fieldType.Name, err)
				}

				// Trim whitespace/newlines from file content
				field.SetString(strings.TrimSpace(string(fileData)))
				logrus.Infof("Loaded %s from file: %s", fieldType.Name, correctPath)
			}

		case reflect.Struct:
			// Recursively process nested structs
			srv.resolveFileReferencesRecursive(field)

		case reflect.Ptr:
			// Recursively process pointer fields
			if !field.IsNil() {
				srv.resolveFileReferencesRecursive(field)
			}
		}
	}
}

func (*Handler) getLog(c *gin.Context) *logrus.Logger {
	url := c.Request.URL
	logger := logrus.New()

	logger.WithFields(logrus.Fields{
		"method": c.Request.Method,
		"url":    url.String(),
	})

	return logger
}
