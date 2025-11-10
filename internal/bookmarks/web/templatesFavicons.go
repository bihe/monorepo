package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/web/html"
	"golang.binggl.net/monorepo/internal/common"
	base "golang.binggl.net/monorepo/pkg/handler/html"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/text"
)

const errorFavicon = `<span id="bookmark_favicon_display" class="error_icon">
<i id="error_tooltip_favicon" class="position_error_icon bi bi-exclamation-square-fill" data-bs-toggle="tooltip" data-bs-title="%s"></i>
</span>

<script type="text/javascript">
[...document.querySelectorAll('[data-bs-toggle="tooltip"]')].map(tooltipTriggerEl => new bootstrap.Tooltip(tooltipTriggerEl))
</script>
`
const favIconImage = `<img id="bookmark_favicon_display" class="bookmark_favicon_preview" src="/bm/favicon/temp/%s">
<input type="hidden" name="bookmark_Favicon" value="%s"/>`

const existingFaviconImage = `<img id="bookmark_favicon_display" class="bookmark_favicon_preview" src="/bm/favicon/raw/%s">
<input type="hidden" name="bookmark_Favicon" value="%s"/>`

// AvailableFaviconsDialog shows the dialog to edit favicons
func (t *TemplateHandler) AvailableFaviconsDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		currFavicon := queryParam(r, "current")
		favicons, err := t.App.GetAvailableFavicons(*user, "")
		if err != nil {
			t.Logger.Error(fmt.Sprintf("could not get available favicons for user '%s'", user.DisplayName), logging.ErrV(err), logging.LogV("username", user.Username))
		}
		faviconDialog := html.FaviconModalDialog(currFavicon, favicons, "")
		faviconDialog.Render(r.Context(), w)
	}
}

// FaviconGridPartial provides the grid of available favicons
func (t *TemplateHandler) FaviconGridPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		filterFavicon := ""
		if r.ParseForm() == nil {
			filterFavicon = r.FormValue("search_favicon")
			t.Logger.Debug("will filter for favicon", logging.LogV("search_favicon", filterFavicon))
		}
		favicons, err := t.App.GetAvailableFavicons(*user, filterFavicon)
		if err != nil {
			t.Logger.Error(fmt.Sprintf("could not get available favicons for user '%s'", user.DisplayName), logging.ErrV(err), logging.LogV("username", user.Username))
		}
		faviconDialog := html.FaviconGrid("", favicons, filterFavicon)
		faviconDialog.Render(r.Context(), w)
	}
}

// SelectExistingFavicon uses an already existing favicon
func (t *TemplateHandler) SelectExistingFavicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base64_id := pathParam(r, "id")
		id := text.DecBase64(base64_id)
		if id == "" {
			t.Logger.ErrorRequest("could not get the favicon ID", r)
			triggerToast(w,
				base.MsgError,
				"Favicon error!",
				"Could not select the favicon")
			w.Write([]byte(fmt.Sprintf(errorFavicon, "Could not select favicon")))
			return
		}
		w.Write([]byte(fmt.Sprintf(existingFaviconImage, base64_id, bookmarks.ExistingFavicon+id)))
	}
}

// FetchCustomFaviconURL fetches the given custom favicon URL and returns a new image
func (t *TemplateHandler) FetchCustomFaviconURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		favURL := r.FormValue("bookmark_CustomFavicon")
		fav, err := t.App.LocalFetchFaviconURL(favURL)
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not fetch the custom favicon; '%v'", err), r)
			triggerToast(w,
				base.MsgError,
				"Favicon error!",
				fmt.Sprintf("Could not fetch favicon: %v", errMsg))
			w.Write([]byte(fmt.Sprintf(errorFavicon, errMsg)))
			return
		}
		w.Write([]byte(fmt.Sprintf(favIconImage, fav.Name, fav.Name)))
	}
}

