package aruba

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"gopkg.in/yaml.v3"
)

// testItem is a minimal Wrapper implementation for list tests.
type testItem struct {
	id  string
	uri string
}

func (i testItem) ID() string  { return i.id }
func (i testItem) URI() string { return i.uri }

func makeTestList(items []testItem, prev, next, first, last string, refetch func(context.Context, string) (*List[testItem], error)) *List[testItem] {
	return newList[testItem](items, int64(len(items)), "", prev, next, first, last, nil, nil, refetch)
}

func TestList_Items(t *testing.T) {
	items := []testItem{{id: "a", uri: "/a"}, {id: "b", uri: "/b"}}
	l := makeTestList(items, "", "", "", "", nil)
	if got := l.Items(); len(got) != 2 {
		t.Errorf("Items() len = %d, want 2", len(got))
	}
	if l.Total() != 2 {
		t.Errorf("Total() = %d, want 2", l.Total())
	}
}

func TestList_HasNextHasPrev(t *testing.T) {
	l := makeTestList(nil, "", "/page2", "", "", nil)
	if !l.HasNext() {
		t.Error("HasNext() should be true")
	}
	if l.HasPrev() {
		t.Error("HasPrev() should be false when prev empty")
	}
}

func TestList_Next(t *testing.T) {
	var capturedURL string
	page2 := makeTestList([]testItem{{id: "c"}}, "", "", "", "", nil)

	refetch := func(_ context.Context, url string) (*List[testItem], error) {
		capturedURL = url
		return page2, nil
	}

	l := makeTestList([]testItem{{id: "a"}}, "", "/page2", "", "", refetch)
	got, err := l.Next(context.Background())
	if err != nil {
		t.Fatalf("Next() error: %v", err)
	}
	if capturedURL != "/page2" {
		t.Errorf("refetch called with %q, want %q", capturedURL, "/page2")
	}
	if got != page2 {
		t.Error("Next() did not return expected page")
	}
}

func TestList_NextNoLink(t *testing.T) {
	l := makeTestList(nil, "", "", "", "", nil)
	if _, err := l.Next(context.Background()); err == nil {
		t.Error("expected error when no next link")
	}
}

func TestList_Prev(t *testing.T) {
	var capturedURL string
	page0 := makeTestList(nil, "", "", "", "", nil)
	refetch := func(_ context.Context, url string) (*List[testItem], error) {
		capturedURL = url
		return page0, nil
	}
	l := makeTestList(nil, "/page0", "", "", "", refetch)
	if _, err := l.Prev(context.Background()); err != nil {
		t.Fatalf("Prev() error: %v", err)
	}
	if capturedURL != "/page0" {
		t.Errorf("refetch URL = %q", capturedURL)
	}
}

func TestList_FirstLast(t *testing.T) {
	var calls []string
	refetch := func(_ context.Context, url string) (*List[testItem], error) {
		calls = append(calls, url)
		return makeTestList(nil, "", "", "", "", nil), nil
	}
	l := makeTestList(nil, "", "", "/first", "/last", refetch)

	if _, err := l.First(context.Background()); err != nil {
		t.Fatalf("First() error: %v", err)
	}
	if _, err := l.Last(context.Background()); err != nil {
		t.Fatalf("Last() error: %v", err)
	}
	if len(calls) != 2 || calls[0] != "/first" || calls[1] != "/last" {
		t.Errorf("calls = %v", calls)
	}
}

func TestList_Cursor(t *testing.T) {
	l := makeTestList(nil, "/prev", "/next", "", "", nil)
	next, prev := l.Cursor()
	if next != "/next" || prev != "/prev" {
		t.Errorf("Cursor() = (%q, %q)", next, prev)
	}
}

func TestList_All_TwoPages(t *testing.T) {
	page2 := newList[testItem](
		[]testItem{{id: "c"}, {id: "d"}},
		4, "", "", "", "", "", nil, nil,
		nil,
	)
	refetch := func(_ context.Context, _ string) (*List[testItem], error) {
		return page2, nil
	}
	page1 := newList[testItem](
		[]testItem{{id: "a"}, {id: "b"}},
		4, "", "", "/page2", "", "", nil, nil,
		refetch,
	)

	var collected []string
	err := page1.All(context.Background(), func(item testItem) bool {
		collected = append(collected, item.id)
		return true
	})
	if err != nil {
		t.Fatalf("All() error: %v", err)
	}
	if len(collected) != 4 {
		t.Errorf("collected = %v, want [a b c d]", collected)
	}
}

