package webserver

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Start serves site in specified port
func Start(rootDir string, port int) error {
	// Create handler
	hdl, err := newHandler(rootDir)
	if err != nil {
		return err
	}

	// Create router
	router := httprouter.New()
	router.GET("/*filepath", hdl.serveSite)

	// Route for panic
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, arg interface{}) {
		http.Error(w, fmt.Sprint(arg), 500)
	}

	// Create server
	url := fmt.Sprintf(":%d", port)
	svr := &http.Server{
		Addr:         url,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	// Serve app
	return svr.ListenAndServe()
}

func panicError(err error) {
	if err == nil {
		return
	}

	// Check for a broken connection, as it is not really a
	// condition that warrants a panic stack trace.
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if se.Err == syscall.EPIPE || se.Err == syscall.ECONNRESET {
				return
			}
		}
	}

	panic(err)
}
