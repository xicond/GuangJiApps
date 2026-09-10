package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

func TestE2E_LookupService(t *testing.T) {
	e, auth := newAuthExpect(t)

	lookups := []string{
		"waktu-ciu-tao",
		"gender",
		"kelas-level",
		"tcs",
		"fotang",
		"kelas",
		"pendidikan",
		"kelas-umum",
		"pekerjaan",
		"keluarga",
		"status",
		"kategori-topic",
		"kategori-event",
		"tipe-sumbangan",
	}

	for _, lk := range lookups {
		t.Run(fmt.Sprintf("lookup_%s_200", lk), func(t *testing.T) {
			path := fmt.Sprintf("/v1/lookup/%s", lk)
			res := e.GET(path).
				WithHeader("Authorization", auth).
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			res.ContainsKey("data")
			res.Value("data").Array()
			res.Value("resource").String().Contains("Lookup")
		})
	}

	t.Run("unauthenticated_401", func(t *testing.T) {
		e.GET("/v1/lookup/gender").
			Expect().
			Status(http.StatusUnauthorized)
	})
}
