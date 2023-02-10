package web_test

import (
	"net/http"
	"testing"

	"github.com/maxmcd/steady/internal/testsuite"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	testsuite.Suite
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (suite *Suite) TestWeb() {
	// t := suite.T()

	addr := suite.NewWebServer()

	suite.Run("index page includes login link", func() {
		resp, doc := suite.WebRequest(http.NewRequest(http.MethodGet, addr, nil))
		suite.Equal(http.StatusOK, resp.StatusCode)

		sel := doc.Find("a[href$='/login']")
		suite.Equal(1, sel.Length())
		suite.Equal("login / signup", sel.Text())
	})
	suite.Run("", func() {
		resp, doc := suite.WebRequest(http.NewRequest(http.MethodGet, addr+"/login", nil))
		suite.Equal(http.StatusOK, resp.StatusCode)

		suite.Equal(1, doc.Find("form[action$='/login'] input[name$='username_or_email']").Length())
		suite.Equal(1, doc.Find("form[action$='/signup'] input[name$='username']").Length())
		suite.Equal(1, doc.Find("form[action$='/signup'] input[name$='email']").Length())
	})

}
