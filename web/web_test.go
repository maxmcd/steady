package web_test

import (
	"fmt"
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
	suite.Suite.BeforeTest("", "")

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
	suite.Require().True(1 == found.Length(), "selector %q not found in doc", selector)
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

func (suite *Suite) TestUserSignup() {
	es, addr := suite.NewWebServer()
	require := suite.Require()
	suite.Run("index page includes login link", func() {
		resp, doc := suite.webRequest(http.NewRequest(http.MethodGet, addr, nil))
		require.Equal(http.StatusOK, resp.StatusCode)

		require.Equal("login / signup", suite.findInDoc(doc, "a[href$='/login']"))
	})

	suite.Run("can sign up", func() {
		resp, doc := suite.webRequest(http.NewRequest(http.MethodGet, addr+"/login", nil))
		require.Equal(http.StatusOK, resp.StatusCode)
		suite.findInDoc(doc, "form[action$='/login'] input[name$='username_or_email']")
		suite.findInDoc(doc, "form[action$='/signup'] input[name$='username']")
		suite.findInDoc(doc, "form[action$='/signup'] input[name$='email']")

		signupForm := url.Values{"username": {"steady"}, "email": {"steady@steady"}}
		resp, doc = suite.postForm(addr+"/signup", signupForm)
		require.Equal(http.StatusOK, resp.StatusCode)
		suite.Contains(suite.findInDoc(doc, ".flash"), "login link is on its way to your inbox")

		resp, doc = suite.webRequest(http.NewRequest(http.MethodGet, addr+es.LatestEmail(), nil))
		fmt.Println(resp.Cookies())
		require.Equal(http.StatusOK, resp.StatusCode)
		require.Equal("profile",
			suite.findInDoc(doc, ".header a[href$='/@steady']"))
		suite.findInDoc(doc, ".header a[href$='/logout']")
	})

	suite.Run("signed in user redirects from login page", func() {
		// Cookie store is still signed in from previous test
		resp, _ := suite.webRequest(http.NewRequest(http.MethodGet, addr+"/login", nil))
		suite.Equal(resp.Request.URL.Path, "/")
	})

	suite.Run("can log out", func() {
		// Cookie store is still signed in from previous test
		resp, doc := suite.webRequest(http.NewRequest(http.MethodGet, addr+"/logout", nil))
		suite.Equal(resp.Request.URL.Path, "/")
		suite.Contains(suite.findInDoc(doc, ".flash"), "logged out")
		suite.Equal("login / signup", suite.findInDoc(doc, "a[href$='/login']"))
	})

	suite.Run("cannot sign up with invalid data", func() {
		for _, tt := range []struct {
			form          url.Values
			expectedError string
		}{
			{
				form:          url.Values{"username": {"stable"}, "email": {"invalid"}},
				expectedError: "email address is invalid",
			},
			{
				form:          url.Values{"username": {"steady"}, "email": {"steady@steady"}},
				expectedError: "a user with this username already exists",
			},
			{
				form:          url.Values{"username": {"stable"}, "email": {""}},
				expectedError: "email address cannot be blank",
			},
			{
				form:          url.Values{"username": {""}, "email": {"foo@bar"}},
				expectedError: "username cannot be blank",
			},
		} {
			resp, doc := suite.postForm(addr+"/signup", tt.form)
			require.Equal(http.StatusBadRequest, resp.StatusCode)
			suite.Contains(suite.findInDoc(doc, "form[action$='/signup'] .error"), tt.expectedError)
		}
	})

	suite.Run("can log in again", func() {
		signupForm := url.Values{"username_or_email": {"steady"}}
		resp, doc := suite.postForm(addr+"/login", signupForm)
		suite.Equal(http.StatusOK, resp.StatusCode)
		suite.Contains(suite.findInDoc(doc, ".flash"), "login link is on its way to your inbox")

		resp, doc = suite.webRequest(http.NewRequest(http.MethodGet, addr+es.LatestEmail(), nil))
		fmt.Println(resp.Cookies())
		suite.Equal(http.StatusOK, resp.StatusCode)
		suite.Equal("profile",
			suite.findInDoc(doc, ".header a[href$='/@steady']"))
		suite.findInDoc(doc, ".header a[href$='/logout']")
	})

	suite.Run("cannot log in with invalid data", func() {
		for _, tt := range []struct {
			form          url.Values
			expectedError string
		}{
			{
				form:          url.Values{"username_or_email": {"stable"}},
				expectedError: "A user with this username or email could not be found",
			},
			{
				form:          url.Values{"username_or_email": {""}},
				expectedError: "username or email cannot be blank",
			},
		} {
			resp, doc := suite.postForm(addr+"/login", tt.form)
			require.Equal(http.StatusBadRequest, resp.StatusCode)
			suite.Contains(suite.findInDoc(doc, "form[action$='/login'] .error"), tt.expectedError)
		}
	})

	suite.Run("can log out again", func() {
		// Cookie store is still signed in from previous test
		resp, doc := suite.webRequest(http.NewRequest(http.MethodGet, addr+"/logout", nil))
		require.Equal(resp.Request.URL.Path, "/")
		require.Contains(suite.findInDoc(doc, ".flash"), "logged out")
		require.Equal("login / signup", suite.findInDoc(doc, "a[href$='/login']"))
	})
}

func (suite *Suite) TestRunApplication() {
	suite.StartMinioServer()
	_, _ = suite.NewDaemon()
	require := suite.Require()
	lb := suite.NewLB()
	_, addr := suite.NewWebServer()

	for _, tt := range []struct {
		form          url.Values
		expectedError string
	}{
		{
			form:          url.Values{"index.ts": {""}},
			expectedError: "Application source cannot be empty",
		},
	} {
		resp, doc := suite.postForm(addr+"/application", tt.form)
		require.Equal(http.StatusBadRequest, resp.StatusCode)
		suite.Contains(suite.findInDoc(doc, "form[action$='/application'] .error"), tt.expectedError)
	}

	signupForm := url.Values{"index.ts": {suite.LoadExampleScript("http")}}
	resp, doc := suite.postForm(addr+"/application", signupForm)
	suite.Equal(http.StatusOK, resp.StatusCode)
	appURL := doc.Find("a.app-url").Text()
	fmt.Println(appURL)

	{
		u, _ := url.Parse(appURL)
		req, err := http.NewRequest(http.MethodGet, "http://"+lb.PublicServerAddr(), nil)
		req.Host = u.Host
		suite.Require().NoError(err)
		resp, err := http.DefaultClient.Do(req)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusOK, resp.StatusCode)
	}
}
