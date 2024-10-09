package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/kensonjohnson/greenlight/internal/data"
	"github.com/kensonjohnson/greenlight/internal/mailer"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("error loading .env file")
	}

	var cfg config

	// env
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// database
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgresql://greenlight:password@localhost/greenlight?sslmode=disable", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// rate limit
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	smtpHost, ok := os.LookupEnv("SMTP_HOST")
	if !ok {
		panic("environment variable SMTP_HOST not set")
	}
	cfg.smtp.host = smtpHost

	smtpPort, ok := os.LookupEnv("SMTP_PORT")
	if !ok {
		panic("environment variable SMTP_PORT not set")
	}
	smtpPortConverted, err := strconv.ParseInt(smtpPort, 10, 0)
	if err != nil {
		panic("invalid SMTP_PORT value")
	}
	cfg.smtp.port = int(smtpPortConverted)

	smtpUsername, ok := os.LookupEnv("SMTP_USERNAME")
	if !ok {
		panic("environment variable SMTP_USERNAME not set")
	}
	cfg.smtp.username = smtpUsername

	smtpPassword, ok := os.LookupEnv("SMTP_PASSWORD")
	if !ok {
		panic("environment variable SMTP_PASSWORD not set")
	}
	cfg.smtp.password = smtpPassword

	smtpSender, ok := os.LookupEnv("SMTP_SENDER")
	if !ok {
		panic("environment variable SMTP_SENDER not set")
	}
	cfg.smtp.sender = smtpSender

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(
			cfg.smtp.host,
			cfg.smtp.port,
			cfg.smtp.username,
			cfg.smtp.password,
			cfg.smtp.sender,
		),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxOpenConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
