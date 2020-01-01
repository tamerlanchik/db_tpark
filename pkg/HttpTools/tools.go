package HttpTools

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/http"
)

func StructFromBody(r http.Request, s interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, s)
}

// Inflate body with struct content
func BodyFromStruct(w http.ResponseWriter, s interface{}) error {
	js, err := json.Marshal(s)
	if err != nil {
		return err
	}
	w.Write(js)	//sets unchangeable status 200 if not set already
	return nil
}


type Response struct {
	body  interface{}
	status int
	writer http.ResponseWriter
}

func NewResponse(writer http.ResponseWriter) *Response {
	return &Response{writer: writer}
}

func (r *Response) SetWriter(writer http.ResponseWriter) *Response {
	r.writer = writer
	return r
}

func (r *Response) SetStatus(status int) *Response {
	r.status = status
	return r
}

func (r *Response) SetError(err interface{}) *Response {
	r.body = err
	return r
}

func (r *Response) SetContent(c interface{}) *Response {
	r.body = c
	return r
}

func (r *Response) Copy() Response {
	return *r
}

func (r *Response) Send() {
	if r.body == nil {
		//fmt.Println("Nil error")
	}
	body, err := json.Marshal(r.body)
	if err != nil {
		//fmt.Println("Cannot encode json")
		return
	}
	r.writer.Header().Set("Content-Type", "application/json")
	if r.status == 0 {
		r.status = http.StatusOK
	}
	r.writer.WriteHeader(r.status)
	r.writer.Write(body)
}

func (r *Response) String() string {
	body, _ := json.Marshal(r.body)
	return string(body)
}
