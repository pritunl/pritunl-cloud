package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/sirupsen/logrus"
)

type NopCloser struct {
	io.Reader
}

func (NopCloser) Close() error {
	return nil
}

var httpErrCodes = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Payload Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	416: "Range Not Satisfiable",
	417: "Expectation Failed",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	510: "Not Extended",
	511: "Network Authentication Required",
}

func CopyBody(r *http.Request) (buffer *bytes.Buffer, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Request read error"),
		}
		return
	}
	_ = r.Body.Close()

	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	buffer = bytes.NewBuffer(body)

	return
}

func StripPort(hostport string) string {
	colon := strings.IndexByte(hostport, ':')
	if colon == -1 {
		return hostport
	}

	n := strings.Count(hostport, ":")
	if n > 1 {
		if i := strings.IndexByte(hostport, ']'); i != -1 {
			return strings.TrimPrefix(hostport[:i], "[")
		}
		return hostport
	}

	return hostport[:colon]
}

func FormatHostPort(hostname string, port int) string {
	if strings.Contains(hostname, ":") {
		hostname = "[" + hostname + "]"
	}
	return fmt.Sprintf("%s:%d", hostname, port)
}

func ParseObjectId(strId string) (objId bson.ObjectID, ok bool) {
	if strId == "" {
		objId = bson.NilObjectID
		return
	}

	objectId, err := bson.ObjectIDFromHex(strId)
	if err != nil {
		objId = bson.NilObjectID
		return
	}

	objId = objectId
	ok = true
	return
}

func ObjectIdHex(strId string) (objId bson.ObjectID) {
	if strId == "" {
		objId = bson.NilObjectID
		return
	}

	objectId, err := bson.ObjectIDFromHex(strId)
	if err != nil {
		objId = bson.NilObjectID
		return
	}

	objId = objectId
	return
}

func GetStatusMessage(code int) string {
	return fmt.Sprintf("%d %s", code, http.StatusText(code))
}

func AbortWithStatus(c *gin.Context, code int) {
	r := render.String{
		Format: GetStatusMessage(code),
	}

	c.Status(code)
	r.WriteContentType(c.Writer)
	c.Writer.WriteHeaderNow()
	r.Render(c.Writer)
	c.Abort()
}

func AbortWithError(c *gin.Context, code int, err error) {
	AbortWithStatus(c, code)
	c.Error(err)
}

func WriteStatus(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, GetStatusMessage(code))
}

func WriteText(w http.ResponseWriter, code int, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, text)
}

func WriteUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(401)
	fmt.Fprintln(w, "401 "+msg)
}

func CloneHeader(src http.Header) (dst http.Header) {
	dst = make(http.Header, len(src))
	for k, vv := range src {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		dst[k] = vv2
	}
	return dst
}

func GetLocation(r *http.Request) string {
	host := ""

	switch {
	case r.Header.Get("X-Host") != "":
		host = r.Header.Get("X-Host")
		break
	case r.Host != "":
		host = r.Host
		break
	case r.URL.Host != "":
		host = r.URL.Host
		break
	}

	return "https://" + host
}

func GetOrigin(r *http.Request) string {
	origin := r.Header.Get("Origin")
	if origin == "" {
		host := ""
		switch {
		case r.Host != "":
			host = r.Host
			break
		case r.URL.Host != "":
			host = r.URL.Host
			break
		}
		origin = "https://" + host
	}

	return origin
}

func CheckRequestN(resp *http.Response, msg string, codes []int) (err error) {
	for _, code := range codes {
		if resp.StatusCode == code {
			return
		}
	}

	bodyStr := ""
	bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 10*1024))
	if readErr != nil {
		bodyStr = fmt.Sprintf("[%v]", readErr)
	} else {
		bodyStr = string(bodyBytes)
	}

	logrus.WithFields(logrus.Fields{
		"body":        bodyStr,
		"status_code": resp.StatusCode,
		"message":     msg,
	}).Error(msg)

	err = &errortypes.RequestError{
		errors.Newf("request: Response status error %d", resp.StatusCode),
	}
	return
}

func CheckRequest(resp *http.Response, msg string) (err error) {
	if resp.StatusCode == 200 {
		return
	}

	bodyStr := ""
	bodyBytes, readErr := io.ReadAll(io.LimitReader(resp.Body, 10*1024))
	if readErr != nil {
		bodyStr = fmt.Sprintf("[%v]", readErr)
	} else {
		bodyStr = string(bodyBytes)
	}

	logrus.WithFields(logrus.Fields{
		"body":        bodyStr,
		"status_code": resp.StatusCode,
		"message":     msg,
	}).Error(msg)

	err = &errortypes.RequestError{
		errors.Newf("request: Response status error %d", resp.StatusCode),
	}
	return
}
