package web

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"github.com/gigawattio/go-commons/pkg/oslib"
)

// AssetProvider defines the function signature for a StaticFilesMiddleware
// asset looker-upper.
//
// While a default implementation of disk-based AssetProvider is provided,
// tools such as go-bindata can also be easily used.
type AssetProvider func(name string) ([]byte, error)

// StaticFilesMiddleware generates a middleware function for serving static
// files from the specified asset provider.
func StaticFilesMiddleware(assetProvider AssetProvider) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == "GET" && len(req.RequestURI) > 1 {
				filePath := req.RequestURI[1:] // Remove leading "/".
				// Attempt to serve the "file".
				data, err := assetProvider(filePath)
				if err != nil {
					next.ServeHTTP(w, req)
					return
				}
				var contentType string
				if strings.Contains(filePath, ".") {
					// Auto-detect the mimetype.
					pieces := strings.Split(filePath, ".")
					ext := "." + pieces[len(pieces)-1]
					contentType = mime.TypeByExtension(ext)
				}
				RespondWith(w, 200, contentType, data)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

func DiskBasedAssetProvider(basePath string) AssetProvider {
	var FileNotOnDiskError = fmt.Errorf("requested file not found on disk under basePath=%s", basePath)

	return func(name string) ([]byte, error) {
		filepath := basePath + "/" + name
		exists, err := oslib.PathExists(filepath)
		if err != nil {
			log.Error("error checking if filepath=%s exists: %s", filepath, err)
			return nil, err
		}
		if exists {
			data, err := ioutil.ReadFile(filepath)
			if err != nil {
				log.Error("error reading filepath=%s: %s", filepath, err)
				return nil, err
			}
			return data, nil
		}
		return nil, FileNotOnDiskError
	}
}
