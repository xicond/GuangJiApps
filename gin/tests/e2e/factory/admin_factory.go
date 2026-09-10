package factory

import (
	"strings"
)

type AdminFactory struct{}

var Admin = AdminFactory{}

func (AdminFactory) ValidPayload() map[string]interface{} {
	u := "test_adm_" + RandDigits(6)
	email := RandEmail(u)
	phone := RandPhone()
	return map[string]interface{}{
		"username":     u,
		"group_id":     3,
		"email":        email,
		"phone_number": phone,
		"flag_use":     true,
	}
}

func (AdminFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"username": "", // required
	}
}

func (AdminFactory) ValidUpdatePayload() map[string]interface{} {
	email := RandEmail("updated")
	phone := RandPhone()
	return map[string]interface{}{
		"email":        email,
		"phone_number": phone,
		"flag_use":     false,
	}
}

func (AdminFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"email": "not-a-valid-email", // invalid email format
	}
}

type AdminGroupFactory struct{}

var AdminGroup = AdminGroupFactory{}

func (AdminGroupFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"group_name": "Group_" + RandDigits(5),
		"group_desc": "E2E Test Group Description",
	}
}

func (AdminGroupFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"group_name": "", // required
	}
}

func (AdminGroupFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"group_desc": "Updated E2E Description",
	}
}

func (AdminGroupFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"group_name": strings.Repeat("x", 55), // max=50
	}
}

type GroupMenuMappingFactory struct{}

var GroupMenu = GroupMenuMappingFactory{}

func (GroupMenuMappingFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"menu_name":   "Menu_" + RandDigits(5),
		"page_url":    "/test/" + RandString(4),
		"sequence":    1,
		"flag_active": true,
	}
}

func (GroupMenuMappingFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"menu_name": "", // required
	}
}

func (GroupMenuMappingFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"menu_desc":   "Updated Menu Desc",
		"flag_active": false,
	}
}

func (GroupMenuMappingFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"menu_name": strings.Repeat("m", 55), // max=50
	}
}

type AdminSubWarehouseFactory struct{}

var AdminSubWarehouse = AdminSubWarehouseFactory{}

func (AdminSubWarehouseFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"full_name":        "SubWarehouse_" + RandDigits(5),
		"wh_id":            1,
		"pic":              "Tester",
		"doc_code":         "DOC_" + RandDigits(3),
		"flag_productions": true,
	}
}

func (AdminSubWarehouseFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"full_name": "", // required
	}
}

func (AdminSubWarehouseFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"pic": "Tester Updated",
	}
}

func (AdminSubWarehouseFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"full_name": strings.Repeat("w", 105), // max=100
	}
}
