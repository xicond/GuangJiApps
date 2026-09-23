package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"guangjiapps/gin/tests/e2e/factory"
)

func TestE2E_MasterData(t *testing.T) {
	e, auth := newAuthExpect(t)

	// 1. Umat CRUD & Reports
	t.Run("umat_crud_and_reports", func(t *testing.T) {
		// GET List
		e.GET("/v1/umats").
			WithHeader("Authorization", auth).
			WithQuery("page", 1).
			WithQuery("limit", 10).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data").
			Value("meta").Object().
			Value("total").Number().Ge(0)

		// GET Report
		e.GET("/v1/umats/report").
			WithHeader("Authorization", auth).
			WithQuery("page", 1).
			WithQuery("limit", 5).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data").
			Value("meta").Object().
			Value("total").Number().Ge(0)

		// GET Report Excel
		e.GET("/v1/umats/report/excel").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			Header("Content-Type").Contains("spreadsheetml")

		// GET PopUp
		e.GET("/v1/umats/popup").
			WithHeader("Authorization", auth).
			WithQuery("page", 1).
			WithQuery("limit", 5).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data").
			Value("meta").Object().
			Value("total").Number().Ge(0)

		// GET PopUp with filters
		e.GET("/v1/umats/popup").
			WithHeader("Authorization", auth).
			WithQuery("page", 1).
			WithQuery("limit", 5).
			WithQuery("nama_indonesia", "Budi").
			WithQuery("fotang_aktif", 0).
			WithQuery("fotang_ciu_tao", 0).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data").
			Value("meta").Object().
			Value("total").Number().Ge(0)

		// POST 400
		e.POST("/v1/umats").
			WithHeader("Authorization", auth).
			WithJSON(factory.Umat.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		createRes := e.POST("/v1/umats").
			WithHeader("Authorization", auth).
			WithJSON(factory.Umat.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		umatID := int(createRes.Value("data").Object().Value("id").Number().Raw())

		// GET By ID - verify qr_token is appended
		getRes := e.GET(fmt.Sprintf("/v1/umats/%d", umatID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		qrToken := getRes.Value("data").Object().Value("qr_token").String().Raw()

		// POST /v1/umats/verify-qr - 400 Bad Request (missing / empty qr_token)
		e.POST("/v1/umats/verify-qr").
			WithHeader("Authorization", auth).
			WithJSON(map[string]string{}).
			Expect().
			Status(http.StatusBadRequest)

		// POST /v1/umats/verify-qr - 400 Bad Request (invalid qr_token)
		e.POST("/v1/umats/verify-qr").
			WithHeader("Authorization", auth).
			WithJSON(map[string]string{"qr_token": "invalid.jwt.token"}).
			Expect().
			Status(http.StatusBadRequest)

		// POST /v1/umats/verify-qr - 200 OK (valid qr_token)
		verifyRes := e.POST("/v1/umats/verify-qr").
			WithHeader("Authorization", auth).
			WithJSON(map[string]string{"qr_token": qrToken}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		verifyData := verifyRes.Value("data").Object()
		verifyData.Value("nama_indonesia").String().NotEmpty()
		verifyData.Value("fotang_ciu_tao").NotNull()
		verifyData.Value("fotang_aktif").NotNull()
		verifyData.Value("claims").Object().NotEmpty()

		// PATCH 400
		e.PATCH(fmt.Sprintf("/v1/umats/%d", umatID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.Umat.InvalidUpdatePayload()).
			Expect().
			Status(http.StatusBadRequest)

		// PATCH 200
		e.PATCH(fmt.Sprintf("/v1/umats/%d", umatID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.Umat.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE(fmt.Sprintf("/v1/umats/%d", umatID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().IsEqual("deleted")
	})

	// 2. Topics CRUD
	t.Run("topics_crud", func(t *testing.T) {
		e.GET("/v1/topics").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data")

		// POST 400
		e.POST("/v1/topic").
			WithHeader("Authorization", auth).
			WithJSON(factory.Topic.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		topicPayload := factory.Topic.ValidPayload()
		topicCode := topicPayload["topic_code"].(string)

		e.POST("/v1/topic").
			WithHeader("Authorization", auth).
			WithJSON(topicPayload).
			Expect().
			Status(http.StatusCreated)

		// GET By Code
		e.GET("/v1/topic").
			WithHeader("Authorization", auth).
			WithQuery("code", topicCode).
			Expect().
			Status(http.StatusOK)

		// PATCH 400
		e.PATCH("/v1/topic").
			WithHeader("Authorization", auth).
			WithQuery("code", topicCode).
			WithJSON(factory.Topic.InvalidUpdatePayload()).
			Expect().
			Status(http.StatusBadRequest)

		// PATCH 200
		e.PATCH("/v1/topic").
			WithHeader("Authorization", auth).
			WithQuery("code", topicCode).
			WithJSON(factory.Topic.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE("/v1/topic").
			WithHeader("Authorization", auth).
			WithQuery("code", topicCode).
			Expect().
			Status(http.StatusOK)
	})

	// 3. Kelas Master CRUD
	t.Run("kelas_master_crud", func(t *testing.T) {
		e.GET("/v1/kelas-masters").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data")

		// POST 400
		e.POST("/v1/kelas-master").
			WithHeader("Authorization", auth).
			WithJSON(factory.KelasMaster.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		createRes := e.POST("/v1/kelas-master").
			WithHeader("Authorization", auth).
			WithJSON(factory.KelasMaster.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		lookupID := createRes.Value("data").Object().Value("lookup_id").String().Raw()

		// GET By ID
		e.GET("/v1/kelas-master").
			WithHeader("Authorization", auth).
			WithQuery("id", lookupID).
			Expect().
			Status(http.StatusOK)

		// PATCH 200
		e.PATCH("/v1/kelas-master").
			WithHeader("Authorization", auth).
			WithQuery("id", lookupID).
			WithJSON(factory.KelasMaster.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE("/v1/kelas-master").
			WithHeader("Authorization", auth).
			WithQuery("id", lookupID).
			Expect().
			Status(http.StatusOK)
	})

	// 4. Activity CRUD
	t.Run("activity_crud", func(t *testing.T) {
		e.GET("/v1/activities").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// POST 400
		e.POST("/v1/activity").
			WithHeader("Authorization", auth).
			WithJSON(factory.Activity.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		actPayload := factory.Activity.ValidPayload()
		actCode := actPayload["event_code"].(string)

		e.POST("/v1/activity").
			WithHeader("Authorization", auth).
			WithJSON(actPayload).
			Expect().
			Status(http.StatusCreated)

		// GET By Code
		e.GET("/v1/activity").
			WithHeader("Authorization", auth).
			WithQuery("code", actCode).
			Expect().
			Status(http.StatusOK)

		// PATCH 200
		e.PATCH("/v1/activity").
			WithHeader("Authorization", auth).
			WithQuery("code", actCode).
			WithJSON(factory.Activity.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE("/v1/activity").
			WithHeader("Authorization", auth).
			WithQuery("code", actCode).
			Expect().
			Status(http.StatusOK)
	})

	// 5. Tim Kerja CRUD & Lookups
	t.Run("tim_kerja_crud_and_lookups", func(t *testing.T) {
		e.GET("/v1/tim-kerja").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		e.GET("/v1/tim-kerja/lookup").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		e.GET("/v1/tim-kerja/lookup/1/sub").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		e.GET("/v1/tim-kerja/lookup-report").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// POST 400
		e.POST("/v1/tim-kerja").
			WithHeader("Authorization", auth).
			WithJSON(factory.TimKerja.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		createRes := e.POST("/v1/tim-kerja").
			WithHeader("Authorization", auth).
			WithJSON(factory.TimKerja.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		tkID := createRes.Value("data").Object().Value("lookup_id").String().Raw()

		// GET By ID
		e.GET(fmt.Sprintf("/v1/tim-kerja/%s", tkID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// PATCH 200
		e.PATCH(fmt.Sprintf("/v1/tim-kerja/%s", tkID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.TimKerja.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE(fmt.Sprintf("/v1/tim-kerja/%s", tkID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)
	})

	// 6. Tahun Ciu Tao CRUD
	t.Run("tahun_ciu_tao_crud", func(t *testing.T) {
		e.GET("/v1/tahun-ciu-tao/list").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// POST 400
		e.POST("/v1/tahun-ciu-tao").
			WithHeader("Authorization", auth).
			WithJSON(factory.TahunCiuTao.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		payload := factory.TahunCiuTao.ValidPayload()
		e.POST("/v1/tahun-ciu-tao").
			WithHeader("Authorization", auth).
			WithJSON(payload).
			Expect().
			Status(http.StatusCreated)

		// GET By ID
		e.GET("/v1/tahun-ciu-tao").
			WithHeader("Authorization", auth).
			WithQuery("tahun_mandarin", payload["tahun_mandarin"].(string)).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE("/v1/tahun-ciu-tao").
			WithHeader("Authorization", auth).
			WithQuery("tahun_mandarin", payload["tahun_mandarin"].(string)).
			Expect().
			Status(http.StatusOK)
	})

	// 7. Penggalang Dana CRUD
	t.Run("penggalang_dana_crud", func(t *testing.T) {
		e.GET("/v1/penggalang-dana").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// POST 400
		e.POST("/v1/penggalang-dana").
			WithHeader("Authorization", auth).
			WithJSON(factory.PenggalangDana.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		createRes := e.POST("/v1/penggalang-dana").
			WithHeader("Authorization", auth).
			WithJSON(factory.PenggalangDana.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		pgID := int(createRes.Value("data").Object().Value("id").Number().Raw())

		// GET By ID
		e.GET(fmt.Sprintf("/v1/penggalang-dana/%d", pgID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// PATCH 200
		e.PATCH(fmt.Sprintf("/v1/penggalang-dana/%d", pgID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.PenggalangDana.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE(fmt.Sprintf("/v1/penggalang-dana/%d", pgID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)
	})

	// 8. Sxy Donatur CRUD
	t.Run("sxy_donatur_crud", func(t *testing.T) {
		e.GET("/v1/sxy-donatur").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// POST 400
		e.POST("/v1/sxy-donatur").
			WithHeader("Authorization", auth).
			WithJSON(factory.SxyDonatur.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		createRes := e.POST("/v1/sxy-donatur").
			WithHeader("Authorization", auth).
			WithJSON(factory.SxyDonatur.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		sdID := int(createRes.Value("data").Object().Value("id").Number().Raw())

		// GET By ID
		e.GET(fmt.Sprintf("/v1/sxy-donatur/%d", sdID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// PATCH 200
		e.PATCH(fmt.Sprintf("/v1/sxy-donatur/%d", sdID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.SxyDonatur.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE(fmt.Sprintf("/v1/sxy-donatur/%d", sdID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)
	})

	// 9. Fotang Lookups
	t.Run("fotang_lookups_200", func(t *testing.T) {
		e.GET("/v1/fotang/lookup").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data")

		e.GET("/v1/fotang/lookup-sxy").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data")
	})
}
