package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// VersionProvider implements an http provider to get worker version
type VersionProvider struct {
	remotePort string
}

// NewVersionProvider instantiates version http provider
func NewVersionProvider(port string) *VersionProvider {
	return &VersionProvider{remotePort: port}
}

// Assignation request worker assignation
func (v *VersionProvider) Assignation(ctx context.Context, IP net.IP) (*config.Workload, error) {
	u := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%s", IP.String(), v.remotePort),
		Path:   "/internal/version",
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected exchange response status code %d", res.StatusCode)
	}

	wk := &config.Workload{}
	err = json.NewDecoder(res.Body).Decode(wk)
	if err != nil {
		return nil, err
	}

	return wk, nil
}
