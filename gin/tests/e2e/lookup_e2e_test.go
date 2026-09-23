package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

// TestE2E_LookupService validates all static and dynamic system lookup endpoints (200),
// confirming proper array data structures and resource identifiers, as well as 401 unauthenticated enforcement.
//
// Parameters:
//   - t: active testing.T pointer.
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
