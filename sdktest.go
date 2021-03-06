package sdktest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/clbanning/x2j"
	"github.com/stretchr/testify/assert"
)

type SDKTester struct {
	mux       *http.ServeMux
	server    *httptest.Server
	respType  string
	t         *testing.T
	xmlRoot   string
	respWant  map[string]interface{}
	dataField string
}

type Options struct {
	RespType  string
	XMLRoot   string
	DataField string
	RespData  []byte
	RespWant  map[string]interface{}
	ReqWant   map[string]interface{}
	URI       string
}

type Stringer interface {
	String() string
}

type Inter interface {
	Int() int
}

func NewSDKTester(t *testing.T, o Options) *SDKTester {
	mux := http.NewServeMux()

	server := httptest.NewServer(mux)

	st := &SDKTester{
		mux:       mux,
		server:    server,
		respType:  o.RespType,
		t:         t,
		xmlRoot:   o.XMLRoot,
		respWant:  o.RespWant,
		dataField: o.DataField,
	}

	if st.xmlRoot == "" {
		st.xmlRoot = "xml"
	}

	st.handleHTTP(o.URI, o.RespData, o.ReqWant)

	return st
}

func (st *SDKTester) checkRequest(req *http.Request, want map[string]interface{}) {
	if want == nil || req == nil {
		st.t.Logf("request want/got is nil, abort\n")
		return
	}

	reqMap := map[string]interface{}{}

	switch st.respType {
	case "query":
		reqData := req.URL.Query()
		for k, v := range reqData {
			reqMap[k] = v[0]
		}
	case "form":
		err := req.ParseForm()
		if err != nil {
			st.t.Errorf("unmarshal request failed %v", err)
			return
		}
		for k, v := range req.Form {
			reqMap[k] = v[0]
		}
	case "json":
		jd := json.NewDecoder(req.Body)
		err := jd.Decode(&reqMap)
		if err != nil {
			st.t.Errorf("unmarshal request failed %v", err)
			return
		}
	case "xml":
		data, _ := ioutil.ReadAll(req.Body)

		err := x2j.Unmarshal(data, &reqMap)
		if err != nil {
			st.t.Errorf("unmarshal request failed %v", err)
			return
		}

		reqMap = reqMap[st.xmlRoot].(map[string]interface{})
	case "formXML":
		data, _ := ioutil.ReadAll(req.Body)
		values, err := url.ParseQuery(string(data))
		if err != nil {
			st.t.Errorf("parse query failed %v", err)
			return
		}
		reqData := []byte(values.Get(st.dataField))
		err = x2j.Unmarshal(reqData, &reqMap)
		if err != nil {
			st.t.Errorf("unmarshal request failed %v", err)
			return
		}

		reqMap = reqMap[st.xmlRoot].(map[string]interface{})
	case "queryJSON": // alipay
		reqData := req.URL.Query().Get(st.dataField)
		err := json.Unmarshal([]byte(reqData), &reqMap)
		if err != nil {
			st.t.Errorf("unmarshal request failed %v", err)
			return
		}
	}

	if len(reqMap) == 0 {
		st.t.Errorf("unmarshal request failed\n")
		return
	}

	for i, v := range want {
		if !assert.Equal(st.t, v, reqMap[i]) {
			st.t.Errorf("%s want %v, got %v", i, v, reqMap[i])
		}
	}
}

func (st *SDKTester) getFieldMap(i interface{}) map[string]interface{} {
	im := map[string]interface{}{}
	val := reflect.ValueOf(i).Elem()
	for i := 0; i < val.NumField(); i++ {
		// anonymous embed struct as first level
		if val.Type().Field(i).Anonymous && val.Type().Field(i).Type.Kind() == reflect.Struct {
			ii := val.Field(i)
			for j := 0; j < ii.NumField(); j++ {

				key := val.Type().Field(i).Type.Field(j).Name
				im[key] = val.Field(i).Field(j).Interface()
			}
			continue
		}

		key := val.Type().Field(i).Name
		im[key] = val.Field(i).Interface()
	}

	return im
}

// CehckResponse check response struct is or not equal want's data
// resp struct pointer
func (st *SDKTester) Test(resp interface{}) {
	if resp == nil || st.respWant == nil {
		st.t.Logf("response want/got is nil, abort\n")
		return
	}

	respMap := st.getFieldMap(resp)
	for i, v := range st.respWant {
		if reflect.DeepEqual(v, respMap[i]) {
			continue
		}
		switch x := respMap[i].(type) {
		case Stringer:
			if !assert.Equal(st.t, v, x.String()) {
				st.t.Errorf("%s want %v, got %v", i, v, respMap[i])
			}
		case map[string]interface{}:
			if value, ok := x["Value"]; ok {
				if !assert.Equal(st.t, v, value) {
					st.t.Errorf("%s want %v, got %v", i, v, respMap[i])
				}
			}
		case Inter:
			if !assert.Equal(st.t, v, x.Int()) {
				st.t.Errorf("%s want %v, got %v", i, v, respMap[i])
			}
		default:
			if !assert.Equal(st.t, v, respMap[i]) {
				st.t.Errorf("%s want %v, got %v", i, v, respMap[i])
			}
		}
	}
}

func (st *SDKTester) handleHTTP(uri string, resp []byte, reqWant map[string]interface{}) {
	st.mux.HandleFunc(uri, func(w http.ResponseWriter, r *http.Request) {
		st.checkRequest(r, reqWant)
		w.Write(resp)
		// w.WriteHeader(200)
	})
}

func (st *SDKTester) Close() {
	st.server.Close()
}

func (st *SDKTester) URL() string {
	return st.server.URL
}
