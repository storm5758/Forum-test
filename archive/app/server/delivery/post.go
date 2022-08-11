package delivery

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 200

	id, ok := mux.Vars(r)["id"]
	if !ok {
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	rel := r.URL.Query()["related"]
	related := make([]string, 0)
	for _, r := range rel {
		related = append(related, strings.Split(r, ",")...)
	}

	posts, err := h.Service.GetPostAccount(idInt, related)

	var answer []byte
	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: err.Error()})
	} else {
		answer, _ = json.Marshal(posts)
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 200

	id, ok := mux.Vars(r)["id"]
	if !ok {
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	var p Post
	err = json.Unmarshal(bytes, &p)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	post, err := h.Service.UpdatePost(idInt, p)

	var answer []byte
	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: err.Error()})
	} else {
		answer, _ = json.Marshal(post)
	}

	w.WriteHeader(code)
	w.Write(answer)
}
