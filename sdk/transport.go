package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const ReqClient = "vault-v0.1.0"

type LogRoundTripper struct {
	debug bool
}

func (l *LogRoundTripper) RoundTrip(request *http.Request) (response *http.Response, err error) {
	var inBytes, outBytes []byte

	if l.debug {
		var start = time.Now()

		defer func() { l.log(inBytes, outBytes, err, start) }()
	}

	bodyReader, err := request.GetBody()
	if err != nil {
		return
	}

	headName := "X-TC-Action"
	request.Header.Set("X-TC-RequestClient", ReqClient)
	inBytes = []byte(fmt.Sprintf("%s, request: ", request.Header[headName]))

	requestBody, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return
	}

	inBytes = append(inBytes, requestBody...)

	headName = "X-TC-Region"
	appendMessage := []byte(fmt.Sprintf(
		", (host %+v, region:%+v)",
		request.Header["Host"],
		request.Header[headName],
	))

	inBytes = append(inBytes, appendMessage...)

	response, err = http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return
	}

	outBytes, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	response.Body = ioutil.NopCloser(bytes.NewBuffer(outBytes))

	return
}

func (l *LogRoundTripper) log(in []byte, out []byte, err error, start time.Time) {
	var buf bytes.Buffer

	if len(in) > 0 {
		buf.WriteString("tencentcloud-sdk-go: ")
		buf.Write(in)
	}

	if len(out) > 0 {
		buf.WriteString("; response:")
		err := json.Compact(&buf, out)
		if err != nil {
			out := bytes.Replace(out,
				[]byte("\n"),
				[]byte(""),
				-1)
			out = bytes.Replace(out,
				[]byte(" "),
				[]byte(""),
				-1)
			buf.Write(out)
		}
	}

	if err != nil {
		buf.WriteString("; error:")
		buf.WriteString(err.Error())
	}

	costFormat := fmt.Sprintf(",cost %s", time.Since(start))
	buf.WriteString(costFormat)

	log.Println(buf.String())
}
