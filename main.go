// Copyright 2021 essquare GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"bookstore.app/api"
	"bookstore.app/database"
	"bookstore.app/model"
	"bookstore.app/storage"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	flagSQLiteFileHelp              = "SQLite file"
	flagListenAddrHelp              = "HTTP Server listen address"
	flagMigrateDBHelp               = "Run database migration"
	flagCreateAdminHelp             = "Create admin user"
	flagCreateAdminUserNameHelp     = "Admin user name"
	flagCreateAdminUserPasswordHelp = "Admin user password"
)

func main() {

	var flagSQLiteFile string
	var flagListenAddr string
	var flagMigrateDB bool
	var flagCreateAdmin bool
	var flagCreateAdminUsername string
	var flagCreateAdminPassword string

	flag.StringVar(&flagSQLiteFile, "sqlite-file", "bookstore.sqlite", flagSQLiteFileHelp)
	flag.StringVar(&flagSQLiteFile, "s", "bookstore.sqlite", flagSQLiteFileHelp)

	flag.StringVar(&flagListenAddr, "listen-address", "0.0.0.0:8080", flagListenAddrHelp)
	flag.StringVar(&flagListenAddr, "l", "0.0.0.0:8080", flagListenAddrHelp)

	flag.BoolVar(&flagMigrateDB, "migrate-database", false, flagMigrateDBHelp)
	flag.BoolVar(&flagMigrateDB, "m", false, flagMigrateDBHelp)

	flag.BoolVar(&flagCreateAdmin, "create-admin", false, flagCreateAdminHelp)

	flag.StringVar(&flagCreateAdminUsername, "create-admin-username", "", flagCreateAdminUserNameHelp)
	flag.StringVar(&flagCreateAdminPassword, "create-admin-password", "", flagCreateAdminUserPasswordHelp)

	flag.Parse()

	db, err := database.NewDatabaseConnection(flagSQLiteFile)
	if err != nil {
		log.Fatalf("Unable to initialize database connection pool: %v", err)
	}
	defer db.Close()

	store := storage.NewStorage(db)
	if err := store.Ping(); err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}

	if flagMigrateDB {
		if err := database.Migrate(db); err != nil {
			log.Fatalf(`%v`, err)
		}
		return
	}
	if err := database.CurrentDBSchema(db); err != nil {
		log.Fatalf("You must run the SQL migrations, %v", err)
	}

	if flagCreateAdmin {
		if flagCreateAdminUsername == "" || flagCreateAdminPassword == "" {
			log.Fatal("Admin user create called, but username or password not supplied")
		}
		user, err := store.CreateUser(&model.UserCreationRequest{
			Username:  flagCreateAdminUsername,
			Password:  flagCreateAdminPassword,
			Pseudonym: flagCreateAdminUsername,
			IsAdmin:   true,
		})
		if err != nil {
			log.Fatalf("Admin user create failed, %v", err)
		}
		log.Infof("Admin user with ID: %d, created!", user.ID)
		return

	}
	r := mux.NewRouter()

	api.Serve(r, store)
	httpServer := &http.Server{
		Addr:         flagListenAddr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handlers.RecoveryHandler()(r),
	}

	go func() {
		log.Infof("Listening on %q without TLS", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Infof("Server failed to start: %v", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	<-stop
	log.Info("Shutting down the process...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	httpServer.Shutdown(ctx)

	log.Info("Process gracefully stopped")
}
