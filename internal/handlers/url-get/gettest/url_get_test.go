package gettests

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	urlget "github.com/FacelessWayfarer/urlshortner/internal/handlers/url-get"
	"github.com/FacelessWayfarer/urlshortner/internal/handlers/url-save/test/mocks"
	discardslogg "github.com/FacelessWayfarer/urlshortner/internal/lib/discard-slogg"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
)

func TestGetHandle(t *testing.T) {
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
			url:   "https://www.google.com/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).Return(tc.url, tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", urlget.New(discardslogg.NewDiscardLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			// ts.URL = "http://127.0.0.1:6060"
			t.Log(ts.URL + "/" + tc.alias)
			redirectedToURL, err := GetRedirectTest(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			// Check the final URL after redirection.
			assert.Equal(t, tc.url, redirectedToURL)
		})

	}
}

// GetRedirectTest returns the final URL after redirection.
func GetRedirectTest(url string) (string, error) {
	var ErrInvalidStatusCode = errors.New("invalid status code")

	const mark = "GetRedirectTest"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // stop after 1st redirect
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("%s: %w: %d", mark, ErrInvalidStatusCode, resp.StatusCode)
	}
	return resp.Header.Get("Location"), nil
}
