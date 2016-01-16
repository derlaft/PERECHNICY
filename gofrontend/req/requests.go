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
)

type Server struct {
	URL         string
	User, Token string
}

func NewServer(addr, nick, token string) *Server {

	s := &Server{
		URL:   addr,
		User:  nick,
		Token: token,
	}

	ok, err := s.CheckLogin()
	if err != nil || !ok {
		s.Token = ""
	}

	return s
}

func (s *Server) request(method string, params url.Values) ([]byte, error) {
	resp, err := http.PostForm(s.URL+method, params)
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
