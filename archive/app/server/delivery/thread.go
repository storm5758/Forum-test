package delivery

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/storm5758/Forum-test/pkg/messages"
)

func (h *Handler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 200

	slugOrID, ok := mux.Vars(r)["slug_or_id"]
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

	id, err := strconv.ParseInt(slugOrID, 10, 64)
	if err == nil {
		t.Id = int32(id)
	} else {
		t.Slug = slugOrID
	}

	t, err = h.Service.UpdateThread(t)
	answer, _ := json.Marshal(t)

	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: err.Error()})
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 200

	slugOrID, ok := mux.Vars(r)["slug_or_id"]
	if !ok {
		return
	}

	var t Thread

	id, err := strconv.ParseInt(slugOrID, 10, 64)
	if err != nil {
		t, err = h.Service.GetThreadBySlug(slugOrID)
	} else {
		t, err = h.Service.GetThreadById(id)
	}

	answer, _ := json.Marshal(t)

	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: err.Error() + slugOrID})
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 201

	slugOrID, ok := mux.Vars(r)["slug_or_id"]
	if !ok {
		return
	}

	var posts []Post

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	err = json.Unmarshal(bytes, &posts)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	posts, err = h.Service.CreatePosts(slugOrID, posts)

	answer, _ := json.Marshal(posts)

	if err != nil {
		switch err.Error() {
		case messages.ThreadNotFound:
			code = 404
			answer, _ = json.Marshal(Error{Message: err.Error() + slugOrID})
			break
		case messages.ParentNotFound:
			code = 409
			answer, _ = json.Marshal(Error{Message: err.Error()})
			break
		default:
			code = 409
			answer, _ = json.Marshal(Error{Message: err.Error()})
		}
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 200

	slugOrID, ok := mux.Vars(r)["slug_or_id"]
	if !ok {
		return
	}
	limit, _ := strconv.ParseInt(r.FormValue("limit"), 10, 64)
	since := r.FormValue("since")
	sort := r.FormValue("sort")
	if sort == "" {
		sort = "flat"
	}
	desc, _ := strconv.ParseBool(r.FormValue("desc"))

	posts, err := h.Service.GetPosts(slugOrID, limit, since, sort, desc)

	answer, _ := json.Marshal(posts)

	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: err.Error()})
	}

	w.WriteHeader(code)
	w.Write(answer)
}

func (h *Handler) CreateVote(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	code := 200

	slugOrID, ok := mux.Vars(r)["slug_or_id"]
	if !ok {
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	var v Vote
	err = json.Unmarshal(bytes, &v)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	thread, err := h.Service.CreateVote(slugOrID, v)
	answer, _ := json.Marshal(thread)

	if err != nil {
		code = 404
		answer, _ = json.Marshal(Error{Message: err.Error()})
	}

	w.WriteHeader(code)
	w.Write(answer)
}
