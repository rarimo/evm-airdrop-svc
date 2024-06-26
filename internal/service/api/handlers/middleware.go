package handlers

import (
	"context"
	"net/http"

	"github.com/rarimo/evm-airdrop-svc/internal/data"
	"github.com/rarimo/evm-airdrop-svc/internal/service/api"
	"gitlab.com/distributed_lab/kit/pgdb"
)

// DBCloneMiddleware is designed to clone DB session on each request. You must
// put all new DB handlers here instead of ape.CtxMiddleware, unless you have a
// reason to do otherwise.
func DBCloneMiddleware(db *pgdb.DB) func(http.Handler) http.Handler {
	type ctxExtender func(context.Context) context.Context

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clone := db.Clone()
			ctx := r.Context()

			extenders := []ctxExtender{
				api.CtxAirdropsQ(data.NewAirdropsQ(clone)),
			}

			for _, extender := range extenders {
				ctx = extender(ctx)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
