package web

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
	"golang.binggl.net/monorepo/internal/bookmarks/web/html"
	"golang.binggl.net/monorepo/internal/common"
	"golang.binggl.net/monorepo/pkg/handler"
	base "golang.binggl.net/monorepo/pkg/handler/html"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/pkg/security"
)

// TemplateHandler takes care of providing HTML templates.
// This is the new approach with a template + htmx based UI to replace the angular frontend
// and have a more go-oriented approach towards UI and user-interaction. This reduces the
// cognitive load because less technology mix is needed. Via template + htmx approach only
// a limited amount of javascript is needed to achieve the frontend.
// As additional benefit the build should be faster, because the nodejs build can be removed
type TemplateHandler struct {
	*handler.TemplateHandler
	App     *bookmarks.Application
	Version string
	Build   string
}

const searchURL = "/bm/search"

// SearchBookmarks performs a search for bookmarks and displays the result using server-side rendering
func (t *TemplateHandler) SearchBookmarks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := queryParam(r, "q")
		user := common.EnsureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by name: '%s' for user: '%s'", search, user.Username), r)

		bms, err := t.App.GetBookmarksByName(search, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for search '%s'; '%v'", search, err), r)
		}

		ell := html.GetEllipsisValues(r)

		base.Layout(
			t.pageModel("Bookmark Search", search, "/public/search_icon.svg", *user),
			html.SearchStyles(),
			html.SearchNavigation(search),
			html.SearchContent(search, bms, ell),
			searchURL,
		).Render(w)
	}
}

// SearchBookmarksPartial is used to return only the bookmark list for a search
func (t *TemplateHandler) SearchBookmarksPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := queryParam(r, "q")
		user := common.EnsureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get search bookmark-list partial by name: '%s' for user: '%s'", search, user.Username), r)

		bms, err := t.App.GetBookmarksByName(search, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for search '%s'; '%v'", search, err), r)
		}
		ell := html.GetEllipsisValues(r)
		html.SearchContent(search, bms, ell).Render(w)
	}
}

// GetBookmarksForPath retrieves and renders the bookmarks for a defined path
func (t *TemplateHandler) GetBookmarksForPath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pathParam(r, "*")
		if path == "" {
			// start with the root path
			path = "/"
		}
		user := common.EnsureUser(r)

		pathHierarchy := make([]html.BookmarkPathEntry, 1)
		// always start with the root item
		pathHierarchy[0] = html.BookmarkPathEntry{
			UrlPath:     "/",
			DisplayName: "/root",
		}

		// split up the path into it's sub-paths
		pathParts := strings.Split(path, "/")
		// now "grow" the paths so that the preceding part is added to the next part
		var lastPath string
		for _, p := range pathParts {
			if p == "" {
				continue
			}
			pathHierarchy = append(pathHierarchy, html.BookmarkPathEntry{
				UrlPath:     lastPath + "/" + p,
				DisplayName: p,
			})
			lastPath = lastPath + "/" + p
		}
		// highlight the last item
		pathHierarchy[len(pathHierarchy)-1].LastItem = true
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks for path: '%s' for user: '%s'", path, user.Username), r)

		bms, err := t.App.GetBookmarksByPath(path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err))
			return
		}

		curFolder := ""
		favicon := ""
		if len(pathParts) > 0 {
			curFolder = pathParts[len(pathParts)-1]
			folder, err := t.App.GetBookmarksFolderByPath(path, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmark folder for path '%s'; '%v'", path, err), r)
				t.RenderErr(r, w, fmt.Sprintf("could not get bookmark folder for path '%s'; '%v'", path, err))
				return
			}
			favicon = fmt.Sprintf("/bm/favicon/%s?t=%s", folder.ID, folder.TStamp())

			// we need to treat the root folder in a "special" way
			if path == "/" && strings.HasSuffix(folder.ID, "ROOT") {
				// use a "default" favicon for the root folder
				favicon = "/public/Clone_white.png"
			}
		}

		ell := html.GetEllipsisValues(r)
		base.Layout(
			t.pageModel(curFolder, "", favicon, *user),
			html.BookmarksByPathStyles(),
			html.BookmarksByPathNavigation(pathHierarchy),
			html.BookmarkList(path, bms, ell),
			searchURL,
		).Render(w)
	}
}

