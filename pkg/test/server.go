package test

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
)

// TInterface covers most of the methods in the testing package's T.
type TInterface interface {
	Fail()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	FailNow()
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Failed() bool
	Parallel()
	Skip(args ...interface{})
	Skipf(format string, args ...interface{})
	SkipNow()
	Skipped() bool
}

// ExpectedCall has information about an expected test server call.
type ExpectedCall struct {
	Path   string
	Method string

	QueryParams map[string]string
	Headers     map[string]string
	Body        string

	Response     string
	ResponseCode int

	// Handler allows the caller a hook to test their own custom expectations
	Handler http.HandlerFunc
}

// ExpectRPCJsonSuccess provides a helper function to return a ExpectedCall
// with generic RPCJSON defaults.
func ExpectRPCJsonSuccess(body, res string) ExpectedCall {
	return ExpectedCall{
		Path:   "/",
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "Application/Json",
		},
		Body:         body,
		Response:     res,
		ResponseCode: http.StatusOK,
	}
}

type server struct {
	call int
	mu   *sync.Mutex

	t           TInterface
	e           *WithT
	recoverFunc func()

	handlers map[int]ExpectedCall
}

// ServeHttp implements the http.Handler interface.
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.recoverFunc != nil {
		defer s.recoverFunc()
	}

	s.mu.Lock()
	c := s.call
	s.call++
	s.mu.Unlock()

	s.e.Expect(s.handlers).To(HaveKey(c))

	expect := s.handlers[c]
	s.e.Expect(r.Method).To(Equal(expect.Method))
	s.e.Expect(r.URL.Path).To(Equal(expect.Path))

	for key, value := range expect.Headers {
		s.e.Expect(strings.ToLower(r.Header.Get(key))).To(Equal(strings.ToLower(value)), fmt.Sprintf("Body param: %s was incorrect", key))
	}

	for key, value := range expect.QueryParams {
		s.e.Expect(strings.ToLower(r.URL.Query().Get(key))).To(Equal(strings.ToLower(value)), fmt.Sprintf("Query param: %s was incorrect", key))
	}

	b, _ := ioutil.ReadAll(r.Body)
	if expect.Body == "" {
		s.e.Expect(b).To(BeEmpty())
	} else {
		s.e.Expect(b).To(MatchJSON(expect.Body))
	}

	if expect.Handler != nil {
		expect.Handler(w, r)
		return
	}

	code := expect.ResponseCode
	if code == 0 {
		code = http.StatusOK
	}

	w.WriteHeader(code)
	_, err := w.Write([]byte(expect.Response))
	s.e.Expect(err).ToNot(HaveOccurred())
}

// Server the main struct which holds the core server implementation and
// reference to the httptest.Server.
type Server struct {
	HttpTest *httptest.Server

	s *server
}

// NewTestServer returns a pointer to a Server with calls and mutex applied.
func NewTestServer(t TInterface, calls ...ExpectedCall) *Server {
	s := &server{
		mu:          &sync.Mutex{},
		recoverFunc: ginkgo.GinkgoRecover,
		t:           t,
		e:           NewWithT(t),
	}

	handlers := make(map[int]ExpectedCall)

	for key, value := range calls {
		handlers[key] = value
	}

	s.handlers = handlers

	httpServer := httptest.NewServer(s)

	return &Server{
		HttpTest: httpServer,
		s:        s,
	}
}

// Close closes the underlying server and fails if the server has not being called the correct amount of times.
func (s *Server) Close() {
	defer s.HttpTest.Close()

	if len(s.s.handlers) != s.s.call {
		s.s.t.Errorf("mismatch in server calls, want: %v got: %v", len(s.s.handlers), s.s.call)
	}
}

// Expect adds an ExpectedCall to the server handlers
func (s *Server) Expect(call ExpectedCall) *Server {
	s.s.handlers[len(s.s.handlers)] = call

	return s
}

// Then is syntactic sugar for Expect.
func (s *Server) Then(call ExpectedCall) *Server {
	return s.Expect(call)
}

// BasicAuth returns a basic auth string
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// FixtureBox is a struct that holds information about the given fixtures.
type FixtureBox struct {
	Base string
}

// LoadFixture loads a json file with the given name.
// All files are relative to the test package, e.g.
// ethereum/eth.json becomes /test/ethereum/eth.json
func (f FixtureBox) LoadFixture(name string, args ...interface{}) (string, error) {
	jsonFile, err := os.Open(path.Join(f.Base, name))
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return "", err
	}

	if args != nil {
		return fmt.Sprintf(string(b), args...), nil
	}

	return string(b), nil
}
