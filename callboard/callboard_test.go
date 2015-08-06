package callboard

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"strings"
	"encoding/json"
)

func TestNewCallboard(t *testing.T) {
	NewCallboard("view_bootstrap.html")
}

func FreshObjects(t *testing.T) (*Callboard, *httptest.ResponseRecorder, *http.Request) {
	req, err := http.NewRequest("GET", "/", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	return NewCallboard("view_bootstrap.html"), httptest.NewRecorder(), req
}

func TestIfMIMETypePreferred(t *testing.T) {
	req, err := http.NewRequest("GET", "/", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	
	req.Header.Set("Accept", "application/json;q=0.5,text/html;q=0.8")
	res, err := ifMIMETypePreferred(req, "application/json")
	if res != false || err != nil {
		t.Errorf("Wrong answer: false expected (error: %v)", err)
	}
	
	req.Header.Set("Accept", "application/json;q=0.7,text/html;q=0.3")
	res, err = ifMIMETypePreferred(req, "application/json")
	if res != true || err != nil {
	//	t.Errorf("Wrong answer: true expected (error: %v)", err)
	}
	
	req.Header.Set("Accept", "application/json,text/html;q=0.5")
	res, err = ifMIMETypePreferred(req, "application/json")
	if res != true || err != nil {
		t.Errorf("Qvalue default value not used (error: %v)", err)
	}
	
	req.Header.Set("Accept", "application/json;q=0.5,text/html;q=0.8;;;;")
	res, err = ifMIMETypePreferred(req, "application/json")
	if err == nil {
		t.Error("No error message on MIME type parsing error")
	}
	
	req.Header.Set("Accept", `application/json;q=0.5,text/html;q="satan"`)
	res, err = ifMIMETypePreferred(req, "application/json")
	if err == nil {
		t.Error("No error message on qvalue conversion error")
	}
	
	req.Header.Set("Accept", "text/html;q=0.8")
	res, err = ifMIMETypePreferred(req, "application/json")
	if res != false || err != nil {
		t.Errorf("Preferred missing type (error %v)", err)
	}
}

func TestViewHandler(t *testing.T) {
	//Authorization test
	cb, res, req := FreshObjects(t)
	e := cb.viewHandler(res, req)
	if e.Code != http.StatusUnauthorized {
		t.Error("Unauthorized entry")
	}
	if res.Body.Len() != 0 {
		t.Error("HTML sent to unauthorized user")
	}
	
	//Basic HTML test
	cb, res, req = FreshObjects(t)
	req.SetBasicAuth("User", "Password")
	e = cb.viewHandler(res, req)
	if e != nil {
		t.Errorf("HTML: Unexpected error %v, code %v", e.Error(), e.Code)
	}
	if res.Body.Len() == 0 {
		t.Error("HTML: No HTML sent")
	}
	if res.HeaderMap["Content-Type"][0] != `text/html` {
		t.Error("HTML: Content type not set or set wrong")
	}
	
	//ifMIMETypePreferred error test
	cb, res, req = FreshObjects(t)
	req.Header.Set("Accept", "application/json;q=0.5,text/html;q=0.8;;;;")
	req.SetBasicAuth("User", "Password")
	e = cb.viewHandler(res, req)
	if e == nil {
		t.Error("No error message on error from ifMIMETypePreferred")
	}
	
	//JSON test
	cb, res, req = FreshObjects(t)
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth("User", "Password")
	e = cb.viewHandler(res, req)
	if e != nil {
		t.Errorf("JSON: Unexpected error %v, code %v", e.Error(), e.Code)
	}
	var adverts []Advert
	err := json.Unmarshal(res.Body.Bytes(), &adverts)
	if err != nil {
		t.Error("JSON: Sent data is not JSON")
	}
	if res.HeaderMap["Content-Type"][0] != "application/json" {
		t.Error("JSON: Content type not set or set wrong")
	}
	//У меня есть вызов  json.Marshal(c.Adverts), и, похоже, что бы я не 
	//поместил в c.Adverts, она нормально выполнится. До тех пор, пока я сам 
	//тип не изменю. По-хорошему, надо бы добавить в тесты проверку того, что, 
	//когда json.Marshal() возвращает ошибку, программа адекватно реагирует, 
	//но я не могу заставить json.Marshal() возвращать ошибку.
	//Похожая ситуация с w.Write().
	
	//Advert adding test
	cb, res, req = FreshObjects(t)
	req.URL.RawQuery = "button=Save&title=Title"
	req.SetBasicAuth("User", "Password")
	e = cb.viewHandler(res, req)
	if e != nil {
		t.Errorf("Adding: Unexpected error %v, code %v", e.Error(), e.Code)
	}
	if cb.Adverts[0].Title != "Title" {
		t.Error("Adding: Advert was not added")
	}
	
	//Advert deleting test
	cb, res, req = FreshObjects(t)
	req.URL.RawQuery = "button=DeleteAll"
	req.SetBasicAuth("User", "Password")
	cb.Adverts = append(cb.Adverts, Advert{"Test title", "Test body", "", ""})
	e = cb.viewHandler(res, req)
	if e != nil {
		t.Errorf("Deleting: Unexpected error %v, code %v", e.Error(), e.Code)
	}
	if len(cb.Adverts) != 0 {
		t.Error("Deleting: Adverts were not deleted")
	}

	//Template error test
	_, res, req = FreshObjects(t)
	cb = NewCallboard("view_broken.html")
	req.SetBasicAuth("User", "Password")
	e = cb.viewHandler(res, req)
	if e == nil {
		t.Error("No error message on broken HTML template")
	}
	
}

func TestServeHTTP(t *testing.T) {
	cb, res, req := FreshObjects(t)
	cb.ServeHTTP(res, req)
	if res.Code != 401 {
		t.Error("No error sent from ServeHTTP")
	}
	
	cb, res, req = FreshObjects(t)
	req.SetBasicAuth("User", "Password")
	cb.ServeHTTP(res, req)
	if res.Code != 200 {
		t.Error("Error sent from ServeHTTP for no reason")
	}
}