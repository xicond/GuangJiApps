import json
import copy

def build_collection():
    """Build and enrich Postman collection with UAT tests and mock sample responses.

    Reads base collection from apps-gin.postman_collection.json.bak, processes each endpoint
    with assertion tests and sample responses, and writes the output back to
    apps-gin.postman_collection.json.
    """
    with open('/Users/xicond/Workspace/www/GuangJiApps/gin/postman/apps-gin.postman_collection.json.bak', 'r') as f:
        col = json.load(f)

    # Mock Data Dictionary for Entities
    MOCK = {
        "admin": {
            "resource": "Admin",
            "item": {
                "id": 1,
                "username": "admin",
                "name": "Administrator",
                "email": "admin@guangji.org",
                "phone_number": "081234567890",
                "group_id": 1,
                "department_id": 1,
                "status": 1
            },
            "created": {
                "id": 2,
                "username": "admin_user",
                "name": "Admin Baru",
                "email": "admin.baru@guangji.org",
                "phone_number": "081234567891",
                "group_id": 1,
                "department_id": 1,
                "status": 1
            },
            "updated": {
                "id": 1,
                "username": "admin",
                "name": "Administrator Updated",
                "email": "admin.updated@guangji.org",
                "phone_number": "081234567890",
                "group_id": 1,
                "department_id": 1,
                "status": 1
            },
            "err_create": {
                "error": "Validation failed: username wajib diisi",
                "details": {"username": ["username wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: format email tidak valid",
                "details": {"email": ["format email tidak valid"]}
            },
            "err_delete": {
                "error": "failed to delete: admin tidak ditemukan atau sedang aktif digunakan",
                "details": {"general": ["failed to delete: admin tidak ditemukan atau sedang aktif digunakan"]}
            }
        },
        "department": {
            "resource": "Department",
            "items": [
                {"id": 1, "code": "SEKR", "name": "Sekretariat"},
                {"id": 2, "code": "DHRM", "name": "Dharma"},
                {"id": 3, "code": "LOG", "name": "Logistik & Konsumsi"}
            ]
        },
        "admin_group": {
            "resource": "Admin Group",
            "item": {
                "group_id": 1,
                "group_name": "Super Admin",
                "description": "Full Access Group Administrator",
                "status": 1
            },
            "created": {
                "group_id": 2,
                "group_name": "Editor",
                "description": "Editor Role Group",
                "status": 1
            },
            "updated": {
                "group_id": 1,
                "group_name": "Super Admin Updated",
                "description": "Updated Administrator Group Description",
                "status": 1
            },
            "err_create": {
                "error": "Validation failed: group_name wajib diisi",
                "details": {"group_name": ["group_name wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: group_id tidak valid",
                "details": {"group_id": ["group_id tidak valid"]}
            },
            "err_delete": {
                "error": "failed to delete: admin group masih memiliki relasi user aktif",
                "details": {"general": ["admin group masih memiliki relasi user aktif"]}
            }
        },
        "group_menu": {
            "resource": "Group Menu Mapping",
            "item": {
                "id": 1,
                "group_id": 1,
                "menu_id": 1,
                "can_read": True,
                "can_write": True,
                "can_delete": True
            },
            "created": {
                "id": 2,
                "group_id": 2,
                "menu_id": 1,
                "can_read": True,
                "can_write": False,
                "can_delete": False
            },
            "updated": {
                "id": 1,
                "group_id": 1,
                "menu_id": 1,
                "can_read": True,
                "can_write": True,
                "can_delete": False
            },
            "err_create": {
                "error": "Validation failed: group_id dan menu_id wajib diisi",
                "details": {"group_id": ["group_id wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: menu_id tidak ditemukan",
                "details": {"menu_id": ["menu_id tidak ditemukan"]}
            },
            "err_delete": {
                "error": "failed to delete: group menu mapping tidak ditemukan",
                "details": {"general": ["group menu mapping tidak ditemukan"]}
            }
        },
        "admin_sub_warehouse": {
            "resource": "Admin Sub Warehouse",
            "item": {
                "id": 1,
                "admin_id": 1,
                "warehouse_id": 1,
                "sub_warehouse_name": "Gudang Utama",
                "status": 1
            },
            "created": {
                "id": 2,
                "admin_id": 1,
                "warehouse_id": 2,
                "sub_warehouse_name": "Gudang Cabang",
                "status": 1
            },
            "updated": {
                "id": 1,
                "admin_id": 1,
                "warehouse_id": 1,
                "sub_warehouse_name": "Gudang Utama Renovasi",
                "status": 1
            },
            "err_create": {
                "error": "Validation failed: warehouse_id wajib diisi",
                "details": {"warehouse_id": ["warehouse_id wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: sub_warehouse_name wajib diisi",
                "details": {"sub_warehouse_name": ["sub_warehouse_name wajib diisi"]}
            },
            "err_delete": {
                "error": "failed to delete: admin sub warehouse tidak ditemukan",
                "details": {"general": ["admin sub warehouse tidak ditemukan"]}
            }
        },
        "umat": {
            "resource": "Umat",
            "item": {
                "id": 1,
                "kode_umat": "UM001",
                "namaindonesia": "Budi Santoso",
                "namamandarin": "武帝",
                "namaalias": "Budi",
                "jenis_kelamin": "001",
                "tempat_lahir": "Jakarta",
                "tanggal_lahir": "1990-01-15",
                "alamat": "Jl. Gajah Mada No. 10",
                "no_ktp": "3171012345670001",
                "no_telp": "081234567890",
                "status_umat": "001",
                "foto_url": "https://storage.example.com/photos/um001.jpg"
            },
            "created": {
                "id": 2,
                "kode_umat": "UM002",
                "namaindonesia": "Siti Rahma",
                "namamandarin": "美兰",
                "namaalias": "Siti",
                "jenis_kelamin": "002",
                "tempat_lahir": "Surabaya",
                "tanggal_lahir": "1992-05-20",
                "alamat": "Jl. Basuki Rahmat No. 25",
                "no_ktp": "3578012345670002",
                "no_telp": "081298765432",
                "status_umat": "001",
                "foto_url": "https://storage.example.com/photos/um002.jpg"
            },
            "updated": {
                "id": 1,
                "kode_umat": "UM001",
                "namaindonesia": "Budi Santoso Updated",
                "namamandarin": "武帝",
                "namaalias": "Budi",
                "jenis_kelamin": "001",
                "tempat_lahir": "Jakarta",
                "tanggal_lahir": "1990-01-15",
                "alamat": "Jl. Hayam Wuruk No. 12",
                "no_ktp": "3171012345670001",
                "no_telp": "081234567899",
                "status_umat": "001",
                "foto_url": "https://storage.example.com/photos/um001.jpg"
            },
            "ocr": {
                "namaindonesia": "BUDI SANTOSO",
                "no_ktp": "3171012345670001",
                "tempat_lahir": "JAKARTA",
                "tanggal_lahir": "1990-01-15",
                "jenis_kelamin": "001",
                "alamat": "JL GAJAH MADA NO 10"
            },
            "report_item": {
                "id": 1,
                "kode_umat": "UM001",
                "namaindonesia": "Budi Santoso",
                "namamandarin": "武帝",
                "jenis_kelamin": "001",
                "tempat_lahir": "Jakarta",
                "tanggal_lahir": "1990-01-15",
                "status_umat": "001",
                "ikrar_ciu_tao": True,
                "tahun_ciu_tao": "2020",
                "fotang_ciu_tao": "Fotang Pusat"
            },
            "err_create": {
                "error": "Validation failed: namaindonesia wajib diisi",
                "details": {"namaindonesia": ["namaindonesia wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: jenis_kelamin tidak valid",
                "details": {"jenis_kelamin": ["jenis_kelamin tidak ditemukan di lookup"]}
            },
            "err_delete": {
                "error": "failed to delete: data umat masih memiliki relasi aktif di kelas",
                "details": {"general": ["data umat masih memiliki relasi aktif di kelas"]}
            },
            "err_ocr": {
                "error": "Validation failed: file gambar KTP wajib diunggah",
                "details": {"file": ["file gambar KTP wajib diunggah"]}
            }
        },
        "topic": {
            "resource": "Topic",
            "item": {
                "topic_code": "TP01",
                "topic_name": "Pengantar Dharma",
                "topic_category": "1",
                "description": "Materi dasar pembinaan dharma",
                "status": True
            },
            "created": {
                "topic_code": "TP02",
                "topic_name": "Etika dan Budi Pekerti",
                "topic_category": "1",
                "description": "Materi pembinaan moral dan kebajikan",
                "status": True
            },
            "updated": {
                "topic_code": "TP01",
                "topic_name": "Pengantar Dharma Lanjutan",
                "topic_category": "1",
                "description": "Materi dasar pembinaan yang disempurnakan",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: topic_code wajib diisi",
                "details": {"topic_code": ["topic_code wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: topic_category tidak valid",
                "details": {"topic_category": ["topic_category tidak valid"]}
            },
            "err_delete": {
                "error": "failed to delete: topic masih digunakan dalam kelas",
                "details": {"general": ["topic masih digunakan dalam kelas"]}
            }
        },
        "kelas_master": {
            "resource": "KelasMaster",
            "item": {
                "lookup_id": "B_KELASKHUSUS001",
                "lookup_name": "Kelas Dharma Dasar",
                "lookup_description": "Kelas Pendalaman Dharma Tingkat 1",
                "status": True
            },
            "created": {
                "lookup_id": "B_KELASKHUSUS002",
                "lookup_name": "Kelas Dharma Menengah",
                "lookup_description": "Kelas Pendalaman Dharma Tingkat 2",
                "status": True
            },
            "updated": {
                "lookup_id": "B_KELASKHUSUS001",
                "lookup_name": "Kelas Dharma Dasar Revisi",
                "lookup_description": "Kelas Pendalaman Dharma Tingkat 1 Terupdate",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: lookup_description wajib diisi",
                "details": {"lookup_description": ["lookup_description wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: lookup_id tidak ditemukan",
                "details": {"lookup_id": ["lookup_id tidak ditemukan"]}
            },
            "err_delete": {
                "error": "failed to delete: kelas master masih digunakan pada transaksi kelas",
                "details": {"general": ["kelas master masih digunakan pada transaksi kelas"]}
            }
        },
        "activity": {
            "resource": "Activity",
            "item": {
                "event_code": "ACT001",
                "event_name": "Bakti Sosial Tahunan",
                "event_category": "001",
                "description": "Kegiatan donor darah dan pembagian sembako",
                "status": True
            },
            "created": {
                "event_code": "ACT002",
                "event_name": "Peringatan Hari Suci",
                "event_category": "002",
                "description": "Perayaan dan doa bersama",
                "status": True
            },
            "updated": {
                "event_code": "ACT001",
                "event_name": "Bakti Sosial Tahunan 2026",
                "event_category": "001",
                "description": "Kegiatan donor darah dan sembako disempurnakan",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: event_code melebihi panjang maksimal 10 karakter",
                "details": {"event_code": ["event_code max:10"]}
            },
            "err_update": {
                "error": "Validation failed: event_name wajib diisi",
                "details": {"event_name": ["event_name wajib diisi"]}
            },
            "err_delete": {
                "error": "failed to delete: activity masih memiliki jadwal aktif",
                "details": {"general": ["activity masih memiliki jadwal aktif"]}
            }
        },
        "tim_kerja": {
            "resource": "Tim Kerja",
            "item": {
                "id": 1,
                "kode_posisi": "POS001",
                "nama_posisi": "Ketua Koordinator",
                "tingkat": 1,
                "parent_id": 0,
                "status": True
            },
            "created": {
                "id": 2,
                "kode_posisi": "POS002",
                "nama_posisi": "Wakil Koordinator",
                "tingkat": 2,
                "parent_id": 1,
                "status": True
            },
            "updated": {
                "id": 1,
                "kode_posisi": "POS001",
                "nama_posisi": "Ketua Koordinator Wilayah",
                "tingkat": 1,
                "parent_id": 0,
                "status": True
            },
            "lookup_items": [
                {"id": 1, "kode": "POS001", "nama": "Ketua Koordinator", "parent_id": 0},
                {"id": 2, "kode": "POS002", "nama": "Wakil Koordinator", "parent_id": 1}
            ],
            "err_create": {
                "error": "Validation failed: kode_posisi wajib diisi",
                "details": {"kode_posisi": ["kode_posisi wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: parent_id tidak valid",
                "details": {"parent_id": ["parent_id tidak valid"]}
            },
            "err_delete": {
                "error": "failed to delete: tim kerja memiliki sub posisi aktif",
                "details": {"general": ["tim kerja memiliki sub posisi aktif"]}
            }
        },
        "tahun_ciu_tao": {
            "resource": "Tahun Ciu Tao",
            "item": {
                "id": 1,
                "tahun": "2026",
                "keterangan": "Tahun Ciu Tao 2026",
                "status": True
            },
            "created": {
                "id": 2,
                "tahun": "2027",
                "keterangan": "Tahun Ciu Tao 2027",
                "status": True
            },
            "updated": {
                "id": 1,
                "tahun": "2026",
                "keterangan": "Tahun Ciu Tao 2026 Terverifikasi",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: tahun wajib diisi",
                "details": {"tahun": ["tahun wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: format tahun harus 4 digit angka",
                "details": {"tahun": ["format tahun harus 4 digit angka"]}
            },
            "err_delete": {
                "error": "failed to delete: tahun ciu tao sudah terikat pada data umat",
                "details": {"general": ["tahun ciu tao sudah terikat pada data umat"]}
            }
        },
        "penggalang_dana": {
            "resource": "Penggalang Dana",
            "item": {
                "id": 1,
                "kode": "PGD001",
                "nama": "Penggalang Dana Pusat",
                "no_telp": "081234567890",
                "status": True
            },
            "created": {
                "id": 2,
                "kode": "PGD002",
                "nama": "Penggalang Dana Wilayah Barat",
                "no_telp": "081234567891",
                "status": True
            },
            "updated": {
                "id": 1,
                "kode": "PGD001",
                "nama": "Penggalang Dana Pusat Updated",
                "no_telp": "081234567899",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: nama penggalang dana wajib diisi",
                "details": {"nama": ["nama wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: format nomor telepon tidak valid",
                "details": {"no_telp": ["format no_telp tidak valid"]}
            },
            "err_delete": {
                "error": "failed to delete: penggalang dana memiliki transaksi aktif",
                "details": {"general": ["penggalang dana memiliki transaksi aktif"]}
            }
        },
        "sxy_donatur": {
            "resource": "Sxy Donatur",
            "item": {
                "id": 1,
                "kode_donatur": "SDN001",
                "nama_donatur": "Yayasan Guang Ji",
                "no_telp": "081122334455",
                "status": True
            },
            "created": {
                "id": 2,
                "kode_donatur": "SDN002",
                "nama_donatur": "Donatur Mitra Mandiri",
                "no_telp": "081122334466",
                "status": True
            },
            "updated": {
                "id": 1,
                "kode_donatur": "SDN001",
                "nama_donatur": "Yayasan Guang Ji Indonesia",
                "no_telp": "081122334499",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: kode_donatur wajib diisi",
                "details": {"kode_donatur": ["kode_donatur wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: nama_donatur tidak boleh kosong",
                "details": {"nama_donatur": ["nama_donatur tidak boleh kosong"]}
            },
            "err_delete": {
                "error": "failed to delete: donatur memiliki riwayat transaksi donasi",
                "details": {"general": ["donatur memiliki riwayat transaksi donasi"]}
            }
        },
        "fotang": {
            "resource": "Fotang",
            "items": [
                {"id": 1, "kode": "FT01", "nama": "Fotang Pusat Maitreya", "alamat": "Jl. Gajah Mada No. 10"},
                {"id": 2, "kode": "FT02", "nama": "Fotang Cabang Harmoni", "alamat": "Jl. Hayam Wuruk No. 25"}
            ]
        },
        "kelas": {
            "resource": "Kelas",
            "item": {
                "id": 1,
                "kode_kelas": "KLS001",
                "nama_kelas": "Kelas Dharma Dasar Angkatan 1",
                "tgl_mulai": "2026-03-01",
                "tgl_selesai": "2026-03-05",
                "fotang_id": 1,
                "lokasi": "Fotang Pusat Maitreya",
                "status": 1
            },
            "created": {
                "id": 2,
                "kode_kelas": "KLS002",
                "nama_kelas": "Kelas Dharma Menengah Angkatan 1",
                "tgl_mulai": "2026-04-01",
                "tgl_selesai": "2026-04-05",
                "fotang_id": 1,
                "lokasi": "Fotang Pusat Maitreya",
                "status": 1
            },
            "updated": {
                "id": 1,
                "kode_kelas": "KLS001",
                "nama_kelas": "Kelas Dharma Dasar Angkatan 1 Revisi",
                "tgl_mulai": "2026-03-01",
                "tgl_selesai": "2026-03-06",
                "fotang_id": 1,
                "lokasi": "Fotang Pusat Maitreya",
                "status": 1
            },
            "err_create": {
                "error": "Validation failed: kode_kelas wajib diisi",
                "details": {"kode_kelas": ["kode_kelas wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: tanggal mulai tidak boleh melebihi tanggal selesai",
                "details": {"tgl_mulai": ["tanggal mulai tidak boleh melebihi tanggal selesai"]}
            },
            "err_delete": {
                "error": "failed to delete: kelas sudah memiliki peserta yang terdaftar",
                "details": {"general": ["kelas sudah memiliki peserta yang terdaftar"]}
            }
        },
        "kelas_peserta": {
            "resource": "KelasPeserta",
            "item": {
                "id": 1,
                "trx_id": 1,
                "id_peserta": 101,
                "nama_peserta": "Budi Santoso",
                "sumbangan": 500000,
                "barang": "Buku Dharma",
                "keterangan": "Peserta Aktif",
                "status": True
            },
            "created": {
                "id": 2,
                "trx_id": 1,
                "id_peserta": 102,
                "nama_peserta": "Siti Rahma",
                "sumbangan": 300000,
                "barang": "",
                "keterangan": "Peserta Baru",
                "status": True
            },
            "updated": {
                "id": 1,
                "trx_id": 1,
                "id_peserta": 101,
                "nama_peserta": "Budi Santoso",
                "sumbangan": 750000,
                "barang": "Buku Dharma & Seragam",
                "keterangan": "Peserta Aktif (Updated)",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: id_peserta wajib diisi dan harus valid",
                "details": {"id_peserta": ["id_peserta wajib diisi"]}
            },
            "err_update": {
                "error": "Validation failed: sumbangan tidak boleh negatif",
                "details": {"sumbangan": ["sumbangan must be greater than or equal to 0"]}
            },
            "err_delete": {
                "error": "failed to delete: peserta tidak terdaftar di kelas",
                "details": {"general": ["peserta tidak terdaftar di kelas"]}
            }
        },
        "kelas_pengabdi": {
            "resource": "KelasPengabdi",
            "item": {
                "id": 1,
                "trx_id": 1,
                "id_pengabdi": 201,
                "nama_pengabdi": "Hendra Wijaya",
                "peran": "Koordinator Lapangan",
                "sumbangan": 250000,
                "keterangan": "Pengabdi Aktif",
                "status": True
            },
            "created": {
                "id": 2,
                "trx_id": 1,
                "id_pengabdi": 202,
                "nama_pengabdi": "Dewi Lestari",
                "peran": "Koordinator Konsumsi",
                "sumbangan": 200000,
                "keterangan": "Pengabdi Dapur",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: id_pengabdi wajib diisi",
                "details": {"id_pengabdi": ["id_pengabdi wajib diisi"]}
            }
        },
        "kelas_topik": {
            "resource": "KelasTopik",
            "item": {
                "id": 1,
                "trx_id": 1,
                "kode_topik": "TP001",
                "nama_topik": "Pengantar Dharma",
                "pembicara": "Pandita Tan",
                "keterangan": "Sesi Pembukaan",
                "status": True
            },
            "created": {
                "id": 2,
                "trx_id": 1,
                "kode_topik": "TP002",
                "nama_topik": "Etika dan Budi Pekerti",
                "pembicara": "Pandita Lim",
                "keterangan": "Sesi Lanjutan",
                "status": True
            },
            "err_create": {
                "error": "Validation failed: kode_topik wajib diisi",
                "details": {"kode_topik": ["kode_topik wajib diisi"]}
            }
        },
        "kelas_sub": {
            "kendaraan": {
                "resource": "KelasKendaraan",
                "items": [{"id": 1, "trx_id": 1, "jenis_kendaraan": "Bus Pariwisata", "plat_nomor": "B 1234 CD", "kapasitas": 45, "status": True}]
            },
            "donasi": {
                "resource": "KelasDonasi",
                "items": [{"id": 1, "trx_id": 1, "id_donatur": 101, "nama_donatur": "Budi Santoso", "nominal": 1000000, "keterangan": "Donasi Buku", "status": True}]
            },
            "donasi_barang": {
                "resource": "KelasDonasiBarang",
                "items": [{"id": 1, "trx_id": 1, "id_donatur": 102, "nama_donatur": "Siti Rahma", "nama_barang": "Beras 50kg", "jumlah": 5, "status": True}]
            },
            "pengeluaran": {
                "resource": "KelasPengeluaran",
                "items": [{"id": 1, "trx_id": 1, "keterangan": "Konsumsi Peserta Hari 1", "nominal": 750000, "tgl_pengeluaran": "2026-03-01", "status": True}]
            },
            "musik": {
                "resource": "KelasMusik",
                "items": [{"id": 1, "trx_id": 1, "judul_lagu": "Dharma Gita Persaudaraan", "pemimpin_lagu": "Sdr. Andi", "status": True}]
            },
            "absensi": {
                "resource": "KelasAbsensi",
                "items": [{"id": 1, "trx_id": 1, "id_peserta": 101, "sesi": 1, "hadir": True, "tgl_absensi": "2026-03-01 08:30:00"}]
            }
        },
        "donasi_sxy": {
            "resource": "Donasi Sxy",
            "item": {
                "id": 1,
                "kode_donasi": "DSX001",
                "nama_donatur": "Yayasan Guang Ji",
                "nominal": 10000000,
                "tgl_donasi": "2026-03-01",
                "keterangan": "Donasi Operasional",
                "status": True
            },
            "created": {
                "id": 2,
                "kode_donasi": "DSX002",
                "nama_donatur": "Donatur Mitra Mandiri",
                "nominal": 5000000,
                "tgl_donasi": "2026-03-02",
                "keterangan": "Donasi Sarana",
                "status": True
            },
            "updated": {
                "id": 1,
                "kode_donasi": "DSX001",
                "nama_donatur": "Yayasan Guang Ji Indonesia",
                "nominal": 12000000,
                "tgl_donasi": "2026-03-01",
                "keterangan": "Donasi Operasional Disesuaikan",
                "status": True
            },
            "report_item": {
                "id": 1,
                "kode_donasi": "DSX001",
                "nama_donatur": "Yayasan Guang Ji",
                "nominal": 10000000,
                "tgl_donasi": "2026-03-01",
                "keterangan": "Donasi Operasional",
                "petugas": "Administrator"
            },
            "err_create": {
                "error": "Validation failed: nominal donasi harus lebih besar dari 0",
                "details": {"nominal": ["nominal must be greater than 0"]}
            },
            "err_update": {
                "error": "Validation failed: id donasi tidak ditemukan",
                "details": {"id": ["id donasi tidak ditemukan"]}
            },
            "err_delete": {
                "error": "failed to delete: donasi sudah dibukukan",
                "details": {"general": ["donasi sudah dibukukan"]}
            }
        }
    }

    # Process all items in collection
    total_processed = 0

    def process_items(items, folder_name=""):
        """Recursively traverse collection items and configure each endpoint request.

        Args:
            items (list[dict]): List of Postman item structures or subfolders.
            folder_name (str, optional): Name of the parent folder. Defaults to "".
        """
        nonlocal total_processed
        for item in items:
            if 'item' in item:
                process_items(item['item'], item.get('name', ''))
            else:
                name = item.get('name', '')
                req = item.get('request', {})
                method = req.get('method', 'GET')
                total_processed += 1
                
                # Configure specific item
                configure_item(item, folder_name, name, method, req, MOCK)

    process_items(col.get('item', []))
    print(f"Total processed endpoints: {total_processed}")

    # Write updated collection
    with open('/Users/xicond/Workspace/www/GuangJiApps/gin/postman/apps-gin.postman_collection.json', 'w') as f:
        json.dump(col, f, indent=2)
    print("apps-gin.postman_collection.json successfully updated!")

def make_test(lines):
    """Generate a Postman test event script structure from a list of test assertions.

    Args:
        lines (list[str]): List of JavaScript assertion and execution code strings.

    Returns:
        list[dict]: Postman collection v2.1.0 test event structure.
    """
    return [
        {
            "listen": "test",
            "script": {
                "exec": lines,
                "type": "text/javascript"
            }
        }
    ]

def make_sample_response(name, original_req, code, status_text, body_data, is_binary=False, filename=""):
    """Construct a Postman sample response structure for documentation and mocking.

    Args:
        name (str): Label for the saved example response.
        original_req (dict): The original Postman request dictionary.
        code (int): HTTP status code (e.g., 200, 201, 400).
        status_text (str): HTTP status string (e.g., "OK", "Created", "Bad Request").
        body_data (dict): JSON response body data.
        is_binary (bool, optional): Whether response is a binary file (e.g. Excel). Defaults to False.
        filename (str, optional): Target filename for attachments. Defaults to "".

    Returns:
        dict: Postman response structure for the example.
    """
    headers = []
    if is_binary:
        headers = [
            {"key": "Content-Type", "value": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
            {"key": "Content-Disposition", "value": f"attachment; filename=\"{filename or 'report.xlsx'}\""}
        ]
        body_str = ""
    else:
        headers = [
            {"key": "Content-Type", "value": "application/json; charset=utf-8"}
        ]
        body_str = json.dumps(body_data, indent=2)

    return {
        "name": name,
        "originalRequest": copy.deepcopy(original_req),
        "status": status_text,
        "code": code,
        "_postman_previewlanguage": "json" if not is_binary else "raw",
        "header": headers,
        "cookie": [],
        "body": body_str
    }

def configure_item(item, folder, name, method, req, MOCK):
    """Configure tests and sample response examples for a single Postman collection item.

    Args:
        item (dict): The Postman request item object to mutate.
        folder (str): Name of the parent folder.
        name (str): Endpoint request name.
        method (str): HTTP method (GET, POST, PATCH, DELETE, etc.).
        req (dict): Postman request dictionary.
        MOCK (dict): Mock payload dictionary for resources.
    """
    # Default containers
    responses = []
    tests = []

    # 1. PUBLIC FOLDER
    if folder == "Public":
        if name == "Ping":
            tests = [
                'pm.test("UAT - Status code is 200 OK", function () {',
                '    pm.response.to.have.status(200);',
                '});',
                '',
                'pm.test("UAT - Response time is acceptable (< 2000ms)", function () {',
                '    pm.expect(pm.response.responseTime).to.be.below(2000);',
                '});',
                '',
                'pm.test("UAT - Ping responds with pong message", function () {',
                '    var jsonData = pm.response.json();',
                '    pm.expect(jsonData).to.be.an("object");',
                '    pm.expect(jsonData.message).to.eql("pong");',
                '});'
            ]
            responses = [
                make_sample_response("200 OK", req, 200, "OK", {"message": "pong"})
            ]
        elif name == "Login":
            # Update request body to admin / password
            req["body"] = {
                "mode": "raw",
                "raw": json.dumps({"username": "admin", "password": "password"}, indent=2),
                "options": {"raw": {"language": "json"}}
            }
            tests = [
                'pm.test("UAT - Status code is 200 OK", function () {',
                '    pm.response.to.have.status(200);',
                '});',
                '',
                'pm.test("UAT - Response time is acceptable (< 2000ms)", function () {',
                '    pm.expect(pm.response.responseTime).to.be.below(2000);',
                '});',
                '',
                'pm.test("UAT - Login successful and token received", function () {',
                '    var jsonData = pm.response.json();',
                '    pm.expect(jsonData).to.have.property("token");',
                '    pm.expect(jsonData.token).to.be.a("string").and.not.empty;',
                '    pm.expect(jsonData).to.have.property("message");',
                '    pm.expect(jsonData.message).to.eql("login successful");',
                '    pm.expect(jsonData).to.have.property("user");',
                '    pm.collectionVariables.set("authToken", jsonData.token);',
                '});'
            ]
            success_body = {
                "message": "login successful",
                "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjEsImV4cCI6MTgwMDAwMDAwMH0.signature",
                "user": {
                    "id": 1,
                    "username": "admin",
                    "name": "Administrator",
                    "group_id": 1,
                    "status": 1
                },
                "main_menu": [
                    {"id": 1, "menu_name": "Master Data", "url": "/master"},
                    {"id": 2, "menu_name": "Transactions", "url": "/transactions"}
                ]
            }
            err_body = {
                "error": "Validation failed: Username is required",
                "details": {"username": ["username wajib diisi"]}
            }
            responses = [
                make_sample_response("200 OK - Login Success", req, 200, "OK", success_body),
                make_sample_response("400 Bad Request - Missing Username/Password", req, 400, "Bad Request", err_body)
            ]
        elif name == "Change Password":
            tests = [
                'pm.test("UAT - Status code is 200 OK", function () {',
                '    pm.response.to.have.status(200);',
                '});',
                '',
                'pm.test("UAT - Response time is acceptable (< 2000ms)", function () {',
                '    pm.expect(pm.response.responseTime).to.be.below(2000);',
                '});',
                '',
                'pm.test("UAT - Confirmation message confirms password changed", function () {',
                '    var jsonData = pm.response.json();',
                '    pm.expect(jsonData).to.have.property("message");',
                '    pm.expect(jsonData.message).to.include("password");',
                '});'
            ]
            success_body = {"message": "password changed successfully"}
            err_body = {
                "error": "Validation failed: password lama tidak sesuai",
                "details": {"old_password": ["password lama tidak sesuai"]}
            }
            responses = [
                make_sample_response("200 OK - Password Updated", req, 200, "OK", success_body),
                make_sample_response("400 Bad Request - Validation Error", req, 400, "Bad Request", err_body)
            ]

    # 2. ADMIN MANAGEMENT FOLDER
    elif folder == "Admin Management":
        if "Department" in name:
            tests = make_get_list_test("Department")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", {"resource": "Department", "data": MOCK["department"]["items"]})
            ]
        elif "Admin Groups" in name:
            res_key = "admin_group"
            res_label = "Admin Group"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Group Menu Mappings" in name:
            res_key = "group_menu"
            res_label = "Group Menu Mapping"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Admin Sub Warehouses" in name:
            res_key = "admin_sub_warehouse"
            res_label = "Admin Sub Warehouse"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Admins" in name:
            res_key = "admin"
            res_label = "Admin"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)

    # 3. MASTER DATA FOLDER
    elif folder == "Master Data":
        if name == "Umat - Report":
            tests = make_get_list_test("Umat Report")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", {
                    "resource": "UmatReport",
                    "data": [MOCK["umat"]["report_item"]],
                    "meta": {"page": 1, "limit": 10, "total": 1}
                })
            ]
        elif name == "Umat - Report Excel":
            tests = make_get_excel_test("Umat Report Excel")
            responses = [
                make_sample_response("200 OK - Excel Download", req, 200, "OK", None, is_binary=True, filename="umat_report.xlsx")
            ]
        elif name == "Umat - OCR Recognition":
            tests = [
                'pm.test("UAT - Status code is 200 OK", function () {',
                '    pm.response.to.have.status(200);',
                '});',
                '',
                'pm.test("UAT - Response time is acceptable (< 3000ms)", function () {',
                '    pm.expect(pm.response.responseTime).to.be.below(3000);',
                '});',
                '',
                'pm.test("UAT - OCR returns parsed KTP fields", function () {',
                '    var jsonData = pm.response.json();',
                '    pm.expect(jsonData).to.have.property("data");',
                '    pm.expect(jsonData.data).to.have.property("namaindonesia");',
                '    pm.expect(jsonData.data).to.have.property("no_ktp");',
                '});'
            ]
            responses = [
                make_sample_response("200 OK - OCR Parsed", req, 200, "OK", {"resource": "Umat", "data": MOCK["umat"]["ocr"]}),
                make_sample_response("400 Bad Request - Missing Image", req, 400, "Bad Request", MOCK["umat"]["err_ocr"])
            ]
        elif "Umat" in name:
            res_key = "umat"
            res_label = "Umat"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Topics" in name:
            res_key = "topic"
            res_label = "Topic"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Kelas Master" in name:
            res_key = "kelas_master"
            res_label = "KelasMaster"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Activity" in name or "Activities" in name:
            res_key = "activity"
            res_label = "Activity"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Tim Kerja" in name:
            if "Lookup" in name:
                tests = make_get_lookup_test("Tim Kerja")
                responses = [
                    make_sample_response("200 OK", req, 200, "OK", {"resource": "Tim Kerja", "data": MOCK["tim_kerja"]["lookup_items"]})
                ]
            else:
                res_key = "tim_kerja"
                res_label = "Tim Kerja"
                handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Tahun Ciu Tao" in name:
            res_key = "tahun_ciu_tao"
            res_label = "Tahun Ciu Tao"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Penggalang Dana" in name:
            res_key = "penggalang_dana"
            res_label = "Penggalang Dana"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Sxy Donatur" in name:
            res_key = "sxy_donatur"
            res_label = "Sxy Donatur"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Fotang" in name:
            tests = make_get_lookup_test("Fotang")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", {"resource": "Fotang", "data": MOCK["fotang"]["items"]})
            ]

    # 4. LOOKUP SERVICE FOLDER
    elif folder == "Lookup Service":
        lookup_title = name.replace("Lookup - ", "")
        tests = make_get_lookup_test(lookup_title)
        responses = [
            make_sample_response("200 OK", req, 200, "OK", {
                "resource": "AppLookup",
                "data": [
                    {"lookup_id": "001", "lookup_name": f"{lookup_title} 1", "lookup_value": "001", "status": True},
                    {"lookup_id": "002", "lookup_name": f"{lookup_title} 2", "lookup_value": "002", "status": True}
                ]
            })
        ]

    # 5. TRANSACTIONS FOLDER
    elif folder == "Transactions":
        if name == "Report - SXY List":
            tests = make_get_list_test("SXY Report")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", {
                    "resource": "DonasiSxyReport",
                    "data": [MOCK["donasi_sxy"]["report_item"]],
                    "meta": {"page": 1, "limit": 10, "total": 1}
                })
            ]
        elif name == "Report - SXY Excel":
            tests = make_get_excel_test("SXY Report Excel")
            responses = [
                make_sample_response("200 OK - Excel Download", req, 200, "OK", None, is_binary=True, filename="sxy_donasi_report.xlsx")
            ]
        elif "Donasi Sxy" in name:
            res_key = "donasi_sxy"
            res_label = "Donasi Sxy"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif name == "Kelas - Lookup":
            tests = make_get_lookup_test("Kelas")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", {
                    "resource": "Kelas",
                    "data": [{"id": 1, "kode_kelas": "KLS001", "nama_kelas": "Kelas Dharma Dasar Angkatan 1"}]
                })
            ]
        elif name == "Kelas - Report":
            tests = make_get_list_test("Kelas Report")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", {
                    "resource": "KelasReport",
                    "data": [MOCK["kelas"]["item"]],
                    "meta": {"page": 1, "limit": 10, "total": 1}
                })
            ]
        elif name == "Kelas - Peserta List":
            tests = make_get_list_test("Kelas Peserta")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", {
                    "resource": "KelasPeserta",
                    "data": [MOCK["kelas_peserta"]["item"]],
                    "meta": {"page": 1, "limit": 10, "total": 1}
                })
            ]
        elif "Kelas Peserta" in name:
            res_key = "kelas_peserta"
            res_label = "KelasPeserta"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)
        elif "Kelas Pengabdi" in name:
            if method == "GET":
                tests = make_get_list_test("Kelas Pengabdi")
                responses = [
                    make_sample_response("200 OK", req, 200, "OK", {
                        "resource": "KelasPengabdi",
                        "data": [MOCK["kelas_pengabdi"]["item"]]
                    })
                ]
            else:
                tests = make_post_create_test("Kelas Pengabdi")
                responses = [
                    make_sample_response("201 Created", req, 201, "Created", {
                        "resource": "KelasPengabdi",
                        "data": MOCK["kelas_pengabdi"]["created"]
                    }),
                    make_sample_response("400 Bad Request - Validation Error", req, 400, "Bad Request", MOCK["kelas_pengabdi"]["err_create"])
                ]
        elif "Kelas Topik" in name:
            if method == "GET":
                tests = make_get_list_test("Kelas Topik")
                responses = [
                    make_sample_response("200 OK", req, 200, "OK", {
                        "resource": "KelasTopik",
                        "data": [MOCK["kelas_topik"]["item"]]
                    })
                ]
            else:
                tests = make_post_create_test("Kelas Topik")
                responses = [
                    make_sample_response("201 Created", req, 201, "Created", {
                        "resource": "KelasTopik",
                        "data": MOCK["kelas_topik"]["created"]
                    }),
                    make_sample_response("400 Bad Request - Validation Error", req, 400, "Bad Request", MOCK["kelas_topik"]["err_create"])
                ]
        elif "Kelas Kendaraan" in name:
            tests = make_get_list_test("Kelas Kendaraan")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", MOCK["kelas_sub"]["kendaraan"])
            ]
        elif "Kelas Donasi Barang" in name:
            tests = make_get_list_test("Kelas Donasi Barang")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", MOCK["kelas_sub"]["donasi_barang"])
            ]
        elif "Kelas Donasi" in name:
            tests = make_get_list_test("Kelas Donasi")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", MOCK["kelas_sub"]["donasi"])
            ]
        elif "Kelas Pengeluaran" in name:
            tests = make_get_list_test("Kelas Pengeluaran")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", MOCK["kelas_sub"]["pengeluaran"])
            ]
        elif "Kelas Musik" in name:
            tests = make_get_list_test("Kelas Musik")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", MOCK["kelas_sub"]["musik"])
            ]
        elif "Kelas Absensi" in name:
            tests = make_get_list_test("Kelas Absensi")
            responses = [
                make_sample_response("200 OK", req, 200, "OK", MOCK["kelas_sub"]["absensi"])
            ]
        elif "Kelas" in name:
            res_key = "kelas"
            res_label = "Kelas"
            handle_crud(name, method, req, res_key, res_label, MOCK, tests, responses)

    # Attach tests and responses to item
    if tests:
        item["event"] = make_test(tests)
    if responses:
        item["response"] = responses

def handle_crud(name, method, req, res_key, res_label, MOCK, tests_out, responses_out):
    """Configure standard CRUD tests and mock responses for RESTful resources.

    Args:
        name (str): Request name.
        method (str): HTTP method.
        req (dict): Postman request dictionary.
        res_key (str): Lookup key in MOCK dictionary.
        res_label (str): Human-readable resource name for test assertions.
        MOCK (dict): Mock payload dictionary.
        tests_out (list): List to append test assertion blocks to.
        responses_out (list): List to append mock response objects to.
    """
    data_dict = MOCK[res_key]
    resource_name = data_dict["resource"]

    if method == "GET":
        if "List" in name or "list" in name:
            t = make_get_list_test(res_label)
            tests_out.extend(t)
            body = {
                "resource": resource_name,
                "data": [data_dict["item"]],
                "meta": {"page": 1, "limit": 10, "total": 1}
            }
            responses_out.append(make_sample_response("200 OK - List", req, 200, "OK", body))
        else: # Get By ID
            t = make_get_detail_test(res_label)
            tests_out.extend(t)
            body = {
                "resource": resource_name,
                "data": data_dict["item"]
            }
            responses_out.append(make_sample_response("200 OK - Detail", req, 200, "OK", body))

    elif method == "POST":
        t = make_post_create_test(res_label)
        tests_out.extend(t)
        success_body = {
            "resource": resource_name,
            "data": data_dict["created"]
        }
        err_body = data_dict["err_create"]
        responses_out.append(make_sample_response("201 Created - Success", req, 201, "Created", success_body))
        responses_out.append(make_sample_response("400 Bad Request - Validation Error", req, 400, "Bad Request", err_body))

    elif method == "PATCH":
        t = make_patch_update_test(res_label)
        tests_out.extend(t)
        success_body = {
            "resource": resource_name,
            "data": data_dict["updated"]
        }
        err_body = data_dict["err_update"]
        responses_out.append(make_sample_response("200 OK - Updated", req, 200, "OK", success_body))
        responses_out.append(make_sample_response("400 Bad Request - Validation Error", req, 400, "Bad Request", err_body))

    elif method == "DELETE":
        t = make_delete_test(res_label)
        tests_out.extend(t)
        success_body = {"message": "deleted"}
        err_body = data_dict["err_delete"]
        responses_out.append(make_sample_response("200 OK - Deleted", req, 200, "OK", success_body))
        responses_out.append(make_sample_response("400 Bad Request - Error", req, 400, "Bad Request", err_body))

def make_get_list_test(resource_name):
    """Generate Postman test script assertions for paginated list endpoints.

    Args:
        resource_name (str): Human-readable resource name.

    Returns:
        list[str]: JavaScript test assertion lines.
    """
    return [
        'pm.test("UAT - Status code is 200 OK", function () {',
        '    pm.response.to.have.status(200);',
        '});',
        '',
        'pm.test("UAT - Response time is acceptable (< 2000ms)", function () {',
        '    pm.expect(pm.response.responseTime).to.be.below(2000);',
        '});',
        '',
        'pm.test("UAT - Response format is valid JSON", function () {',
        '    pm.response.to.be.json;',
        '});',
        '',
        f'pm.test("UAT - Returns list of {resource_name} with pagination meta", function () {{',
        '    var jsonData = pm.response.json();',
        '    pm.expect(jsonData).to.have.property("data");',
        '    pm.expect(jsonData.data).to.be.an("array");',
        '    if (jsonData.meta) {',
        '        pm.expect(jsonData.meta).to.have.property("page");',
        '        pm.expect(jsonData.meta).to.have.property("limit");',
        '        pm.expect(jsonData.meta).to.have.property("total");',
        '    }',
        '});'
    ]

def make_get_detail_test(resource_name):
    """Generate Postman test script assertions for single entity detail endpoints.

    Args:
        resource_name (str): Human-readable resource name.

    Returns:
        list[str]: JavaScript test assertion lines.
    """
    return [
        'pm.test("UAT - Status code is 200 OK", function () {',
        '    pm.response.to.have.status(200);',
        '});',
        '',
        'pm.test("UAT - Response time is acceptable (< 2000ms)", function () {',
        '    pm.expect(pm.response.responseTime).to.be.below(2000);',
        '});',
        '',
        'pm.test("UAT - Response format is valid JSON", function () {',
        '    pm.response.to.be.json;',
        '});',
        '',
        f'pm.test("UAT - Returns valid {resource_name} detail object", function () {{',
        '    var jsonData = pm.response.json();',
        '    pm.expect(jsonData).to.have.property("data");',
        '    pm.expect(jsonData.data).to.be.an("object");',
        '});'
    ]

def make_get_lookup_test(resource_name):
    """Generate Postman test script assertions for lookup/dropdown endpoints.

    Args:
        resource_name (str): Human-readable resource name.

    Returns:
        list[str]: JavaScript test assertion lines.
    """
    return [
        'pm.test("UAT - Status code is 200 OK", function () {',
        '    pm.response.to.have.status(200);',
        '});',
        '',
        'pm.test("UAT - Response time is acceptable (< 2000ms)", function () {',
        '    pm.expect(pm.response.responseTime).to.be.below(2000);',
        '});',
        '',
        f'pm.test("UAT - Returns array of {resource_name} lookup options", function () {{',
        '    var jsonData = pm.response.json();',
        '    pm.expect(jsonData).to.have.property("data");',
        '    pm.expect(jsonData.data).to.be.an("array");',
        '});'
    ]

def make_get_excel_test(report_name):
    """Generate Postman test script assertions for Excel report export endpoints.

    Args:
        report_name (str): Human-readable report name.

    Returns:
        list[str]: JavaScript test assertion lines.
    """
    return [
        'pm.test("UAT - Status code is 200 OK", function () {',
        '    pm.response.to.have.status(200);',
        '});',
        '',
        f'pm.test("UAT - Content-Type is valid Excel format for {report_name}", function () {{',
        '    var contentType = pm.response.headers.get("Content-Type");',
        '    pm.expect(contentType).to.satisfy(function(ct) {',
        '        return ct.includes("spreadsheetml") || ct.includes("octet-stream") || ct.includes("excel");',
        '    });',
        '});',
        '',
        'pm.test("UAT - Content-Disposition header specifies attachment file", function () {',
        '    var disposition = pm.response.headers.get("Content-Disposition");',
        '    if (disposition) {',
        '        pm.expect(disposition).to.include("attachment");',
        '        pm.expect(disposition).to.include(".xlsx");',
        '    }',
        '});'
    ]

def make_post_create_test(resource_name):
    """Generate Postman test script assertions for entity creation endpoints.

    Args:
        resource_name (str): Human-readable resource name.

    Returns:
        list[str]: JavaScript test assertion lines.
    """
    return [
        'pm.test("UAT - Status code is 201 Created or 200 OK", function () {',
        '    pm.expect(pm.response.code).to.be.oneOf([200, 201]);',
        '});',
        '',
        'pm.test("UAT - Response time is acceptable (< 3000ms)", function () {',
        '    pm.expect(pm.response.responseTime).to.be.below(3000);',
        '});',
        '',
        f'pm.test("UAT - Created {resource_name} record returned in data object", function () {{',
        '    var jsonData = pm.response.json();',
        '    pm.expect(jsonData).to.have.property("data");',
        '    pm.expect(jsonData.data).to.be.an("object");',
        '});'
    ]

def make_patch_update_test(resource_name):
    """Generate Postman test script assertions for entity update endpoints.

    Args:
        resource_name (str): Human-readable resource name.

    Returns:
        list[str]: JavaScript test assertion lines.
    """
    return [
        'pm.test("UAT - Status code is 200 OK", function () {',
        '    pm.response.to.have.status(200);',
        '});',
        '',
        'pm.test("UAT - Response time is acceptable (< 3000ms)", function () {',
        '    pm.expect(pm.response.responseTime).to.be.below(3000);',
        '});',
        '',
        f'pm.test("UAT - Updated {resource_name} record returned in data object", function () {{',
        '    var jsonData = pm.response.json();',
        '    pm.expect(jsonData).to.have.property("data");',
        '    pm.expect(jsonData.data).to.be.an("object");',
        '});'
    ]

def make_delete_test(resource_name):
    """Generate Postman test script assertions for entity deletion endpoints.

    Args:
        resource_name (str): Human-readable resource name.

    Returns:
        list[str]: JavaScript test assertion lines.
    """
    return [
        'pm.test("UAT - Status code is 200 OK", function () {',
        '    pm.response.to.have.status(200);',
        '});',
        '',
        'pm.test("UAT - Response time is acceptable (< 3000ms)", function () {',
        '    pm.expect(pm.response.responseTime).to.be.below(3000);',
        '});',
        '',
        f'pm.test("UAT - Confirmation message confirms {resource_name} deleted", function () {{',
        '    var jsonData = pm.response.json();',
        '    pm.expect(jsonData).to.have.property("message");',
        '    pm.expect(jsonData.message).to.satisfy(function(msg) {',
        '        return msg === "deleted" || msg.toLowerCase().includes("success");',
        '    });',
        '});'
    ]

if __name__ == "__main__":
    build_collection()
