package urlget

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/FacelessWayfarer/urlshortner/internal/database"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/response"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/slogg"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (url string, err error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const mark = "handlers.url-get.New"

		log := log.With(
			slog.String("mark", mark),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		fmt.Println(alias)
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, database.ErrURLNotFound) {
				log.Info("url not found", "alias", alias)

				render.JSON(w, r, response.Error("not found"))

				return
			}
			log.Error("failed to get url", slogg.Err(err))

			render.JSON(w, r, response.Error("internal error"))

			return
		}

		log.Info("retrived url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
