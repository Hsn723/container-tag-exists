package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	quayAuthAPI = "https://%s/v2/auth?service=%s&scope=repository:%s:pull"
	authAPI     = "https://%s/token?scope=repository:%s:pull"
	manifestAPI = "https://%s/v2/%s/manifests/%s"
)

type IRegistryClient interface {
	IsTagExist(tag string) (bool, error)
}

type RegistryClient struct {
	RegistryName string
	RegistryURL  string
	ImagePath    string
	HttpClient   *http.Client
	Platforms    []string
}

type tokenResponse struct {
	Token string `json:"token"`
}

type manifestResponse struct {
	Manifests []manifest `json:"manifests"`
}

type manifest struct {
	Platform platform `json:"platform"`
}

type platform struct {
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
}

func (r RegistryClient) retrieve(method, endpoint string, headers map[string]string) (int, []byte, error) {
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return -1, nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	res, err := r.HttpClient.Do(req)
	if err != nil {
		return -1, nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, nil, err
	}
	return res.StatusCode, b, nil
}

func (r RegistryClient) retrieveBearerToken(auth string) (string, error) {
	var endpoint string
	if r.RegistryName == "QUAY_IO" {
		endpoint = fmt.Sprintf(quayAuthAPI, r.RegistryURL, r.RegistryURL, r.ImagePath)
	} else {
		endpoint = fmt.Sprintf(authAPI, r.RegistryURL, r.ImagePath)
	}
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Basic %s", auth),
	}
	status, res, err := r.retrieve(http.MethodGet, endpoint, headers)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("unexpected response code %d", status)
	}
	var token tokenResponse
	if err := json.Unmarshal(res, &token); err != nil {
		return "", err
	}
	return token.Token, nil
}

func (r RegistryClient) hasPlatform(platform string, manifests []manifest) bool {
	for _, m := range manifests {
		p := fmt.Sprintf("%s/%s", m.Platform.Os, m.Platform.Architecture)
		if strings.EqualFold(p, platform) {
			return true
		}
	}
	return false
}

func (r RegistryClient) hasPlatforms(res []byte) (bool, error) {
	var manifests manifestResponse
	if err := json.Unmarshal(res, &manifests); err != nil {
		return false, err
	}
	for _, p := range r.Platforms {
		if !r.hasPlatform(p, manifests.Manifests) {
			return false, nil
		}
	}
	return true, nil
}

func (r RegistryClient) checkManifestForTag(bearer, tag string) (bool, error) {
	endpoint := fmt.Sprintf(manifestAPI, r.RegistryURL, r.ImagePath, tag)
	headers := map[string]string{
		"Accept": "application/vnd.oci.image.index.v1+json",
	}
	if bearer != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", bearer)
	}
	method := http.MethodHead
	if r.Platforms != nil {
		method = http.MethodGet
	}
	status, res, err := r.retrieve(method, endpoint, headers)
	if err != nil {
		return false, err
	}
	if status == http.StatusNotFound {
		return false, nil
	}
	if status == http.StatusOK {
		if r.Platforms == nil {
			return true, nil
		}
		return r.hasPlatforms(res)
	}
	return false, fmt.Errorf("unexpected response registry API: %d", status)
}

func (r RegistryClient) getAuthTokenFromCredentials() (string, error) {
	userEnvName := fmt.Sprintf("%s_USER", r.RegistryName)
	passEnvName := fmt.Sprintf("%s_PASSWORD", r.RegistryName)
	user := os.Getenv(userEnvName)
	pass := os.Getenv(passEnvName)
	if user == "" || pass == "" {
		return "", fmt.Errorf("could not get credentials for %s", r.RegistryName)
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pass)))
	return b64, nil
}

func (r RegistryClient) getBearerTokenFromAuthToken() (string, error) {
	authTokenEnvName := fmt.Sprintf("%s_AUTH", r.RegistryName)
	authToken := os.Getenv(authTokenEnvName)
	if authToken == "" {
		t, err := r.getAuthTokenFromCredentials()
		if err != nil {
			return "", err
		}
		if t == "" {
			return "", fmt.Errorf("could not get auth token for %s", r.RegistryName)
		}
		authToken = t
	}
	return r.retrieveBearerToken(authToken)
}

func (r RegistryClient) getBearerToken() (string, error) {
	bearerTokenEnvName := fmt.Sprintf("%s_TOKEN", r.RegistryName)
	bearerToken := os.Getenv(bearerTokenEnvName)
	if bearerToken != "" {
		return bearerToken, nil
	}
	bearerToken, err := r.getBearerTokenFromAuthToken()
	if err != nil {
		// ghcr.io is a special case where we can use GITHUB_TOKEN as the bearer token.
		if r.RegistryName != "GHCR_IO" {
			return "", err
		}
		github_token := os.Getenv("GITHUB_TOKEN")
		if github_token == "" {
			return "", err
		}
		bearerToken = base64.StdEncoding.EncodeToString([]byte(github_token))
	}
	if bearerToken != "" {
		return bearerToken, nil
	}
	return "", fmt.Errorf("could not get a bearer token for %s", r.RegistryName)
}

func (r RegistryClient) IsTagExist(tag string) (bool, error) {
	// First attempt to retrieve tag anonymously, for public images
	if found, err := r.checkManifestForTag("", tag); err == nil {
		return found, nil
	}
	bearerToken, err := r.getBearerToken()
	if err != nil {
		return false, err
	}
	return r.checkManifestForTag(bearerToken, tag)
}
