package domain

import (
	"time"
)

// ==========================================
// 1. ADMIN & AUTHORIZATION MODELS
// ==========================================

// Admin represents table [dbo].[T_Login_Mst]
type Admin struct {
	ID           int32     `gorm:"primaryKey;column:LoginId" json:"id"`
	Username     string    `gorm:"column:Username" json:"username"`
	Password     string    `gorm:"column:Password" json:"-"` // Hidden from JSON output
	GroupId      int32     `gorm:"column:GroupId" json:"group_id"`
	Email        string    `gorm:"column:email" json:"email"`
	PhoneNumber  string    `gorm:"column:PhoneNumber" json:"phone_number"`
	ImgUrl       string    `gorm:"column:ImgUrl" json:"img_url"`
	FlagUse      bool      `gorm:"column:FlagUse" json:"flag_use"`
	DateStart    time.Time `gorm:"column:DateStart" json:"date_start"`
	DateEnd      time.Time `gorm:"column:DateEnd" json:"date_end"`
	LoginDesc    string    `gorm:"column:LoginDesc" json:"login_desc"`
	LastLogin    time.Time `gorm:"column:LastLogin" json:"last_login"`
	DepartmentId int32     `gorm:"column:DepartmentId" json:"department_id"`
	IsWarehouse  bool      `gorm:"column:IsWarehouse" json:"is_warehouse"`

	AdminGroup AdminGroup     `gorm:"foreignKey:GroupId;references:GroupId" json:"admin_group,omitempty"`
	Department *DepartmentMst `gorm:"foreignKey:DepartmentId;references:DepartmentId" json:"department,omitempty"`
}

func (Admin) TableName() string { return "T_Login_Mst" }

// SubMenuItem represents child menu items returned from SP_Login_View_Mapping_Group
type SubMenuItem struct {
	MenuID   int    `gorm:"column:MenuId" json:"menu_id"`
	ParentID int    `gorm:"column:ParentId" json:"parent_id"`
	Level1   string `gorm:"column:Level1" json:"level1,omitempty"`
	Level2   string `gorm:"column:Level2" json:"level2"`
	Level3   string `gorm:"column:Level3" json:"level3"`
	PageUrl  string `gorm:"column:PageUrl" json:"page_url,omitempty"`
	Sequence int    `gorm:"column:Sequence" json:"sequence"`
}

// MainMenuItem represents item from SP_Login_Create_Xml
type MainMenuItem struct {
	MenuID   int           `gorm:"column:MenuId" json:"menu_id"`
	MainMenu string        `gorm:"column:MainMenu" json:"main_menu"`
	SubMenu  []SubMenuItem `gorm:"-" json:"sub_menu"`
}

// AdminGroup represents table [dbo].[T_Login_Group]
type AdminGroup struct {
	GroupId     int32  `gorm:"primaryKey;column:GroupId" json:"group_id"`
	GroupName   string `gorm:"column:GroupName" json:"group_name"`
	RInsert     bool   `gorm:"column:RInsert" json:"r_insert"`
	REdit       bool   `gorm:"column:REdit" json:"r_edit"`
	RDelete     bool   `gorm:"column:RDelete" json:"r_delete"`
	RReporting  bool   `gorm:"column:RReporting" json:"r_reporting"`
	RPositionId int32  `gorm:"column:RPositionId" json:"r_position_id"`
	GroupDesc   string `gorm:"column:GroupDesc" json:"group_desc"`
}

func (AdminGroup) TableName() string { return "T_Login_Group" }

type DepartmentMst struct {
	DepartmentId   int16      `gorm:"primaryKey;column:DepartmentId;type:smallint;not null" json:"department_id"`
	DepartmentCode *string    `gorm:"column:DepartmentCode;type:varchar(25)" json:"department_code"`
	Departmentname *string    `gorm:"column:Departmentname;type:varchar(100)" json:"department_name"`
	Status         bool       `gorm:"column:Status;type:bit;not null" json:"status"`
	ModAct         *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act"`
	ModBy          *string    `gorm:"column:ModBy;type:varchar(25)" json:"mod_by"`
	ModDate        *time.Time `gorm:"column:ModDate;type:datetime" json:"mod_date"`
}

func (DepartmentMst) TableName() string {
	return "T_BUS_DEPARTMENT_MST"
}

