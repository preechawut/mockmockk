package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	httpClient *http.Client
}

type Project struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Endpoint struct {
	ID        string          `json:"id"`
	ProjectID *string         `json:"projectId,omitempty"`
	Method    string          `json:"method"`
	Path      string          `json:"path"`
	Status    int             `json:"status"`
	Response  json.RawMessage `json:"response"`
	CreatedAt time.Time       `json:"createdAt"`
}

type APIResponse[T any] struct {
	OK    bool   `json:"ok"`
	Data  T      `json:"data"`
	Error *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewClient(baseURL string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &Client{
		BaseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) CheckHealth() (bool, error) {
	resp, err := c.httpClient.Get(c.BaseURL + "/healthz")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func (c *Client) GetProjects() ([]Project, error) {
	var res APIResponse[[]Project]
	err := c.get("/api/v1/projects", &res)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) CreateProject(name, slug string) (Project, error) {
	payload := map[string]string{
		"name": name,
		"slug": slug,
	}
	var res APIResponse[Project]
	err := c.post("/api/v1/projects", payload, &res)
	if err != nil {
		return Project{}, err
	}
	return res.Data, nil
}

func (c *Client) DeleteProject(id string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/v1/projects/"+id, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}
	return nil
}

func (c *Client) GetEndpoints(projectID string) ([]Endpoint, error) {
	var res APIResponse[[]Endpoint]
	path := fmt.Sprintf("/api/v1/projects/%s/endpoints", projectID)
	err := c.get(path, &res)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) GetGlobalEndpoints() ([]Endpoint, error) {
	var res APIResponse[[]Endpoint]
	err := c.get("/api/v1/mockmocks", &res)
	if err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) CreateProjectEndpoint(projectID, method, path string, status int, response string) (Endpoint, error) {
	if response == "" {
		response = "{}"
	}
	payload := map[string]any{
		"method":   method,
		"path":     path,
		"status":   status,
		"response": json.RawMessage(response),
	}
	var res APIResponse[Endpoint]
	err := c.post(fmt.Sprintf("/api/v1/projects/%s/endpoints", projectID), payload, &res)
	if err != nil {
		return Endpoint{}, err
	}
	return res.Data, nil
}

func (c *Client) CreateGlobalEndpoint(method, path string, status int, response string) (Endpoint, error) {
	if response == "" {
		response = "{}"
	}
	payload := map[string]any{
		"method":   method,
		"path":     path,
		"status":   status,
		"response": json.RawMessage(response),
	}
	var res APIResponse[Endpoint]
	err := c.post("/api/v1/mockmocks", payload, &res)
	if err != nil {
		return Endpoint{}, err
	}
	return res.Data, nil
}

func (c *Client) DuplicateEndpoint(projectID *string, id string) (Endpoint, error) {
	var url string
	if projectID != nil && *projectID != "" {
		url = fmt.Sprintf("/api/v1/projects/%s/endpoints/%s/duplicate", *projectID, id)
	} else {
		url = fmt.Sprintf("/api/v1/mockmocks/%s/duplicate", id)
	}
	var res APIResponse[Endpoint]
	err := c.post(url, nil, &res)
	if err != nil {
		return Endpoint{}, err
	}
	return res.Data, nil
}

func (c *Client) DeleteEndpoint(projectID *string, id string) error {
	var url string
	if projectID != nil && *projectID != "" {
		url = fmt.Sprintf("/api/v1/projects/%s/endpoints/%s", *projectID, id)
	} else {
		url = fmt.Sprintf("/api/v1/mockmocks/%s", id)
	}
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}
	return nil
}

type TestResult struct {
	StatusCode int
	Body       string
	Duration   time.Duration
}

func (c *Client) TestMock(method, mockPath string) (TestResult, error) {
	start := time.Now()
	req, err := http.NewRequest(method, c.BaseURL+mockPath, nil)
	if err != nil {
		return TestResult{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TestResult{}, err
	}
	defer resp.Body.Close()
	duration := time.Since(start)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TestResult{}, err
	}

	var formatted bytes.Buffer
	if json.Indent(&formatted, body, "", "  ") == nil {
		return TestResult{
			StatusCode: resp.StatusCode,
			Body:       formatted.String(),
			Duration:   duration,
		}, nil
	}

	return TestResult{
		StatusCode: resp.StatusCode,
		Body:       string(body),
		Duration:   duration,
	}, nil
}

func (c *Client) get(path string, target any) error {
	resp, err := c.httpClient.Get(c.BaseURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) post(path string, payload any, target any) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.parseError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
		return fmt.Errorf("[%s] %s", errResp.Error.Code, errResp.Error.Message)
	}
	return fmt.Errorf("server responded with status %d", resp.StatusCode)
}
