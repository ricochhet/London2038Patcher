package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ricochhet/london2038patcher/pkg/embedutil"
	"github.com/ricochhet/london2038patcher/pkg/errutil"
	"github.com/ricochhet/london2038patcher/pkg/httputil"
	"github.com/ricochhet/london2038patcher/pkg/logutil"
)

const (
	sessionCookieName = "fs_session"
	sessionTTL        = 24 * time.Hour
	loginRoute        = "/auth/login"
	logoutRoute       = "/auth/logout"

	nextQuery       = "next"
	usernameFormKey = "username"
	passwordFormKey = "password"

	loginTmpl     = "login"
	loginTmplHTML = "login.html"
)

type loginPageData struct {
	Error string
	Next  string
}

// newSessionSecret generates a random 32-byte secret for use when the config does not supply one.
func newSessionSecret() []byte {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("auth: cannot generate session secret: " + err.Error())
	}

	return b
}

// signSession produces "<ts>.<hex-hmac>".
func signSession(secret []byte, ts int64) string {
	tsStr := strconv.FormatInt(ts, 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(tsStr))

	return tsStr + "." + hex.EncodeToString(mac.Sum(nil))
}

// verifySession returns true if the cookie value is valid and not expired.
func verifySession(secret []byte, value string) bool {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 {
		return false
	}

	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return false
	}

	if time.Since(time.Unix(ts, 0)) > sessionTTL {
		return false
	}

	expected := signSession(secret, ts)

	return hmac.Equal([]byte(value), []byte(expected))
}

// setSessionCookie writes a signed session cookie to the response.
func setSessionCookie(w http.ResponseWriter, secret []byte) {
	val := signSession(secret, time.Now().Unix())
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	})
}

// clearSessionCookie expires the session cookie.
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// hasValidSession checks the request cookie.
func hasValidSession(r *http.Request, secret []byte) bool {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}

	return verifySession(secret, c.Value)
}

// withFormAuth returns a middleware that redirects unauthenticated requests to
// the login page. It exempts the login/logout routes themselves.
func withFormAuth(secret []byte, publicPrefixes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == loginRoute || path == logoutRoute {
				next.ServeHTTP(w, r)
				return
			}

			for _, prefix := range publicPrefixes {
				if strings.HasPrefix(path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			if !hasValidSession(r, secret) {
				dest := loginRoute + "?" + nextQuery + "=" + url.QueryEscape(r.URL.RequestURI())
				http.Redirect(w, r, dest, http.StatusFound)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// resolveFormAuthSecret resolves the session secret.
func resolveFormAuthSecret(hexSecret string) []byte {
	if hexSecret != "" {
		b, err := hex.DecodeString(hexSecret)
		if err == nil && len(b) >= 16 {
			return b
		}

		logutil.Errorf(logutil.Get(), "formAuth.secret is invalid hex; generating random secret\n")
	}

	return newSessionSecret()
}

// registerAuthRoutes registers GET /auth/login, POST /auth/login, and
// GET /auth/logout on the provided mux using the supplied credentials and HMAC secret.
func (c *Context) registerAuthRoutes(
	handle func(pattern string, h http.Handler),
	user, password string,
	secret []byte,
) {
	bytes := embedutil.MaybeRead(c.FS, loginTmplHTML)
	tmpl := template.Must(template.New(loginTmpl).Parse(string(bytes)))

	serveLoginPage := func(w http.ResponseWriter, errMsg, next string) {
		httputil.ContentType(w, httputil.ContentTypeHTML)

		if errMsg != "" {
			w.WriteHeader(http.StatusUnauthorized)
		}

		if err := tmpl.Execute(w, loginPageData{Error: errMsg, Next: next}); err != nil {
			logutil.Errorf(logutil.Get(), "login tmpl: %v\n", err)
		}
	}

	handle(loginRoute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if hasValidSession(r, secret) {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			next := r.URL.Query().Get(nextQuery)
			if next == "" {
				next = "/"
			}

			serveLoginPage(w, "", next)

		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				errutil.HTTPBadRequest(w)
				return
			}

			u := r.FormValue(usernameFormKey)
			p := r.FormValue(passwordFormKey)

			next := r.FormValue(nextQuery)
			if next == "" || !strings.HasPrefix(next, "/") {
				next = "/"
			}

			if u != user || p != password {
				serveLoginPage(w, "invalid username or password.", next)
				return
			}

			setSessionCookie(w, secret)
			http.Redirect(w, r, next, http.StatusFound)

		default:
			http.NotFound(w, r)
		}
	}))

	handle(logoutRoute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clearSessionCookie(w)
		http.Redirect(w, r, loginRoute, http.StatusFound)
	}))
}
