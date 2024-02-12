package server

import (
	"fmt"
	"net/http"

	"simpleFTP/pkg/middlewares/handlers"
	"simpleFTP/pkg/middlewares/logger"
)

type ServerConfig struct {
	Addr string
	Port int
}

func (sc *ServerConfig) serveAt() string {
	return fmt.Sprintf("%s:%d", sc.Addr, sc.Port)
}

func Serve(conf *ServerConfig) {
	logger.UseDevelopmentLogger()
	log := logger.SugaredLogger()

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		reader, err := r.MultipartReader()
		if err != nil {
			log.Errorf("error intializing reader: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = handlers.Upload(reader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write([]byte("File uploaded sccessfully"))
	})

	log.Infof("Starting server at %s...", conf.serveAt())
	http.Handle("/", http.FileServer(http.Dir("./static")))
	err := http.ListenAndServe(conf.serveAt(), nil)
	if err != nil {
		log.Info("err starting server")
	}
}
