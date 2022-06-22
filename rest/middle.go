package rest

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/auturnn/peer-base-nodes/utils"
	"github.com/gorilla/handlers"
)

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	//adaptor pattern
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-type", "application/json")
		next.ServeHTTP(rw, r)
	})
}

const windowLogName = "2006_01_02"
const defaultLogName = "2006-01-02"

func loggerMiddleware(next http.Handler) http.Handler {
	var f *os.File
	t := time.Now().Local()

	switch runtime.GOOS {
	case "windows":
		f = loggingFileOpen(t.Format(windowLogName))
	default:
		f = loggingFileOpen(t.Format(defaultLogName))
	}

	return handlers.LoggingHandler(f, next)
}

func loggingFileOpen(fileName string) *os.File {
	f, err := os.OpenFile(fmt.Sprintf("./log/%s.log", fileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0655)
	utils.HandleError(err)

	return f
}
