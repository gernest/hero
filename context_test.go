package hero

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContext(t *testing.T) {
	w := httptest.NewRecorder()
	ctx := newContext(w)

	link := "http://www.example.com"

	// SetData
	ctx.SetData("hello", "world")
	if _, ok := ctx.Data["hello"]; !ok {
		t.Error("Expcted key value pairs to be set")
	}

	// ClearData
	ctx.ClearData()
	if len(ctx.Data) != 0 {
		t.Errorf("expcted 0 got %d", len(ctx.Data))
	}

	// Error
	ctx.SetErrorURI(errorsKeys.InvalidRequest, "", link, "")

	if !ctx.HasError {
		t.Error("expcted has error to be true")
	}

	if ctx.ErrID != errorsKeys.InvalidRequest {
		t.Errorf("expcted %s got %s", errorsKeys.InvalidRequest, ctx.ErrID)
	}

	// Redirect
	ctx.SetRedirect(link)

	rdir, err := ctx.GetRedirectURL()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(rdir, link+"?") {
		t.Errorf("expected a normal query got %s", rdir)
	}

	// Fragment
	ctx.SetRedirectFragment(true)
	rdir, err = ctx.GetRedirectURL()
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(rdir, link+"#") {
		t.Errorf("expected a fragment query got %s", rdir)
	}

	// commitJSON
	ctx.CommitJSON()

}
