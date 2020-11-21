package recoverer

import (
	"log"
	"net/http"
)

func Recoverer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Println(err)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}