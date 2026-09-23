package factory

import "strings"

// UmatFactory provides test payload generator methods for congregant (Umat) profile testing.
type UmatFactory struct{}

// Umat is the global singleton instance of UmatFactory.
var Umat = UmatFactory{}

// ValidPayload creates a valid payload map for creating a new congregant (Umat).
//
// Returns:
//   - map[string]interface{}: valid umat creation payload.
func (UmatFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"nama_indonesia": "Umat " + RandString(6),
		"jenis_kelamin":  "001",
		"status_umat":    "001",
		"tempat_lahir":   "Jakarta",
		"tanggal_lahir":  "1995-05-12",
	}
}

// InvalidPayload creates an invalid payload (missing required nama_indonesia) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid umat payload.
func (UmatFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"nama_indonesia": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating an existing umat record.
//
// Returns:
//   - map[string]interface{}: valid umat update payload.
func (UmatFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"nama_indonesia": "Umat Updated " + RandString(4),
		"jenis_kelamin":  "1",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max name length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid umat update payload.
func (UmatFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"nama_indonesia": strings.Repeat("A", 60), // max=50
	}
}

// TopicFactory provides test payload generator methods for dharma lecture topic testing.
type TopicFactory struct{}

// Topic is the global singleton instance of TopicFactory.
var Topic = TopicFactory{}

// ValidPayload creates a valid payload map for creating a topic.
//
// Returns:
//   - map[string]interface{}: valid topic creation payload.
func (TopicFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"topic_code":     RandCode("TP", 4),
		"topic_name":     "Topic " + RandString(8),
		"topic_category": "1",
		"description":    "E2E Test Topic Description",
		"status":         true,
	}
}

// InvalidPayload creates an invalid payload (missing required topic_code) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid topic payload.
func (TopicFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"topic_code": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a topic.
//
// Returns:
//   - map[string]interface{}: valid topic update payload.
func (TopicFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"description": "Updated Topic Description",
		"status":      true,
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max name length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid topic update payload.
func (TopicFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"topic_name": strings.Repeat("x", 310), // max=300
	}
}

// KelasMasterFactory provides test payload generator methods for class master lookup testing.
type KelasMasterFactory struct{}

// KelasMaster is the global singleton instance of KelasMasterFactory.
var KelasMaster = KelasMasterFactory{}

// ValidPayload creates a valid payload map for creating a master class lookup item.
//
// Returns:
//   - map[string]interface{}: valid class lookup creation payload.
func (KelasMasterFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "Kelas Dharma " + RandString(6),
	}
}

// InvalidPayload creates an invalid payload (empty description) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid class lookup payload.
func (KelasMasterFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a master class lookup item.
//
// Returns:
//   - map[string]interface{}: valid class lookup update payload.
func (KelasMasterFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "Kelas Revisi " + RandString(6),
		"status":             true,
	}
}

// InvalidUpdatePayload creates an invalid update payload (empty description) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid class lookup update payload.
func (KelasMasterFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "",
	}
}

// ActivityFactory provides test payload generator methods for temple activity and event testing.
type ActivityFactory struct{}

// Activity is the global singleton instance of ActivityFactory.
var Activity = ActivityFactory{}

// ValidPayload creates a valid payload map for creating an activity event.
//
// Returns:
//   - map[string]interface{}: valid activity creation payload.
func (ActivityFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"event_code":     RandCode("AC", 3),
		"event_name":     "Activity " + RandString(6),
		"event_category": "001",
		"description":    "E2E Activity Description",
		"status":         true,
	}
}

// InvalidPayload creates an invalid payload (event_code exceeding length limit) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid activity payload.
func (ActivityFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"event_code": "EVENT_CODE_LONG_123", // max 10
	}
}

// ValidUpdatePayload creates a valid payload map for updating an activity event.
//
// Returns:
//   - map[string]interface{}: valid activity update payload.
func (ActivityFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"description": "Updated Activity Description",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max name length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid activity update payload.
func (ActivityFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"event_name": strings.Repeat("e", 105), // max 100
	}
}

