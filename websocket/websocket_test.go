package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gorillaws "github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// Default CheckOrigin: Gorilla's safe same-host check (Fix 3)
// ---------------------------------------------------------------------------

func TestWebSocket_DefaultOrigin_SameHost_Accepted(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := gorillaws.Upgrader{} // nil CheckOrigin → gorilla same-host default
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, resp, err := gorillaws.DefaultDialer.Dial(wsURL, http.Header{
		"Origin": []string{ts.URL}, // same host
	})
	if err != nil {
		t.Fatalf("expected same-origin connection to succeed, got: %v (status %v)", err, resp)
	}
	conn.Close()
}

func TestWebSocket_DefaultOrigin_CrossSite_Rejected(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := gorillaws.Upgrader{} // nil CheckOrigin → gorilla same-host default
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	_, resp, err := gorillaws.DefaultDialer.Dial(wsURL, http.Header{
		"Origin": []string{"http://evil.example.com"},
	})
	if err == nil {
		t.Error("expected cross-origin connection to be rejected, but it succeeded")
	}
	if resp != nil && resp.StatusCode == http.StatusSwitchingProtocols {
		t.Error("cross-origin connection should not return 101 Switching Protocols")
	}
}

func TestWebSocket_CustomCheckOrigin_Override(t *testing.T) {
	// When the user provides CheckOrigin, it should be used instead of the default.
	allowAll := func(r *http.Request) bool { return true }

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := gorillaws.Upgrader{CheckOrigin: allowAll}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")
	conn, _, err := gorillaws.DefaultDialer.Dial(wsURL, http.Header{
		"Origin": []string{"http://evil.example.com"},
	})
	if err != nil {
		t.Fatalf("custom CheckOrigin(allowAll) should permit any origin, got: %v", err)
	}
	conn.Close()
}
