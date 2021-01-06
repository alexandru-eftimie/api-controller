package apicontroller

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// HTTPHandler is the handler that is called when path is accessed
type HTTPHandler func(w http.ResponseWriter, r *http.Request)

// AuthCallback is the function called when doing bearer authentication
type AuthCallback func(token string) (id string, err error)

// Controller runs the controller
type Controller struct {
	router               *mux.Router
	server               *http.Server
	useDefaultMiddleware bool
	AuthCallback         AuthCallback
}

// NewController creates a new HTTP API controller
func NewController() *Controller {
	c := Controller{
		router:               mux.NewRouter(),
		useDefaultMiddleware: true,
	}
	return &c
}

// AddHandler adds a handler
func (c *Controller) AddHandler(path string, fn HTTPHandler, methods ...string) {
	c.router.HandleFunc(path, fn).Methods(methods...)
}

// Run runs the controller and the listener
func (c *Controller) Run(addr string) {
	if c.useDefaultMiddleware {
		c.router.Use(c.defaultAuthMiddleware)
	}
	c.server = &http.Server{Addr: addr, Handler: c.router}

	log.Println("Running at http://" + addr)
	// log.Fatal(http.ListenAndServe(addr, nil))
	if err := c.server.ListenAndServe(); err != nil {
		// handle err
	}
}

// Stop stops the http listener
func (c *Controller) Stop() {
	if c.server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c.server.Shutdown(ctx)
	c.server = nil
}

func (c *Controller) setMiddleware(h ...mux.MiddlewareFunc) {
	c.useDefaultMiddleware = false
	c.router.Use(h...)
}
func (c *Controller) defaultAuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// if there's no auth callback then skip auth
		if c.AuthCallback == nil {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("X-ID", "")

		header := r.Header.Get("Authorization")
		parts := strings.Split(header, " ")
		if parts[0] != "Bearer" {
			w.Header().Add("X-Error", "Only Authorization: Bearer Allowed")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token := parts[1]

		fn := c.AuthCallback
		id, err := fn(token)
		if err != nil {
			w.Header().Add("X-Error", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("X-ID", id)

		// continue from here
		next.ServeHTTP(w, r)
		return
	})
}
