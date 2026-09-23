package factory

import "strings"

// KelasFactory provides test payload generator methods for class session testing.
type KelasFactory struct{}

// Kelas is the global singleton instance of KelasFactory.
var Kelas = KelasFactory{}

// ValidPayload creates a valid payload map for creating a class session.
//
// Returns:
//   - map[string]interface{}: valid class creation payload.
func (KelasFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"kode_kelas": "001",
		"start_date": "2026-03-01",
		"end_date":   "2026-03-05",
		"lokasi":     "Fotang Pusat Maitreya",
		"pic":        "PIC " + RandString(5),
		"keterangan": "Kelas E2E Test",
	}
}

// InvalidPayload creates an invalid payload (kode_kelas exceeding max length) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid class payload.
func (KelasFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"kode_kelas": strings.Repeat("K", 10), // max=3
	}
}

// ValidUpdatePayload creates a valid payload map for updating a class session.
//
// Returns:
//   - map[string]interface{}: valid class update payload.
func (KelasFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Kelas Updated E2E",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid class update payload.
func (KelasFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"kode_kelas": strings.Repeat("k", 10), // max=3
	}
}

// KelasPesertaFactory provides test payload generator methods for class attendee testing.
type KelasPesertaFactory struct{}

// KelasPeserta is the global singleton instance of KelasPesertaFactory.
var KelasPeserta = KelasPesertaFactory{}

// ValidPayload creates a valid payload map for enrolling a participant into a class session.
//
// Parameters:
//   - trxId: class transaction ID.
//   - idPeserta: congregant (Umat) ID.
//
// Returns:
//   - map[string]interface{}: valid attendee enrollment payload.
func (KelasPesertaFactory) ValidPayload(trxId int32, idPeserta int32) map[string]interface{} {
	sumb := 250000.0
	return map[string]interface{}{
		"trx_id":     trxId,
		"id_peserta": idPeserta,
		"sumbangan":  sumb,
		"barang":     "Buku Dharma",
		"keterangan": "Peserta E2E Test",
	}
}

// InvalidPayload creates an invalid payload (missing id_peserta) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid attendee payload.
func (KelasPesertaFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"id_peserta": nil, // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a participant record.
//
// Returns:
//   - map[string]interface{}: valid attendee update payload.
func (KelasPesertaFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Peserta Updated E2E",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max description length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid attendee update payload.
func (KelasPesertaFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"barang": strings.Repeat("b", 120), // max=100
	}
}

// KelasPengabdiFactory provides test payload generator methods for class server/volunteer testing.
type KelasPengabdiFactory struct{}

// KelasPengabdi is the global singleton instance of KelasPengabdiFactory.
var KelasPengabdi = KelasPengabdiFactory{}

// ValidPayload creates a valid payload map for assigning a volunteer/server to a class event.
//
// Parameters:
//   - trxId: class transaction ID.
//   - idPengabdi: congregant (Umat) ID serving as volunteer.
//
// Returns:
//   - map[string]interface{}: valid server assignment payload.
func (KelasPengabdiFactory) ValidPayload(trxId int32, idPengabdi int32) map[string]interface{} {
	sumb := 150000.0
	return map[string]interface{}{
		"trx_id":      trxId,
		"id_pengabdi": idPengabdi,
		"sumbangan":   sumb,
		"keterangan":  "Pengabdi E2E Test",
	}
}

// InvalidPayload creates an invalid payload (missing id_pengabdi) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid server assignment payload.
func (KelasPengabdiFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"id_pengabdi": nil, // required
	}
}

// KelasTopikFactory provides test payload generator methods for class curriculum session testing.
type KelasTopikFactory struct{}

// KelasTopik is the global singleton instance of KelasTopikFactory.
var KelasTopik = KelasTopikFactory{}

// ValidPayload creates a valid payload map for attaching a lecture topic to a class session.
//
// Parameters:
//   - trxId: class transaction ID.
//   - kodeTopik: lecture topic code.
//
// Returns:
//   - map[string]interface{}: valid class topic payload.
func (KelasTopikFactory) ValidPayload(trxId int32, kodeTopik string) map[string]interface{} {
	return map[string]interface{}{
		"trx_id":     trxId,
		"kode_topik": kodeTopik,
		"keterangan": "Topik E2E Test",
	}
}

// InvalidPayload creates an invalid payload (kode_topik exceeding max length) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid class topic payload.
func (KelasTopikFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"kode_topik": strings.Repeat("T", 30), // max=20
	}
}

// DonasiSxyFactory provides test payload generator methods for SXY charitable transaction testing.
type DonasiSxyFactory struct{}

// DonasiSxy is the global singleton instance of DonasiSxyFactory.
var DonasiSxy = DonasiSxyFactory{}

// ValidPayload creates a valid payload map for recording a charitable donation transaction.
//
// Returns:
//   - map[string]interface{}: valid donation creation payload.
func (DonasiSxyFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no_kwitansi": RandCode("KW", 6),
		"tanggal":     "2026-03-01",
		"jumlah":      500000.0,
		"keterangan":  "Donasi E2E Test",
	}
}

// InvalidPayload creates an invalid payload (empty no_kwitansi) to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid donation payload.
func (DonasiSxyFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no_kwitansi": "", // required
	}
}

// ValidUpdatePayload creates a valid payload map for updating a charitable donation record.
//
// Returns:
//   - map[string]interface{}: valid donation update payload.
func (DonasiSxyFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Donasi Updated E2E",
	}
}

// InvalidUpdatePayload creates an invalid update payload exceeding max length to test validation failure.
//
// Returns:
//   - map[string]interface{}: invalid donation update payload.
func (DonasiSxyFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"no_kwitansi": strings.Repeat("k", 60), // max=50
	}
}
