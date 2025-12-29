package must_test

import (
	"net/url"
	"testing"

	"github.com/merlindorin/go-shared/pkg/must"
)

func TestGet(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log("Test passed, panic was caught!")
		}
	}()

	must.Get(url.Parse(":wrong")) //nolint:staticcheck // expected error for panic test

	t.Errorf("Test failed, panic was expected")
}
