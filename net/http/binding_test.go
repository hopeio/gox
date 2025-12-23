package http

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type User struct {
	ID    int    `uri:"id"`
	Name  string `json:"name"`
	Age   int    `header:"age"`
	Phone string `query:"phone"`
}

func TestBind(t *testing.T) {
	done := make(chan struct{}, 1)
	http.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		var u User
		err := Bind(r, &u)
		if err != nil {
			t.Error(err)
		}
		t.Log(u)
		assert.Equal(t, 1, u.ID)
		assert.Equal(t, "test", u.Name)
		done <- struct{}{}
	})
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			t.Fatal(err)
		}
	}()
	time.Sleep(time.Second)
	req, err := http.NewRequest("POST", "http://localhost:8080/user/1?phone=123", bytes.NewBufferString(`{"name":"test"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Age", "16")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	<-done
}

type User2 struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Phone string `json:"phone"`
	User  User   `json:"user"`
}

func TestBind2(t *testing.T) {

	http.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		var u User2
		err := Bind(r, &u)
		if err != nil {
			t.Error(err)
		}
		t.Log(u)
		if u.ID == 1 {
			assert.Equal(t, "test", u.Name)
		}
		if u.ID == 2 {
			assert.Equal(t, "test2", u.Name)
		}

	})
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			t.Fatal(err)
		}
	}()
	time.Sleep(time.Second)
	req, err := http.NewRequest("POST", "http://localhost:8080/user/1?phone=123", bytes.NewBufferString(`{"name":"test","user":{"id":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Age", "16")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest("POST", "http://localhost:8080/user/2?phone=007", bytes.NewBufferString(`{"name":"test2"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Age", "18")
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
}
