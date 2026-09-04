package document

import (
	"context"
	"regexp"
	"testing"
)

func TestProcessSubmittedHTMLFileDoesNotReadFilesystem(t *testing.T) {
	d := &Document{
		URL:  "file:///path/that/does/not/exist/page.html",
		HTML: "<html><body>saved content</body></html>",
	}

	err := d.ProcessContext(context.Background(), nil, func(_ context.Context, got *Document) error {
		got.Text = "saved content"
		return nil
	})
	if err != nil {
		t.Fatalf("ProcessContext() unexpected error: %v", err)
	}
	if d.Text != "saved content" {
		t.Errorf("Text = %q, want extracted saved content", d.Text)
	}
	if d.Type != Web {
		t.Errorf("Type = %v, want web", d.Type)
	}
	if d.Domain != "local" {
		t.Errorf("Domain = %q, want local", d.Domain)
	}
	if !d.Processed {
		t.Error("document was not marked as processed")
	}
}

func TestProcessRemoteFilePreservesImportedSnapshot(t *testing.T) {
	d := &Document{
		URL:     "remote-file://workstation/home/alice/notes.md",
		Text:    "imported content",
		Updated: 1234,
		Type:    RemoteFile,
	}

	err := d.ProcessContext(context.Background(), nil, func(_ context.Context, _ *Document) error {
		t.Fatal("remote file processing called the web extractor")
		return nil
	})
	if err != nil {
		t.Fatalf("ProcessContext() unexpected error: %v", err)
	}
	if d.Type != RemoteFile {
		t.Errorf("Type = %v, want remote", d.Type)
	}
	if d.Domain != "workstation" {
		t.Errorf("Domain = %q, want workstation", d.Domain)
	}
	if d.Title != "alice/notes.md" {
		t.Errorf("Title = %q, want alice/notes.md", d.Title)
	}
	if d.Updated != 1234 {
		t.Errorf("Updated = %d, want 1234", d.Updated)
	}
	if !d.Processed {
		t.Error("document was not marked as processed")
	}
}

func TestProcessRemoteFileValidation(t *testing.T) {
	t.Run("invalid URL", func(t *testing.T) {
		d := &Document{URL: "file:///tmp/notes.txt", Text: "content", Type: RemoteFile}
		if err := d.Process(nil, nil); err == nil {
			t.Fatal("Process() accepted a local file URL for a remote file")
		}
	})

	t.Run("sensitive text", func(t *testing.T) {
		d := &Document{URL: "remote-file://client/tmp/notes.txt", Text: "secret token", Type: RemoteFile}
		err := d.ProcessWithSensitivePattern(nil, nil, regexp.MustCompile("secret"))
		if err != ErrSensitiveContent {
			t.Fatalf("ProcessWithSensitivePattern() error = %v, want %v", err, ErrSensitiveContent)
		}
	})
}
