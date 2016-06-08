package daemon

import (
	"bytes"
	"encoding/json"
	"github.com/asiainfoLDP/datahub/cmd"
	"github.com/asiainfoLDP/datahub/ds"
	log "github.com/asiainfoLDP/datahub/utils/clog"
	"github.com/asiainfoLDP/datahub/utils/logq"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	loginLogged       = false
	loginAuthStr      string
	loginBasicAuthStr string
	gstrUsername      string
	DefaultServer     = "https://hub.dataos.io/api"
)

type UserForJson struct {
	Username string `json:"username", omitempty`
}

type tk struct {
	Token string `json:"token"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := DefaultServer + "/" //r.URL.Path
	//r.ParseForm()

	if _, ok := r.Header["Authorization"]; !ok {

		if !loginLogged {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	userjsonbody, _ := ioutil.ReadAll(r.Body)
	userforjson := UserForJson{}
	if err := json.Unmarshal(userjsonbody, &userforjson); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	gstrUsername = userforjson.Username
	log.Println("login to", url, "Authorization:", r.Header.Get("Authorization"), gstrUsername)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", r.Header.Get("Authorization"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	result := &ds.Result{}
	log.Println("login return", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(body))

		result.Data = &tk{}
		if err = json.Unmarshal(body, result); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusServiceUnavailable)

			l := log.Println(resp.StatusCode, string(body))
			logq.LogPutqueue(l)
			return
		} else {

			loginAuthStr = "Token " + result.Data.(*tk).Token //must be pointer
			loginLogged = true
			log.Println(loginAuthStr)
			loginBasicAuthStr = r.Header.Get("Authorization")
		}
	} else if resp.StatusCode == http.StatusForbidden {
		body, _ := ioutil.ReadAll(resp.Body)
		l := log.Println(string(body))
		logq.LogPutqueue(l)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	} else {
		w.WriteHeader(resp.StatusCode)
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("logout.")
	if loginAuthStr == "" {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		loginAuthStr = ""
		w.WriteHeader(http.StatusOK)
	}
}

func commToServer(method, path string, buffer []byte, w http.ResponseWriter) (body []byte, err error) {
	//Trace()
	s := log.Info("daemon: connecting to", DefaultServer+path)
	logq.LogPutqueue(s)
	req, err := http.NewRequest(strings.ToUpper(method), DefaultServer+path, bytes.NewBuffer(buffer))
	if len(loginAuthStr) > 0 {
		req.Header.Set("Authorization", loginAuthStr)
	}

	//req.Header.Set("User", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		d := ds.Result{Code: cmd.ErrorServiceUnavailable, Msg: err.Error()}
		body, e := json.Marshal(d)
		if e != nil {
			log.Error(e)
			return body, e
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write(body)
		return body, err
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	body, err = ioutil.ReadAll(resp.Body)
	w.Write(body)
	log.Info(resp.StatusCode, string(body))
	return
}

func commToServerGetRsp(method, path string, buffer []byte) (resp *http.Response, err error) {

	s := log.Info("daemon: connecting to", DefaultServer+path)
	logq.LogPutqueue(s)
	req, err := http.NewRequest(strings.ToUpper(method), DefaultServer+path, bytes.NewBuffer(buffer))
	if len(loginAuthStr) > 0 {
		req.Header.Set("Authorization", loginAuthStr)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		return resp, err
	}

	return resp, nil
}

func whoamiHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	code := 0
	msg := "OK"
	httpcode := http.StatusOK
	userstru := &ds.User{}
	if len(loginAuthStr) > 0 {
		userstru.Username = gstrUsername
	} else {
		userstru.Username = ""
		code = cmd.ErrorUnAuthorization
		msg = "Not login."
		httpcode = http.StatusUnauthorized
	}

	b, _ := buildResp(code, msg, userstru)
	w.WriteHeader(httpcode)
	w.Write(b)
}
