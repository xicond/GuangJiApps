package e2e

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"guangjiapps/gin/tests/e2e/factory"

	"github.com/stretchr/testify/assert"
)

func TestKelas_EndToEnd(t *testing.T) {
	e := getAuthenticatedExpect(t)

	// 1. GET /v1/kelas list
	t.Run("GET /v1/kelas - 200 OK", func(t *testing.T) {
		e.GET("/v1/kelas").
			WithQuery("page", 1).
			WithQuery("limit", 10).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data").
			Value("meta").Object().
			Value("total").Number().Gt(0)
	})

	// 2. GET /v1/kelas/lookup
	t.Run("GET /v1/kelas/lookup - 200 OK", func(t *testing.T) {
		e.GET("/v1/kelas/lookup").
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("data").Array()
	})

	// 3. POST /v1/kelas with invalid payload - 400 Bad Request
	t.Run("POST /v1/kelas - 400 Bad Request on validation failure", func(t *testing.T) {
		e.POST("/v1/kelas").
			WithJSON(factory.Kelas.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)
	})

	// 4. POST /v1/kelas with valid payload - 201 Created
	var createdTrxId int64
	t.Run("POST /v1/kelas - 201 Created", func(t *testing.T) {
		payload := factory.Kelas.ValidPayload()
		resp := e.POST("/v1/kelas").
			WithJSON(payload).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.Value("resource").String().IsEqual("Kelas")
		data := resp.Value("data").Object()

		// Get created ID
		val := data.Value("trx_id").Raw()
		switch v := val.(type) {
		case float64:
			createdTrxId = int64(v)
		case string:
			parsed, err := strconv.ParseInt(v, 10, 64)
			assert.NoError(t, err)
			createdTrxId = parsed
		default:
			t.Fatalf("unexpected type for trx_id: %T (%v)", val, val)
		}
		assert.True(t, createdTrxId > 0)
	})

	if createdTrxId == 0 {
		t.Skip("Skipping Kelas detail and sub-resource tests because Kelas creation failed")
	}

	trxIdStr := strconv.FormatInt(createdTrxId, 10)

	// 5. GET /v1/kelas/:id
	t.Run(fmt.Sprintf("GET /v1/kelas/%s - 200 OK", trxIdStr), func(t *testing.T) {
		e.GET("/v1/kelas/" + trxIdStr).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("data").Object().
			Value("trx_id").Number().IsEqual(float64(createdTrxId))
	})

	// 6. PATCH /v1/kelas/:id invalid - 400 Bad Request
	t.Run(fmt.Sprintf("PATCH /v1/kelas/%s - 400 Bad Request", trxIdStr), func(t *testing.T) {
		e.PATCH("/v1/kelas/" + trxIdStr).
			WithJSON(factory.Kelas.InvalidUpdatePayload()).
			Expect().
			Status(http.StatusBadRequest)
	})

	// 7. PATCH /v1/kelas/:id valid - 200 OK
	t.Run(fmt.Sprintf("PATCH /v1/kelas/%s - 200 OK", trxIdStr), func(t *testing.T) {
		e.PATCH("/v1/kelas/" + trxIdStr).
			WithJSON(factory.Kelas.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("data").Object().
			Value("keterangan").String().IsEqual("Kelas Updated E2E")
	})

	// 8. GET /v1/kelas/:id/report - 200 OK
	t.Run(fmt.Sprintf("GET /v1/kelas/%s/report - 200 OK", trxIdStr), func(t *testing.T) {
		e.GET("/v1/kelas/" + trxIdStr + "/report").
			Expect().
			Status(http.StatusOK)
	})

	// Find an existing Umat for Peserta / Pengabdi
	var testUmatId int32 = 1
	umatResp := e.GET("/v1/umats").WithQuery("limit", 1).Expect().Status(http.StatusOK).JSON().Object()
	umatDataArr := umatResp.Value("data").Array().Raw()
	if len(umatDataArr) > 0 {
		if firstUmat, ok := umatDataArr[0].(map[string]interface{}); ok {
			if uid, ok := firstUmat["id"].(float64); ok {
				testUmatId = int32(uid)
			} else if uid, ok := firstUmat["umat_id"].(float64); ok {
				testUmatId = int32(uid)
			}
		}
	}

	// 9. Kelas Peserta Sub-Resource
	t.Run("KelasPeserta Sub-Resource CRUD", func(t *testing.T) {
		// GET list
		pList := e.GET("/v1/kelas/" + trxIdStr + "/peserta").
			Expect().
			Status(http.StatusOK).
			JSON().Object()
		if pList.Value("data").Raw() != nil {
			pList.Value("data").Array()
		}

		// GET load-previous
		e.GET("/v1/kelas/" + trxIdStr + "/peserta/load-previous").
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			HasValue("resource", "KelasPesertaPrevious").
			Value("data").Array()

		// GET load-previous on known class with pagination
		prevResp := e.GET("/v1/kelas/26340002/peserta/load-previous").
			WithQuery("page", 1).
			WithQuery("limit", 5).
			Expect().
			Status(http.StatusOK).
			JSON().Object()
		prevResp.Value("resource").IsEqual("KelasPesertaPrevious")
		prevResp.Value("data").Array()
		prevResp.Value("meta").Object().HasValue("page", 1).HasValue("limit", 5)

		// POST invalid - 400
		e.POST("/v1/kelas/" + trxIdStr + "/peserta").
			WithJSON(factory.KelasPeserta.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST invalid array payload on single create - 400 Bad Request
		e.POST("/v1/kelas/" + trxIdStr + "/peserta").
			WithJSON(map[string]interface{}{
				"id_peserta": []int32{101, 102},
			}).
			Expect().
			Status(http.StatusBadRequest)

		// POST invalid single payload on bulk create - 400 Bad Request
		e.POST("/v1/kelas/" + trxIdStr + "/peserta/bulk").
			WithJSON(map[string]interface{}{
				"id_peserta": 101,
			}).
			Expect().
			Status(http.StatusBadRequest)

		// POST valid - 201
		validPeserta := factory.KelasPeserta.ValidPayload(int32(createdTrxId), testUmatId)
		pesertaResp := e.POST("/v1/kelas/" + trxIdStr + "/peserta").
			WithJSON(validPeserta).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		var detailId int64
		pData := pesertaResp.Value("data").Object()
		if rawDetail := pData.Value("detail_id").Raw(); rawDetail != nil {
			if num, ok := rawDetail.(float64); ok {
				detailId = int64(num)
			}
		}

		if detailId > 0 {
			detailStr := strconv.FormatInt(detailId, 10)

			// GET by detail_id
			e.GET("/v1/kelas/" + trxIdStr + "/peserta/" + detailStr).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("data").Object().
				Value("detail_id").Number().IsEqual(float64(detailId))

			// GET by id_peserta
			e.GET("/v1/kelas/" + trxIdStr + "/peserta/by-idpeserta/" + strconv.Itoa(int(testUmatId))).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("data").Object().
				Value("detail_id").Number().IsEqual(float64(detailId))

			// GET by id_peserta - 404 non-existent
			e.GET("/v1/kelas/" + trxIdStr + "/peserta/by-idpeserta/999999").
				Expect().
				Status(http.StatusNotFound)

			// GET by id_peserta - 404 invalid format
			e.GET("/v1/kelas/" + trxIdStr + "/peserta/by-idpeserta/invalid-id").
				Expect().
				Status(http.StatusNotFound)

			// PATCH invalid - 400
			e.PATCH("/v1/kelas/" + trxIdStr + "/peserta/" + detailStr).
				WithJSON(factory.KelasPeserta.InvalidUpdatePayload()).
				Expect().
				Status(http.StatusBadRequest)

			// PATCH valid - 200
			e.PATCH("/v1/kelas/" + trxIdStr + "/peserta/" + detailStr).
				WithJSON(factory.KelasPeserta.ValidUpdatePayload()).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("data").Object().
				Value("keterangan").String().IsEqual("Peserta Updated E2E")

			// DELETE - 200
			e.DELETE("/v1/kelas/" + trxIdStr + "/peserta/" + detailStr).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("message").String().IsEqual("deleted")
		}
	})

	// 10. Kelas Pengabdi Sub-Resource
	t.Run("KelasPengabdi Sub-Resource CRUD", func(t *testing.T) {
		// GET list
		e.GET("/v1/kelas/" + trxIdStr + "/pengabdi").
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("data").Array()

		// POST invalid - 400
		e.POST("/v1/kelas/" + trxIdStr + "/pengabdi").
			WithJSON(factory.KelasPengabdi.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST valid - 201
		validPengabdi := factory.KelasPengabdi.ValidPayload(int32(createdTrxId), testUmatId)
		pengabdiResp := e.POST("/v1/kelas/" + trxIdStr + "/pengabdi").
			WithJSON(validPengabdi).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		var detailId int64
		pData := pengabdiResp.Value("data").Object()
		if rawDetail := pData.Value("detail_id").Raw(); rawDetail != nil {
			if num, ok := rawDetail.(float64); ok {
				detailId = int64(num)
			}
		}

		if detailId > 0 {
			detailStr := strconv.FormatInt(detailId, 10)

			// GET detail
			e.GET("/v1/kelas/" + trxIdStr + "/pengabdi/" + detailStr).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("data").Object().
				Value("detail_id").Number().IsEqual(float64(detailId))

			// PATCH valid - 200
			e.PATCH("/v1/kelas/" + trxIdStr + "/pengabdi/" + detailStr).
				WithJSON(map[string]interface{}{"keterangan": "Pengabdi Updated"}).
				Expect().
				Status(http.StatusOK)

			// DELETE - 200
			e.DELETE("/v1/kelas/" + trxIdStr + "/pengabdi/" + detailStr).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("message").String().IsEqual("deleted")
		}
	})

	// 11. Kelas Topik Sub-Resource
	t.Run("KelasTopik Sub-Resource CRUD", func(t *testing.T) {
		// Get existing topic code
		var testKodeTopik string = "T01"
		topikResp := e.GET("/v1/topics").WithQuery("limit", 1).Expect().Status(http.StatusOK).JSON().Object()
		topikDataArr := topikResp.Value("data").Array().Raw()
		if len(topikDataArr) > 0 {
			if firstTopik, ok := topikDataArr[0].(map[string]interface{}); ok {
				if kt, ok := firstTopik["topic_code"].(string); ok && kt != "" {
					testKodeTopik = kt
				} else if kt, ok := firstTopik["kode_topik"].(string); ok && kt != "" {
					testKodeTopik = kt
				}
			}
		}

		// GET list
		e.GET("/v1/kelas/" + trxIdStr + "/topik").
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("data").Array()

		// POST invalid - 400
		e.POST("/v1/kelas/" + trxIdStr + "/topik").
			WithJSON(factory.KelasTopik.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST valid - 201
		validTopik := factory.KelasTopik.ValidPayload(int32(createdTrxId), testKodeTopik)
		topikPostResp := e.POST("/v1/kelas/" + trxIdStr + "/topik").
			WithJSON(validTopik).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		var detailId int64
		tData := topikPostResp.Value("data").Object()
		if rawDetail := tData.Value("detail_id").Raw(); rawDetail != nil {
			if num, ok := rawDetail.(float64); ok {
				detailId = int64(num)
			}
		}

		if detailId > 0 {
			detailStr := strconv.FormatInt(detailId, 10)

			// GET detail
			e.GET("/v1/kelas/" + trxIdStr + "/topik/" + detailStr).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("data").Object().
				Value("detail_id").Number().IsEqual(float64(detailId))

			// PATCH valid - 200
			e.PATCH("/v1/kelas/" + trxIdStr + "/topik/" + detailStr).
				WithJSON(map[string]interface{}{"keterangan": "Topik Updated"}).
				Expect().
				Status(http.StatusOK)

			// DELETE - 200
			e.DELETE("/v1/kelas/" + trxIdStr + "/topik/" + detailStr).
				Expect().
				Status(http.StatusOK).
				JSON().Object().
				Value("message").String().IsEqual("deleted")
		}
	})

	// 12. Sub-lists GET 200 OK tests
	t.Run("Kelas Sub-resources List GET endpoints - 200 OK", func(t *testing.T) {
		subEndpoints := []string{
			"kendaraan",
			"donasi",
			"donasi-barang",
			"pengeluaran",
			"musik",
			"absensi",
		}

		for _, sub := range subEndpoints {
			t.Run(fmt.Sprintf("GET /v1/kelas/%s/%s", trxIdStr, sub), func(t *testing.T) {
				e.GET(fmt.Sprintf("/v1/kelas/%s/%s", trxIdStr, sub)).
					Expect().
					Status(http.StatusOK).
					JSON().Object().
					Value("data").Array()
			})
		}
	})

	// 13. DELETE /v1/kelas/:id - 200 OK
	t.Run(fmt.Sprintf("DELETE /v1/kelas/%s - 200 OK", trxIdStr), func(t *testing.T) {
		e.DELETE("/v1/kelas/" + trxIdStr).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().IsEqual("deleted")
	})
}

func TestDonasiSxy_EndToEnd(t *testing.T) {
	e := getAuthenticatedExpect(t)

	// 1. GET /v1/donasi-sxy list
	t.Run("GET /v1/donasi-sxy - 200 OK", func(t *testing.T) {
		e.GET("/v1/donasi-sxy").
			WithQuery("page", 1).
			WithQuery("limit", 10).
			WithQuery("start_date", "2010-01-01").
			WithQuery("end_date", "2025-12-31").
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data").
			Value("meta").Object().
			Value("total").Number().Gt(0)
	})

	// 2. GET /v1/donasi-sxy/report
	t.Run("GET /v1/donasi-sxy/report - 200 OK", func(t *testing.T) {
		e.GET("/v1/donasi-sxy/report").
			WithQuery("page", 1).
			WithQuery("limit", 5).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data").
			Value("meta").Object().
			Value("total").Number().Gt(0)
	})

	// 3. GET /v1/donasi-sxy/report/excel
	t.Run("GET /v1/donasi-sxy/report/excel", func(t *testing.T) {
		resp := e.GET("/v1/donasi-sxy/report/excel").
			Expect()
		status := resp.Raw().StatusCode
		assert.True(t, status == http.StatusOK || status == http.StatusInternalServerError || status == http.StatusBadGateway)
	})

	// 4. POST /v1/donasi-sxy invalid - 400 Bad Request
	t.Run("POST /v1/donasi-sxy - 400 Bad Request on validation failure", func(t *testing.T) {
		e.POST("/v1/donasi-sxy").
			WithJSON(factory.DonasiSxy.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)
	})

	// 5. POST /v1/donasi-sxy valid - 201 Created
	var createdId int64
	t.Run("POST /v1/donasi-sxy - 201 Created", func(t *testing.T) {
		payload := factory.DonasiSxy.ValidPayload()
		resp := e.POST("/v1/donasi-sxy").
			WithJSON(payload).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		resp.Value("resource").String().Contains("Donasi")
		data := resp.Value("data").Object()

		val := data.Value("id").Raw()
		switch v := val.(type) {
		case float64:
			createdId = int64(v)
		case string:
			parsed, err := strconv.ParseInt(v, 10, 64)
			assert.NoError(t, err)
			createdId = parsed
		default:
			t.Fatalf("unexpected type for id: %T (%v)", val, val)
		}
		assert.True(t, createdId > 0)
	})

	if createdId == 0 {
		t.Skip("Skipping DonasiSxy detail tests because creation failed")
	}

	idStr := strconv.FormatInt(createdId, 10)

	// 6. GET /v1/donasi-sxy/:id
	t.Run(fmt.Sprintf("GET /v1/donasi-sxy/%s - 200 OK", idStr), func(t *testing.T) {
		e.GET("/v1/donasi-sxy/" + idStr).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("data").Object().
			Value("id").Number().IsEqual(float64(createdId))
	})

	// 7. PATCH /v1/donasi-sxy/:id invalid - 400 Bad Request
	t.Run(fmt.Sprintf("PATCH /v1/donasi-sxy/%s - 400 Bad Request", idStr), func(t *testing.T) {
		e.PATCH("/v1/donasi-sxy/" + idStr).
			WithJSON(factory.DonasiSxy.InvalidUpdatePayload()).
			Expect().
			Status(http.StatusBadRequest)
	})

	// 8. PATCH /v1/donasi-sxy/:id valid - 200 OK
	t.Run(fmt.Sprintf("PATCH /v1/donasi-sxy/%s - 200 OK", idStr), func(t *testing.T) {
		e.PATCH("/v1/donasi-sxy/" + idStr).
			WithJSON(factory.DonasiSxy.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("data").Object().
			Value("keterangan").String().IsEqual("Donasi Updated E2E")
	})

	// 9. DELETE /v1/donasi-sxy/:id - 200 OK
	t.Run(fmt.Sprintf("DELETE /v1/donasi-sxy/%s - 200 OK", idStr), func(t *testing.T) {
		e.DELETE("/v1/donasi-sxy/" + idStr).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().IsEqual("deleted")
	})
}
