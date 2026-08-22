package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"depresolve/internal/httpapi"
	"depresolve/internal/model"
	"depresolve/internal/service"
	"depresolve/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "depresolve.db", "SQLite path")
	smokeFlag := flag.Bool("smoke-test", false, "run self check")
	rebuild := flag.Bool("rebuild-index", false, "rebuild catalog index and exit")
	flag.Parse()
	st, err := store.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()
	svc, err := service.New(st)
	if err != nil {
		log.Fatal(err)
	}
	if *rebuild {
		if _, err := svc.Recover(context.Background()); err != nil {
			log.Fatal(err)
		}
		fmt.Println("index rebuilt")
		return
	}
	if *smokeFlag {
		if err := smoke(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("smoke-test passed")
		return
	}
	srv := &http.Server{Addr: *addr, Handler: httpapi.New(svc).Handler(), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("depresolve listening on %s", *addr)
	log.Fatal(srv.ListenAndServe())
}

func smoke() error {
	dir, err := os.MkdirTemp("", "task151-smoke-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	db := filepath.Join(dir, "smoke.db")
	st, err := store.Open(db)
	if err != nil {
		return err
	}
	svc, err := service.New(st)
	if err != nil {
		return err
	}
	ctx := context.Background()
	for _, c := range []model.CreateComponentRequest{{ID: "app", Name: "Application", Ecosystem: "go"}, {ID: "lib", Name: "Library", Ecosystem: "go"}, {ID: "codec", Name: "Codec", Ecosystem: "go"}} {
		if _, err = svc.CreateComponent(ctx, c); err != nil {
			return err
		}
	}
	if _, err = svc.CreateRelease(ctx, "app", model.CreateReleaseRequest{ID: "app-1", Version: "1.0.0", Platforms: []string{"linux/amd64"}, Dependencies: []model.DependencyInput{{ID: "app-lib", TargetComponent: "lib", Constraint: "^1.0.0"}}}); err != nil {
		return err
	}
	for _, r := range []struct{ c, id, v string }{{"lib", "lib-1", "1.0.0"}, {"lib", "lib-2", "1.2.0"}, {"codec", "codec-1", "2.0.0"}} {
		if _, err = svc.CreateRelease(ctx, r.c, model.CreateReleaseRequest{ID: r.id, Version: r.v, Platforms: []string{"linux/amd64"}, Capabilities: []string{"codec"}}); err != nil {
			return err
		}
	}
	ts := httptest.NewServer(httpapi.New(svc).Handler())
	defer ts.Close()
	if err = expect(ts.URL+"/static/index.html", http.StatusOK); err != nil {
		return err
	}
	if err = expect(ts.URL+"/api/stats", http.StatusOK); err != nil {
		return err
	}
	body, _ := json.Marshal(model.ResolveRequest{ID: "smoke-run", RootComponent: "app", RootConstraint: "*", Platform: "linux/amd64"})
	resp, err := http.Post(ts.URL+"/api/resolve", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	raw, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("resolve status=%d body=%s", resp.StatusCode, raw)
	}
	var out model.ResolveResponse
	if err = json.Unmarshal(raw, &out); err != nil {
		return err
	}
	if len(out.Selections) != 2 || out.Run.Status != model.RunResolved {
		return fmt.Errorf("unexpected resolve %+v", out)
	}
	snapBody, _ := json.Marshal(model.CreateSnapshotRequest{Name: "smoke"})
	resp, err = http.Post(ts.URL+"/api/resolve/"+out.Run.ID+"/snapshot", "application/json", bytes.NewReader(snapBody))
	if err != nil {
		return err
	}
	raw, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("snapshot status=%d body=%s", resp.StatusCode, raw)
	}
	var snap model.Snapshot
	if err = json.Unmarshal(raw, &snap); err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/snapshots/"+snap.ID+"/activate", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("activate status=%d", resp.StatusCode)
	}
	if err = st.Close(); err != nil {
		return err
	}
	st, err = store.Open(db)
	if err != nil {
		return err
	}
	defer st.Close()
	recovered, err := service.New(st)
	if err != nil {
		return err
	}
	active, err := st.ActiveSnapshot(ctx)
	if err != nil {
		return err
	}
	if active.ID != snap.ID {
		return fmt.Errorf("active snapshot lost: got %s want %s", active.ID, snap.ID)
	}
	if _, err = recovered.GetResolve(ctx, out.Run.ID); err != nil {
		return err
	}
	return nil
}
func expect(url string, status int) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != status {
		return fmt.Errorf("GET %s: status=%d want=%d", url, resp.StatusCode, status)
	}
	return nil
}
