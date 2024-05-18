package compress

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/Stern-Ritter/gophermart/internal/utils"
)

var compressedContentTypes = []string{"application/json", "text/plain", "text/html"}

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(body []byte) (int, error) {
	contentType := c.Header().Values("Content-type")
	needCompress := utils.Contains(contentType, compressedContentTypes...)
	if needCompress {
		c.w.Header().Set("Content-Encoding", "gzip")
		return c.zw.Write(body)
	}
	return c.w.Write(body)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(body []byte) (n int, err error) {
	return c.zr.Read(body)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Values("Content-Encoding")
		sendsGzip := utils.Contains(contentEncoding, "gzip")
		contentType := r.Header.Values("Content-type")
		needUncompressed := utils.Contains(compressedContentTypes, contentType...)

		if sendsGzip && needUncompressed {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
		}

		ow := w
		acceptEncoding := r.Header.Values("Accept-Encoding")
		supportsGzip := utils.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		next.ServeHTTP(ow, r)
	})
}
