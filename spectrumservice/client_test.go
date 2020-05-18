package spectrumservice

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"go.uber.org/zap"
)

var (
	logger, _ = zap.NewDevelopment()

	cookie = &http.Cookie{
		Name:     "session",
		Value:    "logged in",
		Domain:   "example.net",
		Path:     "/",
		MaxAge:   60 * 60,
		HttpOnly: true,
	}

	httpTest = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.URL.EscapedPath() == authenticate {
			b, err := ioutil.ReadFile("testdata/authentication.json")
			if err != nil {
				logger.Sugar().Panicf("Error reading testdata/authentication.json file.", err)
			}

			fmt.Fprint(w, string(b))

			http.SetCookie(w, cookie)
		}
	}))
)
