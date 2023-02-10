package web_test

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/maxmcd/steady/internal/testsuite"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	testsuite.Suite
	httpClient *http.Client
}

func TestSuite(t *testing.T) { suite.Run(t, new(Suite)) }

func (suite *Suite) BeforeTest(_, _ string) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.httpClient = &http.Client{
		Jar: jar,
	}
}

func (suite *Suite) findInDoc(doc *goquery.Document, selector string) (text string) {
	found := doc.Find(selector)
	suite.True(1 == found.Length(), "selector %q not found in doc", selector)
	return found.Text()
}

func (suite *Suite) webRequest(req *http.Request, err error) (resp *http.Response, doc *goquery.Document) {
	t := suite.T()
	if err != nil {
		t.Fatal(err)
	}
	return suite.webResponse(suite.httpClient.Do(req))
}

func (suite *Suite) webResponse(resp *http.Response, err error) (_ *http.Response, doc *goquery.Document) {
	t := suite.T()
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if doc, err = goquery.NewDocumentFromReader(resp.Body); err != nil {
		t.Fatal(err)
	}
	return resp, doc
}

func (suite *Suite) postForm(url string, body url.Values) (_ *http.Response, doc *goquery.Document) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body.Encode()))
	if err != nil {
		suite.T().Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return suite.webRequest(req, nil)
}

func (suite *Suite) TestWeb() {
	// t := suite.T()

	es, addr := suite.NewWebServer()

	suite.Run("index page includes login link", func() {
		resp, doc := suite.webRequest(http.NewRequest(http.MethodGet, addr, nil))
		suite.Equal(http.StatusOK, resp.StatusCode)

		sel := doc.Find("a[href$='/login']")
		suite.Equal(1, sel.Length())
		suite.Equal("login / signup", sel.Text())
	})

	suite.Run("can sign up", func() {
		resp, doc := suite.webRequest(http.NewRequest(http.MethodGet, addr+"/login", nil))
		suite.Equal(http.StatusOK, resp.StatusCode)
		suite.findInDoc(doc, "form[action$='/login'] input[name$='username_or_email']")
		suite.findInDoc(doc, "form[action$='/signup'] input[name$='username']")
		suite.findInDoc(doc, "form[action$='/signup'] input[name$='email']")

		signupForm := url.Values{"username": {"steady"}, "email": {"steady"}}
		resp, doc = suite.postForm(addr+"/signup", signupForm)
		suite.Equal(http.StatusBadRequest, resp.StatusCode)
		suite.Contains(doc.Find("form[action$='/signup'] .error").Text(), "email address is invalid")

		signupForm = url.Values{"username": {""}, "email": {""}}
		resp, doc = suite.postForm(addr+"/signup", signupForm)
		suite.Equal(http.StatusBadRequest, resp.StatusCode)
		suite.Contains(doc.Find("form[action$='/signup'] .error").Text(), "blank")

		signupForm = url.Values{"username": {"steady"}, "email": {"steady@steady"}}
		resp, doc = suite.postForm(addr+"/signup", signupForm)
		suite.Equal(http.StatusOK, resp.StatusCode)
		suite.Contains(doc.Find(".flash").Text(), "login link is on its way to your inbox")

		resp, doc = suite.webRequest(http.NewRequest(http.MethodGet, addr+es.Emails[0], nil))
		suite.Equal(http.StatusOK, resp.StatusCode)
		suite.Equal("profile",
			suite.findInDoc(doc, ".header a[href$='/@steady']"))
		suite.findInDoc(doc, ".header a[href$='/logout']")
	})

}
