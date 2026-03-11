package udpscan

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScanSign(t *testing.T) {
	require.Equal(t, "UDPSCAN", ScanSign)
	require.NotEmpty(t, ScanSign)
}

func TestScanPort(t *testing.T) {
	require.Equal(t, 42388, ScanPort)
	require.Positive(t, ScanPort)
}

func TestResponse(t *testing.T) {
	now := time.Now()
	resp := Response{Name: "test-host", Time: now}
	require.Equal(t, "test-host", resp.Name)
	require.Equal(t, now, resp.Time)
}

func TestResponseJSON(t *testing.T) {
	resp := Response{Name: "xiaozhixiang", Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}

	data, err := json.Marshal(resp)
	require.NoError(t, err)
	require.Contains(t, string(data), `"name":"xiaozhixiang"`)

	var decoded Response
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, resp.Name, decoded.Name)
	require.Equal(t, resp.Time, decoded.Time)
}
