package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/mux"
	"github.com/openaustralia/yinyo/pkg/commands"
	"github.com/openaustralia/yinyo/pkg/protocol"
)

func (server *Server) create(w http.ResponseWriter, r *http.Request) error {
	// TODO: Pass options in POST body instead of url
	options := protocol.CreateRunOptions{
		APIKey:      r.URL.Query().Get("api_key"),
		CallbackURL: r.URL.Query().Get("callback_url"),
	}
	createResult, err := server.app.CreateRun(options)
	if err != nil {
		if errors.Is(err, commands.ErrNotAllowed) {
			return newHTTPError(err, http.StatusUnauthorized, err.Error())
		}
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(createResult)
}

func (server *Server) getApp(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]
	w.Header().Set("Content-Type", "application/gzip")
	reader, err := server.app.GetApp(runID)
	if err != nil {
		// Returns 404 if there is no app
		if errors.Is(err, commands.ErrNotFound) {
			return newHTTPError(err, http.StatusNotFound, err.Error())
		}
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

func (server *Server) putApp(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]
	err := server.app.PutApp(runID, r.Body, r.ContentLength)
	if errors.Is(err, commands.ErrArchiveFormat) {
		return newHTTPError(err, http.StatusBadRequest, err.Error())
	}
	return err
}

func (server *Server) getCache(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]
	reader, err := server.app.GetCache(runID)
	if err != nil {
		// Returns 404 if there is no cache
		if errors.Is(err, commands.ErrNotFound) {
			return newHTTPError(err, http.StatusNotFound, err.Error())
		}
		return err
	}
	w.Header().Set("Content-Type", "application/gzip")
	_, err = io.Copy(w, reader)
	return err
}

func (server *Server) putCache(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]
	return server.app.PutCache(runID, r.Body, r.ContentLength)
}

func (server *Server) getOutput(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]
	reader, err := server.app.GetOutput(runID)
	if err != nil {
		// Returns 404 if there is no output
		if errors.Is(err, commands.ErrNotFound) {
			return newHTTPError(err, http.StatusNotFound, err.Error())
		}
		return err
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, reader)
	return err
}

func (server *Server) putOutput(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]
	return server.app.PutOutput(runID, r.Body, r.ContentLength)
}

func (server *Server) getExitData(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]

	exitData, err := server.app.GetExitData(runID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(exitData)
}

func (server *Server) start(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]

	decoder := json.NewDecoder(r.Body)
	var options protocol.StartRunOptions
	err := decoder.Decode(&options)
	if err != nil {
		return newHTTPError(err, http.StatusBadRequest, "JSON in body not correctly formatted")
	}

	if options.MaxRunTime == 0 {
		options.MaxRunTime = server.maxRunTime
	} else if options.MaxRunTime > server.maxRunTime {
		return newHTTPError(err, http.StatusBadRequest, fmt.Sprintf("max_run_time should not be larger than %v", server.maxRunTime))
	}

	env := make(map[string]string)
	for _, keyvalue := range options.Env {
		env[keyvalue.Name] = keyvalue.Value
	}

	err = server.app.StartRun(runID, options)
	if errors.Is(err, commands.ErrAppNotAvailable) {
		err = newHTTPError(err, http.StatusBadRequest, "app needs to be uploaded before starting a run")
	}
	return err
}

func (server *Server) getEvents(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]
	lastID := r.URL.Query().Get("last_id")
	if lastID == "" {
		lastID = "0"
	}
	w.Header().Set("Content-Type", "application/ld+json")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("couldn't access the flusher")
	}

	events := server.app.GetEvents(runID, lastID)
	enc := json.NewEncoder(w)
	for events.More() {
		e, err := events.Next()
		if err != nil {
			return err
		}
		err = enc.Encode(e)
		if err != nil {
			return err
		}
		flusher.Flush()
	}
	return nil
}

func (server *Server) createEvent(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]

	// Read json message as is into a string
	// TODO: Switch over to json decoder
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Check the form of the JSON by interpreting it
	var event protocol.Event
	err = json.Unmarshal(buf, &event)
	if err != nil {
		return newHTTPError(err, http.StatusBadRequest, "JSON in body not correctly formatted")
	}

	return server.app.CreateEvent(runID, event)
}

func (server *Server) delete(w http.ResponseWriter, r *http.Request) error {
	runID := mux.Vars(r)["id"]

	return server.app.DeleteRun(runID)
}

func (server *Server) whoAmI(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintln(w, "Hello from Yinyo!")
	return nil
}

// isPrivate reports whether `ip' is a private address, according to
// RFC 1918 (IPv4 addresses) and RFC 4193 (IPv6 addresses).
// TODO: Switch over to implementation in https://github.com/golang/go/issues/29146 when available
func isPrivate(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		// Local IPv4 addresses are defined in https://tools.ietf.org/html/rfc1918
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1]&0xf0 == 16) ||
			(ip4[0] == 192 && ip4[1] == 168)
	}
	// Local IPv6 addresses are defined in https://tools.ietf.org/html/rfc4193
	return len(ip) == net.IPv6len && ip[0]&0xfe == 0xfc
}