// GroupMenuMapping represents table [dbo].[T_Login_Menu]
type GroupMenuMapping struct {
	MenuId       int32  `gorm:"primaryKey;column:MenuId" json:"menu_id"`
	ParentId     int32  `gorm:"column:ParentId" json:"parent_id"`
	MenuName     string `gorm:"column:MenuName" json:"menu_name"`
	PageUrl      string `gorm:"column:PageUrl" json:"page_url"`
	Sequence     int32  `gorm:"column:Squence" json:"sequence"` // Matches 'Squence' typo in schema
	MenuDesc     string `gorm:"column:MenuDesc" json:"menu_desc"`
	ParentLevel1 int32  `gorm:"column:ParentLevel1" json:"parent_level_1"`
	FlagActive   bool   `gorm:"column:FlagActive" json:"flag_active"`
}

func (GroupMenuMapping) TableName() string { return "T_Login_Menu" }

// AdminSubWarehouse represents table [dbo].[T_WH_SUBWH_MST]
type AdminSubWarehouse struct {
	SubWhId         int32     `gorm:"primaryKey;column:SUBWHID" json:"sub_wh_id"`
	WhId            int64     `gorm:"column:WHID" json:"wh_id"`
	FullName        string    `gorm:"column:FULL_NAME" json:"full_name"`
	Pic             string    `gorm:"column:PIC" json:"pic"`
	DocCode         string    `gorm:"column:DOCCODE" json:"doc_code"`
	CruId           int64     `gorm:"column:CRUID" json:"cru_id"`
	UpdateUId       int64     `gorm:"column:UPDATEUID" json:"update_uid"`
	LstUpdate       time.Time `gorm:"column:LSTUPDATE" json:"lst_update"`
	FlagProductions bool      `gorm:"column:FlagProductions" json:"flag_productions"`
	SubWhType       string    `gorm:"column:SubWhType" json:"sub_wh_type"`
}

func (AdminSubWarehouse) TableName() string { return "T_WH_SUBWH_MST" }

// ==========================================
// 2. CORE SYSTEM & MASTER RECORD MODELS
// ==========================================

