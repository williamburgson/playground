// package main

// import (
// 	"bufio"
// 	"io"
// 	"mime/multipart"
// 	"os"
// 	"simpleFTP/pkg/middlewares/logger"

// 	flag "github.com/spf13/pflag"
// 	"go.uber.org/zap"
// )

// var log *zap.SugaredLogger = logger.SugaredLogger()

// func multiPartRead(fname string) {
// 	// f, err := os.Open(fname)
// 	// if err != nil {
// 	// 	log.Info(err)
// 	// }

// 	// fi, err := f.Stat()
// 	// if err != nil {
// 	// 	log.Info(err)
// 	// }

// 	// buf := bufio.NewReader(f)
// 	// r, w := io.Pipe()

// 	// multi := multipart.NewReader()
// }

// func main() {
// 	logger.UseDevelopmentLogger()
// 	fname := flag.String("filename", "", "File name")
// 	flag.Parse()

// 	log.Infof("loading file %s", *fname)

// }