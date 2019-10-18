package dogbin

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var serverUrl string

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		slug := r.URL.Path[len("/documents/"):]
		switch slug {
		case "exists":
			io.WriteString(w, `{"data":"works","document":{"_id":"exists","content":"works","isUrl":false,"owner":{"$oid":"5b20334e5e7034132c431e78"},"version":2,"viewCount":12},"key":"exists"}`)
		case "existshaste":
			io.WriteString(w, `{"data":"works", "key":"existshaste"}`)
		case "broken":
			io.WriteString(w, "whoops")
		case "broken2":
			io.WriteString(w, "{}")
		case "notexist":
			w.WriteHeader(404)
			io.WriteString(w, `{"message":"Document not found."}`)
		case "whoops":
			w.WriteHeader(500)
		case "whoops2":
			w.WriteHeader(500)
			io.WriteString(w, "Internal Server Error")
		case "whoops3":
			w.WriteHeader(500)
			io.WriteString(w, "{}")
		default:
			println(slug)
			w.WriteHeader(404)
			io.WriteString(w, `{"message":"Document not found."}`)
		}
	} else {
		content := new(UploadRequest)
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(content)
		if err != nil {
			io.WriteString(w, `{"key":"hasteup"}`)
		} else {
			switch content.Slug {
			case "works":
				io.WriteString(w, `{"key":"works", "isUrl": false}`)
			case "url":
				io.WriteString(w, `{"key":"url", "isUrl": true}`)
			case "duplicate":
				w.WriteHeader(409)
				io.WriteString(w, `{"message":"This URL is already in use, please choose a different one"}`)
			}
		}
	}
}

var local Server

func TestMain(m *testing.M) {
	flag.Parse()
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()
	serverUrl = ts.URL
	local = NewServer(ts.URL)
	os.Exit(m.Run())
}

func TestNewServer(t *testing.T) {
	assert := assert.New(t)

	custom := NewServer("http://test")
	assert.Equal("http://test", custom.server)
	base, err := custom.baseUrl()
	assert.Nil(err)
	assert.Equal("http://test", base)

	customPort := NewServer("http://test:90")
	assert.Equal("http://test:90", customPort.server)
	base, err = customPort.baseUrl()
	assert.Nil(err)
	assert.Equal("http://test:90", base)

	dogbin := Dogbin()
	assert.Equal("del.dog", dogbin.server)
	base, err = dogbin.baseUrl()
	assert.Nil(err)
	assert.Equal("https://del.dog", base)

	haste := Hastebin()
	assert.Equal("hastebin.com", haste.server)
}

func TestGetExists(t *testing.T) {
	assert := assert.New(t)

	got, err := local.Get("exists")
	require.Nil(t, err)

	assert.Equal("exists", got.Slug, "Slug didn't match")
	assert.Equal("works", got.Content, "Content didn't match")
	assert.Equal(12, got.ViewCount, "ViewCount didn't match")
	assert.Equal(false, got.IsUrl, "IsUrl didn't match")
}

func TestGetExistsHaste(t *testing.T) {
	assert := assert.New(t)

	got, err := local.Get("existshaste")
	require.Nil(t, err)

	assert.Equal("existshaste", got.Slug, "Slug didn't match")
	assert.Equal("works", got.Content, "Content didn't match")
	assert.Equal(0, got.ViewCount, "ViewCount should always be zero for haste servers")
	assert.Equal(false, got.IsUrl, "IsUrl should always be false for haste servers")
}

func TestGetNotExists(t *testing.T) {
	got, err := local.Get("notexist")
	require.NotNil(t, err)

	assert.Equal(t, "unable to make request: Document not found.", err.Error())
	assert.Nil(t, got)
}

func TestGetBrokenResponse(t *testing.T) {
	got, err := local.Get("broken")
	require.NotNil(t, err)

	assert.Equal(t, "unable to decode response: invalid character 'w' looking for beginning of value", err.Error())
	assert.Nil(t, got)
}

func TestGetBrokenResponseValidJson(t *testing.T) {
	got, err := local.Get("broken2")
	require.NotNil(t, err)

	assert.Equal(t, "unable to decode response: document is empty", err.Error())
	assert.Nil(t, got)
}

func TestServerErrorEmpty(t *testing.T) {
	got, err := local.Get("whoops")
	require.NotNil(t, err)

	assert.Equal(t, "unable to make request (500 Internal Server Error) and decode response: EOF", err.Error())
	assert.Nil(t, got)
}

func TestServerError(t *testing.T) {
	got, err := local.Get("whoops2")
	require.NotNil(t, err)

	assert.Equal(t, "unable to make request (500 Internal Server Error) and decode response: invalid character 'I' looking for beginning of value", err.Error())
	assert.Nil(t, got)
}

func TestServerErrorValidJson(t *testing.T) {
	got, err := local.Get("whoops3")
	require.NotNil(t, err)

	assert.Equal(t, "unable to make request: 500 Internal Server Error", err.Error())
	assert.Nil(t, got)
}

func TestPutWorks(t *testing.T) {
	assert := assert.New(t)
	got, err := local.Put("works", "random content")

	require.Nil(t, err)

	assert.Equal("works", got.Slug, "Slug didn't match")
	assert.Equal(false, got.IsUrl, "IsUrl didn't match")
	assert.Equal(serverUrl+"/works", got.Url, "Url didn't match")
}

func TestPutUrlWorks(t *testing.T) {
	assert := assert.New(t)
	got, err := local.Put("url", "https://github.com/dogbin/inu")

	require.Nil(t, err)

	assert.Equal("url", got.Slug, "Slug didn't match")
	assert.Equal(true, got.IsUrl, "IsUrl didn't match")
	assert.Equal(serverUrl+"/url", got.Url, "Url didn't match")
}

func TestPutLegacyWorks(t *testing.T) {
	assert := assert.New(t)
	got, err := local.Put("", "random content")

	require.Nil(t, err)

	assert.Equal("hasteup", got.Slug, "Slug didn't match")
	assert.Equal(false, got.IsUrl, "IsUrl didn't match")
	assert.Equal(serverUrl+"/hasteup", got.Url, "Url didn't match")
}

func TestPutDuplicate(t *testing.T) {
	got, err := local.Put("duplicate", "random content")

	require.NotNil(t, err)

	assert.Equal(t, "This URL is already in use, please choose a different one", err.Error())
	assert.Nil(t, got)
}

func TestPutEmpty(t *testing.T) {
	got, err := local.Put("empty", "")

	require.NotNil(t, err)

	assert.Equal(t, "no content was provided", err.Error())
	assert.Nil(t, got)
}