// Umat represents table [dbo].[T_BUS_UMAT]
type Umat struct {
	ID                   int32      `gorm:"primaryKey;column:id" json:"id"`
	Kode                 string     `gorm:"column:kode" json:"kode"`
	Alias                string     `gorm:"column:alias" json:"alias"`
	NamaIndonesia        string     `gorm:"column:namaindonesia" json:"nama_indonesia"`
	Marga                string     `gorm:"column:marga" json:"marga"`
	NamaMandarin         string     `gorm:"column:namamandarin" json:"nama_mandarin"`
	Alamat               string     `gorm:"column:alamat" json:"alamat"`
	Alamat2              string     `gorm:"column:alamat2" json:"alamat2"`
	Telepon              string     `gorm:"column:telepon" json:"telepon"`
	Mobile               string     `gorm:"column:mobile" json:"mobile"`
	TempatLahir          string     `gorm:"column:tempatlahir" json:"tempat_lahir"`
	TanggalLahir         time.Time  `gorm:"column:tanggallahir" json:"tanggal_lahir"`
	Usia                 int32      `gorm:"column:usia" json:"usia"`
	Wilayah              string     `gorm:"column:wilayah" json:"wilayah"`
	JenisKelamin         string     `gorm:"column:jeniskelamin" json:"jenis_kelamin"`
	JenisKelaminInfo     *AppLookup `gorm:"foreignKey:JenisKelamin;references:LookupValue" json:"jenis_kelamin_info,omitempty"`
	Pekerjaan            string     `gorm:"column:pekerjaan" json:"pekerjaan"`
	Pendidikan           string     `gorm:"column:pendidikan" json:"pendidikan"`
	TanggalChiutaoInt    time.Time  `gorm:"column:tanggalchiutaoint" json:"tanggal_chiutao_int"`
	TanggalChiutaoMan    string     `gorm:"column:tanggalchiutaoman" json:"tanggal_chiutao_man"`
	TahunChiutaoMandarin string     `gorm:"column:tahunchiutaomandarin" json:"tahun_chiutao_mandarin"`
	WaktuChiutaoMandarin string     `gorm:"column:waktuchiutaomandarin" json:"waktu_chiutao_mandarin"`
	Pengajak             string     `gorm:"column:pengajak" json:"pengajak"`
	// PengajakUmat         *Umat      `gorm:"foreignKey:Pengajak;references:Kode" json:"pengajak_umat,omitempty"`
	PengajakManual string `gorm:"column:pengajakmanual" json:"pengajak_manual"`
	Penanggung     string `gorm:"column:penanggung" json:"penanggung"`
	// PenanggungUmat       *Umat      `gorm:"foreignKey:Penanggung;references:Kode" json:"penanggung_umat,omitempty"`
	PenanggungManual    string    `gorm:"column:penanggungmanual" json:"penanggung_manual"`
	Tcs                 string    `gorm:"column:tcs" json:"tcs"`
	UangPahala          float64   `gorm:"column:uangpahala" json:"uang_pahala"`
	FotangChiutao       string    `gorm:"column:fotangciutao" json:"fotang_chiutao"`
	FotangAktif         string    `gorm:"column:fotangaktif" json:"fotang_aktif"`
	Sd2                 bool      `gorm:"column:sd2" json:"sd2"`
	TempatSd2           string    `gorm:"column:tempatsd2" json:"tempat_sd2"`
	TanggalSd2          time.Time `gorm:"column:tanggalsd2" json:"tanggal_sd2"`
	Sd3                 bool      `gorm:"column:sd3" json:"sd3"`
	TempatSd3           string    `gorm:"column:tempatsd3" json:"tempat_sd3"`
	TanggalSd3          time.Time `gorm:"column:tanggalsd3" json:"tanggal_sd3"`
	KelasUmum           string    `gorm:"column:kelasumum" json:"kelas_umum"`
	KelasKhusus         string    `gorm:"column:kelaskhusus" json:"kelas_khusus"`
	ChingKhou           bool      `gorm:"column:chingkhou" json:"ching_khou"`
	TanggalChingKhou    time.Time `gorm:"column:tanggalchingkhou" json:"tanggal_ching_khou"`
	TanggalAncuo        time.Time `gorm:"column:tanggalancuo" json:"tanggal_ancuo"`
	NamaCetyaRumah      string    `gorm:"column:namacetyarumah" json:"nama_cetya_rumah"`
	Meninggal           bool      `gorm:"column:meninggal" json:"meninggal"`
	TanggalMeninggal    time.Time `gorm:"column:tanggalmeninggal" json:"tanggal_meninggal"`
	TimKerja            string    `gorm:"column:timkerja" json:"tim_kerja"`
	Posisi              string    `gorm:"column:posisi" json:"posisi"`
	StatusUmat          string    `gorm:"column:statusumat" json:"status_umat"`
	Keterangan          string    `gorm:"column:keterangan" json:"keterangan"`
	Email               string    `gorm:"column:email" json:"email"`
	ImagePath           string    `gorm:"column:imagepath" json:"image_path"`
	Status              bool      `gorm:"column:status" json:"status"`
	ModAct              string    `gorm:"column:modact" json:"mod_act"`
	ModBy               int32     `gorm:"column:modby" json:"mod_by"`
	ModDate             time.Time `gorm:"column:moddate" json:"mod_date"`
	Ikrar1              bool      `gorm:"column:ikrar1" json:"ikrar_1"`
	Ikrar2              bool      `gorm:"column:ikrar2" json:"ikrar_2"`
	Ikrar3              bool      `gorm:"column:ikrar3" json:"ikrar_3"`
	Ikrar4              bool      `gorm:"column:ikrar4" json:"ikrar_4"`
	Ikrar5              bool      `gorm:"column:ikrar5" json:"ikrar_5"`
	Ikrar6              bool      `gorm:"column:ikrar6" json:"ikrar_6"`
	RenChaiPan          bool      `gorm:"column:RenChaiPan" json:"ren_chai_pan"`
	TanggalRenChaiPan   time.Time `gorm:"column:TanggalRenChaiPan" json:"tanggal_ren_chai_pan"`
	LienCiangPan        bool      `gorm:"column:LienCiangPan" json:"lien_ciang_pan"`
	TanggalLienCiangPan time.Time `gorm:"column:TanggalLienCiangPan" json:"tanggal_lien_ciang_pan"`
	CiangYenPan         bool      `gorm:"column:CiangYenPan" json:"ciang_yen_pan"`
	TanggalCiangYenPan  time.Time `gorm:"column:TanggalCiangYenPan" json:"tanggal_ciang_yen_pan"`
	ActiveStatus        bool      `gorm:"column:ActiveStatus" json:"active_status"`
	NamaFotangLain      string    `gorm:"column:NamaFotangLain" json:"nama_fotang_lain"`
	NamaTcsLain         string    `gorm:"column:NamaTcsLain" json:"nama_tcs_lain"`
	KodeBuku            string    `gorm:"column:KodeBuku" json:"kode_buku"`
}

