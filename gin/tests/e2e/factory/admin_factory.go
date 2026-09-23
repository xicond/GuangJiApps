package factory

import (
	"strings"
)

// AdminFactory provides test payload generator methods for administrator account testing.
type AdminFactory struct{}

// Admin is the global singleton instance of AdminFactory.
var Admin = AdminFactory{}

// ValidPayload creates a valid payload map for creating an administrator account.
//
// Returns:
//   - map[string]interface{}: valid creation payload.
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

// InvalidPayload creates an intentionally invalid payload (missing required username) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid creation payload.
func (AdminFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"username": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating an existing administrator account.
//
// Returns:
//   - map[string]interface{}: valid update payload.
func (AdminFactory) ValidUpdatePayload() map[string]interface{} {
	email := RandEmail("updated")
	phone := RandPhone()
	return map[string]interface{}{
		"email":        email,
		"phone_number": phone,
		"flag_use":     false,
	}
}

// InvalidUpdatePayload creates an invalid update payload (malformed email) to test patch validation failure.
//
// Returns:
//   - map[string]interface{}: invalid update payload.
func (AdminFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"email": "not-a-valid-email", // invalid email format
	}
}

// AdminGroupFactory provides test payload generator methods for admin role group testing.
type AdminGroupFactory struct{}

// AdminGroup is the global singleton instance of AdminGroupFactory.
var AdminGroup = AdminGroupFactory{}

// ValidPayload creates a valid payload map for creating an admin group.
//
// Returns:
//   - map[string]interface{}: valid group creation payload.
func (AdminGroupFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"group_name": "Group_" + RandDigits(5),
		"group_desc": "E2E Test Group Description",
	}
}

// InvalidPayload creates an invalid payload (missing required group_name) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid group creation payload.
func (AdminGroupFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"group_name": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating an admin group.
//
// Returns:
//   - map[string]interface{}: valid group update payload.
func (AdminGroupFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"group_desc": "Updated E2E Description",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid group update payload.
func (AdminGroupFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"group_name": strings.Repeat("x", 55), // max=50
	}
}

// GroupMenuMappingFactory provides test payload generator methods for menu mapping and permission testing.
type GroupMenuMappingFactory struct{}

// GroupMenu is the global singleton instance of GroupMenuMappingFactory.
var GroupMenu = GroupMenuMappingFactory{}

// ValidPayload creates a valid payload map for creating a group menu mapping.
//
// Returns:
//   - map[string]interface{}: valid menu mapping creation payload.
func (GroupMenuMappingFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"menu_name":   "Menu_" + RandDigits(5),
		"page_url":    "/test/" + RandString(4),
		"sequence":    1,
		"flag_active": true,
	}
}

// InvalidPayload creates an invalid payload (missing required menu_name) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid menu mapping payload.
func (GroupMenuMappingFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"menu_name": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a group menu mapping.
//
// Returns:
//   - map[string]interface{}: valid menu mapping update payload.
func (GroupMenuMappingFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"menu_desc":   "Updated Menu Desc",
		"flag_active": false,
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid menu mapping update payload.
func (GroupMenuMappingFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"menu_name": strings.Repeat("m", 55), // max=50
	}
}

// AdminSubWarehouseFactory provides test payload generator methods for sub-warehouse testing.
type AdminSubWarehouseFactory struct{}

// AdminSubWarehouse is the global singleton instance of AdminSubWarehouseFactory.
var AdminSubWarehouse = AdminSubWarehouseFactory{}

// ValidPayload creates a valid payload map for creating an admin sub-warehouse.
//
// Returns:
//   - map[string]interface{}: valid sub-warehouse creation payload.
func (AdminSubWarehouseFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"full_name":        "SubWarehouse_" + RandDigits(5),
		"wh_id":            1,
		"pic":              "Tester",
		"doc_code":         "DOC_" + RandDigits(3),
		"flag_productions": true,
	}
}

// InvalidPayload creates an invalid payload (missing required full_name) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid sub-warehouse creation payload.
func (AdminSubWarehouseFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"full_name": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating an admin sub-warehouse.
//
// Returns:
//   - map[string]interface{}: valid sub-warehouse update payload.
func (AdminSubWarehouseFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"pic": "Tester Updated",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid sub-warehouse update payload.
func (AdminSubWarehouseFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"full_name": strings.Repeat("w", 105), // max=100
	}
}
