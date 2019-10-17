package dogbin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Super basic dogbin/hastebin client library

type Dogbin struct {
	server string
	legacy bool
}

type UploadRequest struct {
	Slug    string `json:"slug"`
	Content string `json:"content"`
}

type UploadResult struct {
	IsUrl bool   `json:"isUrl"`
	Slug  string `json:"key"`
	Url   string `json:"-"`
}

type Message struct {
	Message string `json:"message"`
}

type Wrapper struct {
	Content  string    `json:"data"`
	Document *Document `json:"document,omitempty"`
	Slug     string    `json:"key"`
}

type Document struct {
	Slug      string `json:"_id"`
	IsUrl     bool   `json:"isUrl"`
	Content   string `json:"content"`
	ViewCount int    `json:"viewCount"`
}

func (d Dogbin) Put(slug string, content string) (*UploadResult, error) {
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
		err = json.NewDecoder(r.Body).Decode(message)
		if message.Message == "" {
			message.Message = r.Status
		}
		return nil, errors.New(message.Message)
	}

	result := new(UploadResult)
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(result)

	result.Url, err = d.slugUrl(result.Slug)

	return result, err
}

func (d Dogbin) Get(slug string) (*Document, error) {
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
		err = json.NewDecoder(r.Body).Decode(message)
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

func (d Dogbin) baseUrl() (string, error) {
	srv, err := url.Parse(d.server)
	if err != nil {
		return "", err
	}
	if srv.Scheme == "" {
		srv.Scheme = "https"
	}
	return srv.String(), nil
}

func (d Dogbin) slugUrl(slug string) (string, error) {
	base, err := d.baseUrl()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s/%s", base, slug), nil
}

func (d Dogbin) getUrl(slug string) (string, error) {
	base, err := d.baseUrl()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s/documents/%s", base, slug), nil
}

func (d Dogbin) putUrl() (string, error) {
	base, err := d.baseUrl()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s/documents", base), nil
}

func New(server string) Dogbin {
	return Dogbin{server: server}
}

// The del.dog public dogbin instance
func Default() Dogbin {
	return Dogbin{server: "del.dog"}
}

func Hastebin() Dogbin {
	return Dogbin{server: "hastebin.com"}
}
