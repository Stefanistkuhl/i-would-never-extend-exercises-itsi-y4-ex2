package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/stefanistkuhl/i-would-never-extend-exercises-itsi-y4-ex2/pkg/config"
)

type Client struct {
	baseURL    string
	password   string
	httpClient *http.Client
	socketPath string
}

type ClientConfig struct {
	Server   string `json:"server"`
	Password string `json:"password"`
	Port     int    `json:"port,omitempty"`
	Socket   string `json:"socket,omitempty"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pcapstore"), nil
}

func LoadClientConfig() (*ClientConfig, error) {
	cfgPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := &ClientConfig{
		Server:   os.Getenv("PCAPSTORE_SERVER"),
		Password: os.Getenv("PCAPSTORE_PASSWORD"),
	}

	if cfg.Password == "" {
		cfg.Password = os.Getenv("SORTER_PASSWORD")
	}

	if portStr := os.Getenv("PCAPSTORE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
	}

	if socket := os.Getenv("PCAPSTORE_SOCKET"); socket != "" {
		cfg.Socket = socket
	}

	if _, err := os.Stat(cfgPath); err == nil {
		data, err := os.ReadFile(cfgPath)
		if err == nil {
			var fileCfg ClientConfig
			if err := json.Unmarshal(data, &fileCfg); err == nil {
				if cfg.Server == "" {
					cfg.Server = fileCfg.Server
				}
				if cfg.Password == "" {
					cfg.Password = fileCfg.Password
				}
				if cfg.Port == 0 && fileCfg.Port != 0 {
					cfg.Port = fileCfg.Port
				}
				if cfg.Socket == "" && fileCfg.Socket != "" {
					cfg.Socket = fileCfg.Socket
				}
			}
		}
	}

	return cfg, nil
}

func SaveClientConfig(cfg *ClientConfig) error {
	cfgPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cfgPath, data, 0600)
}

func NewClient(server, password string, port int, socket string) (*Client, error) {
	cfg, _ := LoadClientConfig()
	if cfg == nil {
		cfg = &ClientConfig{}
	}

	loadedServer := cfg.Server
	loadedPassword := cfg.Password
	loadedPort := cfg.Port
	loadedSocket := cfg.Socket

	if envServer := os.Getenv("PCAPSTORE_SERVER"); envServer != "" {
		loadedServer = envServer
	}
	if envPassword := os.Getenv("PCAPSTORE_PASSWORD"); envPassword != "" {
		loadedPassword = envPassword
	} else if envPassword := os.Getenv("SORTER_PASSWORD"); envPassword != "" {
		loadedPassword = envPassword
	}
	if portStr := os.Getenv("PCAPSTORE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			loadedPort = p
		}
	}
	if envSocket := os.Getenv("PCAPSTORE_SOCKET"); envSocket != "" {
		loadedSocket = envSocket
	}

	if server != "" {
		loadedServer = server
	}
	if password != "" {
		loadedPassword = password
	}
	if port != 0 {
		loadedPort = port
	}
	if socket != "" {
		loadedSocket = socket
	}

	if loadedSocket == "" && loadedServer == "" {
		loadedSocket = "/tmp/pcap-sorter.sock"
	}

	var baseURL string
	var socketPath string
	var httpClient *http.Client

	if loadedSocket != "" || strings.HasPrefix(loadedServer, "/") || strings.HasPrefix(loadedServer, "unix://") {
		if strings.HasPrefix(loadedServer, "unix://") {
			socketPath = strings.TrimPrefix(loadedServer, "unix://")
		} else if strings.HasPrefix(loadedServer, "/") {
			socketPath = loadedServer
		} else {
			socketPath = loadedSocket
		}
		if socketPath == "" {
			socketPath = "/tmp/pcap-sorter.sock"
		}

		transport := &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		}

		httpClient = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}

		baseURL = "http://unix"
	} else {
		if loadedServer == "" {
			if loadedPort == 0 {
				return nil, fmt.Errorf("server URL or port is required (use --server flag, --port flag, PCAPSTORE_SERVER env var, or set in ~/.pcapstore)")
			}
			loadedServer = fmt.Sprintf("http://localhost:%d", loadedPort)
		} else {
			if !strings.HasPrefix(loadedServer, "http://") && !strings.HasPrefix(loadedServer, "https://") {
				if loadedPort != 0 {
					loadedServer = fmt.Sprintf("http://%s:%d", loadedServer, loadedPort)
				} else {
					loadedServer = "http://" + loadedServer
				}
			} else {
				parsedURL, err := url.Parse(loadedServer)
				if err == nil {
					_, existingPort, err := net.SplitHostPort(parsedURL.Host)
					if err != nil || existingPort == "" {
						// No port in URL, add it if we have one
						if loadedPort != 0 {
							parsedURL.Host = net.JoinHostPort(parsedURL.Host, strconv.Itoa(loadedPort))
							loadedServer = parsedURL.String()
						}
					}
					// If URL already has a port, use it as-is (don't override)
				}
			}
		}

		httpClient = &http.Client{Timeout: 30 * time.Second}
		baseURL = loadedServer
	}

	// Save config: normalize server URL (remove port if we have separate port field)
	saveCfg := &ClientConfig{
		Password: loadedPassword,
		Socket:   socketPath,
	}

	if socketPath == "" && loadedServer != "" {
		parsedURL, err := url.Parse(loadedServer)
		if err == nil {
			host, portStr, err := net.SplitHostPort(parsedURL.Host)
			if err == nil && portStr != "" {
				// URL has port, extract it and save separately
				saveCfg.Server = fmt.Sprintf("%s://%s", parsedURL.Scheme, host)
				if port, err := strconv.Atoi(portStr); err == nil {
					saveCfg.Port = port
				}
			} else {
				// URL doesn't have port, save as-is
				saveCfg.Server = loadedServer
				if loadedPort != 0 {
					saveCfg.Port = loadedPort
				}
			}
		} else {
			saveCfg.Server = loadedServer
			if loadedPort != 0 {
				saveCfg.Port = loadedPort
			}
		}
	} else if loadedPort != 0 {
		saveCfg.Port = loadedPort
	}

	_ = SaveClientConfig(saveCfg)

	return &Client{
		baseURL:    baseURL,
		password:   loadedPassword,
		httpClient: httpClient,
		socketPath: socketPath,
	}, nil
}

func (c *Client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	var fullURL string

	if c.socketPath != "" {
		fullURL = c.baseURL + path
	} else {
		baseURL, err := url.Parse(c.baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		reqURL, err := url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}
		fullURL = baseURL.ResolveReference(reqURL).String()
	}

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		return nil, err
	}

	if c.socketPath != "" {
		req.Host = c.socketPath
	} else if c.password != "" {
		req.Header.Set("Authorization", c.password)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) doJSONRequest(method, path string, requestBody any, responseBody any) error {
	var body io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonData)
	}

	resp, err := c.doRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	if responseBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Health() (map[string]any, error) {
	var result map[string]any
	err := c.doJSONRequest("GET", "/api/health", nil, &result)
	return result, err
}

func (c *Client) Status() (map[string]any, error) {
	var result map[string]any
	err := c.doJSONRequest("GET", "/api/status", nil, &result)
	return result, err
}

func (c *Client) Version() (map[string]any, error) {
	var result map[string]any
	err := c.doJSONRequest("GET", "/api/version", nil, &result)
	return result, err
}

func (c *Client) ListFiles() ([]any, error) {
	var result []any
	err := c.doJSONRequest("GET", "/api/files", nil, &result)
	return result, err
}

func (c *Client) GetFile(id int64) (any, error) {
	var result any
	err := c.doJSONRequest("GET", fmt.Sprintf("/api/file/%d", id), nil, &result)
	return result, err
}

func (c *Client) DownloadFile(id int64, outputPath string) error {
	resp, err := c.doRequest("GET", fmt.Sprintf("/api/files/%d/download", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}

func (c *Client) GetFileStats(id int64) (any, error) {
	var result any
	err := c.doJSONRequest("GET", fmt.Sprintf("/api/files/%d/stats", id), nil, &result)
	return result, err
}

func (c *Client) DeleteFile(id int64) error {
	return c.doJSONRequest("DELETE", fmt.Sprintf("/api/file/%d", id), nil, nil)
}

func (c *Client) GetArchive() (any, error) {
	var result any
	err := c.doJSONRequest("GET", "/api/archive", nil, &result)
	return result, err
}

func (c *Client) ArchiveFile(id int64) error {
	return c.doJSONRequest("POST", fmt.Sprintf("/api/archive/%d", id), nil, nil)
}

func (c *Client) ArchiveStatus() (any, error) {
	var result any
	err := c.doJSONRequest("GET", "/api/archive/status", nil, &result)
	return result, err
}

func (c *Client) GetConfig() (*config.Config, error) {
	var result config.Config
	err := c.doJSONRequest("GET", "/api/config", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateConfig(cfg *config.Config) error {
	return c.doJSONRequest("PUT", "/api/config", cfg, nil)
}

func (c *Client) GetCleanupCandidates() (any, error) {
	var result any
	err := c.doJSONRequest("GET", "/api/cleanup/candidates", nil, &result)
	return result, err
}

func (c *Client) CleanupExecute() (any, error) {
	var result any
	err := c.doJSONRequest("POST", "/api/cleanup/execute", nil, &result)
	return result, err
}

func (c *Client) GetSummary() (any, error) {
	var result any
	err := c.doJSONRequest("GET", "/api/summary", nil, &result)
	return result, err
}

func (c *Client) GetStatsByHostname() (any, error) {
	var result any
	err := c.doJSONRequest("GET", "/api/stats/by-hostname", nil, &result)
	return result, err
}

func (c *Client) GetStatsByScenario() (any, error) {
	var result any
	err := c.doJSONRequest("GET", "/api/stats/by-scenario", nil, &result)
	return result, err
}

func (c *Client) CompressFile(id int64) error {
	return c.doJSONRequest("POST", fmt.Sprintf("/api/compression/%d", id), nil, nil)
}

func (c *Client) CompressTrigger() (any, error) {
	var result any
	err := c.doJSONRequest("POST", "/api/compression/trigger", nil, &result)
	return result, err
}

// Search & Query
func (c *Client) Search(query string) (any, error) {
	var result any
	path := "/api/search"
	if query != "" {
		path = fmt.Sprintf("/api/search?q=%s", url.QueryEscape(query))
	}
	err := c.doJSONRequest("GET", path, nil, &result)
	return result, err
}

func (c *Client) GetFilesByHostname(hostname string) (any, error) {
	var result any
	err := c.doJSONRequest("GET", fmt.Sprintf("/api/files/by-hostname/%s", url.PathEscape(hostname)), nil, &result)
	return result, err
}

func (c *Client) GetFilesByScenario(scenario string) (any, error) {
	var result any
	err := c.doJSONRequest("GET", fmt.Sprintf("/api/files/by-scenario/%s", url.PathEscape(scenario)), nil, &result)
	return result, err
}

func (c *Client) QuerySQL(query string) (any, error) {
	var result any
	reqBody := map[string]string{"query": query}
	err := c.doJSONRequest("POST", "/api/query", reqBody, &result)
	return result, err
}

func (c *Client) ExportStore(outputPath string) error {
	resp, err := c.doRequest("GET", "/api/export", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	return err
}