// TimKerjaFactory provides test payload generator methods for work team / committee lookup testing.
type TimKerjaFactory struct{}

// TimKerja is the global singleton instance of TimKerjaFactory.
var TimKerja = TimKerjaFactory{}

// ValidPayload creates a valid payload map for creating a work team lookup.
//
// Returns:
//   - map[string]interface{}: valid work team creation payload.
func (TimKerjaFactory) ValidPayload() map[string]interface{} {
	code := RandCode("TK", 4)
	return map[string]interface{}{
		"lookup_id":          code,
		"lookup_value":       code,
		"lookup_description": "Tim Kerja " + RandString(6),
		"status":             true,
	}
}

// InvalidPayload creates an invalid payload (empty lookup_value) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid work team payload.
func (TimKerjaFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_value": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a work team lookup.
//
// Returns:
//   - map[string]interface{}: valid work team update payload.
func (TimKerjaFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "Tim Kerja Updated " + RandString(4),
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max description length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid work team update payload.
func (TimKerjaFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": strings.Repeat("t", 155), // max 150
	}
}

// TahunCiuTaoFactory provides test payload generator methods for Chiu Tao lunar year calendar testing.
type TahunCiuTaoFactory struct{}

// TahunCiuTao is the global singleton instance of TahunCiuTaoFactory.
var TahunCiuTao = TahunCiuTaoFactory{}

// ValidPayload creates a valid payload map for creating a Chiu Tao lunar year entry.
//
// Returns:
//   - map[string]interface{}: valid year entry creation payload.
func (TahunCiuTaoFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"tahun_mandarin": "20" + RandDigits(2) + RandString(2),
		"description":    "Tahun Ciu Tao E2E",
		"status":         true,
	}
}

// InvalidPayload creates an invalid payload (missing required tahun_mandarin) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid year entry payload.
func (TahunCiuTaoFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"tahun_mandarin": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a Chiu Tao lunar year entry.
//
// Returns:
//   - map[string]interface{}: valid year entry update payload.
func (TahunCiuTaoFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"description": "Tahun Ciu Tao Updated",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max year length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid year entry update payload.
func (TahunCiuTaoFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"tahun_mandarin": strings.Repeat("y", 25), // max 20
	}
}

// PenggalangDanaFactory provides test payload generator methods for fundraiser profile testing.
type PenggalangDanaFactory struct{}

// PenggalangDana is the global singleton instance of PenggalangDanaFactory.
var PenggalangDana = PenggalangDanaFactory{}

// ValidPayload creates a valid payload map for creating a fundraiser.
//
// Returns:
//   - map[string]interface{}: valid fundraiser creation payload.
func (PenggalangDanaFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no":     RandCode("PG", 4),
		"nama":   "Penggalang " + RandString(6),
		"status": true,
	}
}

// InvalidPayload creates an invalid payload (missing required nama) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid fundraiser payload.
func (PenggalangDanaFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"nama": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a fundraiser.
//
// Returns:
//   - map[string]interface{}: valid fundraiser update payload.
func (PenggalangDanaFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Penggalang Updated",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max name length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid fundraiser update payload.
func (PenggalangDanaFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"nama": strings.Repeat("p", 55), // max 50
	}
}

// SxyDonaturFactory provides test payload generator methods for SXY charitable donor profile testing.
type SxyDonaturFactory struct{}

// SxyDonatur is the global singleton instance of SxyDonaturFactory.
var SxyDonatur = SxyDonaturFactory{}

// ValidPayload creates a valid payload map for creating a charitable donor profile.
//
// Returns:
//   - map[string]interface{}: valid donor creation payload.
func (SxyDonaturFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no":     RandCode("SD", 4),
		"nama":   "Donatur " + RandString(6),
		"status": true,
	}
}

// InvalidPayload creates an invalid payload (missing required no) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid donor payload.
func (SxyDonaturFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a charitable donor profile.
//
// Returns:
//   - map[string]interface{}: valid donor update payload.
func (SxyDonaturFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Donatur Updated",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max name length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid donor update payload.
func (SxyDonaturFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"nama": strings.Repeat("s", 55), // max 50
	}
}
