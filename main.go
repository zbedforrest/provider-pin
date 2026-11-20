// provider-pin forwards OpenRouter calls and pins requests for matching models
// to one provider with fallbacks off, for clients that cannot send the provider field.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8790", "listen address")
	provider := flag.String("provider", "", "OpenRouter provider slug to pin")
	match := flag.String("match", "deepseek/", "pin only models with this prefix")
	flag.Parse()
	target, _ := url.Parse("https://openrouter.ai")
	var pinned, passed atomic.Int64
	rp := httputil.NewSingleHostReverseProxy(target)
	rp.FlushInterval = -1
	base := rp.Director
	rp.Director = func(r *http.Request) {
		base(r)
		r.Host = target.Host
		if r.Body == nil || r.Method != http.MethodPost {
			return
		}
		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			r.Body = io.NopCloser(bytes.NewReader(nil))
			return
		}
		var m map[string]any
		if json.Unmarshal(body, &m) == nil {
			if model, _ := m["model"].(string); strings.HasPrefix(model, *match) {
				m["provider"] = map[string]any{"only": []string{*provider}, "allow_fallbacks": false}
				if nb, err := json.Marshal(m); err == nil {
					body = nb
					pinned.Add(1)
				}
			} else {
				passed.Add(1)
			}
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		r.Header.Set("Content-Length", itoa(len(body)))
	}
	rp.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode >= 400 {
			log.Printf("upstream %d for %s", resp.StatusCode, resp.Request.URL.Path)
		}
		return nil
	}
	http.HandleFunc("/stats", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]int64{"pinned": pinned.Load(), "passed": passed.Load()})
	})
	http.Handle("/", rp)
	log.Printf("pinning %s* to %s on %s", *match, *provider, *listen)
	log.Fatal(http.ListenAndServe(*listen, nil))
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }
