package urldelete

import (
	"log/slog"
	"net/http"

	"github.com/FacelessWayfarer/urlshortner/internal/lib/response"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/slogg"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLRemover interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, URLRemover URLRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const mark = "handlers.url-delete.New"

		log := log.With(
			slog.String("mark", mark),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")

		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, response.Error("invalid request"))

			return
		}

		err := URLRemover.DeleteURL(alias)
		if err != nil {
			log.Error("failed to remove url", slogg.Err(err))

			render.JSON(w, r, response.Error("internal error"))

			return
		}

		log.Info("removed url")

	}
}
