package reason

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"sync"

	"github.com/julienschmidt/httprouter"
)

// Server test
type Server struct {
	router *httprouter.Router

	formCacheLock sync.RWMutex
	formCache     map[reflect.Type][]formField
}

// New creates a new instance of Server.
func New() *Server {
	s := &Server{}
	s.router = httprouter.New()
	s.formCache = make(map[reflect.Type][]formField)

	s.router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return s
}

// Add a resource to be handled.
func (s *Server) Add(resourceSchema interface{}, handler ResourceHandler) {
	path := handler.Path()

	if getter, ok := handler.(Getter); ok {
		s.router.GET("/"+path+"/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			s.getRequest(w, r, ps.ByName("id"), getter)
		})
	}
	if lister, ok := handler.(Lister); ok {
		s.router.GET("/"+path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			s.listRequest(w, r, lister)
		})
	}
	if creator, ok := handler.(Creator); ok {
		fn := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			data, err := s.parseForm(r, resourceSchema)
			if err != nil {
				s.writeError(w, err)
			} else {
				s.createRequest(w, r, creator, data)
			}
		}
		s.router.POST("/"+path, fn)
		s.router.PUT("/"+path, fn)
	}
	if updater, ok := handler.(Updater); ok {
		fn := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			data, err := s.parseForm(r, resourceSchema)
			if err != nil {
				s.writeError(w, err)
			} else {
				s.updateRequest(w, r, ps.ByName("id"), updater, data)
			}
		}
		s.router.POST("/"+path+"/:id", fn)
	}
	if deleter, ok := handler.(Deleter); ok {
		fn := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			s.deleteRequest(w, r, ps.ByName("id"), deleter)
		}
		s.router.DELETE("/"+path+"/:id", fn)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) getRequest(w http.ResponseWriter, r *http.Request, id string, getter Getter) {
	res, err := getter.GetResource(id)
	if err != nil {
		s.writeError(w, err)
	} else {
		s.writeResource(w, http.StatusOK, res)
	}
}

func (s *Server) listRequest(w http.ResponseWriter, r *http.Request, lister Lister) {
	list, err := lister.ListResource()
	if err != nil {
		s.writeError(w, err)
	} else {
		s.writeResourceList(w, http.StatusOK, list)
	}
}

func (s *Server) createRequest(w http.ResponseWriter, r *http.Request, creator Creator, data interface{}) {
	response, err := creator.CreateResource(data)
	if err != nil {
		s.writeError(w, err)
	} else {
		s.writeResource(w, http.StatusCreated, response)
	}
}

func (s *Server) updateRequest(w http.ResponseWriter, r *http.Request, id string, updater Updater, data interface{}) {
	res, err := updater.GetResource(id)
	if err != nil {
		s.writeError(w, err)
		return
	}

	response, err := updater.UpdateResource(res, data)
	if err != nil {
		s.writeError(w, err)
		return
	}

	s.writeResource(w, http.StatusOK, response)
}

func (s *Server) deleteRequest(w http.ResponseWriter, r *http.Request, id string, deleter Deleter) {
	res, err := deleter.GetResource(id)
	if err != nil {
		s.writeError(w, err)
		return
	}

	err = deleter.DeleteResource(res)
	if err != nil {
		s.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) writeResource(w http.ResponseWriter, status int, res interface{}) {
	out, err := json.Marshal(res)
	if err != nil {
		log.Printf("Failed to marshal resource to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Write(out)
}

func (s *Server) writeResourceList(w http.ResponseWriter, status int, list []interface{}) {
	out, err := json.Marshal(list)
	if err != nil {
		log.Printf("Failed to marshal resource to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Write(out)
}

func (s *Server) writeError(w http.ResponseWriter, err error) {
	if err == ErrNotFound {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		log.Printf("Unhandled error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
