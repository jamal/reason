package reason

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

type TestResource struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

var testData = []TestResource{
	{1, "The Test"},
	{2, "The Other"},
}

type TestResourceHandler struct {
}

func (trh TestResourceHandler) Path() string {
	return "test"
}

func (trh TestResourceHandler) GetResource(id string) (interface{}, error) {
	for _, data := range testData {
		if strconv.FormatInt(data.ID, 10) == id {
			return data, nil
		}
	}
	return nil, ErrNotFound
}

func (trh TestResourceHandler) ListResource() ([]interface{}, error) {
	list := make([]interface{}, len(testData))
	for k, v := range testData {
		list[k] = interface{}(v)
	}
	return list, nil
}

func (trh TestResourceHandler) CreateResource(resource interface{}) (interface{}, error) {
	if tr, ok := resource.(TestResource); ok {
		tr.ID = 3
		return tr, nil
	}
	return nil, fmt.Errorf("Invalid resource type")
}

func (trh TestResourceHandler) UpdateResource(resource interface{}, data interface{}) (interface{}, error) {
	if tr, ok := resource.(TestResource); ok {
		if v, ok := data.(TestResource); ok {
			tr.Name = v.Name
		}
		return tr, nil
	}
	return nil, fmt.Errorf("Invalid resource type")
}

func (trh TestResourceHandler) DeleteResource(resource interface{}) error {
	return nil
}

type NoHandler struct{}

func (n NoHandler) Path() string {
	return "no"
}

func TestGetter(t *testing.T) {
	var requests = []struct {
		Path       string
		StatusCode int
		Body       string
	}{
		{"/test/1", 200, `{"id":1,"name":"The Test"}`},
		{"/test/3", 404, ``},
		{"/other/1", 404, ``},
		{"/no/1", 404, ``},
	}

	s := New()
	s.Add(TestResource{}, TestResourceHandler{})
	s.Add(TestResource{}, NoHandler{})
	ts := httptest.NewServer(s)
	defer ts.Close()

	for _, request := range requests {
		res, err := http.Get(ts.URL + request.Path)
		if err != nil {
			t.Errorf("%s: expected no error from Get, got %s", request.Path, err.Error())
		}

		if res.StatusCode != request.StatusCode {
			t.Errorf("%s: expected status code %d, got %d", request.Path, request.StatusCode, res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf("%s: expected no error from read, got %s", request.Path, err.Error())
		}

		if string(body) != request.Body {
			t.Errorf("%s: expected body '%s', got '%s'", request.Path, request.Body, body)
		}
	}
}

func TestLister(t *testing.T) {
	var requests = []struct {
		Path       string
		StatusCode int
		Body       string
	}{
		{"/test", 200, `[{"id":1,"name":"The Test"},{"id":2,"name":"The Other"}]`},
		{"/test/", 200, `[{"id":1,"name":"The Test"},{"id":2,"name":"The Other"}]`},
		{"/other", 404, ``},
		{"/no", 404, ``},
	}

	s := New()
	s.Add(TestResource{}, TestResourceHandler{})
	s.Add(TestResource{}, NoHandler{})
	ts := httptest.NewServer(s)
	defer ts.Close()

	for _, request := range requests {
		res, err := http.Get(ts.URL + request.Path)
		if err != nil {
			t.Errorf("%s: expected no error from Get, got %s", request.Path, err.Error())
		}

		if res.StatusCode != request.StatusCode {
			t.Errorf("%s: expected status code %d, got %d", request.Path, request.StatusCode, res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf("%s: expected no error from read, got %s", request.Path, err.Error())
		}

		if string(body) != request.Body {
			t.Errorf("%s: expected body '%s', got '%s'", request.Path, request.Body, body)
		}
	}
}

func TestCreator(t *testing.T) {
	form := url.Values{}
	form.Add("name", "New Test")

	var requests = []struct {
		Path       string
		StatusCode int
		Body       string
		Data       url.Values
	}{
		{"/test", 201, `{"id":3,"name":"New Test"}`, form},
		{"/test/", 307, ``, form},
		{"/other", 404, ``, nil},
		{"/no", 404, ``, nil},
	}

	s := New()
	s.Add(TestResource{}, TestResourceHandler{})
	s.Add(TestResource{}, NoHandler{})
	ts := httptest.NewServer(s)
	defer ts.Close()

	for _, request := range requests {
		res, err := http.PostForm(ts.URL+request.Path, request.Data)
		if err != nil {
			t.Errorf("%s: expected no error from PostForm, got %s", request.Path, err.Error())
		}

		if res.StatusCode != request.StatusCode {
			t.Errorf("%s: expected status code %d, got %d", request.Path, request.StatusCode, res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf("%s: expected no error from read, got %s", request.Path, err.Error())
		}

		if string(body) != request.Body {
			t.Errorf("%s: expected body '%s', got '%s'", request.Path, request.Body, body)
		}
	}
}

func TestUpdater(t *testing.T) {
	form := url.Values{}
	form.Add("name", "Updated Test")

	var requests = []struct {
		Path       string
		StatusCode int
		Body       string
		Data       url.Values
	}{
		{"/test/1", 200, `{"id":1,"name":"Updated Test"}`, form},
		{"/test/3", 404, ``, form},
		{"/other", 404, ``, nil},
		{"/no", 404, ``, nil},
	}

	s := New()
	s.Add(TestResource{}, TestResourceHandler{})
	s.Add(TestResource{}, NoHandler{})
	ts := httptest.NewServer(s)
	defer ts.Close()

	for _, request := range requests {
		res, err := http.PostForm(ts.URL+request.Path, request.Data)
		if err != nil {
			t.Errorf("%s: expected no error from PostForm, got %s", request.Path, err.Error())
		}

		if res.StatusCode != request.StatusCode {
			t.Errorf("%s: expected status code %d, got %d", request.Path, request.StatusCode, res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf("%s: expected no error from read, got %s", request.Path, err.Error())
		}

		if string(body) != request.Body {
			t.Errorf("%s: expected body '%s', got '%s'", request.Path, request.Body, body)
		}
	}
}

func TestDeleter(t *testing.T) {
	var requests = []struct {
		Path       string
		StatusCode int
		Body       string
	}{
		{"/test/1", 200, ``},
		{"/test/3", 404, ``},
		{"/other", 404, ``},
		{"/no", 404, ``},
	}

	s := New()
	s.Add(TestResource{}, TestResourceHandler{})
	s.Add(TestResource{}, NoHandler{})
	ts := httptest.NewServer(s)
	defer ts.Close()

	client := &http.Client{}

	for _, request := range requests {
		req, err := http.NewRequest("DELETE", ts.URL+request.Path, nil)
		if err != nil {
			t.Errorf("%s: expected no error from http.NewRequest, got %s", request.Path, err.Error())
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("%s: expected no error from client.Do, got %s", request.Path, err.Error())
		}

		if res.StatusCode != request.StatusCode {
			t.Errorf("%s: expected status code %d, got %d", request.Path, request.StatusCode, res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Errorf("%s: expected no error from read, got %s", request.Path, err.Error())
		}

		if string(body) != request.Body {
			t.Errorf("%s: expected body '%s', got '%s'", request.Path, request.Body, body)
		}
	}
}
