package aruba

import (
	"errors"
	"net/http"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// --------------------------------------------------------------------------
// errMixin — setter-time error accumulator
// --------------------------------------------------------------------------

type errMixin struct {
	errs []error
}

func (m *errMixin) addErr(err error) {
	if err != nil {
		m.errs = append(m.errs, err)
	}
}

// Err returns the joined setter-time errors, or nil if none were recorded.
func (m *errMixin) Err() error {
	return errors.Join(m.errs...)
}

// --------------------------------------------------------------------------
// metadataMixin — resource name and tags
// --------------------------------------------------------------------------

type metadataMixin struct {
	name string
	tags []string
}

func (m *metadataMixin) named(name string) {
	m.name = name
}

func (m *metadataMixin) addTag(tag string) {
	for _, t := range m.tags {
		if t == tag {
			return
		}
	}
	m.tags = append(m.tags, tag)
}

func (m *metadataMixin) removeTag(tag string) {
	out := m.tags[:0]
	for _, t := range m.tags {
		if t != tag {
			out = append(out, t)
		}
	}
	m.tags = out
}

func (m *metadataMixin) replaceTags(tags ...string) {
	m.tags = append([]string(nil), tags...)
}

// Name returns the name set via named.
func (m *metadataMixin) Name() string { return m.name }

// Tags returns a copy of the current tag slice.
func (m *metadataMixin) Tags() []string {
	if len(m.tags) == 0 {
		return nil
	}
	out := make([]string, len(m.tags))
	copy(out, m.tags)
	return out
}

func (m *metadataMixin) toMetadata() types.ResourceMetadataRequest {
	return types.ResourceMetadataRequest{Name: m.name, Tags: m.Tags()}
}

// --------------------------------------------------------------------------
// regionalMixin — resource location / region
// --------------------------------------------------------------------------

type regionalMixin struct {
	region Region
}

func (m *regionalMixin) inRegion(region Region) { m.region = region }

// Region returns the region value.
func (m *regionalMixin) Region() Region { return m.region }

func (m *regionalMixin) toLocation() *types.LocationRequest {
	if m.region == "" {
		return nil
	}
	return &types.LocationRequest{Value: m.region}
}

// --------------------------------------------------------------------------
// zonalMixin — resource zone (extends regionalMixin)
// --------------------------------------------------------------------------

// zonalMixin extends regionalMixin with zone tracking. Zones are always within
// a region (e.g. "ITBG-1" lives in region "ITBG"), so zonalMixin embeds
// regionalMixin and inherits its setter/getter/toLocation helper.
//
// The zone wire field is NOT part of types.LocationRequest — every zonal
// resource carries it on its own *PropertiesRequest under JSON tag "dataCenter".
// This mixin therefore only owns the value; each wrapper's toRequest() reads it
// via Zone() (for required Zone wire fields) or zonePtr() (for *Zone omitempty
// fields) and places it itself.
type zonalMixin struct {
	regionalMixin
	zone *Zone
}

func (m *zonalMixin) inZone(z Zone) { m.zone = &z }

// Zone returns the configured zone, or "" if InZone was never called.
func (m *zonalMixin) Zone() Zone {
	if m.zone == nil {
		return ""
	}
	return *m.zone
}

// zonePtr returns the underlying *Zone for resources whose wire field is
// *Zone with omitempty (e.g. BlockStorage, DBaaS). Returns nil if InZone
// was not called.
func (m *zonalMixin) zonePtr() *Zone { return m.zone }

// --------------------------------------------------------------------------
// responseMetadataMixin — post-server-reply metadata
// --------------------------------------------------------------------------

type responseMetadataMixin struct {
	meta *types.ResourceMetadataResponse
}

func (m *responseMetadataMixin) setMeta(meta *types.ResourceMetadataResponse) {
	m.meta = meta
}

// ID returns the resource's server-assigned ID, or "" if not yet received.
func (m *responseMetadataMixin) ID() string {
	if m.meta == nil || m.meta.ID == nil {
		return ""
	}
	return *m.meta.ID
}

// RespURI returns the resource's server-assigned URI, or "" if not yet received.
// Named RespURI to avoid collision with the Ref.URI() method on wrapper types that
// derive their URI from the response.
func (m *responseMetadataMixin) RespURI() string {
	if m.meta == nil || m.meta.URI == nil {
		return ""
	}
	return *m.meta.URI
}

// Project returns the owning project's ID from the response metadata, or "".
func (m *responseMetadataMixin) Project() string {
	if m.meta == nil || m.meta.ProjectMetadataResponse == nil {
		return ""
	}
	return m.meta.ProjectMetadataResponse.ID
}

// CreatedAt returns the resource creation time, or zero time.
func (m *responseMetadataMixin) CreatedAt() time.Time {
	if m.meta == nil || m.meta.CreationDate == nil {
		return time.Time{}
	}
	return *m.meta.CreationDate
}

// UpdatedAt returns the last update time, or zero time.
func (m *responseMetadataMixin) UpdatedAt() time.Time {
	if m.meta == nil || m.meta.UpdateDate == nil {
		return time.Time{}
	}
	return *m.meta.UpdateDate
}

// Version returns the resource version string, or "".
func (m *responseMetadataMixin) Version() string {
	if m.meta == nil || m.meta.Version == nil {
		return ""
	}
	return *m.meta.Version
}

// CreatedBy returns the identifier of the actor that created the resource, or "".
func (m *responseMetadataMixin) CreatedBy() string {
	if m.meta == nil || m.meta.CreatedBy == nil {
		return ""
	}
	return *m.meta.CreatedBy
}

// UpdatedBy returns the identifier of the actor that last updated the resource, or "".
func (m *responseMetadataMixin) UpdatedBy() string {
	if m.meta == nil || m.meta.UpdatedBy == nil {
		return ""
	}
	return *m.meta.UpdatedBy
}

// CreatedUser returns the user display name that created the resource, or "".
func (m *responseMetadataMixin) CreatedUser() string {
	if m.meta == nil || m.meta.CreatedUser == nil {
		return ""
	}
	return *m.meta.CreatedUser
}

// UpdatedUser returns the user display name that last updated the resource, or "".
func (m *responseMetadataMixin) UpdatedUser() string {
	if m.meta == nil || m.meta.UpdatedUser == nil {
		return ""
	}
	return *m.meta.UpdatedUser
}

// Raw returns the underlying *types.ResourceMetadataResponse, or nil.
func (m *responseMetadataMixin) Raw() *types.ResourceMetadataResponse {
	return m.meta
}

// --------------------------------------------------------------------------
// linkedMixin — linked resources
// --------------------------------------------------------------------------

type linkedMixin struct {
	linked []types.LinkedResourceCommon
}

func (m *linkedMixin) setLinked(l []types.LinkedResourceCommon) { m.linked = l }

// LinkedResources returns the slice of linked resources.
func (m *linkedMixin) LinkedResources() []types.LinkedResourceCommon { return m.linked }

// --------------------------------------------------------------------------
// httpEnvelopeMixin — HTTP response metadata
// --------------------------------------------------------------------------

type httpEnvelopeMixin struct {
	statusCode int
	headers    http.Header
	rawBody    []byte
	httpResp   *http.Response
	errResp    *types.ErrorResponse
}

// populateHTTPEnvelope fills an httpEnvelopeMixin from a typed *types.Response[T].
// Defined as a package-level generic function because Go does not allow generic methods.
func populateHTTPEnvelope[T any](m *httpEnvelopeMixin, resp *types.Response[T]) {
	if resp == nil {
		return
	}
	m.statusCode = resp.StatusCode
	m.headers = resp.Headers
	m.rawBody = resp.RawBody
	m.httpResp = resp.HTTPResponse
	m.errResp = resp.Error
}

// StatusCode returns the HTTP status code, or 0 before any response.
func (m *httpEnvelopeMixin) StatusCode() int { return m.statusCode }

// Headers returns the HTTP response headers, or nil.
func (m *httpEnvelopeMixin) Headers() http.Header { return m.headers }

// RawHTTP returns the underlying *http.Response and raw body bytes.
func (m *httpEnvelopeMixin) RawHTTP() (*http.Response, []byte) {
	return m.httpResp, m.rawBody
}

// RawError returns the parsed error response body for non-2xx replies, or nil.
func (m *httpEnvelopeMixin) RawError() *types.ErrorResponse { return m.errResp }
