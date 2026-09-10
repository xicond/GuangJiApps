package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"guangjiapps/gin/tests/e2e/factory"
)

func TestE2E_AdminManagement(t *testing.T) {
	e, auth := newAuthExpect(t)

	// 1. Department
	t.Run("department_list_200", func(t *testing.T) {
		res := e.GET("/v1/department").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.ContainsKey("data")
		res.Value("data").Array()
	})

	// 2. Admins CRUD
	t.Run("admins_crud", func(t *testing.T) {
		// GET List
		listRes := e.GET("/v1/admins").
			WithHeader("Authorization", auth).
			WithQuery("page", 1).
			WithQuery("limit", 10).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		listRes.ContainsKey("data")
		listRes.Value("data").Array()
		listRes.ContainsKey("meta")

		// POST 400 (Invalid)
		e.POST("/v1/admins").
			WithHeader("Authorization", auth).
			WithJSON(factory.Admin.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("error")

		// POST 201 (Valid)
		createRes := e.POST("/v1/admins").
			WithHeader("Authorization", auth).
			WithJSON(factory.Admin.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		createRes.ContainsKey("data")
		createdID := int(createRes.Value("data").Object().Value("id").Number().Raw())

		// GET By ID 200
		getRes := e.GET(fmt.Sprintf("/v1/admins/%d", createdID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		getRes.Value("data").Object().Value("id").Number().IsEqual(createdID)

		// PATCH 400 (Invalid update)
		e.PATCH(fmt.Sprintf("/v1/admins/%d", createdID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.Admin.InvalidUpdatePayload()).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("error")

		// PATCH 200 (Valid update)
		e.PATCH(fmt.Sprintf("/v1/admins/%d", createdID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.Admin.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data")

		// DELETE 200
		e.DELETE(fmt.Sprintf("/v1/admins/%d", createdID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().IsEqual("deleted")

		// DELETE 404 (non-existent)
		e.DELETE("/v1/admins/99999999").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusNotFound)
	})

	// 3. Admin Groups CRUD
	t.Run("admin_groups_crud", func(t *testing.T) {
		// GET List
		e.GET("/v1/admin-groups").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data")

		// POST 400
		e.POST("/v1/admin-groups").
			WithHeader("Authorization", auth).
			WithJSON(factory.AdminGroup.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		createRes := e.POST("/v1/admin-groups").
			WithHeader("Authorization", auth).
			WithJSON(factory.AdminGroup.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		groupID := int(createRes.Value("data").Object().Value("group_id").Number().Raw())

		// GET By ID
		e.GET(fmt.Sprintf("/v1/admin-group/%d", groupID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// PATCH 400
		e.PATCH(fmt.Sprintf("/v1/admin-groups/%d", groupID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.AdminGroup.InvalidUpdatePayload()).
			Expect().
			Status(http.StatusBadRequest)

		// PATCH 200
		e.PATCH(fmt.Sprintf("/v1/admin-groups/%d", groupID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.AdminGroup.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE(fmt.Sprintf("/v1/admin-groups/%d", groupID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().IsEqual("deleted")
	})

	// 4. Group Menu Mappings CRUD
	t.Run("group_menu_mappings_crud", func(t *testing.T) {
		// GET List
		e.GET("/v1/group-menu-mappings").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data")

		// POST 400
		e.POST("/v1/group-menu-mappings").
			WithHeader("Authorization", auth).
			WithJSON(factory.GroupMenu.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		createRes := e.POST("/v1/group-menu-mappings").
			WithHeader("Authorization", auth).
			WithJSON(factory.GroupMenu.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		menuID := int(createRes.Value("data").Object().Value("menu_id").Number().Raw())

		// GET By ID
		e.GET(fmt.Sprintf("/v1/group-menu-mapping/%d", menuID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// PATCH 400
		e.PATCH(fmt.Sprintf("/v1/group-menu-mappings/%d", menuID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.GroupMenu.InvalidUpdatePayload()).
			Expect().
			Status(http.StatusBadRequest)

		// PATCH 200
		e.PATCH(fmt.Sprintf("/v1/group-menu-mappings/%d", menuID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.GroupMenu.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE(fmt.Sprintf("/v1/group-menu-mappings/%d", menuID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().IsEqual("deleted")
	})

	// 5. Admin Sub Warehouses CRUD
	t.Run("admin_sub_warehouses_crud", func(t *testing.T) {
		// GET List
		e.GET("/v1/admin-sub-warehouses").
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("data")

		// POST 400
		e.POST("/v1/admin-sub-warehouses").
			WithHeader("Authorization", auth).
			WithJSON(factory.AdminSubWarehouse.InvalidPayload()).
			Expect().
			Status(http.StatusBadRequest)

		// POST 201
		createRes := e.POST("/v1/admin-sub-warehouses").
			WithHeader("Authorization", auth).
			WithJSON(factory.AdminSubWarehouse.ValidPayload()).
			Expect().
			Status(http.StatusCreated).
			JSON().Object()

		subWhID := int(createRes.Value("data").Object().Value("sub_wh_id").Number().Raw())

		// GET By ID
		e.GET(fmt.Sprintf("/v1/admin-sub-warehouse/%d", subWhID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK)

		// PATCH 400
		e.PATCH(fmt.Sprintf("/v1/admin-sub-warehouses/%d", subWhID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.AdminSubWarehouse.InvalidUpdatePayload()).
			Expect().
			Status(http.StatusBadRequest)

		// PATCH 200
		e.PATCH(fmt.Sprintf("/v1/admin-sub-warehouses/%d", subWhID)).
			WithHeader("Authorization", auth).
			WithJSON(factory.AdminSubWarehouse.ValidUpdatePayload()).
			Expect().
			Status(http.StatusOK)

		// DELETE 200
		e.DELETE(fmt.Sprintf("/v1/admin-sub-warehouses/%d", subWhID)).
			WithHeader("Authorization", auth).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().IsEqual("deleted")
	})
}
