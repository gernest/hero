package hero

import (
	"bytes"
	"testing"
)

func TestView(t *testing.T) {

	sampleBadTplDir := []string{
		"bogus", "fixtures/views/hello.tpl",
	}

	for _, v := range sampleBadTplDir {
		_, err := NewDefaultView(v)
		if err == nil {
			t.Error("expected an error")
		}
	}

	view, err := NewDefaultView("fixtures/views")
	if err != nil {
		t.FailNow()
	}
	sampleVies := []struct {
		name, context, result string
	}{
		{"hello.tpl", "hero", "hello hero"},
		{"nest/hello.tpl", "hero", "nest hero"},
	}

	for _, v := range sampleVies {
		data := make(map[string]interface{})
		data["Name"] = v.context
		out := &bytes.Buffer{}
		err = view.Render(out, v.name, data)
		if err != nil {
			t.Error(err)
		}
		if out.String() != v.result {
			t.Errorf("expected %s got %s", v.result, out)
		}
	}
}
