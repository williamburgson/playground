package handlers

import (
	"bufio"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"simpleFTP/pkg/middlewares/logger"
)

func Upload(reader *multipart.Reader) error {
	log := logger.SugaredLogger()
	log.Infof("uploading file ...")

	p, err := reader.NextPart()
	if err != nil && err != io.EOF {
		log.Errorf("err uploading file: %s", err)
		return err
	}

	buf := bufio.NewReader(p)
	log.Infof("uploaded %d bytes ...", buf.Size())
	sniff, _ := buf.Peek(512)
	contentType := http.DetectContentType(sniff)
	log.Infof("uploaded file type: %s", contentType)
	f, err := ioutil.TempFile("tmp", "upload-*")
	if err != nil {
		log.Errorf("err creating temp file for upload: %s", err)
		return err
	}
	defer f.Close()
	var maxSize int64 = 32
	lmt := io.MultiReader(io.LimitReader(p, maxSize))
	_, err = io.Copy(f, lmt)
	if err != nil && err != io.EOF {
		log.Errorf("err writing uploaded file: %s", err)
		return err
	}
	return nil
}
