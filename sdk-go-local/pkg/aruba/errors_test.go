package aruba

import (
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func pstr(s string) *string { return &s }

func TestHTTPError_NoErrResp(t *testing.T) {
	e := &HTTPError{StatusCode: 500}
	if e.Error() != "HTTP 500" {
		t.Errorf("Error() = %q", e.Error())
	}
}

func TestHTTPError_TitleOnly(t *testing.T) {
	e := &HTTPError{
		StatusCode: 404,
		ErrResp:    &types.ErrorResponse{Title: pstr("Not Found")},
	}
	got := e.Error()
	if !strings.Contains(got, "404") || !strings.Contains(got, "Not Found") {
		t.Errorf("Error() = %q", got)
	}
}

func TestHTTPError_TitleAndDetail(t *testing.T) {
	e := &HTTPError{
		StatusCode: 422,
		ErrResp: &types.ErrorResponse{
			Title:  pstr("Unprocessable Entity"),
			Detail: pstr("name already taken"),
		},
	}
	got := e.Error()
	if !strings.Contains(got, "422") || !strings.Contains(got, "Unprocessable Entity") || !strings.Contains(got, "name already taken") {
		t.Errorf("Error() = %q", got)
	}
}

func TestHTTPError_NilTitle(t *testing.T) {
	e := &HTTPError{
		StatusCode: 403,
		ErrResp:    &types.ErrorResponse{},
	}
	if e.Error() != "HTTP 403" {
		t.Errorf("Error() = %q", e.Error())
	}
}
