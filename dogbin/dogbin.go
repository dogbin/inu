// Package dogbin provides a simple go client library for dogbin and hastebin.
package dogbin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Server defines the dogbin or hastebin server to communicate with
type Server struct {
	server string
}

// UploadRequest represents the json used internally for uploads to the dogbin extended API
type UploadRequest struct {
	Slug    string `json:"slug"`
	Content string `json:"content"`
}

// UploadResult represents the json returned for upload requests
type UploadResult struct {
	IsUrl bool   `json:"isUrl"`
	Slug  string `json:"key"`
	Url   string `json:"-"`
}

// Message represents the json format used by the server for errors
type Message struct {
	Message string `json:"message"`
}

// Wrapper represents the JSON response from hastebin/dogbin which in the case of dogbin simply exists for legacy purposes and
// wraps around the actual document.
type Wrapper struct {
	Content  string    `json:"data"`
	Document *Document `json:"document,omitempty"`
	Slug     string    `json:"key"`
}

// Document represents the dogbin document structure and is used for both dogbin and hastebin here
type Document struct {
	Slug      string `json:"_id"`
	IsUrl     bool   `json:"isUrl"`
	Content   string `json:"content"`
	ViewCount int    `json:"viewCount"`
}

// Put uploads content to the server,
// if a slug is supplied it is assumed that the server supports
// the extended api used by dogbin.
func (d Server) Put(slug string, content string) (*UploadResult, error) {
	if content == "" {
		return nil, errors.New("no content was provided")
	}
	u, err := d.putUrl()
	if err != nil {
		return nil, err
	}

	var mime = "text/plain"
	var data = []byte(content)

	if slug != "" {
		// Make json requests for dogbin servers with custom slug support
		mime = "application/json"
		data, _ = json.Marshal(UploadRequest{
			Slug:    slug,
			Content: content,
		})
	}
	r, err := http.Post(u, mime, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	if r.StatusCode != 200 {
		message := new(Message)
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(message)
		if message.Message == "" {
			message.Message = r.Status
		}
		return nil, errors.New(message.Message)
	}

	result := new(UploadResult)
	defer r.Body.Close()
	_ = json.NewDecoder(r.Body).Decode(result)

	result.Url, err = d.slugUrl(result.Slug)

	return result, err
}

// Get gets a *Document from the server for the supplied slug
func (d Server) Get(slug string) (*Document, error) {
	u, err := d.getUrl(slug)
	if err != nil {
		return nil, err
	}
	r, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != 200 {
		message := new(Message)
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(message)
		if message.Message == "" {
			message.Message = r.Status
		}
		return nil, errors.New(message.Message)
	}

	wrapper := new(Wrapper)
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(wrapper)

	document := Document{
		Slug:    wrapper.Slug,
		Content: wrapper.Content,
	}
	if wrapper.Document != nil {
		document.IsUrl = wrapper.Document.IsUrl
		document.ViewCount = wrapper.Document.ViewCount
	}

	return &document, err
}

// baseUrl returns the base Url for the server, assuming https if no scheme has been supplied
func (d Server) baseUrl() (string, error) {
	srv, err := url.Parse(d.server)
	if err != nil {
		return "", err
	}
	if srv.Scheme == "" {
		srv.Scheme = "https"
	}
	return srv.String(), nil
}

// slugUrl returns the Url of the document with the supplied slug
func (d Server) slugUrl(slug string) (string, error) {
	base, err := d.baseUrl()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s/%s", base, slug), nil
}

// getUrl returns the Url to get details about the document with the supplied slug
func (d Server) getUrl(slug string) (string, error) {
	base, err := d.baseUrl()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s/documents/%s", base, slug), nil
}

// putUrl returns the Url of the upload endpoint
func (d Server) putUrl() (string, error) {
	base, err := d.baseUrl()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s/documents", base), nil
}

// NewServer returns a new Server configured for the supplied dogbin/hastebin instance
func NewServer(server string) Server {
	return Server{server: server}
}

// Dogbin returns a Server instance configured for the public del.dog dogbin instance
func Dogbin() Server {
	return Server{server: "del.dog"}
}

// Hastebin returns a Server instance configured for the public hastebin.com hastebin instance
func Hastebin() Server {
	return Server{server: "hastebin.com"}
}
