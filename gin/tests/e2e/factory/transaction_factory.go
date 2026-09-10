package factory

import "strings"

type KelasFactory struct{}

var Kelas = KelasFactory{}

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

func (KelasFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"kode_kelas": strings.Repeat("K", 10), // max=3
	}
}

func (KelasFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Kelas Updated E2E",
	}
}

func (KelasFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"kode_kelas": strings.Repeat("k", 10), // max=3
	}
}

type KelasPesertaFactory struct{}

var KelasPeserta = KelasPesertaFactory{}

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

func (KelasPesertaFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"id_peserta": nil, // required
	}
}

func (KelasPesertaFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Peserta Updated E2E",
	}
}

func (KelasPesertaFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"barang": strings.Repeat("b", 120), // max=100
	}
}

type KelasPengabdiFactory struct{}

var KelasPengabdi = KelasPengabdiFactory{}

func (KelasPengabdiFactory) ValidPayload(trxId int32, idPengabdi int32) map[string]interface{} {
	sumb := 150000.0
	return map[string]interface{}{
		"trx_id":      trxId,
		"id_pengabdi": idPengabdi,
		"sumbangan":   sumb,
		"keterangan":  "Pengabdi E2E Test",
	}
}

func (KelasPengabdiFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"id_pengabdi": nil, // required
	}
}

type KelasTopikFactory struct{}

var KelasTopik = KelasTopikFactory{}

func (KelasTopikFactory) ValidPayload(trxId int32, kodeTopik string) map[string]interface{} {
	return map[string]interface{}{
		"trx_id":     trxId,
		"kode_topik": kodeTopik,
		"keterangan": "Topik E2E Test",
	}
}

func (KelasTopikFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"kode_topik": strings.Repeat("T", 30), // max=20
	}
}

type DonasiSxyFactory struct{}

var DonasiSxy = DonasiSxyFactory{}

func (DonasiSxyFactory) ValidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no_kwitansi": RandCode("KW", 6),
		"tanggal":     "2026-03-01",
		"jumlah":      500000.0,
		"keterangan":  "Donasi E2E Test",
	}
}

func (DonasiSxyFactory) InvalidPayload() map[string]interface{} {
	return map[string]interface{}{
		"no_kwitansi": "", // required
	}
}

func (DonasiSxyFactory) ValidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"keterangan": "Donasi Updated E2E",
	}
}

func (DonasiSxyFactory) InvalidUpdatePayload() map[string]interface{} {
	return map[string]interface{}{
		"no_kwitansi": strings.Repeat("k", 60), // max=50
	}
}