type AppLookup struct {
	LookupId          string     `gorm:"primaryKey;column:LookupId;type:varchar(25);not null"`
	CategoryId        *string    `gorm:"column:CategoryId;type:varchar(25)"`
	LookupValue       *string    `gorm:"column:LookupValue;type:varchar(50)"`
	LookupDescription *string    `gorm:"column:LookupDescription;type:nvarchar(150)"`
	Status            *bool      `gorm:"column:Status;type:bit"`
	ModAct            *string    `gorm:"column:ModAct;type:char(1)"`
	ModBy             *string    `gorm:"column:ModBy;type:varchar(25)"`
	ModDate           *time.Time `gorm:"column:ModDate;type:datetime"`
}

// TableName menentukan nama tabel secara eksplisit di database
func (AppLookup) TableName() string {
	return "T_APP_LOOKUP"
}

func (Umat) TableName() string { return "T_BUS_UMAT" }

// Topic represents table [dbo].[T_BUS_TOPIC]
type Topic struct {
	TopicCode     string    `gorm:"primaryKey;column:TopicCode" json:"topic_code"`
	TopicName     string    `gorm:"column:TopicName" json:"topic_name"`
	TopicCategory string    `gorm:"column:TopicCategory" json:"topic_category"`
	Description   string    `gorm:"column:Description" json:"description"`
	Status        bool      `gorm:"column:Status" json:"status"`
	ModAct        string    `gorm:"column:ModAct" json:"mod_act"`
	ModBy         string    `gorm:"column:ModBy" json:"mod_by"`
	ModDate       time.Time `gorm:"column:ModDate" json:"mod_date"`
}

func (Topic) TableName() string { return "T_BUS_TOPIC" }

// Activity represents table [dbo].[T_BUS_EVENT]
type Activity struct {
	EventCode     string    `gorm:"primaryKey;column:EventCode" json:"event_code"`
	EventName     string    `gorm:"column:EventName" json:"event_name"`
	EventCategory string    `gorm:"column:EventCategory" json:"event_category"`
	Description   string    `gorm:"column:Description" json:"description"`
	Status        bool      `gorm:"column:Status" json:"status"`
	ModAct        string    `gorm:"column:ModAct" json:"mod_act"`
	ModBy         string    `gorm:"column:ModBy" json:"mod_by"`
	ModDate       time.Time `gorm:"column:ModDate" json:"mod_date"`
}

func (Activity) TableName() string { return "T_BUS_EVENT" } // TimKerja represents records inside [dbo].[T_APP_LOOKUP] filtered by CategoryId = 'B_POSISI'

type TimKerja struct {
	LookupId          string    `gorm:"primaryKey;column:LookupId" json:"lookup_id"`
	CategoryId        string    `gorm:"column:CategoryId;default:B_POSISI" json:"category_id"`
	LookupValue       string    `gorm:"column:LookupValue" json:"lookup_value"`
	LookupDescription string    `gorm:"column:LookupDescription" json:"lookup_description"`
	Status            bool      `gorm:"column:Status" json:"status"`
	ModAct            string    `gorm:"column:ModAct" json:"mod_act"`
	ModBy             string    `gorm:"column:ModBy" json:"mod_by"`
	ModDate           time.Time `gorm:"column:ModDate" json:"mod_date"`
}

func (TimKerja) TableName() string { return "T_APP_LOOKUP" } // TahunCiuTao represents table [dbo].[T_BUS_TAHUN_CIUTAO]

type TahunCiuTao struct {
	TahunMandarin string    `gorm:"primaryKey;column:TahunMandarin" json:"tahun_mandarin"`
	StartDate     time.Time `gorm:"column:StartDate" json:"start_date"`
	EndDate       time.Time `gorm:"column:EndDate" json:"end_date"`
	ModAct        string    `gorm:"column:ModAct" json:"mod_act"`
	ModBy         string    `gorm:"column:ModBy" json:"mod_by"`
	ModDate       time.Time `gorm:"column:ModDate" json:"mod_date"`
	Status        bool      `gorm:"column:status" json:"status"`
	Description   string    `gorm:"column:description" json:"description"`
}

func (TahunCiuTao) TableName() string { return "T_BUS_TAHUN_CIUTAO" } // PenggalangDana represents table [dbo].[T_SXY_MST_PENGGALANG]