// GetBookmarksForPathPartial only returns the bookmark list of the whole page
func (t *TemplateHandler) GetBookmarksForPathPartial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := pathParam(r, "*")
		user := common.EnsureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmark-list partial for path: '%s' for user: '%s'", path, user.Username), r)

		bms, err := t.App.GetBookmarksByPath(path, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for path '%s'; '%v'", path, err), r)
		}
		html.BookmarkList(path, bms, html.GetEllipsisValues(r)).Render(w)
	}
}

// DeleteConfirm shows a confirm dialog before an item is deleted
func (t *TemplateHandler) DeleteConfirm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := common.EnsureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by id: '%s' for user: '%s'", id, user.Username), r)

		bm, err := t.App.GetBookmarkByID(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
		}
		base.DialogConfirmDeleteHx(bm.DisplayName, "/bm/delete/"+bm.ID).Render(w)
	}
}

// DeleteBookmark deletes the given bookmark and replaces the bookmark-list with a new list without the deleted bookmark
func (t *TemplateHandler) DeleteBookmark() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := common.EnsureUser(r)

		t.Logger.InfoRequest(fmt.Sprintf("get bookmark by id: '%s' for user: '%s'", id, user.Username), r)
		bm, err := t.App.GetBookmarkByID(id, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmark for id '%s'; '%v'", id, err), r)
		}
		err = t.App.Delete(bm.ID, *user)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not delete bookmark with id '%s'; '%v'", id, err), r)
			triggerRefreshWithToast(w,
				common.MsgError,
				"Bookmark delete error!",
				fmt.Sprintf("Error: '%s'", err))
			return
		}
		triggerRefreshWithToast(w,
			common.MsgSuccess,
			"Bookmark deleted!",
			fmt.Sprintf("The bookmark '%s' was deleted.", bm.DisplayName))

	}
}

// EditBookmarkDialog shows a dialog to edit a bookmark
func (t *TemplateHandler) EditBookmarkDialog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathParam(r, "id")
		user := common.EnsureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("get bookmarks by id: '%s' for user: '%s'", id, user.Username), r)

		var (
			bm    html.Bookmark
			b     *bookmarks.Bookmark
			paths []string
			err   error
		)
		if id == "-1" {
			bm.ID = html.ValidatorInput{Val: "-1", Valid: true}
			bm.Path = html.ValidatorInput{Val: queryParam(r, "path"), Valid: true}
			bm.DisplayName = html.ValidatorInput{Valid: true}
			bm.URL = html.ValidatorInput{Valid: true}
			bm.CustomFavicon = html.ValidatorInput{Valid: true}
			bm.Type = bookmarks.Node
			bm.TStamp = fmt.Sprintf("%d", time.Now().Unix())
		} else {
			// fetch an existing bookmark
			b, err = t.App.GetBookmarkByID(id, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err), r)
				t.RenderErr(r, w, fmt.Sprintf("could not get bookmarks for id '%s'; '%v'", id, err))
				return
			}
			bm.ID = html.ValidatorInput{Val: b.ID, Valid: true}
			bm.Path = html.ValidatorInput{Val: b.Path, Valid: true}
			bm.DisplayName = html.ValidatorInput{Val: b.DisplayName, Valid: true}
			bm.URL = html.ValidatorInput{Val: b.URL, Valid: true}
			bm.Type = b.Type
			bm.CustomFavicon = html.ValidatorInput{Valid: true}
			bm.InvertFaviconColor = (b.InvertFaviconColor == 1)
			bm.TStamp = b.TStamp()

			paths, err = t.App.GetAllPaths(*user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get all paths for bookmarks; '%v'", err), r)
			}
		}
		html.EditBookmarks(bm, paths).Render(w)
	}
}

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
		favicons, err := t.App.GetAvailableFavicons(*user)
		if err != nil {
			t.Logger.Error(fmt.Sprintf("could not get available favicons for user '%s'", user.DisplayName), logging.ErrV(err), logging.LogV("username", user.Username))
		}
		html.FaviconDialog(favicons).Render(w)
	}
}

func decodeBase64(input string) string {
	decode, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return ""
	}
	return string(decode)
}

