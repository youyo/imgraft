package reference_test

import (
	"testing"

	"github.com/youyo/imgraft/internal/errs"
	"github.com/youyo/imgraft/internal/reference"
)

func TestValidateURL_AllowPublicHTTPS(t *testing.T) {
	urls := []string{
		"https://example.com/image.png",
		"https://cdn.example.com/img/test.jpg",
		"https://1.2.3.4/image.png", // パブリックIP
		"http://example.com/image.png",
	}
	for _, u := range urls {
		if err := reference.ValidateURL(u); err != nil {
			t.Errorf("ValidateURL(%q) = %v; want nil", u, err)
		}
	}
}

func TestValidateURL_RejectLocalhost(t *testing.T) {
	urls := []string{
		"http://localhost/image.png",
		"http://localhost:8080/image.png",
		"https://localhost/image.png",
	}
	for _, u := range urls {
		err := reference.ValidateURL(u)
		if err == nil {
			t.Errorf("ValidateURL(%q) = nil; want error", u)
			continue
		}
		code := errs.CodeOf(err)
		if code != errs.ErrReferenceURLForbidden {
			t.Errorf("ValidateURL(%q) code = %q; want %q", u, code, errs.ErrReferenceURLForbidden)
		}
	}
}

func TestValidateURL_RejectLoopback(t *testing.T) {
	urls := []string{
		"http://127.0.0.1/image.png",
		"http://127.0.0.1:8080/image.png",
		"http://127.1.2.3/image.png",
	}
	for _, u := range urls {
		err := reference.ValidateURL(u)
		if err == nil {
			t.Errorf("ValidateURL(%q) = nil; want error", u)
			continue
		}
		code := errs.CodeOf(err)
		if code != errs.ErrReferenceURLForbidden {
			t.Errorf("ValidateURL(%q) code = %q; want %q", u, code, errs.ErrReferenceURLForbidden)
		}
	}
}

func TestValidateURL_RejectPrivateIP_10(t *testing.T) {
	urls := []string{
		"http://10.0.0.1/image.png",
		"http://10.255.255.255/image.png",
	}
	for _, u := range urls {
		err := reference.ValidateURL(u)
		if err == nil {
			t.Errorf("ValidateURL(%q) = nil; want error", u)
			continue
		}
		code := errs.CodeOf(err)
		if code != errs.ErrReferenceURLForbidden {
			t.Errorf("ValidateURL(%q) code = %q; want %q", u, code, errs.ErrReferenceURLForbidden)
		}
	}
}

func TestValidateURL_RejectPrivateIP_172(t *testing.T) {
	urls := []string{
		"http://172.16.0.1/image.png",
		"http://172.31.255.255/image.png",
	}
	for _, u := range urls {
		err := reference.ValidateURL(u)
		if err == nil {
			t.Errorf("ValidateURL(%q) = nil; want error", u)
			continue
		}
		code := errs.CodeOf(err)
		if code != errs.ErrReferenceURLForbidden {
			t.Errorf("ValidateURL(%q) code = %q; want %q", u, code, errs.ErrReferenceURLForbidden)
		}
	}
}

func TestValidateURL_RejectPrivateIP_192(t *testing.T) {
	urls := []string{
		"http://192.168.0.1/image.png",
		"http://192.168.255.255/image.png",
	}
	for _, u := range urls {
		err := reference.ValidateURL(u)
		if err == nil {
			t.Errorf("ValidateURL(%q) = nil; want error", u)
			continue
		}
		code := errs.CodeOf(err)
		if code != errs.ErrReferenceURLForbidden {
			t.Errorf("ValidateURL(%q) code = %q; want %q", u, code, errs.ErrReferenceURLForbidden)
		}
	}
}

func TestValidateURL_RejectIPv6Loopback(t *testing.T) {
	urls := []string{
		"http://[::1]/image.png",
		"http://[::1]:8080/image.png",
	}
	for _, u := range urls {
		err := reference.ValidateURL(u)
		if err == nil {
			t.Errorf("ValidateURL(%q) = nil; want error", u)
			continue
		}
		code := errs.CodeOf(err)
		if code != errs.ErrReferenceURLForbidden {
			t.Errorf("ValidateURL(%q) code = %q; want %q", u, code, errs.ErrReferenceURLForbidden)
		}
	}
}

func TestValidateURL_RejectUnsupportedScheme(t *testing.T) {
	urls := []string{
		"ftp://example.com/image.png",
		"data:image/png;base64,abc",
		"file:///etc/passwd",
		"s3://bucket/image.png",
	}
	for _, u := range urls {
		err := reference.ValidateURL(u)
		if err == nil {
			t.Errorf("ValidateURL(%q) = nil; want error", u)
		}
	}
}
