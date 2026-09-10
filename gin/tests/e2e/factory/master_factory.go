package factory

import "strings"

type UmatFactory struct{}

var Umat = UmatFactory{}

func (UmatFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"nama_indonesia": "Umat " + RandString(6),
		"jenis_kelamin":  "001",
		"status_umat":    "001",
		"tempat_lahir":   "Jakarta",
		"tanggal_lahir":  "1995-05-12",
	}
}

func (UmatFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"nama_indonesia": "", // required
	}
}

func (UmatFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"nama_indonesia": "Umat Updated " + RandString(4),
		"jenis_kelamin":  "1",
	}
}

func (UmatFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"nama_indonesia": strings.Repeat("A", 60), // max=50
	}
}

type TopicFactory struct{}

var Topic = TopicFactory{}

func (TopicFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"topic_code":     RandCode("TP", 4),
		"topic_name":     "Topic " + RandString(8),
		"topic_category": "1",
		"description":    "E2E Test Topic Description",
		"status":         true,
	}
}

func (TopicFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"topic_code": "", // required
	}
}

func (TopicFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"description": "Updated Topic Description",
		"status":      true,
	}
}

func (TopicFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"topic_name": strings.Repeat("x", 310), // max=300
	}
}

type KelasMasterFactory struct{}

var KelasMaster = KelasMasterFactory{}

func (KelasMasterFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "Kelas Dharma " + RandString(6),
	}
}

func (KelasMasterFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "", // required
	}
}

func (KelasMasterFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "Kelas Revisi " + RandString(6),
		"status":             true,
	}
}

func (KelasMasterFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "",
	}
}

type ActivityFactory struct{}

var Activity = ActivityFactory{}

func (ActivityFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"event_code":     RandCode("AC", 3),
		"event_name":     "Activity " + RandString(6),
		"event_category": "001",
		"description":    "E2E Activity Description",
		"status":         true,
	}
}

func (ActivityFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"event_code": "EVENT_CODE_LONG_123", // max 10
	}
}

func (ActivityFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"description": "Updated Activity Description",
	}
}

func (ActivityFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"event_name": strings.Repeat("e", 105), // max 100
	}
}

type TimKerjaFactory struct{}

var TimKerja = TimKerjaFactory{}

func (TimKerjaFactory) ValidPayload() map[string]interface{} {
	code := RandCode("TK", 4)
	return map[string]interface{}{
		"lookup_id":          code,
		"lookup_value":       code,
		"lookup_description": "Tim Kerja " + RandString(6),
		"status":             true,
	}
}

func (TimKerjaFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_value": "", // required
	}
}

func (TimKerjaFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": "Tim Kerja Updated " + RandString(4),
	}
}

func (TimKerjaFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"lookup_description": strings.Repeat("t", 155), // max 150
	}
}

type TahunCiuTaoFactory struct{}

var TahunCiuTao = TahunCiuTaoFactory{}

func (TahunCiuTaoFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"tahun_mandarin": "20" + RandDigits(2) + RandString(2),
		"description":    "Tahun Ciu Tao E2E",
		"status":         true,
	}
}

func (TahunCiuTaoFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"tahun_mandarin": "", // required
	}
}

func (TahunCiuTaoFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"description": "Tahun Ciu Tao Updated",
	}
}

func (TahunCiuTaoFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"tahun_mandarin": strings.Repeat("y", 25), // max 20
	}
}

type PenggalangDanaFactory struct{}

var PenggalangDana = PenggalangDanaFactory{}

func (PenggalangDanaFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no":     RandCode("PG", 4),
		"nama":   "Penggalang " + RandString(6),
		"status": true,
	}
}

func (PenggalangDanaFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"nama": "", // required
	}
}

func (PenggalangDanaFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Penggalang Updated",
	}
}

func (PenggalangDanaFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"nama": strings.Repeat("p", 55), // max 50
	}
}

type SxyDonaturFactory struct{}

var SxyDonatur = SxyDonaturFactory{}

func (SxyDonaturFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no":     RandCode("SD", 4),
		"nama":   "Donatur " + RandString(6),
		"status": true,
	}
}

func (SxyDonaturFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no": "", // required
	}
}

func (SxyDonaturFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Donatur Updated",
	}
}

func (SxyDonaturFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"nama": strings.Repeat("s", 55), // max 50
	}
}
