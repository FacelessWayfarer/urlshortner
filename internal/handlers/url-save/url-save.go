package urlsave

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/FacelessWayfarer/urlshortner/internal/database"
	urlget "github.com/FacelessWayfarer/urlshortner/internal/handlers/url-get"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/random"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/response"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/slogg"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

// interface of the database decleared where it is used

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(longURL, alias string) (int64, error)
}

const (
	retrycount = 5 // number of trys to generate psudo-random alias without collision
)

func New(log *slog.Logger, urlSaver URLSaver, urlGetter urlget.URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const mark = "handlers.url-save.New"

		log := log.With(
			slog.String("mark", mark),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")

				render.JSON(w, r, response.Error("empty request"))

				return
			}
			err = render.DefaultDecoder(r, &req)
			if err != nil {
				log.Error("failed to deecode request body")

				render.JSON(w, r, response.Error("failed to decode request"))

				return
			}
			log.Info("request body decoded; format:unknown", slog.Any("request", req))
		}

		log.Info("request body decoded; format:JSON", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", slogg.Err(err))

			render.JSON(w, r, response.ValidationError(validateErr))

			return
		}

		const aliasLength = 5

		alias := req.Alias
		if alias == "" {
			for i := 0; i < retrycount; i++ {
				alias = random.NewRandomString(aliasLength)
				_, err := urlGetter.GetURL(alias)
				if errors.Is(err, database.ErrURLNotFound) {
					break
				}
			}
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if err != nil {
			if errors.Is(err, database.ErrURLAlreadyExists) {
				log.Info("url already exists", slog.String("url", req.URL))

				render.JSON(w, r, response.Error("url already exists"))

				return
			}
			log.Error("failed to add url", slogg.Err(err))

			render.JSON(w, r, response.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})

	}
}
