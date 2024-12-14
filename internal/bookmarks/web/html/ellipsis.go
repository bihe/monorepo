package html

import (
	"net/http"
	"strconv"
	"strings"
)

type EllipsisValues struct {
	PathLen   int
	NodeLen   int
	FolderLen int
}

// --------------------------------------------------------------------------
//  UI Ellipsis Handling
// --------------------------------------------------------------------------

const cookieViewPortName = "viewport"

func getViewPort(r *http.Request) (x, y int) {
	cookie, _ := r.Cookie(cookieViewPortName)
	if cookie != nil && cookie.Value != "" {
		dim := strings.Split(cookie.Value, ":")
		if len(dim) == 2 {
			if v, err := strconv.Atoi(dim[0]); err == nil {
				x = v
			}
			if v, err := strconv.Atoi(dim[1]); err == nil {
				y = v
			}
		}
	}
	return
}

// Desktop Browser
var StdEllipsis = EllipsisValues{
	PathLen:   50,
	NodeLen:   60,
	FolderLen: 50,
}

// Mobile View
var MobileEllipsis = EllipsisValues{
	PathLen:   5,
	NodeLen:   30,
	FolderLen: 20,
}

func GetEllipsisValues(r *http.Request) (ell EllipsisValues) {
	vX, _ := getViewPort(r)
	ell = StdEllipsis
	if vX == 0 {
		// this looks odd - use the std
		return
	}
	// iPhone 12 Pro
	if vX <= 390 {
		ell = MobileEllipsis
	}
	return
}
