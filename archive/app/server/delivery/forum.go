package delivery

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/storm5758/Forum-test/pkg/messages"
)

func (h *Handler) CreateForum(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 201

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	var newForum NewForum
	err = json.Unmarshal(bytes, &newForum)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	forum, err := h.Service.CreateForum(newForum)

	var answer []byte

	if err != nil {
		if err.Error() == messages.UserNotFound {
			code = 404
			answer, _ = json.Marshal(Error{Message: err.Error() + newForum.User})
		} else {
			code = 409
			answer, _ = json.Marshal(forum)
		}

	} else {
		answer, _ = json.Marshal(forum)
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) GetForum(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	code := 200

	slug, ok := mux.Vars(r)["slug"]
	if !ok {
		return
	}

	forum, err := h.Service.GetForum(slug)

	var answer []byte

	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: messages.ForumNotFound + slug})
	} else {
		answer, _ = json.Marshal(forum)
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) CreateThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 201

	forum, ok := mux.Vars(r)["slug"]
	if !ok {
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	var t Thread
	err = json.Unmarshal(bytes, &t)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	t.Forum = forum
	if t.Created == "" {
		t.Created = time.Now().Format(time.RFC3339Nano)
	}

	thread, err := h.Service.CreateThread(t)

	var answer []byte

	if err != nil {
		if err.Error() == messages.UserNotFound || err.Error() == messages.ForumNotFound {
			code = 404
			answer, _ = json.Marshal(Error{Message: err.Error()})
		} else {
			code = 409
			answer, _ = json.Marshal(thread)
		}
	} else {
		answer, _ = json.Marshal(thread)
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) GetThreads(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 200

	forum, ok := mux.Vars(r)["slug"]
	if !ok {
		return
	}

	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64)
	since := r.FormValue("since")
	desc, err := strconv.ParseBool(r.FormValue("desc"))
	if err != nil {
		desc = false
	}

	threads, err := h.Service.GetThreads(forum, limit, since, desc)

	var answer []byte

	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: messages.ForumNotFound + forum})
	} else {
		answer, _ = json.Marshal(threads)
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) GetUsersByForum(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 200

	forum, ok := mux.Vars(r)["slug"]
	if !ok {
		return
	}

	limit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64)
	since := r.FormValue("since")
	desc, err := strconv.ParseBool(r.FormValue("desc"))
	if err != nil {
		desc = false
	}

	_, err = h.Service.GetForum(forum)

	var answer []byte
	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: messages.ForumNotFound + forum})
	} else {
		users, _ := h.Service.GetUsersByForum(forum, limit, since, desc)
		answer, _ = json.Marshal(users)
	}
	w.WriteHeader(code)
	w.Write(answer)
}
