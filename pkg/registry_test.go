package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type mockRegistry struct {
	t      *testing.T
	tags   []string
	scope  string
	basic  string
	bearer string
	server *httptest.Server
}

type mockTransport struct {
	Transport http.RoundTripper
}

func (m *mockRegistry) init() {
	r := mux.NewRouter()
	handleToken := func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		scope := params.Get("scope")
		auth := r.Header.Get("Authorization")
		if scope != m.scope || auth != fmt.Sprintf("Basic %s", m.basic) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		tokenResp := tokenResponse{Token: m.bearer}
		resp, err := json.Marshal(tokenResp)
		if err != nil {
			m.t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Add("content-type", "application/json")
		if _, err := w.Write(resp); err != nil {
			m.t.Fatal(err)
		}
	}
	r.HandleFunc("/token", handleToken)
	r.HandleFunc("/v2/hsn723/hoge/manifests/{tag}", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != fmt.Sprintf("Bearer %s", m.bearer) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		vars := mux.Vars(r)
		rt := vars["tag"]
		for _, tag := range m.tags {
			if tag == rt {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})
	r.HandleFunc("/v2/hsn723/public-hoge/manifests/{tag}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rt := vars["tag"]
		for _, tag := range m.tags {
			if tag == rt {
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(r)
	m.server = server
}

func (t mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}

func TestMain(m *testing.M) {
	http.DefaultClient.Transport = mockTransport{}
	os.Exit(m.Run())
}

func TestRetrieveBearerToken(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title    string
		registry mockRegistry
		auth     string
		expect   string
		isErr    bool
	}{
		{
			title: "AuthSuccess",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			auth:   "aG9nZTpoaWdl",
			expect: "aG9nZWJlYXJlcg==",
		},
		{
			title: "AuthFailure",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			auth:  "hoge",
			isErr: true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.title, func(t *testing.T) {
			t.Parallel()
			c.registry.init()
			url := c.registry.server.Listener.Addr().String()
			client := RegistryClient{
				RegistryName: NormalizeRegistryName(url),
				RegistryURL:  url,
				ImagePath:    "hsn723/hoge",
				HttpClient:   http.DefaultClient,
			}
			actual, err := client.retrieveBearerToken(c.auth)
			assertExpectedErr(t, err, c.isErr)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestCheckManifestForTag(t *testing.T) {
	t.Parallel()
	cases := []struct {
		title    string
		registry mockRegistry
		bearer   string
		tag      string
		expect   bool
		isErr    bool
	}{
		{
			title: "Exists",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
				tags:   []string{"1.0.0", "1.0.1", "0.1.0"},
			},
			bearer: "aG9nZWJlYXJlcg==",
			tag:    "1.0.1",
			expect: true,
		},
		{
			title: "NotExists",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
				tags:   []string{"1.0.0", "1.0.1", "0.1.0"},
			},
			bearer: "aG9nZWJlYXJlcg==",
			tag:    "1.0.2",
		},
		{
			title: "Unauthorized",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
				tags:   []string{"1.0.0", "1.0.1", "0.1.0"},
			},
			bearer: "hoge==",
			tag:    "1.0.2",
			isErr:  true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.title, func(t *testing.T) {
			t.Parallel()
			c.registry.init()
			url := c.registry.server.Listener.Addr().String()
			client := RegistryClient{
				RegistryName: NormalizeRegistryName(url),
				RegistryURL:  url,
				ImagePath:    "hsn723/hoge",
				HttpClient:   http.DefaultClient,
			}
			actual, err := client.checkManifestForTag(c.bearer, c.tag)
			assertExpectedErr(t, err, c.isErr)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestGetAuthTokenFromCredentials(t *testing.T) {
	cases := []struct {
		title        string
		env          map[string]string
		registryName string
		expect       string
		isErr        bool
	}{
		{
			title: "CredentialsExist",
			env: map[string]string{
				"HOGE_DEV_USER":     "hoge",
				"HOGE_DEV_PASSWORD": "hige",
			},
			registryName: "HOGE_DEV",
			expect:       "aG9nZTpoaWdl",
		},
		{
			title: "MissingUsername",
			env: map[string]string{
				"HOGE_DEV_PASSWORD": "hige",
			},
			registryName: "HOGE_DEV",
			expect:       "",
			isErr:        true,
		},
		{
			title: "MissingPassword",
			env: map[string]string{
				"HOGE_DEV_USER": "hoge",
			},
			registryName: "HOGE_DEV",
			expect:       "",
			isErr:        true,
		},
		{
			title: "DifferentRegistry",
			env: map[string]string{
				"HOGE_DEV_USER":     "hoge",
				"HOGE_DEV_PASSWORD": "hige",
			},
			registryName: "HOGE_IO",
			expect:       "",
			isErr:        true,
		},
	}
	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			t.Helper()
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			client := RegistryClient{
				RegistryName: c.registryName,
			}
			actual, err := client.getAuthTokenFromCredentials()
			assertExpectedErr(t, err, c.isErr)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestGetBearerTokenFromAuthToken(t *testing.T) {
	cases := []struct {
		title    string
		authEnv  string
		userEnv  string
		passEnv  string
		registry mockRegistry
		expect   string
		isErr    bool
	}{
		{
			title: "TokenInEnv",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			authEnv: "aG9nZTpoaWdl",
			expect:  "aG9nZWJlYXJlcg==",
		},
		{
			title: "TokenFromCreds",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			userEnv: "hoge",
			passEnv: "hige",
			expect:  "aG9nZWJlYXJlcg==",
		},
		{
			title: "WrongToken",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			authEnv: "hoge",
			isErr:   true,
		},
		{
			title: "WrongCredentials",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			isErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			t.Helper()
			c.registry.init()
			url := c.registry.server.Listener.Addr().String()
			client := RegistryClient{
				RegistryName: NormalizeRegistryName(url),
				RegistryURL:  url,
				ImagePath:    "hsn723/hoge",
				HttpClient:   http.DefaultClient,
			}
			t.Setenv(fmt.Sprintf("%s_AUTH", client.RegistryName), c.authEnv)
			t.Setenv(fmt.Sprintf("%s_USER", client.RegistryName), c.userEnv)
			t.Setenv(fmt.Sprintf("%s_PASSWORD", client.RegistryName), c.passEnv)
			actual, err := client.getBearerTokenFromAuthToken()
			assertExpectedErr(t, err, c.isErr)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		title        string
		bearerEnv    string
		githubEnv    string
		registryName string
		registry     mockRegistry
		expect       string
		isErr        bool
	}{
		{
			title: "TokenInEnv",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			bearerEnv: "aG9nZWJlYXJlcg==",
			expect:    "aG9nZWJlYXJlcg==",
		},
		{
			title: "GithubToken",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			registryName: "GHCR_IO",
			githubEnv:    "ghp_hogebearer",
			expect:       "Z2hwX2hvZ2ViZWFyZXI=",
		},
		{
			title: "WrongCredentials",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				basic:  "aG9nZTpoaWdl",
				bearer: "aG9nZWJlYXJlcg==",
			},
			isErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			t.Helper()
			c.registry.init()
			url := c.registry.server.Listener.Addr().String()
			registryName := c.registryName
			if registryName == "" {
				registryName = NormalizeRegistryName(url)
			}
			client := RegistryClient{
				RegistryName: registryName,
				RegistryURL:  url,
				ImagePath:    "hsn723/hoge",
				HttpClient:   http.DefaultClient,
			}
			t.Setenv(fmt.Sprintf("%s_TOKEN", client.RegistryName), c.bearerEnv)
			t.Setenv("GITHUB_TOKEN", c.githubEnv)
			actual, err := client.getBearerToken()
			assertExpectedErr(t, err, c.isErr)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestIsTagExist(t *testing.T) {
	cases := []struct {
		title     string
		bearerEnv string
		path      string
		registry  mockRegistry
		tag       string
		expect    bool
		isErr     bool
	}{
		{
			title:     "Found",
			bearerEnv: "aG9nZWJlYXJlcg==",
			path:      "hsn723/hoge",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				bearer: "aG9nZWJlYXJlcg==",
				tags:   []string{"1.0.0", "1.0.1", "0.1.0"},
			},
			tag:    "0.1.0",
			expect: true,
		},
		{
			title: "PublicImage",
			path:  "hsn723/public-hoge",
			registry: mockRegistry{
				t:      t,
				bearer: "aG9nZWJlYXJlcg==",
				tags:   []string{"1.0.0", "1.0.1", "0.1.0"},
			},
			tag:    "0.1.0",
			expect: true,
		},
		{
			title:     "NotFound",
			bearerEnv: "aG9nZWJlYXJlcg==",
			path:      "hsn723/hoge",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				bearer: "aG9nZWJlYXJlcg==",
				tags:   []string{"1.0.0", "1.0.1", "0.1.0"},
			},
			tag: "0.2.0",
		},
		{
			title:     "Unauthorized",
			bearerEnv: "",
			path:      "hsn723/hoge",
			registry: mockRegistry{
				t:      t,
				scope:  "repository:hsn723/hoge:pull",
				bearer: "aG9nZWJlYXJlcg==",
				tags:   []string{"1.0.0", "1.0.1", "0.1.0"},
			},
			tag:   "0.1.0",
			isErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			t.Helper()
			c.registry.init()
			url := c.registry.server.Listener.Addr().String()
			client := RegistryClient{
				RegistryName: NormalizeRegistryName(url),
				RegistryURL:  url,
				ImagePath:    c.path,
				HttpClient:   http.DefaultClient,
			}
			t.Setenv(fmt.Sprintf("%s_TOKEN", client.RegistryName), c.bearerEnv)
			actual, err := client.IsTagExist(c.tag)
			assertExpectedErr(t, err, c.isErr)
			assert.Equal(t, c.expect, actual)
		})
	}
}