// SelectExistingFavicon uses an already existing favicon
func (t *TemplateHandler) SelectExistingFavicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		base64_id := pathParam(r, "id")
		id := decodeBase64(base64_id)
		if id == "" {
			t.Logger.ErrorRequest("could not get the favicon ID", r)
			triggerToast(w,
				common.MsgError,
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
				common.MsgError,
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
				common.MsgError,
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
				common.MsgError,
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
				common.MsgError,
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
				common.MsgError,
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
				common.MsgError,
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
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedID)
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not get decode ID; '%v'", err), r)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id := string(decodedBytes)

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

type triggerDef struct {
	common.ToastMessage
	Refresh string `json:"refreshBookmarkList,omitempty"`
}

// SaveBookmark receives the values from the edit form, validates and persists the data
func (t *TemplateHandler) SaveBookmark() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			recv       bookmarks.Bookmark
			formBm     html.Bookmark
			paths      []string
			err        error
			formPrefix string = "bookmark_"
		)

		err = r.ParseForm()
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not parse supplied form data; '%v'", err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not parse supplied form data; '%v'", err))
			return
		}
		user := common.EnsureUser(r)
		t.Logger.InfoRequest(fmt.Sprintf("save the bookmark data for user: '%s'", user.Username), r)

		customFavicon := false
		recv.ID = r.FormValue(formPrefix + "ID")
		recv.Path = r.FormValue(formPrefix + "Path")
		recv.DisplayName = r.FormValue(formPrefix + "DisplayName")
		recv.URL = r.FormValue(formPrefix + "URL")
		recv.InvertFaviconColor = getIntFromString(r.FormValue(formPrefix + "InvertFaviconColor"))
		recv.Type = bookmarks.Node
		recv.Favicon = r.FormValue(formPrefix + "Favicon")
		if r.FormValue(formPrefix+"Type") == "Folder" {
			recv.Type = bookmarks.Folder
		}

		// validation
		validData := true
		formBm.TStamp = fmt.Sprintf("%d", time.Now().Unix())
		formBm.ID = html.ValidatorInput{Val: recv.ID, Valid: true}
		formBm.DisplayName = html.ValidatorInput{Val: recv.DisplayName, Valid: true}
		if recv.DisplayName == "" {
			formBm.DisplayName.Valid = false
			formBm.DisplayName.Message = "missing value!"
			validData = false
		}
		formBm.Path = html.ValidatorInput{Val: recv.Path, Valid: true}
		if recv.Path == "" {
			formBm.Path.Valid = false
			formBm.Path.Message = "missing value!"
			validData = false
		}
		formBm.Type = recv.Type
		if formBm.Type == bookmarks.Node {
			formBm.URL = html.ValidatorInput{Val: recv.URL, Valid: true}
			if recv.URL == "" {
				formBm.URL.Valid = false
				formBm.URL.Message = "missing value!"
				validData = false
			}
		}
		if customFavicon {
			formBm.UseCustomFavicon = customFavicon
			formBm.CustomFavicon = html.ValidatorInput{Val: recv.Favicon, Valid: true}
			if recv.Favicon == "" {
				formBm.CustomFavicon.Valid = false
				formBm.CustomFavicon.Message = "missing value!"
				validData = false
			}
		}
		formBm.InvertFaviconColor = (recv.InvertFaviconColor == 1)

		if !validData {
			// show the same form again!
			t.Logger.ErrorRequest("the supplied data for creating bookmark entry is not valid!", r)
			html.EditBookmarks(formBm, paths).Render(w)
			return
		}

		if recv.ID == "-1" {
			// create a new bookmark entry
			bm := bookmarks.Bookmark{
				Path:               recv.Path,
				DisplayName:        recv.DisplayName,
				Type:               recv.Type,
				URL:                recv.URL,
				InvertFaviconColor: recv.InvertFaviconColor,
				Favicon:            recv.Favicon,
			}
			if customFavicon {
				// the provided favicon needs to be an ID
				bm.Favicon = recv.Favicon
			}
			created, err := t.App.CreateBookmark(bm, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not create a new bookmark entry; '%v'", err), r)
				formBm.Error = "Error: " + err.Error()
				html.EditBookmarks(formBm, paths).Render(w)
				return
			}
			t.Logger.Info("new bookmark created", logging.LogV("ID", created.ID))

			triggerRefreshWithToast(w,
				common.MsgSuccess,
				"Bookmark saved!",
				fmt.Sprintf("The bookmark '%s' was created.", created.DisplayName))

			formBm.Close = true
			html.EditBookmarks(formBm, paths).Render(w)
			return

		} else {
			paths, err = t.App.GetAllPaths(*user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not get all paths for bookmarks; '%v'", err), r)
			}

			// update an exiting entry
			existing, err := t.App.GetBookmarkByID(recv.ID, *user)
			if err != nil {
				t.Logger.Error("the given bookmark ID is not available", logging.ErrV(err), logging.LogV("ID", recv.ID))
				t.RenderErr(r, w, fmt.Sprintf("the given bookmark '%s' is not available; %v", recv.ID, err))
				return
			}

			// update the existing bookmark with the supplied data
			existing.DisplayName = recv.DisplayName
			existing.Path = recv.Path
			existing.InvertFaviconColor = recv.InvertFaviconColor
			existing.URL = recv.URL
			existing.Favicon = recv.Favicon
			formBm.TStamp = existing.TStamp()
			updated, err := t.App.UpdateBookmark(*existing, *user)
			if err != nil {
				t.Logger.ErrorRequest(fmt.Sprintf("could not update bookmark entry '%s'; '%v'", recv.ID, err), r)
				formBm.Error = "Error: " + err.Error()
				html.EditBookmarks(formBm, paths).Render(w)
				return
			}

			t.Logger.Info("bookmark updated", logging.LogV("ID", updated.ID))

			triggerRefreshWithToast(w,
				common.MsgSuccess,
				"Bookmark saved!",
				fmt.Sprintf("The bookmark '%s' (%s) was updated.", existing.DisplayName, existing.ID))

			formBm.Close = true
			html.EditBookmarks(formBm, paths).Render(w)
			return
		}
	}
}