// isExternal returns true if the request has arrived via the public internet. This relies
// on the source IP address being preserved which does require the Kubernetes load
// balancer to be set up in a particular way.
// This is used in measuring network traffic
func isExternal(request *http.Request) (bool, error) {
	// First get the ip address from the string of the form "host:port"
	ipString, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return false, err
	}
	ip := net.ParseIP(ipString)
	return !isPrivate(ip), nil
}

// Middleware that logs the request uri
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var source string
		e, err := isExternal(r)
		switch {
		case err != nil:
			source = "?"
		case e:
			source = "external"
		default:
			source = "internal"
		}
		log.Println(source, r.Method, r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

type readMeasurer struct {
	rc        io.ReadCloser
	BytesRead int64
}

func newReadMeasurer(rc io.ReadCloser) *readMeasurer {
	return &readMeasurer{rc: rc}
}

func (r *readMeasurer) Read(p []byte) (n int, err error) {
	n, err = r.rc.Read(p)
	atomic.AddInt64(&r.BytesRead, int64(n))
	return
}

func (r *readMeasurer) Close() error {
	return r.rc.Close()
}

func (server *Server) recordTraffic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runID := mux.Vars(r)["id"]
		readMeasurer := newReadMeasurer(r.Body)
		r.Body = readMeasurer
		m := httpsnoop.CaptureMetrics(next, w, r)
		// TODO: Don't ignore any errors from isExternal
		external, err := isExternal(r)
		if err != nil {
			// TODO: Will this actually work here
			logAndReturnError(err, w)
			return
		}
		if runID != "" {
			err = server.app.RecordAPINetworkUsage(runID, external, readMeasurer.BytesRead, m.Written)
			if err != nil {
				// TODO: Will this actually work here
				logAndReturnError(err, w)
				return
			}
		}
	})
}

// Middleware function, which will be called for each request
// TODO: Refactor checkRunCreated method to return an error
func (server *Server) checkRunCreated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runID := mux.Vars(r)["id"]

		created, err := server.app.IsRunCreated(runID)
		if err != nil {
			log.Println(err)
			logAndReturnError(err, w)
			return
		}
		if !created {
			err = newHTTPError(err, http.StatusNotFound, fmt.Sprintf("run %v: not found", runID))
			logAndReturnError(err, w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func logAndReturnError(err error, w http.ResponseWriter) {
	log.Println(err)
	err2, ok := err.(clientError)
	if !ok {
		// TODO: Factor out common code with other error handling
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		//nolint:errcheck // ignore error while logging an error
		//skipcq: GSC-G104
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	}
	body, err := err2.ResponseBody()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	status, headers := err2.ResponseHeaders()
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	//nolint:errcheck // ignore error while logging an error
	//skipcq: GSC-G104
	w.Write(body)
}

type appHandler func(http.ResponseWriter, *http.Request) error

// Error handling
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil {
		logAndReturnError(err, w)
	}
}

// Server holds the internal state for the server
type Server struct {
	router     *mux.Router
	app        commands.App
	maxRunTime int64 // the global maximum run time in seconds that every run can not exceed
}

// Initialise the server's state
func (server *Server) Initialise(startupOptions *commands.StartupOptions, maxRunTime int64) error {
	app, err := commands.New(startupOptions)
	if err != nil {
		return err
	}
	server.app = app
	server.maxRunTime = maxRunTime
	server.InitialiseRoutes()
	return nil
}

// InitialiseRoutes sets up the routes
func (server *Server) InitialiseRoutes() {
	server.router = mux.NewRouter().StrictSlash(true)
	server.router.Handle("/", appHandler(server.whoAmI))
	server.router.Handle("/runs", appHandler(server.create)).Methods("POST")

	runRouter := server.router.PathPrefix("/runs/{id}").Subrouter()
	runRouter.Handle("/app", appHandler(server.getApp)).Methods("GET")
	runRouter.Handle("/app", appHandler(server.putApp)).Methods("PUT")
	runRouter.Handle("/cache", appHandler(server.getCache)).Methods("GET")
	runRouter.Handle("/cache", appHandler(server.putCache)).Methods("PUT")
	runRouter.Handle("/output", appHandler(server.getOutput)).Methods("GET")
	runRouter.Handle("/output", appHandler(server.putOutput)).Methods("PUT")
	runRouter.Handle("/exit-data", appHandler(server.getExitData)).Methods("GET")
	runRouter.Handle("/start", appHandler(server.start)).Methods("POST")
	runRouter.Handle("/events", appHandler(server.getEvents)).Methods("GET")
	runRouter.Handle("/events", appHandler(server.createEvent)).Methods("POST")
	runRouter.Handle("", appHandler(server.delete)).Methods("DELETE")
	server.router.Use(server.recordTraffic)
	runRouter.Use(server.checkRunCreated)
	server.router.Use(logRequests)
}

// Run runs the server. This blocks until the server quits
func (server *Server) Run(addr string) {
	log.Println("Yinyo is ready and waiting.")
	log.Fatal(http.ListenAndServe(addr, server.router))
}
