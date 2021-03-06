/*
This file is part of GO gallery.

GO gallery is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GO gallery is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GO gallery.  If not, see <http://www.gnu.org/licenses/>.
*/

package handler

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/gorilla/mux"
	imageupload "github.com/olahol/go-imageupload"
	"gitlab.com/coliss86/go-gallery/pkg/conf"
	"gitlab.com/coliss86/go-gallery/pkg/file"
	"gitlab.com/coliss86/go-gallery/pkg/img"
)

var urlImgRE = regexp.MustCompile("(.*)/([^/]*)/?")

func RenderImg(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	vars := mux.Vars(r)

	serveFile(w, r, file.PathJoin(conf.Config.Images, vars["img"]))
}

type convert func(string, string)

func RenderThumb(w http.ResponseWriter, r *http.Request) {
	renderImageResize(w, r, img.ConvertThumbnail, "thumb")
}

func RenderSmall(w http.ResponseWriter, r *http.Request) {
	renderImageResize(w, r, img.ConvertSmall, "small")
}

func renderImageResize(w http.ResponseWriter, r *http.Request, fn convert, cacheName string) {
	r.ParseForm()
	vars := mux.Vars(r)
	thumb := vars["img"]

	ir := file.PathJoin(conf.Config.Images, thumb)
	ic := file.PathJoin(conf.Config.Cache, cacheName, thumb)
	// Return a 404 if the template doesn't exist
	infoir, err := os.Stat(ir)
	if err != nil && os.IsNotExist(err) || infoir.IsDir() {
		http.NotFound(w, r)
		return
	}

	dir := path.Dir(ic)
	os.MkdirAll(dir, os.ModePerm)

	infoic, err := os.Stat(ic)
	if err != nil && os.IsNotExist(err) || infoir.ModTime().After(infoic.ModTime()) {
		fn(ir, ic)
	}

	serveFile(w, r, ic)
}

func RenderDownload(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	vars := mux.Vars(r)
	img := vars["img"]
	matches := urlImgRE.FindStringSubmatch(img)
	title := ""
	if len(matches) > 0 {
		title = matches[2]
	} else {
		title = img
	}
	w.Header().Set("Content-Disposition", "attachment; filename="+title)
	serveFile(w, r, file.PathJoin(conf.Config.Images, img))
}

func RenderUpload(w http.ResponseWriter, r *http.Request) {
	img, err := imageupload.Process(r, "file")
	if err != nil {
		http.Error(w, "invalid file", http.StatusUnprocessableEntity)
		return
	}
	img.Save(fmt.Sprintf("pics/%s", img.Filename))
	http.Redirect(w, r, "/ui/", http.StatusMovedPermanently)
}

func serveFile(w http.ResponseWriter, r *http.Request, file string) {
	// Return a 404 if the template doesn't exist
	info, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	fileOs, err := os.Open(file)
	defer fileOs.Close()
	http.ServeContent(w, r, info.Name(), info.ModTime(), fileOs)
}
