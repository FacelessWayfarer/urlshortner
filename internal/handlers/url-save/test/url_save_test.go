package savetests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacelessWayfarer/urlshortner/internal/database"
	urlsave "github.com/FacelessWayfarer/urlshortner/internal/handlers/url-save"
	"github.com/FacelessWayfarer/urlshortner/internal/handlers/url-save/test/mocks"
	discardslogg "github.com/FacelessWayfarer/urlshortner/internal/lib/discard-slogg"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/random"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/response"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/slogg"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveHandle(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
		},
		{
			name:  "Empty alias",
			alias: "",
			url:   "https://google.com",
		},
		{
			name:      "Empty URL",
			url:       "",
			alias:     "some_alias",
			respError: "field URL is a required field",
		},
		{
			name:      "Invalid URL",
			url:       "some invalid URL",
			alias:     "some_alias",
			respError: "field URL is not a valid URL",
		},
		{
			name:      "SaveURL Error",
			alias:     "test_alias",
			url:       "https://google.com",
			respError: "failed to add url",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()

			}

			handler := NewMock(discardslogg.NewDiscardLogger(), urlSaverMock)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			var resp urlsave.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

			// TODO: add more checks
		})
	}
}

// NewMock() almost equal to urlsave.New() with the exeption of disabled check for random generated alias been equal to an alias already existing in the database
func NewMock(log *slog.Logger, urlSaver urlsave.URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const mark = "handlers.url-save.New"

		log := log.With(
			slog.String("mark", mark),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req urlsave.Request

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
			// for i := 0; i < retrycount; i++ {
			alias = random.NewRandomString(aliasLength)
			// 	_, err := urlGetter.GetURL(alias)
			// 	if errors.Is(err, database.ErrURLNotFound) {
			// 		break
			// 	}
			// }
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

		render.JSON(w, r, urlsave.Response{
			Response: response.OK(),
			Alias:    alias,
		})

	}
}
