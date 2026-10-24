package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type application struct {
	config     *config
	db         *sql.DB
	httpServer *http.Server
	handler    *handler
}

func newApplication(cfg *config) (*application, error) {
	app := new(application)
	app.config = cfg

	if err := app.initPostgres(); err != nil {
		return nil, err
	}

	if err := app.initHttpServer(); err != nil {
		return nil, err
	}

	if err := app.setupHandler(); err != nil {
		return nil, err
	}

	return app, nil
}

func (app *application) initHttpServer() error {
	router := gin.Default()

	app.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", app.config.http.host, app.config.http.port),
		Handler: router,
	}

	return nil
}

func (app *application) initPostgres() error {
	conf := app.config.database

	db, err := sql.Open("postgres", conf.dsn())
	if err != nil {
		return err
	}

	log.Println("postgres connection established")

	app.db = db

	return nil
}

func (app *application) setupHandler() error {
	svc := newVehicleService(app.db)
	h := newHandler(svc)
	h.registerRoute(app.httpServer.Handler.(*gin.Engine))

	return nil
}

func (app *application) start(ctx context.Context) error {
	log.Println("starting http server")
	if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (app *application) shutdown(ctx context.Context) error {
	if err := app.httpServer.Shutdown(ctx); err != nil {
		log.Println("error while shutting down server: ", err)
	}
	log.Println("http server closed")

	if err := app.db.Close(); err != nil {
		log.Println("error while closing postgres: ", err)
	}
	log.Println("postgres closed")

	return nil
}