// SortBookmarks performs a reordering of the bookmark list
func (t *TemplateHandler) SortBookmarks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := common.EnsureUser(r)
		err := r.ParseForm()
		if err != nil {
			t.Logger.ErrorRequest(fmt.Sprintf("could not parse supplied form data; '%v'", err), r)
			t.RenderErr(r, w, fmt.Sprintf("could not parse supplied form data; '%v'", err))
			return
		}

		idList := r.Form["ID"]
		indexList := make([]int, len(idList))
		t.Logger.Info("will save the new bookmark list", logging.LogV("ID_List", fmt.Sprintf("%v", idList)))
		// we jus ust the order of supplied IDs as the "natural" sort-order
		for i := range idList {
			indexList[i] = i
		}

		sortOrder := bookmarks.BookmarksSortOrder{
			IDs:       idList,
			SortOrder: indexList,
		}

		updates, err := t.App.UpdateSortOrder(sortOrder, *user)
		if err != nil {

			triggerToast(w,
				common.MsgError,
				"Error sorting!",
				fmt.Sprintf("Could not perform sorting: %v", err))
			return
		}

		triggerRefreshWithToast(w,
			common.MsgSuccess,
			"List sorted!",
			fmt.Sprintf("%d bookmarks were successfully sorted", updates))
	}
}

// --------------------------------------------------------------------------
//  Internals
// --------------------------------------------------------------------------

func (t *TemplateHandler) versionString() string {
	return fmt.Sprintf("%s-%s", t.Version, t.Build)
}

func (t *TemplateHandler) pageModel(pageTitle, searchStr, favicon string, user security.User) base.LayoutModel {
	return common.CreatePageModel("/bm", pageTitle, searchStr, favicon, t.versionString(), t.Env, user)
}

func queryParam(r *http.Request, name string) string {
	keys, ok := r.URL.Query()[name]
	if !ok || len(keys[0]) < 1 {
		return ""
	}
	return keys[0]
}

func pathParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

func getIntFromString(val string) int {
	if val == "1" {
		return 1
	}
	return 0
}

// define a header for htmx to trigger events
// https://htmx.org/headers/hx-trigger/
const htmxHeaderTrigger = "HX-Trigger"

func triggerRefreshWithToast(w http.ResponseWriter, errType, title, text string) {
	triggerEvent := triggerDef{
		ToastMessage: common.ToastMessage{
			Event: common.ToastMessageContent{
				Type:  errType,
				Title: title,
				Text:  text,
			},
		},
		Refresh: "now",
	}
	triggerJson := common.Json(triggerEvent)
	w.Header().Add(htmxHeaderTrigger, triggerJson)
}

func triggerToast(w http.ResponseWriter, errType, title, text string) {
	triggerEvent := triggerDef{
		ToastMessage: common.ToastMessage{
			Event: common.ToastMessageContent{
				Type:  errType,
				Title: title,
				Text:  text,
			},
		},
	}
	triggerJson := common.Json(triggerEvent)
	w.Header().Add(htmxHeaderTrigger, triggerJson)
}