func TestList_All_EarlyStop(t *testing.T) {
	l := makeTestList([]testItem{{id: "a"}, {id: "b"}, {id: "c"}}, "", "", "", "", nil)
	var count int
	_ = l.All(context.Background(), func(_ testItem) bool {
		count++
		return count < 2 // stop after second item
	})
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestList_Raw(t *testing.T) {
	raw := struct{ n int }{42}
	l := newList[testItem](nil, 0, "", "", "", "", "", raw, nil, nil)
	got, ok := l.Raw().(struct{ n int })
	if !ok || got.n != 42 {
		t.Errorf("Raw() = %v", l.Raw())
	}
}

// TestList_Raw_JSONMarshalable is the regression test for #298: prior to the
// fix Raw() returned *types.Response[L] whose embedded *http.Response.GetBody
// is a non-serialisable func, breaking json.Marshal. Raw() now returns only
// the JSON-safe wire payload (*types.XxxList).
func TestList_Raw_JSONMarshalable(t *testing.T) {
	resp := &types.Response[types.VPCListResponse]{
		Data: &types.VPCListResponse{
			ListResponse: types.ListResponse{Total: 2, Self: "/self", Next: "/next"},
			Values:       []types.VPCResponse{},
		},
		StatusCode: 200,
	}
	l := newListFromResponse[testItem, types.VPCListResponse](nil, resp, nil, nil)

	b, err := json.Marshal(l.Raw())
	if err != nil {
		t.Fatalf("json.Marshal(list.Raw()): %v", err)
	}
	var back types.VPCListResponse
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if back.Total != 2 || back.Self != "/self" || back.Next != "/next" {
		t.Errorf("round-trip lost fields: %+v", back)
	}
}

// TestList_HTTPEnvelopeAccessors verifies the accessors promoted from the
// embedded httpEnvelopeMixin (parity with single-resource wrappers).
func TestList_HTTPEnvelopeAccessors(t *testing.T) {
	resp := &types.Response[types.VPCListResponse]{
		Data:       &types.VPCListResponse{ListResponse: types.ListResponse{Total: 0}},
		StatusCode: 200,
		Headers:    http.Header{"X-Trace-Id": []string{"abc-123"}},
		RawBody:    []byte(`{"total":0,"values":[]}`),
	}
	l := newListFromResponse[testItem, types.VPCListResponse](nil, resp, nil, nil)

	if l.StatusCode() != 200 {
		t.Errorf("StatusCode() = %d, want 200", l.StatusCode())
	}
	if got := l.Headers().Get("X-Trace-Id"); got != "abc-123" {
		t.Errorf("Headers X-Trace-Id = %q", got)
	}
	if _, body := l.RawHTTP(); string(body) != `{"total":0,"values":[]}` {
		t.Errorf("RawHTTP body = %q", string(body))
	}
	if l.RawError() != nil {
		t.Errorf("RawError() = %v, want nil for 2xx", l.RawError())
	}
}

// TestList_NewListFromResponse_NilSafe ensures the helper tolerates a nil
// response and a nil resp.Data.
func TestList_NewListFromResponse_NilSafe(t *testing.T) {
	l1 := newListFromResponse[testItem, types.VPCListResponse](nil, nil, nil, nil)
	if l1 == nil || l1.Total() != 0 || l1.HasNext() || l1.HasPrev() {
		t.Errorf("nil resp produced bad list: %+v", l1)
	}
	l2 := newListFromResponse[testItem, types.VPCListResponse](
		nil,
		&types.Response[types.VPCListResponse]{StatusCode: 200},
		nil, nil,
	)
	if l2.Raw() != nil {
		t.Errorf("Raw() should be nil when resp.Data is nil, got %v", l2.Raw())
	}
	if l2.StatusCode() != 200 {
		t.Errorf("StatusCode() = %d, want 200", l2.StatusCode())
	}
}

func TestList_RawJSON_RoundTrip(t *testing.T) {
	resp := &types.Response[types.VPCListResponse]{
		Data: &types.VPCListResponse{
			ListResponse: types.ListResponse{Total: 3, Self: "/self", Next: "/next"},
			Values:       []types.VPCResponse{},
		},
	}
	l := newListFromResponse[testItem, types.VPCListResponse](nil, resp, nil, nil)
	b := l.RawJSON()
	if len(b) == 0 {
		t.Fatal("RawJSON() returned empty")
	}
	var back types.VPCListResponse
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if back.Total != 3 || back.Self != "/self" || back.Next != "/next" {
		t.Errorf("round-trip lost fields: %+v", back)
	}
}

func TestList_RawYAML_RoundTrip(t *testing.T) {
	resp := &types.Response[types.VPCListResponse]{
		Data: &types.VPCListResponse{
			ListResponse: types.ListResponse{Total: 5, Self: "/s", Next: "/n"},
			Values:       []types.VPCResponse{},
		},
	}
	l := newListFromResponse[testItem, types.VPCListResponse](nil, resp, nil, nil)
	b := l.RawYAML()
	if len(b) == 0 {
		t.Fatal("RawYAML() returned empty")
	}
	var back types.VPCListResponse
	if err := yaml.Unmarshal(b, &back); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if back.Total != 5 || back.Self != "/s" || back.Next != "/n" {
		t.Errorf("round-trip lost fields: %+v", back)
	}
}

func TestList_RawJSON_NilSafe(t *testing.T) {
	l := newListFromResponse[testItem, types.VPCListResponse](nil, nil, nil, nil)
	if b := l.RawJSON(); b != nil {
		t.Errorf("RawJSON() = %q, want nil", b)
	}
}

func TestList_RawYAML_NilSafe(t *testing.T) {
	l := newListFromResponse[testItem, types.VPCListResponse](nil, nil, nil, nil)
	if b := l.RawYAML(); b != nil {
		t.Errorf("RawYAML() = %q, want nil", b)
	}
}