// FetchCustomFaviconFromPage tries to fetch the favicon from the given Page-URL
func (t *TemplateHandler) FetchCustomFaviconFromPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pageUrl := r.FormValue("bookmark_URL")
		fav, err := t.App.LocalExtractFaviconFromURL(pageUrl)
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not fetch the page favicon; '%v'", err), r)
			triggerToast(w,
				base.MsgError,
				"Favicon error!",
				fmt.Sprintf("Could not fetch favicon: %v", errMsg))
			w.Write([]byte(fmt.Sprintf(errorFavicon, errMsg)))
			return
		}
		w.Write([]byte(fmt.Sprintf(favIconImage, fav.Name, fav.Name)))
	}
}

// UploadCustomFavicon takes a file upload and stores the payload locally
func (t *TemplateHandler) UploadCustomFavicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse our multipart form, 10 << 20 specifies a maximum
		// upload of 10 MB files.
		r.ParseMultipartForm(10 << 20)

		file, meta, err := r.FormFile("bookmark_customFaviconUpload")
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not upload the custom favicon; '%v'", err), r)
			triggerToast(w,
				base.MsgError,
				"Favicon error!",
				fmt.Sprintf("Could not upload the custom favicon: %v", errMsg))
			w.Write([]byte(fmt.Sprintf(errorFavicon, errMsg)))
			return
		}
		defer file.Close()

		cType := meta.Header.Get("Content-Type")
		if !strings.HasPrefix(cType, "image") {
			t.Logger.ErrorRequest(fmt.Sprintf("only image types are supported - got '%s'", cType), r)
			triggerToast(w,
				base.MsgError,
				"Favicon error!",
				"Only an image mimetype is supported!")
			w.Write([]byte(fmt.Sprintf(errorFavicon, "Only an image mimetype is supported!")))
			return
		}
		payload, err := io.ReadAll(file)
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not read data of upload '%s'", cType), r)
			triggerToast(w,
				base.MsgError,
				"Favicon error!",
				fmt.Sprintf("Could not read data of upload: %v!", errMsg))
			w.Write([]byte(fmt.Sprintf(errorFavicon, errMsg)))
			return
		}
		fav, err := t.App.WriteLocalFavicon(meta.Filename, cType, payload)
		if err != nil {
			errMsg := strings.ReplaceAll(err.Error(), "\"", "'")
			t.Logger.ErrorRequest(fmt.Sprintf("could not write the custom favicon; '%v'", err), r)
			triggerToast(w,
				base.MsgError,
				"Favicon error!",
				fmt.Sprintf("Could not write the custom favicon: %v!", errMsg))
			w.Write([]byte(fmt.Sprintf(errorFavicon, errMsg)))
			return
		}
		w.Write([]byte(fmt.Sprintf(favIconImage, fav.Name, fav.Name)))
	}
}

// GetFaviconByBookmarkID returns the favicon of a given bookmark
func (t *TemplateHandler) GetFaviconByBookmarkID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		id := pathParam(r, "id")
		t.Logger.InfoRequest(fmt.Sprintf("try to get bookmark's favicon with the given ID '%s'", id), r)
		favicon, err := t.App.GetBookmarkFavicon(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get favicon by ID; '%v'", err), r)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeContent(w, r, favicon.Name, favicon.Modified, bytes.NewReader(favicon.Payload))
	}
}

// GetFaviconByID returns a stored favicon by the provided object-ID
// the object-ID is base64 encoded
func (t *TemplateHandler) GetFaviconByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodedID := pathParam(r, "id")
		id := text.DecBase64(encodedID)
		if id == "" {
			t.Logger.ErrorRequest("could not get decode ID", r)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		t.Logger.InfoRequest(fmt.Sprintf("try to get favicon with the given ID '%s'", id), r)
		favicon, err := t.App.GetFaviconByID(id)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get favicon by ID; '%v'", err), r)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeContent(w, r, favicon.Name, favicon.Modified, bytes.NewReader(favicon.Payload))
	}
}

// GetTempFaviconByID returns a temporarily saved favicon specified by the given ID
func (t *TemplateHandler) GetTempFaviconByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		t.Logger.InfoRequest(fmt.Sprintf("try to get locally stored favicon by ID '%s'", id), r)
		favicon, err := t.App.GetLocalFaviconByID(id)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get locally stored favicon by ID; '%v'", err), r)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeContent(w, r, favicon.Name, favicon.Modified, bytes.NewReader(favicon.Payload))
	}
}
