package gh

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

const (
	defaultBaseURL = "https://api.github.com/"
	uploadBaseURL  = "https://uploads.github.com/"
)

type Gh struct {
	client *github.Client
	owner  string
	repo   string
}

// New return Gh
func New(owner, repo string) (*Gh, error) {
	c := github.NewClient(httpClient())
	if baseURL := os.Getenv("GITHUB_BASE_URL"); baseURL != "" {
		baseEndpoint, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(baseEndpoint.Path, "/") {
			baseEndpoint.Path += "/"
		}
		c.BaseURL = baseEndpoint
	}
	if uploadURL := os.Getenv("GITHUB_UPLOAD_URL"); uploadURL != "" {
		uploadEndpoint, err := url.Parse(uploadURL)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(uploadEndpoint.Path, "/") {
			uploadEndpoint.Path += "/"
		}
		c.UploadURL = uploadEndpoint
	}
	return &Gh{
		client: c,
		owner:  owner,
		repo:   repo,
	}, nil
}

func (g *Gh) PutPrComment(ctx context.Context, n int, comment string) error {
	c := &github.IssueComment{Body: &comment}
	if _, _, err := g.client.Issues.CreateComment(ctx, g.owner, g.repo, n, c); err != nil {
		return err
	}
	return nil
}

type roundTripper struct {
	transport   *http.Transport
	accessToken string
}

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", fmt.Sprintf("token %s", rt.accessToken))
	return rt.transport.RoundTrip(r)
}

func httpClient() *http.Client {
	t := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	rt := roundTripper{
		transport:   t,
		accessToken: os.Getenv("GITHUB_TOKEN"),
	}
	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: rt,
	}
}
