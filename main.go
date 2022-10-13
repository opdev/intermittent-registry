package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/go-containerregistry/pkg/registry"
)

var (
	// reqcount is the number of seen requests. This is periodically reset.
	reqcount int = 0
	// disruption is what at what multiple to disrupt the request.
	disruption int = 50
)

func main() {
	fmt.Println("Starting")
	defer fmt.Println("Ending")
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	m := incrementRequestCount(
		useIncorrectHandlerPeriodically(
			registry.New().ServeHTTP,
		),
	)

	go func() {
		fmt.Println("Listening")
		if err := http.ListenAndServe("0.0.0.0:80", m); err != nil {
			log.Fatalln("FATAL", err)
		}
	}()

	<-done

}

func incrementRequestCount(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if reqcount > 100 {
			// a little reset.
			reqcount = 0
		}
		reqcount++
		// fmt.Println(reqcount)
		next.ServeHTTP(w, r)
	}
}

// useIncorrectHandlerPeriodically will throw a 500 response at request multiples
// as defined by disruption.
func useIncorrectHandlerPeriodically(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if reqcount%disruption == 0 {
			log.Println("Simulating intermittent failure!")
			bad := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Simulated intermittent failure"))
			}

			bad(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
}
