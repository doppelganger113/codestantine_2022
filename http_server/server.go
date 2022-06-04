package http_server

import (
	"api/http_server/authenticator"
	coremiddleware "api/http_server/middleware"
	"api/http_server/openapi"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"net/http"
	"time"
)

type Server struct {
	Port       uint
	router     *chi.Mux
	httpServer *http.Server
	logger     *zerolog.Logger
}

type Handlers struct {
	ImagesHandler ImagesHandler
	Authenticator authenticator.Authenticator
}

func NewServer(logger *zerolog.Logger, config Config, handlers Handlers) (*Server, error) {
	port := fmt.Sprintf(":%d", config.Port)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(coremiddleware.Secure)
	r.Use(middleware.Timeout(config.Timeout))
	r.Use(middleware.Heartbeat(config.HeartbeatUrl))

	log := logger.With().
		Timestamp().
		Str("service", "api").
		Str("host", "http://demo.com").
		Logger()

	c := alice.New()

	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(log))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to that handler, all our logs will come with some prepopulated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))

	r.Use(func(next http.Handler) http.Handler {
		// Here is your final handler
		return c.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info().Msg("HANDLING!")
			// Get the logger from the request's context. You can safely assume it
			// will be always there: if the handler is removed, hlog.FromRequest
			// will return a no-op logger.
			hlog.FromRequest(r).Info().
				Str("user", "current user").
				Str("status", "ok").
				Msg("Request")

			// Output: {"level":"info","time":"2001-02-03T04:05:06Z","role":"my-service","host":"local-hostname","req_id":"b4g0l5t6tfid6dtrapu0","user":"current user","status":"ok","message":"Something happened"}
			next.ServeHTTP(w, r)
		}))
	})

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.CorsAllowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	//if config.DebugRoutes {
	//	r.Use(middleware.Logger)
	//}

	// Routing
	swaggerRouter, err := openapi.NewOpenApi3Router(openapi.Config{
		BasicAuthUsername:          config.BasicAuthUsername,
		BasicAuthPassword:          config.BasicAuthPassword,
		BasicAuthRealm:             config.BasicAuthRealm,
		Domain:                     config.Domain,
		OAuth2TokenUrl:             config.OAuth2TokenUrl,
		OAuth2AuthorizationCodeUrl: config.OAuth2AuthorizationCodeUrl,
	})
	if err != nil {
		return nil, err
	}
	r.Route("/docs", swaggerRouter)
	r.Route("/api/v1/images", ImagesRouter(
		handlers.ImagesHandler,
		logger,
		handlers.Authenticator,
	))

	httpServer := &http.Server{
		Addr:              port,
		Handler:           r,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
	}

	return &Server{
		router:     r,
		httpServer: httpServer,
		Port:       config.Port,
		logger:     &log,
	}, nil
}

// StartNewConfiguredAndListenChannel boots configuration, creates and starts the server with
// err channel which is used to signal when the server closes
func StartNewConfiguredAndListenChannel(
	logger *zerolog.Logger, handlers Handlers, errChannel chan<- error,
) (*Server, error) {
	var server *Server

	httpConfig := NewDefaultConfig()
	if err := httpConfig.LoadFromEnv(); err != nil {
		return nil, err
	}

	server, err := NewServer(logger, httpConfig, handlers)
	if err != nil {
		return nil, err
	}

	go func() {
		errChannel <- server.StartAndListen()
	}()

	return server, nil
}

func (s *Server) StartAndListen() error {
	s.logger.Info().Msgf("Server started on port :%d", s.Port)
	if err := s.httpServer.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