type PenggalangDana struct {
	ID            int32     `gorm:"primaryKey;column:id" json:"id"`
	No            string    `gorm:"column:no" json:"no"`
	Nama          string    `gorm:"column:nama" json:"nama"`
	Mandarin      string    `gorm:"column:mandarin" json:"mandarin"`
	Keterangan    string    `gorm:"column:keterangan" json:"keterangan"`
	LookupFothang int32     `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Alamat        string    `gorm:"column:alamat" json:"alamat"`
	Telepon       string    `gorm:"column:telepon" json:"telepon"`
	Mobile        string    `gorm:"column:mobile" json:"mobile"`
	Email         string    `gorm:"column:email" json:"email"`
	Status        bool      `gorm:"column:STATUS" json:"status"`
	CreatedBy     int32     `gorm:"column:createdby" json:"created_by"`
	CreatedDate   time.Time `gorm:"column:createddate" json:"created_date"`
	UpdatedBy     int32     `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate   time.Time `gorm:"column:updateddate" json:"updated_date"`
}

func (PenggalangDana) TableName() string { return "T_SXY_MST_PENGGALANG" } // SxyDonatur represents table [dbo].[T_SXY_MST_DONATUR]

type SxyDonatur struct {
	ID            int32     `gorm:"primaryKey;column:id" json:"id"`
	No            string    `gorm:"column:no" json:"no"`
	Nama          string    `gorm:"column:nama" json:"nama"`
	Mandarin      string    `gorm:"column:mandarin" json:"mandarin"`
	Keterangan    string    `gorm:"column:keterangan" json:"keterangan"`
	LookupFothang int32     `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Alamat        string    `gorm:"column:alamat" json:"alamat"`
	Telepon       string    `gorm:"column:telepon" json:"telepon"`
	Mobile        string    `gorm:"column:mobile" json:"mobile"`
	Email         string    `gorm:"column:email" json:"email"`
	Status        bool      `gorm:"column:STATUS" json:"status"`
	CreatedBy     int32     `gorm:"column:createdby" json:"created_by"`
	CreatedDate   time.Time `gorm:"column:createddate" json:"created_date"`
	UpdatedBy     int32     `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate   time.Time `gorm:"column:updateddate" json:"updated_date"`
}

func (SxyDonatur) TableName() string { return "T_SXY_MST_DONATUR" }

// ==========================================// 3. SPECIAL CUSTOM TYPE & DONATION MODELS// ==========================================// Kelas represents records inside [dbo].[T_APP_LOOKUP] filtered by CategoryId = 'B_KELASKHUSUS'

type Kelas struct {
	LookupId          string    `gorm:"primaryKey;column:LookupId" json:"lookup_id"`
	CategoryId        string    `gorm:"column:CategoryId;default:B_KELASKHUSUS" json:"category_id"`
	LookupValue       string    `gorm:"column:LookupValue" json:"lookup_value"`
	LookupDescription string    `gorm:"column:LookupDescription" json:"lookup_description"`
	Status            bool      `gorm:"column:Status" json:"status"`
	ModAct            string    `gorm:"column:ModAct" json:"mod_act"`
	ModBy             string    `gorm:"column:ModBy" json:"mod_by"`
	ModDate           time.Time `gorm:"column:ModDate" json:"mod_date"`
}

func (Kelas) TableName() string { return "T_APP_LOOKUP" } // DonasiSxy represents table [dbo].[T_SXY_TRANSAKSI]

type DonasiSxy struct {
	ID              int32     `gorm:"primaryKey;column:id" json:"id"`
	NoKwitansi      string    `gorm:"column:nokwitansi" json:"no_kwitansi"`
	Tanggal         time.Time `gorm:"column:tanggal" json:"tanggal"`
	Donatur         int32     `gorm:"column:donatur" json:"donatur_id"`
	Penggalang      int32     `gorm:"column:penggalang" json:"penggalang_id"`
	Jumlah          float64   `gorm:"column:jumlah" json:"jumlah"` // Maps NUMERIC(18,0) cleanly
	TipeSumbangan   int32     `gorm:"column:tipesumbangan" json:"tipe_sumbangan"`
	NoKupon         string    `gorm:"column:nokupon" json:"no_kupon"`
	Keterangan      string    `gorm:"column:keterangan" json:"keterangan"`
	Status          bool      `gorm:"column:STATUS" json:"status"`
	CreatedBy       int32     `gorm:"column:createdby" json:"created_by"`
	CreatedDate     time.Time `gorm:"column:createddate" json:"created_date"`
	UpdatedBy       int32     `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate     time.Time `gorm:"column:updateddate" json:"updated_date"`
	TanggalTransfer time.Time `gorm:"column:tanggaltransfer" json:"tanggal_transfer"`
	AtasNama        string    `gorm:"column:atasnama" json:"atas_nama"`
	TtkSent         bool      `gorm:"column:ttksent" json:"ttk_sent"`
}

func (DonasiSxy) TableName() string { return "T_SXY_TRANSAKSI" }
