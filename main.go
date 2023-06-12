package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"
)

type Retry struct {
	RetryAfter time.Duration `json:"retry_after"`
}

func main() {

	fmt.Println("Start Listening")
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := os.OpenFile("output", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		io.WriteString(f, string(body))
		//
		//err = json.NewEncoder(w).Encode(Retry{
		//	RetryAfter: time.Minute,
		//})
		//if err != nil {
		//	fmt.Println(err.Error())
		//	return
		//}

	})

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:53377",
	}

	errChan := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1) // the channel used with signal.Notify should be buffered (SA1017)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	gracefulShutdown := func() {
		timedContext, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		_ = srv.Shutdown(timedContext)
	}

	select {
	case <-errChan:
		gracefulShutdown()

	case <-sigChan:
		gracefulShutdown()
	}

}
