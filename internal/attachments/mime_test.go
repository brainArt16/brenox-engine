package attachments

import "testing"

func TestValidateUpload(t *testing.T) {
	if err := validateUpload("doc.pdf", "application/pdf", 1024, 10*1024*1024); err != nil {
		t.Fatalf("expected valid upload, got %v", err)
	}
}

func TestValidateUploadRejectsLargeFile(t *testing.T) {
	if err := validateUpload("big.bin", "application/pdf", 20*1024*1024, 10*1024*1024); err != ErrFileTooLarge {
		t.Fatalf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestValidateUploadRejectsMime(t *testing.T) {
	if err := validateUpload("run.exe", "application/x-msdownload", 100, 1024); err != ErrMimeNotAllowed {
		t.Fatalf("expected ErrMimeNotAllowed, got %v", err)
	}
}
