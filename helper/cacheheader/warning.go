package cacheheader

import (
	"fmt"
	"net/http"
	"time"
)

// Warning represents an HTTP Warning: http://tools.ietf.org/html/rfc7234#section-5.5
type Warning int

// Warning response code
const (
	// Response is Stale
	// A cache SHOULD generate this whenever the sent response is stale.
	WarningResponseIsStale Warning = 110

	// Revalidation Failed
	// A cache SHOULD generate this when sending a stale
	// response because an attempt to validate the response failed, due to an
	// inability to reach the server.
	WarningRevalidationFailed Warning = 111

	// Disconnected Operation
	// A cache SHOULD generate this if it is intentionally disconnected from
	// the rest of the network for a period of time.
	WarningDisconnectedOperation Warning = 112

	// Heuristic Expiration
	//
	// A cache SHOULD generate this if it heuristically chose a freshness
	// lifetime greater than 24 hours and the response's age is greater than
	// 24 hours.
	WarningHeuristicExpiration Warning = 113

	// Miscellaneous Warning
	//
	// The warning text can include arbitrary information to be presented to
	// a human user or logged.  A system receiving this warning MUST NOT
	// take any automated action, besides presenting the warning to the
	// user.
	WarningMiscellaneousWarning Warning = 199

	// Transformation Applied
	//
	// This Warning code MUST be added by a proxy if it applies any
	// transformation to the representation, such as changing the
	// content-coding, media-type, or modifying the representation data,
	// unless this Warning code already appears in the response.
	WarningTransformationApplied Warning = 214

	// Miscellaneous Persistent Warning
	//
	// The warning text can include arbitrary information to be presented to
	// a human user or logged.  A system receiving this warning MUST NOT
	// take any automated action.
	WarningMiscellaneousPersistentWarning Warning = 299
)

// HeaderString will convert the warning to string format
func (w Warning) HeaderString(agent string, date time.Time) string {
	if agent == "" {
		agent = "-"
	} else {
		// TODO(pquerna): this doesn't escape agent if it contains bad things.
		agent = `"` + agent + `"`
	}
	return fmt.Sprintf(`%d %s "%s" %s`, w, agent, w.String(), date.Format(http.TimeFormat))
}

func (w Warning) String() string {
	switch w {
	case WarningResponseIsStale:
		return "Response is Stale"
	case WarningRevalidationFailed:
		return "Revalidation Failed"
	case WarningDisconnectedOperation:
		return "Disconnected Operation"
	case WarningHeuristicExpiration:
		return "Heuristic Expiration"
	case WarningMiscellaneousWarning:
		// TODO(pquerna): ideally had a better way to override this one code.
		return "Miscellaneous Warning"
	case WarningTransformationApplied:
		return "Transformation Applied"
	case WarningMiscellaneousPersistentWarning:
		// TODO(pquerna): same as WarningMiscellaneousWarning
		return "Miscellaneous Persistent Warning"
	}

	panic(w)
}
