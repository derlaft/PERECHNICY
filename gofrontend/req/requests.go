package req

import (
	"../../entity/entities"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Server struct {
	URL string
	Credentials
}

type Credentials struct {
	User, Token string
}

func (c *Credentials) Save(filename string) error {
	body, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, body, 0600)
	if err != nil {
		return err
	}

	return nil
}

func Load(addr, filename string) (*Server, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := Credentials{}
	err = json.Unmarshal(body, &c)

	if err != nil {
		return nil, err
	}

	s := &Server{addr, c}
	s.Check()

	return s, nil
}

func NewServer(addr, nick, token string) *Server {

	s := &Server{
		addr,
		Credentials{
			User:  nick,
			Token: token,
		},
	}

	s.Check()

	return s
}

func (s *Server) Check() {
	ok, err := s.CheckLogin()
	if err != nil || !ok {
		s.Token = ""
	}

}

func (s *Server) request(method string, params url.Values) ([]byte, error) {
	return requestUrl(s.URL+method, params)
}

func requestUrl(url string, params url.Values) ([]byte, error) {
	resp, err := http.PostForm(url, params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "Request %v\n", params)
	fmt.Fprintf(os.Stderr, "Got body %v\n", string(body))
	return body, nil
}

func (s *Server) values(v url.Values) url.Values {

	if v == nil {
		v = url.Values{}
	}

	v["User"] = []string{s.User}
	if s.Token != "" {
		v["Token"] = []string{s.Token}
	}
	return v
}

func (s *Server) Register() error {

	body, err := s.request("register", s.values(nil))

	if err != nil {
		return err
	}

	tr := make(map[string]interface{})
	err = json.Unmarshal(body, &tr)
	if err != nil {
		return err
	}

	switch tok := tr["Token"].(type) {
	case string:
		s.Token = tok
	case nil:
		switch reason := tr["Reason"].(type) {
		case string:
			return errors.New(reason)
		case nil:
			return errors.New("No token in response")
		}
	}
	return nil
}

func (s *Server) CheckLogin() (bool, error) {

	body, err := s.request("ping", s.values(nil))

	if err != nil {
		return false, err
	}

	tr := make(map[string]bool)
	err = json.Unmarshal(body, &tr)

	if err != nil {
		return false, err
	}

	return tr["Result"], err
}

func (s *Server) GetMap(x, y, w, h int) ([]int, error) {
	body, err := s.request("map", url.Values{
		"X": {strconv.Itoa(x)}, "Y": {strconv.Itoa(y)},
		"W": {strconv.Itoa(w)}, "H": {strconv.Itoa(h)},
	})

	if err != nil {
		return nil, err
	}

	mp := make([]int, 0, 0)
	err = json.Unmarshal(body, &mp)

	if err != nil {
		return nil, err
	}

	return mp, nil
}

func (s *Server) GetData() (*entities.JSONOutput, error) {
	body, err := s.request("get", s.values(nil))

	if err != nil {
		return nil, err
	}

	newState := entities.JSONOutput{}
	err = json.Unmarshal(body, &newState)

	if err != nil {
		return nil, err
	}

	return &newState, nil
}

func (s *Server) Start(prog string) error {

	body, err := s.request("start", s.values(url.Values{
		"Prog": {prog},
	}))

	tr := make(map[string]interface{})
	err = json.Unmarshal(body, &tr)

	if err != nil {
		return err
	}

	result := false
	switch res := tr["Result"].(type) {
	case bool:
		result = res
	}

	if !result {
		return errors.New(fmt.Sprintf("Not started (%v)", tr["Reason"]))
	}

	return nil

}
