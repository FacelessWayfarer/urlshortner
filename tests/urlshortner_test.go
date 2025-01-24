package tests

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"testing"

	urlsave "github.com/FacelessWayfarer/urlshortner/internal/handlers/url-save"
	"github.com/FacelessWayfarer/urlshortner/internal/lib/random"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:6060"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(urlsave.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth("CoolAdmin69", "6996").
		Expect().
		Status(200).
		JSON().Object().
		ContainsKey("alias")
}

func TestURLShortner_SaveGet(t *testing.T) {
	testCases := []struct {
		name  string
		alias string
		url   string
		err   string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			err:   "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{Scheme: "http", Host: host}

			e := httpexpect.Default(t, u.String())

			r := e.POST("/url").WithJSON(urlsave.Request{URL: tc.url, Alias: tc.alias}).
				WithBasicAuth("CoolAdmin69", "6996").Expect().Status(http.StatusOK).JSON().Object()

			if tc.err != "" {
				r.NotContainsKey("alias")

				r.Value("error").String().IsEqual(tc.err)

				return
			}

			alias := tc.alias

			if tc.alias != "" {
				r.Value("alias").String().IsEqual(tc.alias)
			} else {
				r.Value("alias").String().NotEmpty()

				alias = r.Value("alias").String().Raw()
			}

			testGet(t, alias, tc.url)
			testGet(t, alias, tc.url)
			_ = e.DELETE("/"+path.Join("url", alias)).WithBasicAuth("CoolAdmin69", "6996").WithHeader("Content-Type", "application/json").
				Expect().Status(http.StatusOK)

			testRedirectNotFound(t, alias)
		})
	}
}

func testGet(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := GetRedirectTest(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)
}

// GetRedirectTest returns the final URL after redirection.
func GetRedirectTest(url string) (string, error) {
	var ErrInvalidStatusCode = errors.New("invalid status code")

	// const mark = "GetRedirectTest"

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
		return "", ErrInvalidStatusCode
	}
	return resp.Header.Get("Location"), nil
}

func testRedirectNotFound(t *testing.T, alias string) {

	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := GetRedirectTest(u.String())
	require.Error(t, err)
}
