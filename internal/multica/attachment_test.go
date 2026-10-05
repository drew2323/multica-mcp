package multica

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadAttachmentMultipartContract(t *testing.T) {
	const markdown = "# Ahoj\n\nPříloha – přesně.\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/upload-file" { t.Errorf("request %s %s", r.Method, r.URL.Path) }
		if r.Header.Get("Authorization") != "Bearer secret" || r.Header.Get("X-Workspace-ID") != "workspace" { t.Errorf("headers: %v", r.Header) }
		_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type")); if err != nil { t.Fatal(err) }
		reader := multipart.NewReader(r.Body, params["boundary"])
		form, err := reader.ReadForm(1<<20); if err != nil { t.Fatal(err) }
		defer form.RemoveAll()
		f, err := form.File["file"][0].Open(); if err != nil { t.Fatal(err) }; defer f.Close()
		data, err := io.ReadAll(f); if err != nil { t.Fatal(err) }
		if string(data) != markdown || form.File["file"][0].Filename != "requirements.md" { t.Errorf("file: %q %q", form.File["file"][0].Filename, data) }
		if form.Value["issue_id"][0] != "issue-1" || form.Value["comment_id"][0] != "comment-1" { t.Errorf("fields: %#v", form.Value) }
		w.Header().Set("Content-Type", "application/json"); _, _ = io.WriteString(w, `{"id":"attachment-1"}`)
	}))
	defer server.Close()
	c := NewClient(server.URL, "secret", "test-version"); got, err := c.UploadAttachment(context.Background(), "requirements.md", "text/markdown", []byte(markdown), "issue-1", "comment-1", "workspace")
	if err != nil || string(got) != `{"id":"attachment-1"}` { t.Fatalf("upload = %s, %v", got, err) }
}

func TestAttachmentContentReadsExactBytesAndErrors(t *testing.T) {
	for _, tc := range []struct{name string; body string; status int; wantErr bool}{
		{"text", "# Requirements\nPříloha", 200, false}, {"binary", "\x00\xffdata", 200, false}, {"limit", strings.Repeat("x", (2<<20)+1), 200, true}, {"api-error", `{"error":"unsupported"}`, 415, true},
	} { t.Run(tc.name, func(t *testing.T) {
		server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/api/attachments/a%2Fb/content" && r.URL.EscapedPath()!="/api/attachments/a%2Fb/content" { t.Errorf("path %q",r.URL.EscapedPath()) }; w.Header().Set("X-Original-Content-Type","text/markdown");w.WriteHeader(tc.status);_,_=io.WriteString(w,tc.body)}));defer server.Close()
		c:=NewClient(server.URL,"secret","test");got,ct,err:=c.AttachmentContent(context.Background(),"a/b","w");if (err!=nil)!=tc.wantErr{t.Fatalf("error=%v",err)};if err==nil&&(string(got)!=tc.body||ct!="text/markdown"){t.Fatalf("got %q type %q",got,ct)}
	}) }
}
