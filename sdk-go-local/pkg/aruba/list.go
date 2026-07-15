package aruba

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"gopkg.in/yaml.v3"
)

// Wrapper is the constraint for List items: every resource wrapper satisfies it.
type Wrapper interface {
	URI() string
	ID() string
}

// List is a paginated collection of resource wrappers. Per-resource clients
// construct it after a server List call; callers use Next/Prev/All to iterate.
type List[T Wrapper] struct {
	httpEnvelopeMixin // StatusCode / Headers / RawHTTP / RawError — parity with single-resource wrappers

	items      []T
	total      int64
	self       string
	prev       string
	next       string
	first      string
	last       string
	callerOpts []CallOption
	refetch    func(ctx context.Context, url string) (*List[T], error)
	raw        any // JSON-safe payload only: *types.XxxList (resp.Data)
}

// listPayload is the type constraint satisfied by every per-resource list type
// (e.g. types.VPCListResponse, types.AlertsListResponse). All such types embed
// types.ListResponse, which carries a BaseList() method that is automatically
// promoted, so no per-type implementation is required.
type listPayload interface {
	BaseList() types.ListResponse
}

// newListFromResponse is the high-level constructor used by per-resource
// adapters. It extracts pagination links from resp.Data.BaseList(), stores
// resp.Data as the JSON-safe raw payload, and populates the HTTP envelope from
// resp. Returns a usable empty list when resp or resp.Data is nil.
func newListFromResponse[T Wrapper, L listPayload](
	items []T,
	resp *types.Response[L],
	opts []CallOption,
	refetch func(ctx context.Context, url string) (*List[T], error),
) *List[T] {
	l := &List[T]{
		items:      items,
		callerOpts: opts,
		refetch:    refetch,
	}
	if resp == nil {
		return l
	}
	populateHTTPEnvelope(&l.httpEnvelopeMixin, resp)
	if resp.Data != nil {
		l.raw = resp.Data
		base := (*resp.Data).BaseList()
		l.total = base.Total
		l.self = base.Self
		l.prev = base.Prev
		l.next = base.Next
		l.first = base.First
		l.last = base.Last
	}
	return l
}

// newList constructs a List from a server reply. lr holds the pagination links,
// raw is the JSON-safe wire payload (typically *types.XxxList, i.e. resp.Data),
// and refetch is provided by the per-resource client to fetch adjacent pages.
func newList[T Wrapper](
	items []T,
	total int64,
	self, prev, next, first, last string,
	raw any,
	opts []CallOption,
	refetch func(ctx context.Context, url string) (*List[T], error),
) *List[T] {
	return &List[T]{
		items:      items,
		total:      total,
		self:       self,
		prev:       prev,
		next:       next,
		first:      first,
		last:       last,
		callerOpts: opts,
		refetch:    refetch,
		raw:        raw,
	}
}

// Items returns the items on the current page.
func (l *List[T]) Items() []T { return l.items }

// Total returns the server-reported total item count across all pages.
func (l *List[T]) Total() int { return int(l.total) }

// HasNext reports whether a next page is available.
func (l *List[T]) HasNext() bool { return l.next != "" }

// HasPrev reports whether a previous page is available.
func (l *List[T]) HasPrev() bool { return l.prev != "" }

// Next fetches the next page.
func (l *List[T]) Next(ctx context.Context) (*List[T], error) {
	if !l.HasNext() {
		return nil, fmt.Errorf("no next page")
	}
	return l.refetch(ctx, l.next)
}

// Prev fetches the previous page.
func (l *List[T]) Prev(ctx context.Context) (*List[T], error) {
	if !l.HasPrev() {
		return nil, fmt.Errorf("no previous page")
	}
	return l.refetch(ctx, l.prev)
}

// First fetches the first page.
func (l *List[T]) First(ctx context.Context) (*List[T], error) {
	if l.first == "" {
		return nil, fmt.Errorf("no first page link")
	}
	return l.refetch(ctx, l.first)
}

// Last fetches the last page.
func (l *List[T]) Last(ctx context.Context) (*List[T], error) {
	if l.last == "" {
		return nil, fmt.Errorf("no last page link")
	}
	return l.refetch(ctx, l.last)
}

// Cursor returns the next and previous page URLs.
func (l *List[T]) Cursor() (next, prev string) {
	return l.next, l.prev
}

// Raw returns the JSON-marshal-safe wire payload, typically a *types.XxxList
// containing pagination metadata and the typed Values slice. Cast to the
// concrete *types.XxxList type to inspect fields not promoted to wrappers.
// HTTP envelope (status, headers, raw body, error envelope) is exposed via
// StatusCode(), Headers(), RawHTTP(), RawError() on the same list value.
func (l *List[T]) Raw() any { return l.raw }

// RawJSON returns the wire payload marshaled as JSON, or nil if the list has
// no payload (Raw() == nil).
func (l *List[T]) RawJSON() []byte {
	if l.raw == nil {
		return nil
	}
	b, err := json.Marshal(l.raw)
	if err != nil {
		return nil
	}
	return b
}

// RawYAML returns the wire payload marshaled as YAML, or nil if the list has
// no payload (Raw() == nil).
func (l *List[T]) RawYAML() []byte {
	if l.raw == nil {
		return nil
	}
	b, err := yaml.Marshal(l.raw)
	if err != nil {
		return nil
	}
	return b
}

// All iterates all pages, calling yield for each item. Iteration stops early
// if yield returns false. Returns the first error encountered while fetching.
func (l *List[T]) All(ctx context.Context, yield func(T) bool) error {
	current := l
	for {
		for _, item := range current.items {
			if !yield(item) {
				return nil
			}
		}
		if !current.HasNext() {
			return nil
		}
		next, err := current.Next(ctx)
		if err != nil {
			return err
		}
		current = next
	}
}
