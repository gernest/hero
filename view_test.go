package hero

import (
	"bytes"
	"testing"
)

func TestView(t *testing.T) {
	view, err := NewDefaultView("fixtures/views", false)
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
