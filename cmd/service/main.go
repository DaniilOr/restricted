package main

import (
	"context"
	"github.com/DaniilOr/restricted/cmd/service/app"
	"github.com/DaniilOr/restricted/pkg/business"
	"github.com/DaniilOr/restricted/pkg/payments"
	"github.com/DaniilOr/restricted/pkg/security"
	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"

	//"context"
	//"github.com/jackc/pgx/v4/pgxpool"
	//"log"
	//"net"
	//"net/http"
	//"os"
	//"service/cmd/service/app"
	//"service/pkg/business"
	//"service/pkg/security"
	"net"
	"os"

)

const (
	defaultPort = "9999"
	defaultHost = "0.0.0.0"
	defaultDSN  = "postgres://app:pass@localhost:5432/db"
)

func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	dsn, ok := os.LookupEnv("APP_DSN")
	if !ok {
		dsn = defaultDSN
	}

	if err := execute(net.JoinHostPort(host, port), dsn); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, dsn string) error {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		log.Print(err)
		return err
	}
	defer pool.Close()

	securitySvc := security.NewService(pool)
	businessSvc := business.NewService(pool)
	paymentsSvc := payments.NewService(pool)
	router := chi.NewRouter()
	application := app.NewServer(securitySvc, businessSvc, paymentsSvc, router)
	err = application.Init()
	if err != nil {
		log.Print(err)
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}