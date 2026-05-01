package db

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Connection() {
	db, err := sql.Open("pgx", config.Envs.DatabaseURL)
	if err != nil {
		panic(err)
	}

	DB = db
}

func PingConnection(shortenUrlService *service.ShortenURLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := DB.PingContext(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
