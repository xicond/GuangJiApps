package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ==========================================
// 1. ADMIN & AUTHORIZATION MODELS
// ==========================================

// Admin represents table [dbo].[T_Login_Mst]
type Admin struct {
	ID           int32     `gorm:"primaryKey;column:LoginId" json:"id"`
	Username     string    `gorm:"column:Username" json:"username" validate:"required,max=50"`
	Password     string    `gorm:"column:Password" json:"-" validate:"omitempty,max=50"`
	GroupId      int32     `gorm:"column:GroupId" json:"group_id,omitempty"`
	Email        *string   `gorm:"column:email" json:"email,omitempty" validate:"omitempty,max=50,email"`
	PhoneNumber  *string   `gorm:"column:PhoneNumber" json:"phone_number,omitempty" validate:"omitempty,max=50"`
	ImgUrl       string    `gorm:"column:ImgUrl" json:"img_url" validate:"omitempty,max=50"`
	FlagUse      bool      `gorm:"column:FlagUse" json:"flag_use"`
	DateStart    time.Time `gorm:"column:DateStart" json:"date_start,omitempty"`
	DateEnd      time.Time `gorm:"column:DateEnd" json:"date_end,omitempty"`
	LoginDesc    *string   `gorm:"column:LoginDesc" json:"login_desc,omitempty" validate:"omitempty,max=350"`
	LastLogin    time.Time `gorm:"column:LastLogin" json:"last_login"`
	DepartmentId int32     `gorm:"column:DepartmentId" json:"department_id"`
	IsWarehouse  bool      `gorm:"column:IsWarehouse" json:"is_warehouse"`

	AdminGroup AdminGroup     `gorm:"foreignKey:GroupId;references:GroupId;constraint:false" json:"admin_group" validate:"-"`
	Department *DepartmentMst `gorm:"foreignKey:DepartmentId;references:DepartmentId;constraint:false" json:"department,omitempty" validate:"-"`
}

func (Admin) TableName() string { return "T_Login_Mst" }

type AdminMatrix struct {
	CruID       int64  `gorm:"primaryKey;column:CRUID;type:bigint;not null" json:"cru_id"`
	LoginId     *int64 `gorm:"column:LOGINID;type:bigint" json:"login_id,omitempty"`
	WarehouseId *int64 `gorm:"column:WAREHOUSEID;type:bigint" json:"warehouse_id,omitempty"`
	SubWhId     *int64 `gorm:"column:SUBWHID;type:bigint" json:"sub_wh_id,omitempty"`
}

// TableName menentukan nama tabel secara eksplisit di database
func (AdminMatrix) TableName() string {
	return "T_WH_USER_MATRIX_MST"
}

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
	GroupName   string `gorm:"column:GroupName" json:"group_name" validate:"required,max=50"`
	RInsert     bool   `gorm:"column:RInsert" json:"-"`
	REdit       bool   `gorm:"column:REdit" json:"-"`
	RDelete     bool   `gorm:"column:RDelete" json:"-"`
	RReporting  bool   `gorm:"column:RReporting" json:"-"`
	RPositionId int32  `gorm:"column:RPositionId" json:"-"`
	GroupDesc   string `gorm:"column:GroupDesc" json:"group_desc" validate:"omitempty,max=350"`
}

func (AdminGroup) TableName() string { return "T_Login_Group" }

type DepartmentMst struct {
	DepartmentId   int16      `gorm:"primaryKey;column:DepartmentId;type:smallint;not null" json:"department_id"`
	DepartmentCode *string    `gorm:"column:DepartmentCode;type:varchar(25)" json:"department_code" validate:"omitempty,max=25"`
	Departmentname *string    `gorm:"column:Departmentname;type:varchar(100)" json:"department_name" validate:"omitempty,max=100"`
	Status         bool       `gorm:"column:Status;type:bit;not null" json:"status"`
	ModAct         *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,max=1"`
	ModBy          *string    `gorm:"column:ModBy;type:varchar(25)" json:"mod_by" validate:"omitempty,max=25"`
	ModDate        *time.Time `gorm:"column:ModDate;type:datetime" json:"mod_date"`
}

func (DepartmentMst) TableName() string {
	return "T_BUS_DEPARTMENT_MST"
}

// GroupMenuMapping represents table [dbo].[T_Login_Menu]
type GroupMenuMapping struct {
	MenuId       int32  `gorm:"primaryKey;column:MenuId" json:"menu_id"`
	ParentId     int32  `gorm:"column:ParentId" json:"parent_id"`
	MenuName     string `gorm:"column:MenuName" json:"menu_name" validate:"required,max=50"`
	PageUrl      string `gorm:"column:PageUrl" json:"page_url" validate:"omitempty,max=250"`
	Sequence     int32  `gorm:"column:Squence" json:"sequence"` // Matches 'Squence' typo in schema
	MenuDesc     string `gorm:"column:MenuDesc" json:"menu_desc" validate:"omitempty,max=350"`
	ParentLevel1 int32  `gorm:"column:ParentLevel1" json:"parent_level_1"`
	FlagActive   bool   `gorm:"column:FlagActive" json:"flag_active"`
}

func (GroupMenuMapping) TableName() string { return "T_Login_Menu" }

// AdminSubWarehouse represents table [dbo].[T_WH_SUBWH_MST]
type AdminSubWarehouse struct {
	SubWhId         int32     `gorm:"primaryKey;column:SUBWHID" json:"sub_wh_id"`
	WhId            int64     `gorm:"column:WHID" json:"wh_id"`
	FullName        string    `gorm:"column:FULL_NAME" json:"full_name" validate:"required,max=100"`
	Pic             string    `gorm:"column:PIC" json:"pic" validate:"omitempty,max=75"`
	DocCode         string    `gorm:"column:DOCCODE" json:"doc_code" validate:"omitempty,max=50"`
	CruId           int64     `gorm:"column:CRUID" json:"cru_id"`
	UpdateUId       int64     `gorm:"column:UPDATEUID" json:"update_uid"`
	LstUpdate       time.Time `gorm:"column:LSTUPDATE" json:"lst_update"`
	FlagProductions bool      `gorm:"column:FlagProductions" json:"flag_productions"`
	SubWhType       string    `gorm:"column:SubWhType" json:"sub_wh_type" validate:"omitempty,max=3"`
}

func (AdminSubWarehouse) TableName() string { return "T_WH_SUBWH_MST" }

// ==========================================
// 2. CORE SYSTEM & MASTER RECORD MODELS
// ==========================================

// Umat represents table [dbo].[T_BUS_UMAT]
type Umat struct {
	ID                   int32      `gorm:"primaryKey;column:id" json:"id"`
	Kode                 string     `gorm:"column:kode" json:"kode" validate:"omitempty,max=50"`
	Alias                string     `gorm:"column:alias" json:"alias" validate:"omitempty,max=50"`
	NamaIndonesia        string     `gorm:"column:namaindonesia" json:"nama_indonesia" validate:"required,max=50"`
	Marga                string     `gorm:"column:marga" json:"marga" validate:"omitempty,max=10"`
	NamaMandarin         string     `gorm:"column:namamandarin" json:"nama_mandarin" validate:"omitempty,max=50"`
	Alamat               string     `gorm:"column:alamat" json:"alamat" validate:"omitempty,max=200"`
	Alamat2              string     `gorm:"column:alamat2" json:"alamat2" validate:"omitempty,max=100"`
	Telepon              string     `gorm:"column:telepon" json:"telepon" validate:"omitempty,max=50"`
	Mobile               string     `gorm:"column:mobile" json:"mobile" validate:"omitempty,max=50"`
	TempatLahir          string     `gorm:"column:tempatlahir" json:"tempat_lahir" validate:"omitempty,max=50"`
	TanggalLahir         time.Time  `gorm:"column:tanggallahir" json:"tanggal_lahir"`
	Usia                 int32      `gorm:"column:usia" json:"usia" validate:"omitempty,gte=0,lte=150"`
	Wilayah              string     `gorm:"column:wilayah" json:"wilayah" validate:"omitempty,max=100"`
	JenisKelamin         string     `gorm:"column:jeniskelamin" json:"jenis_kelamin" validate:"required,omitempty,max=3"`
	JenisKelaminInfo     *AppLookup `gorm:"foreignKey:JenisKelamin;references:LookupValue;constraint:false" json:"jenis_kelamin_info,omitempty" validate:"-"`
	Pekerjaan            string     `gorm:"column:pekerjaan" json:"pekerjaan" validate:"omitempty,max=3"`
	Pendidikan           string     `gorm:"column:pendidikan" json:"pendidikan" validate:"omitempty,max=3"`
	TanggalChiutaoInt    time.Time  `gorm:"column:tanggalchiutaoint" json:"tanggal_chiutao_int"`
	TanggalChiutaoMan    string     `gorm:"column:tanggalchiutaoman" json:"tanggal_chiutao_man" validate:"omitempty,max=6"`
	TahunChiutaoMandarin string     `gorm:"column:tahunchiutaomandarin" json:"tahun_chiutao_mandarin" validate:"omitempty,max=20"`
	WaktuChiutaoMandarin string     `gorm:"column:waktuchiutaomandarin" json:"waktu_chiutao_mandarin" validate:"omitempty,max=3"`
	Pengajak             string     `gorm:"column:pengajak" json:"pengajak" validate:"omitempty,max=20"`
	// PengajakUmat         *Umat      `gorm:"foreignKey:Pengajak;references:Kode" json:"pengajak_umat,omitempty"`
	PengajakManual string `gorm:"column:pengajakmanual" json:"pengajak_manual" validate:"omitempty,max=50"`
	Penanggung     string `gorm:"column:penanggung" json:"penanggung" validate:"omitempty,max=20"`
	// PenanggungUmat       *Umat      `gorm:"foreignKey:Penanggung;references:Kode" json:"penanggung_umat,omitempty"`
	PenanggungManual    string    `gorm:"column:penanggungmanual" json:"penanggung_manual" validate:"omitempty,max=50"`
	Tcs                 string    `gorm:"column:tcs" json:"tcs" validate:"omitempty,max=3"`
	UangPahala          float64   `gorm:"column:uangpahala" json:"uang_pahala" validate:"omitempty,gte=0"`
	FotangChiutao       string    `gorm:"column:fotangciutao" json:"fotang_chiutao" validate:"omitempty,max=3"`
	FotangAktif         string    `gorm:"column:fotangaktif" json:"fotang_aktif" validate:"omitempty,max=3"`
	Sd2                 bool      `gorm:"column:sd2" json:"sd2"`
	TempatSd2           string    `gorm:"column:tempatsd2" json:"tempat_sd2" validate:"omitempty,max=3"`
	TanggalSd2          time.Time `gorm:"column:tanggalsd2" json:"tanggal_sd2"`
	Sd3                 bool      `gorm:"column:sd3" json:"sd3"`
	TempatSd3           string    `gorm:"column:tempatsd3" json:"tempat_sd3" validate:"omitempty,max=3"`
	TanggalSd3          time.Time `gorm:"column:tanggalsd3" json:"tanggal_sd3"`
	KelasUmum           string    `gorm:"column:kelasumum" json:"kelas_umum" validate:"omitempty,max=3"`
	KelasKhusus         string    `gorm:"column:kelaskhusus" json:"kelas_khusus" validate:"omitempty,max=3"`
	ChingKhou           bool      `gorm:"column:chingkhou" json:"ching_khou"`
	TanggalChingKhou    time.Time `gorm:"column:tanggalchingkhou" json:"tanggal_ching_khou"`
	TanggalAncuo        time.Time `gorm:"column:tanggalancuo" json:"tanggal_ancuo"`
	NamaCetyaRumah      string    `gorm:"column:namacetyarumah" json:"nama_cetya_rumah" validate:"omitempty,max=50"`
	Meninggal           bool      `gorm:"column:meninggal" json:"meninggal"`
	TanggalMeninggal    time.Time `gorm:"column:tanggalmeninggal" json:"tanggal_meninggal"`
	TimKerja            string    `gorm:"column:timkerja" json:"tim_kerja" validate:"omitempty,max=3"`
	Posisi              string    `gorm:"column:posisi" json:"posisi" validate:"omitempty,max=3"`
	StatusUmat          string    `gorm:"column:statusumat" json:"status_umat" validate:"omitempty,max=3"`
	Keterangan          string    `gorm:"column:keterangan" json:"keterangan" validate:"omitempty,max=200"`
	Email               string    `gorm:"column:email" json:"email" validate:"omitempty,max=50,email"`
	ImagePath           string    `gorm:"column:imagepath" json:"image_path" validate:"omitempty,max=50"`
	Status              bool      `gorm:"column:status" json:"status"`
	ModAct              string    `gorm:"column:modact" json:"mod_act" validate:"omitempty,max=1"`
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
	NamaFotangLain      string    `gorm:"column:NamaFotangLain" json:"nama_fotang_lain" validate:"omitempty,max=50"`
	NamaTcsLain         string    `gorm:"column:NamaTcsLain" json:"nama_tcs_lain" validate:"omitempty,max=50"`
	KodeBuku            string    `gorm:"column:KodeBuku" json:"kode_buku" validate:"omitempty,max=50"`
}

type AppLookup struct {
	LookupId          string     `gorm:"primaryKey;column:LookupId;type:varchar(25);not null" json:"lookup_id" validate:"omitempty,max=25"`
	CategoryId        *string    `gorm:"column:CategoryId;type:varchar(25)" json:"category_id,omitempty" validate:"omitempty,max=25"`
	LookupValue       *string    `gorm:"column:LookupValue;type:varchar(50)" json:"lookup_value,omitempty" validate:"omitempty,max=50"`
	LookupDescription *string    `gorm:"column:LookupDescription;type:nvarchar(150)" json:"lookup_description,omitempty" validate:"omitempty,max=150"`
	Status            *bool      `gorm:"column:Status;type:bit" json:"status,omitempty"`
	ModAct            *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act,omitempty" validate:"omitempty,max=1"`
	ModBy             *string    `gorm:"column:ModBy;type:varchar(25)" json:"mod_by,omitempty" validate:"omitempty,max=25"`
	ModDate           *time.Time `gorm:"column:ModDate;type:datetime" json:"mod_date,omitempty"`

	Category *AppLookupCategory `gorm:"foreignKey:CategoryId;references:CategoryId;constraint:false" json:"category,omitempty" validate:"-"`
}

// TableName menentukan nama tabel secara eksplisit di database
func (AppLookup) TableName() string {
	return "T_APP_LOOKUP"
}

type AppLookupCategory struct {
	CategoryId          string     `gorm:"primaryKey;column:CategoryId;type:varchar(25);not null" json:"category_id" validate:"omitempty,max=25"`
	CategoryType        string     `gorm:"column:CategoryType;type:char(1);not null" json:"category_type" validate:"omitempty,max=1"`
	CategoryDescription string     `gorm:"column:CategoryDescription;type:varchar(150);not null" json:"category_description" validate:"omitempty,max=150"`
	Status              bool       `gorm:"column:Status;type:bit;not null" json:"status"`
	ModAct              *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act,omitempty" validate:"omitempty,max=1"`
	ModBy               *string    `gorm:"column:ModBy;type:varchar(25)" json:"mod_by,omitempty" validate:"omitempty,max=25"`
	ModDate             *time.Time `gorm:"column:ModDate;type:datetime" json:"mod_date,omitempty"`
}

// TableName memaksa GORM menggunakan nama tabel spesifik tanpa pluralisasi otomatis
func (AppLookupCategory) TableName() string {
	return "T_APP_LOOKUPCATEGORY"
}

func (Umat) TableName() string { return "T_BUS_UMAT" }

// Topic represents table [dbo].[T_BUS_TOPIC]
type Topic struct {
	TopicCode     string    `gorm:"primaryKey;column:TopicCode" json:"topic_code" validate:"required,max=20"`
	TopicName     string    `gorm:"column:TopicName" json:"topic_name" validate:"required,max=300"`
	TopicCategory string    `gorm:"column:TopicCategory" json:"topic_category" validate:"omitempty,max=3"`
	Description   string    `gorm:"column:Description" json:"description" validate:"omitempty,max=200"`
	Status        bool      `gorm:"column:Status" json:"status"`
	ModAct        string    `gorm:"column:ModAct" json:"mod_act" validate:"omitempty,max=1"`
	ModBy         string    `gorm:"column:ModBy" json:"mod_by" validate:"omitempty,max=25"`
	ModDate       time.Time `gorm:"column:ModDate" json:"mod_date"`

	TopicCategoryInfo *AppLookup `gorm:"foreignKey:TopicCategory;references:LookupValue;constraint:false" json:"topic_category_info,omitempty" validate:"-"`
}

func (Topic) TableName() string { return "T_BUS_TOPIC" }

// Activity represents table [dbo].[T_BUS_EVENT]
type Activity struct {
	EventCode     string    `gorm:"primaryKey;column:EventCode" json:"event_code" validate:"required,max=10"`
	EventName     string    `gorm:"column:EventName" json:"event_name" validate:"required,max=100"`
	EventCategory string    `gorm:"column:EventCategory" json:"event_category" validate:"omitempty,max=3"`
	Description   string    `gorm:"column:Description" json:"description" validate:"omitempty,max=200"`
	Status        bool      `gorm:"column:Status" json:"status"`
	ModAct        string    `gorm:"column:ModAct" json:"mod_act" validate:"omitempty,max=1"`
	ModBy         string    `gorm:"column:ModBy" json:"mod_by" validate:"omitempty,max=25"`
	ModDate       time.Time `gorm:"column:ModDate" json:"mod_date"`
}

func (Activity) TableName() string { return "T_BUS_EVENT" } // TimKerja represents records inside [dbo].[T_APP_LOOKUP] filtered by CategoryId = 'B_POSISI'

type TimKerja struct {
	LookupId          string    `gorm:"primaryKey;column:LookupId" json:"lookup_id" validate:"omitempty,max=25"`
	CategoryId        string    `gorm:"column:CategoryId;default:B_POSISI" json:"category_id" validate:"omitempty,max=25"`
	LookupValue       string    `gorm:"column:LookupValue" json:"lookup_value" validate:"required,max=50"`
	LookupDescription string    `gorm:"column:LookupDescription" json:"lookup_description" validate:"required,max=150"`
	Status            bool      `gorm:"column:Status" json:"status"`
	ModAct            string    `gorm:"column:ModAct" json:"mod_act" validate:"omitempty,max=1"`
	ModBy             string    `gorm:"column:ModBy" json:"mod_by" validate:"omitempty,max=25"`
	ModDate           time.Time `gorm:"column:ModDate" json:"mod_date"`
}

func (TimKerja) TableName() string { return "T_APP_LOOKUP" } // TahunCiuTao represents table [dbo].[T_BUS_TAHUN_CIUTAO]

type DateOnly struct {
	time.Time
}

// Scan mendeteksi data dari database SQL Server
func (d *DateOnly) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case string:
		return d.parseString(v)
	case []byte:
		return d.parseString(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into DateOnly", value)
	}
}

func (d *DateOnly) parseString(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	layouts := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			d.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse date string %q into DateOnly", s)
}

// Value untuk menyimpan kembali ke database jika diperlukan
func (d DateOnly) Value() (driver.Value, error) {
	return d.Time, nil
}

// UnmarshalJSON mendeteksi format string JSON (misal "2026-07-27" atau "2026-07-27T00:00:00Z")
func (d *DateOnly) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	return d.parseString(s)
}

// MarshalJSON memotong format jam dan hanya menampilkan YYYY-MM-DD
func (d DateOnly) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(d.Format("2006-01-02"))
}

// IntBool adalah tipe kustom untuk mengubah int/bit/string (0/1) dari DB menjadi bool di Go
type IntBool bool

// Scan mengonversi nilai dari database (bool/int/int64/uint8/[]byte/string/etc) ke bool
func (ib *IntBool) Scan(value interface{}) error {
	if value == nil {
		*ib = false
		return nil
	}

	switch v := value.(type) {
	case bool:
		*ib = IntBool(v)
	case int64:
		*ib = v != 0
	case int32:
		*ib = v != 0
	case int16:
		*ib = v != 0
	case int8:
		*ib = v != 0
	case int:
		*ib = v != 0
	case uint64:
		*ib = v != 0
	case uint32:
		*ib = v != 0
	case uint16:
		*ib = v != 0
	case uint8:
		*ib = v != 0
	case uint:
		*ib = v != 0
	case float64:
		*ib = v != 0
	case float32:
		*ib = v != 0
	case []byte:
		str := strings.TrimSpace(string(v))
		*ib = IntBool(str == "1" || strings.EqualFold(str, "true") || strings.EqualFold(str, "t") || strings.EqualFold(str, "y") || strings.EqualFold(str, "ya"))
	case string:
		str := strings.TrimSpace(v)
		*ib = IntBool(str == "1" || strings.EqualFold(str, "true") || strings.EqualFold(str, "t") || strings.EqualFold(str, "y") || strings.EqualFold(str, "ya"))
	default:
		return fmt.Errorf("cannot scan type %T into IntBool", value)
	}
	return nil
}

// Value mengonversi kembali bool ke database jika diperlukan
func (ib IntBool) Value() (driver.Value, error) {
	if ib {
		return 1, nil
	}
	return 0, nil
}

// UnmarshalJSON mendeteksi format boolean (true/false), integer (1/0), string ("1"/"0"/"true"/"false"), atau null
func (ib *IntBool) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" {
		*ib = false
		return nil
	}
	if s == "true" || s == "1" || s == `"1"` || s == `"true"` || s == `"t"` || s == `"y"` || s == `"ya"` {
		*ib = true
		return nil
	}
	if s == "false" || s == "0" || s == `"0"` || s == `"false"` || s == `"f"` || s == `"n"` || s == `"tidak"` {
		*ib = false
		return nil
	}

	var bVal bool
	if err := json.Unmarshal(b, &bVal); err == nil {
		*ib = IntBool(bVal)
		return nil
	}
	var iVal int
	if err := json.Unmarshal(b, &iVal); err == nil {
		*ib = iVal != 0
		return nil
	}
	var strVal string
	if err := json.Unmarshal(b, &strVal); err == nil {
		strVal = strings.TrimSpace(strVal)
		*ib = IntBool(strVal == "1" || strings.EqualFold(strVal, "true") || strings.EqualFold(strVal, "t") || strings.EqualFold(strVal, "y") || strings.EqualFold(strVal, "ya"))
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into IntBool", s)
}

// MarshalJSON mengembalikan boolean true/false untuk JSON output
func (ib IntBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(ib))
}

type TahunCiuTao struct {
	TahunMandarin string    `gorm:"primaryKey;column:TahunMandarin" json:"tahun_mandarin" validate:"required,max=20"`
	StartDate     DateOnly  `gorm:"column:StartDate" json:"start_date"`
	EndDate       DateOnly  `gorm:"column:EndDate" json:"end_date"`
	ModAct        string    `gorm:"column:ModAct" json:"mod_act" validate:"omitempty,max=1"`
	ModBy         string    `gorm:"column:ModBy" json:"mod_by" validate:"omitempty,max=25"`
	ModDate       time.Time `gorm:"column:ModDate" json:"mod_date"`
	Status        bool      `gorm:"column:status" json:"status"`
	Description   string    `gorm:"column:description" json:"description" validate:"omitempty,max=50"`
}

func (TahunCiuTao) TableName() string { return "T_BUS_TAHUN_CIUTAO" } // PenggalangDana represents table [dbo].[T_SXY_MST_PENGGALANG]

type PenggalangDana struct {
	ID            int32     `gorm:"primaryKey;column:id" json:"id"`
	No            string    `gorm:"column:no" json:"no" validate:"required,max=50"`
	Nama          string    `gorm:"column:nama" json:"nama" validate:"required,max=50"`
	Mandarin      string    `gorm:"column:mandarin" json:"mandarin" validate:"omitempty,max=50"`
	Keterangan    string    `gorm:"column:keterangan" json:"keterangan" validate:"omitempty,max=255"`
	LookupFothang int32     `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Alamat        string    `gorm:"column:alamat" json:"alamat" validate:"omitempty,max=255"`
	Telepon       string    `gorm:"column:telepon" json:"telepon" validate:"omitempty,max=50"`
	Mobile        string    `gorm:"column:mobile" json:"mobile" validate:"omitempty,max=50"`
	Email         string    `gorm:"column:email" json:"email" validate:"omitempty,max=100,email"`
	Status        bool      `gorm:"column:STATUS" json:"status"`
	CreatedBy     int32     `gorm:"column:createdby" json:"created_by"`
	CreatedDate   time.Time `gorm:"column:createddate" json:"created_date"`
	UpdatedBy     int32     `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate   time.Time `gorm:"column:updateddate" json:"updated_date"`
}
type PenggalangDanaResponse struct {
	ID            int32  `gorm:"primaryKey;column:id" json:"id"`
	No            string `gorm:"column:no" json:"no"`
	Nama          string `gorm:"column:nama" json:"nama"`
	Mandarin      string `gorm:"column:mandarin" json:"mandarin"`
	Keterangan    string `gorm:"column:keterangan" json:"keterangan"`
	LookupFothang int32  `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Fotang        string `gorm:"column:fothang" json:"fotang"`
	Alamat        string `gorm:"column:alamat" json:"alamat"`
	Telepon       string `gorm:"column:telepon" json:"telepon"`
	Mobile        string `gorm:"column:mobile" json:"mobile"`
	Email         string `gorm:"column:email" json:"email"`
}

func (PenggalangDana) TableName() string { return "T_SXY_MST_PENGGALANG" } // SxyDonatur represents table [dbo].[T_SXY_MST_DONATUR]

type SxyDonatur struct {
	ID            int32     `gorm:"primaryKey;column:id" json:"id"`
	No            string    `gorm:"column:no" json:"no" validate:"required,max=10"`
	Nama          string    `gorm:"column:nama" json:"nama" validate:"required,max=50"`
	Mandarin      string    `gorm:"column:mandarin" json:"mandarin" validate:"omitempty,max=50"`
	Keterangan    string    `gorm:"column:keterangan" json:"keterangan" validate:"omitempty,max=255"`
	LookupFothang int32     `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Alamat        string    `gorm:"column:alamat" json:"alamat" validate:"omitempty,max=255"`
	Telepon       string    `gorm:"column:telepon" json:"telepon" validate:"omitempty,max=50"`
	Mobile        string    `gorm:"column:mobile" json:"mobile" validate:"omitempty,max=50"`
	Email         string    `gorm:"column:email" json:"email" validate:"omitempty,max=100,email"`
	Status        bool      `gorm:"column:STATUS" json:"status"`
	CreatedBy     int32     `gorm:"column:createdby" json:"created_by"`
	CreatedDate   time.Time `gorm:"column:createddate" json:"created_date"`
	UpdatedBy     int32     `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate   time.Time `gorm:"column:updateddate" json:"updated_date"`
}

type SxyDonaturResponse struct {
	ID            int32     `gorm:"primaryKey;column:id" json:"id"`
	No            string    `gorm:"column:no" json:"no"`
	Nama          string    `gorm:"column:nama" json:"nama"`
	Mandarin      string    `gorm:"column:mandarin" json:"mandarin"`
	Keterangan    string    `gorm:"column:keterangan" json:"keterangan"`
	LookupFothang int32     `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Fotang        string    `gorm:"column:fothang" json:"fotang"`
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
	TrxId      int32      `gorm:"primaryKey;column:trxid;type:int;not null" json:"trx_id,omitempty"`
	KodeKelas  *string    `gorm:"column:kodekelas;type:varchar(3)" json:"kode_kelas,omitempty" validate:"required,max=3"`
	StartDate  *time.Time `gorm:"column:startdate;type:date" json:"start_date,omitempty"`
	EndDate    *time.Time `gorm:"column:enddate;type:date" json:"end_date,omitempty"`
	KodeFotang *string    `gorm:"column:kodefotang;type:varchar(3)" json:"kode_fotang,omitempty" validate:"omitempty,max=3"`
	Lokasi     *string    `gorm:"column:lokasi;type:varchar(50)" json:"lokasi,omitempty" validate:"omitempty,max=50"`
	PIC        *string    `gorm:"column:PIC;type:nvarchar(100)" json:"pic,omitempty" validate:"omitempty,max=100"`
	Keterangan *string    `gorm:"column:keterangan;type:varchar(200)" json:"keterangan,omitempty" validate:"omitempty,max=200"`
	Status     *bool      `gorm:"column:status;type:bit" json:"status,omitempty"`
	ModAct     *string    `gorm:"column:modact;type:char(1)" json:"mod_act,omitempty" validate:"omitempty,max=1"`
	ModBy      *int32     `gorm:"column:modby;type:int" json:"mod_by,omitempty"`
	ModDate    *time.Time `gorm:"column:moddate;type:datetime" json:"mod_date,omitempty"`
	Level      *string    `gorm:"column:Level;type:varchar(3)" json:"level,omitempty" validate:"omitempty,max=3"`
	Mc1        *string    `gorm:"column:Mc1;type:nvarchar(100)" json:"mc1,omitempty" validate:"omitempty,max=100"`
	Mc2        *string    `gorm:"column:Mc2;type:nvarchar(100)" json:"mc2,omitempty" validate:"omitempty,max=100"`
	Mc3        *string    `gorm:"column:Mc3;type:nvarchar(100)" json:"mc3,omitempty" validate:"omitempty,max=100"`
	Mc4        *string    `gorm:"column:Mc4;type:nvarchar(100)" json:"mc4,omitempty" validate:"omitempty,max=100"`
	Mc5        *string    `gorm:"column:Mc5;type:nvarchar(100)" json:"mc5,omitempty" validate:"omitempty,max=100"`
	Deadline   *time.Time `gorm:"column:deadline;type:date" json:"deadline,omitempty"`

	KelasName  *AppLookup `gorm:"foreignKey:KodeKelas;references:LookupValue;constraint:false" json:"kelas_name,omitempty" validate:"-"`   // CategoryId = B_KELASKHUSUS
	FotangName *AppLookup `gorm:"foreignKey:KodeFotang;references:LookupValue;constraint:false" json:"fotang_name,omitempty" validate:"-"` // CategoryId = B_FOTHANG
}

func (Kelas) TableName() string { return "T_TRX_KELAS" } // DonasiSxy represents table [dbo].[T_SXY_TRANSAKSI]

type KelasResponse struct {
	TrxId      string   `gorm:"column:trxid" json:"trx_id"`
	KodeKelas  string   `gorm:"column:kodekelas" json:"kode_kelas"`
	StartDate  DateOnly `gorm:"column:startdate" json:"start_date"`
	EndDate    DateOnly `gorm:"column:enddate" json:"end_date"`
	KodeFotang string   `gorm:"column:kodefotang" json:"kode_fotang"`
	Lokasi     string   `gorm:"column:lokasi" json:"lokasi"`
	Pic        string   `gorm:"column:PIC" json:"pic"`
	Keterangan string   `gorm:"column:keterangan" json:"keterangan"`
	KelasDesc  string   `gorm:"column:KelasDesc" json:"kelas_desc"`
	FotangDesc string   `gorm:"column:FotangDesc" json:"fotang_desc"`
}

type KelasPeserta struct {
	DetailId        int32      `gorm:"primaryKey;column:detailid;type:int;not null" json:"detail_id"`
	TrxId           int32      `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	IdPeserta       *int32     `gorm:"column:idpeserta;type:int" json:"id_peserta" validate:"required"`
	Sumbangan       *float64   `gorm:"column:sumbangan;type:decimal(13,2)" json:"sumbangan" validate:"omitempty"`
	Barang          *string    `gorm:"column:barang;type:varchar(100)" json:"barang" validate:"omitempty,max=100"`
	TimKerja        *string    `gorm:"column:timkerja;type:varchar(3)" json:"tim_kerja" validate:"omitempty,max=3"`
	Keterangan      *string    `gorm:"column:keterangan;type:nvarchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Status          *bool      `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct          *string    `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy           *int32     `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate         *time.Time `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`
	Lulus           *bool      `gorm:"column:lulus;type:bit" json:"lulus" validate:"omitempty"`
	KeteranganLulus *string    `gorm:"column:keteranganlulus;type:varchar(200)" json:"keterangan_lulus" validate:"omitempty,max=200"`
	Anak            *string    `gorm:"column:Anak;type:varchar(30)" json:"anak" validate:"omitempty,max=30"`
	Suster          *string    `gorm:"column:Suster;type:varchar(30)" json:"suster" validate:"omitempty,max=30"`
	Menginap        *string    `gorm:"column:Menginap;type:varchar(30)" json:"menginap" validate:"omitempty,max=30"`
	MakananPagi     *string    `gorm:"column:MakananPagi;type:varchar(30)" json:"makanan_pagi" validate:"omitempty,max=30"`
	MakananSiang    *string    `gorm:"column:MakananSiang;type:varchar(30)" json:"makanan_siang" validate:"omitempty,max=30"`
	MakananMalam    *string    `gorm:"column:MakananMalam;type:varchar(30)" json:"makanan_malam" validate:"omitempty,max=30"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

// TableName menentukan nama tabel secara eksplisit di database
func (KelasPeserta) TableName() string {
	return "T_TRX_KELAS_PESERTA"
}

type KelasPesertaList struct {
	DetailId        string  `gorm:"column:detailid" json:"detailid"`
	TrxId           string  `gorm:"column:trxid" json:"trx_id"`
	Keterangan      string  `gorm:"column:keterangan" json:"keterangan"`
	IdPeserta       string  `gorm:"column:idpeserta" json:"id_peserta"`
	Sumbangan       float32 `gorm:"column:sumbangan" json:"sumbangan"`
	Barang          string  `gorm:"column:barang" json:"barang"`
	TimKerja        string  `gorm:"column:timkerja" json:"tim_kerja"`
	ID              int64   `gorm:"column:id" json:"id"`
	Kode            string  `gorm:"column:kode" json:"kode"`
	Marga           string  `gorm:"column:marga" json:"marga"`
	NamaIndonesia   string  `gorm:"column:namaindonesia" json:"nama_indonesia"`
	NamaMandarin    string  `gorm:"column:namamandarin" json:"nama_mandarin"`
	FotangAktifDesc string  `gorm:"column:FotangAktifDesc" json:"fotang_aktif_desc"`
	Anak            string  `gorm:"column:Anak" json:"nak"`
	Suster          string  `gorm:"column:Suster" json:"suster"`
	Menginap        string  `gorm:"column:Menginap" json:"menginap"`
	MakananPagi     string  `gorm:"column:MakananPagi" json:"makanan_pagi"`
	MakananSiang    string  `gorm:"column:MakananSiang" json:"makanan_siang"`
	MakananMalam    string  `gorm:"column:MakananMalam" json:"makanan_malam"`
}

type KelasPesertaResponse struct {
	DetailId         string   `gorm:"column:detailid" json:"detailid"`
	TrxId            string   `gorm:"column:trxid" json:"trx_id"`
	IdPeserta        string   `gorm:"column:idpeserta" json:"id_peserta"`
	NamaIndonesia    string   `gorm:"column:namaindonesia" json:"nama_indonesia"`
	NamaMandarin     string   `gorm:"column:namamandarin" json:"nama_mandarin"`
	FotangAktifDesc  string   `gorm:"column:FotangAktifDesc" json:"fotang_aktif_desc"`
	FotangCiuTaoDesc string   `gorm:"column:FotangCiuTaoDesc" json:"fotang_ciutao_desc"`
	TanggalCiuTaoInt DateOnly `gorm:"column:tanggalciutaoint" json:"tanggal_ciu_tao_int"`
	Pengajak         string   `gorm:"column:pengajak" json:"pengajak"`
	Penanggung       string   `gorm:"column:penanggung" json:"penanggung"`
	Lulus            IntBool  `gorm:"column:lulus" json:"lulus"`
	KeteranganLulus  string   `gorm:"column:keteranganlulus" json:"keterangan_lulus"`
	Ikrar1           IntBool  `gorm:"column:ikrar1" json:"ikrar1"`
	Ikrar2           IntBool  `gorm:"column:ikrar2" json:"ikrar2"`
	Ikrar3           IntBool  `gorm:"column:ikrar3" json:"ikrar3"`
	Ikrar4           IntBool  `gorm:"column:ikrar4" json:"ikrar4"`
	Ikrar5           IntBool  `gorm:"column:ikrar5" json:"ikrar5"`
	Ikrar6           IntBool  `gorm:"column:ikrar6" json:"ikrar6"`
}

type KelasPengabdi struct {
	DetailId       int32      `gorm:"primaryKey;column:detailid;type:int;not null" json:"detail_id"`
	TrxId          int32      `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	IdPengabdi     *int32     `gorm:"column:idpengabdi;type:int" json:"id_pengabdi" validate:"required"`
	Sumbangan      *float64   `gorm:"column:sumbangan;type:decimal(13,2)" json:"sumbangan" validate:"omitempty"`
	Barang         *string    `gorm:"column:barang;type:varchar(100)" json:"barang" validate:"omitempty,max=100"`
	TimKerja       *string    `gorm:"column:timkerja;type:varchar(3)" json:"tim_kerja" validate:"omitempty,max=3"`
	TimKerjaReport *string    `gorm:"column:timkerjareport;type:varchar(3)" json:"tim_kerja_report" validate:"omitempty,max=3"`
	Keterangan     *string    `gorm:"column:keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=500"`
	Status         *bool      `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct         *string    `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy          *int32     `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate        *time.Time `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`
	Hari           *string    `gorm:"column:hari;type:varchar(30)" json:"hari" validate:"omitempty,max=30"`
	SubKerja       *string    `gorm:"column:SubKerja;type:varchar(3)" json:"sub_kerja" validate:"omitempty,max=3"`
	Anak           *string    `gorm:"column:Anak;type:varchar(30)" json:"anak" validate:"omitempty,max=30"`
	Suster         *string    `gorm:"column:Suster;type:varchar(30)" json:"suster" validate:"omitempty,max=30"`
	Menginap       *string    `gorm:"column:Menginap;type:varchar(30)" json:"menginap" validate:"omitempty,max=30"`
	MakananPagi    *string    `gorm:"column:MakananPagi;type:varchar(30)" json:"makanan_pagi" validate:"omitempty,max=30"`
	MakananSiang   *string    `gorm:"column:MakananSiang;type:varchar(30)" json:"makanan_siang" validate:"omitempty,max=30"`
	MakananMalam   *string    `gorm:"column:MakananMalam;type:varchar(30)" json:"makanan_malam" validate:"omitempty,max=30"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

// TableName menentukan nama tabel secara eksplisit di database
func (KelasPengabdi) TableName() string {
	return "T_TRX_KELAS_PENGABDI"
}

type KelasTopik struct {
	DetailId      int32      `gorm:"primaryKey;column:detailid;type:int;not null" json:"detail_id"`
	TrxId         int32      `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	KodeTopik     *string    `gorm:"column:kodetopik;type:varchar(20)" json:"kode_topik" validate:"omitempty,max=20"`
	Urutan        *int32     `gorm:"column:urutan;type:int" json:"urutan" validate:"omitempty"`
	TopikDate     *time.Time `gorm:"column:topikdate;type:date" json:"topik_date" validate:"omitempty"`
	Penceramah    *int32     `gorm:"column:penceramah;type:int" json:"penceramah" validate:"omitempty"`
	PenceramahExt *string    `gorm:"column:penceramahext;type:nvarchar(100)" json:"penceramah_ext" validate:"omitempty,max=100"`
	Keterangan    *string    `gorm:"column:keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Status        *bool      `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct        *string    `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy         *int32     `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate       *time.Time `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`
	Durasi        *int32     `gorm:"column:Durasi;type:int" json:"durasi" validate:"omitempty"`
	Penterjemah   *string    `gorm:"column:Penterjemah;type:nvarchar(200)" json:"penterjemah" validate:"omitempty,max=200"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasTopik) TableName() string {
	return "T_TRX_KELAS_TOPIK"
}

type KelasKendaraan struct {
	DetailId      int32      `gorm:"primaryKey;column:detailid;type:int;not null" json:"detail_id"`
	TrxId         int32      `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	NoPolisi      *string    `gorm:"column:nopolisi;type:varchar(15)" json:"no_polisi" validate:"omitempty,max=15"`
	Pengendara    *string    `gorm:"column:pengendara;type:nvarchar(100)" json:"pengendara" validate:"omitempty,max=100"`
	TipeKendaraan *string    `gorm:"column:tipekendaraan;type:varchar(50)" json:"tipe_kendaraan" validate:"omitempty,max=50"`
	Fotang        *string    `gorm:"column:fotang;type:char(3)" json:"fotang" validate:"omitempty,len=3"`
	Hari          *string    `gorm:"column:hari;type:varchar(30)" json:"hari" validate:"omitempty,max=30"`
	Keterangan    *string    `gorm:"column:keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Status        *bool      `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct        *string    `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy         *int32     `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate       *time.Time `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasKendaraan) TableName() string {
	return "T_TRX_KELAS_KENDARAAN"
}

type KelasDonasi struct {
	DetailId int32      `gorm:"primaryKey;column:DetailId;type:int;not null" json:"detail_id"`
	TrxId    int32      `gorm:"column:TrxId;type:int;not null" json:"trx_id" validate:"required"`
	Donatur  string     `gorm:"column:Donatur;type:nvarchar(100);not null" json:"donatur" validate:"required,max=100"`
	Donasi   float64    `gorm:"column:Donasi;type:decimal(15,2);not null" json:"donasi" validate:"required"`
	Status   *bool      `gorm:"column:Status;type:bit" json:"status" validate:"omitempty"`
	ModAct   *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy    *string    `gorm:"column:ModBy;type:varchar(25)" json:"mod_by" validate:"omitempty,max=25"`
	ModDate  *time.Time `gorm:"column:ModDate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasDonasi) TableName() string {
	return "T_TRX_KELAS_DONASI"
}

type KelasDonasiBarang struct {
	DetailId int32      `gorm:"primaryKey;column:DetailId;type:int;not null" json:"detail_id"`
	TrxId    int32      `gorm:"column:TrxId;type:int;not null" json:"trx_id" validate:"required"`
	Donatur  string     `gorm:"column:Donatur;type:nvarchar(100);not null" json:"donatur" validate:"required,max=100"`
	Barang   string     `gorm:"column:Barang;type:varchar(100);not null" json:"barang" validate:"required,max=100"`
	Status   *bool      `gorm:"column:Status;type:bit" json:"status" validate:"omitempty"`
	ModAct   *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy    *string    `gorm:"column:ModBy;type:varchar(25)" json:"mod_by" validate:"omitempty,max=25"`
	ModDate  *time.Time `gorm:"column:ModDate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasDonasiBarang) TableName() string {
	return "T_TRX_KELAS_DONASI_BARANG"
}

type KelasPengeluaran struct {
	DetailId   int32      `gorm:"primaryKey;column:detailid;type:int;not null" json:"detail_id"`
	TrxId      int32      `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	TimKerja   *string    `gorm:"column:timkerja;type:varchar(3)" json:"tim_kerja" validate:"omitempty,max=3"`
	Keterangan *string    `gorm:"column:keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Biaya      *float64   `gorm:"column:biaya;type:decimal(13,2)" json:"biaya" validate:"omitempty"`
	Status     *bool      `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct     *string    `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy      *int32     `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate    *time.Time `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasPengeluaran) TableName() string {
	return "T_TRX_KELAS_PENGELUARAN"
}

type KelasMusik struct {
	DetailId   int32      `gorm:"primaryKey;column:DetailId;type:int;not null" json:"detail_id"`
	TrxId      int32      `gorm:"column:TrxId;type:int;not null" json:"trx_id" validate:"required"`
	MusicId    int32      `gorm:"column:MusicId;type:int;not null" json:"music_id" validate:"required"`
	Urutan     *int32     `gorm:"column:Urutan;type:int" json:"urutan" validate:"omitempty"`
	MusicDate  *time.Time `gorm:"column:MusicDate;type:date" json:"music_date" validate:"omitempty"`
	Keterangan *string    `gorm:"column:Keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Status     *bool      `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct     *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy      *int32     `gorm:"column:ModBy;type:int" json:"mod_by" validate:"omitempty"`
	ModDate    *time.Time `gorm:"column:ModDate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasMusik) TableName() string {
	return "T_TRX_MUSIK"
}

type KelasAbsensi struct {
	Id        int32      `gorm:"primaryKey;column:Id;type:int;not null" json:"id"`
	TrxId     int32      `gorm:"column:TrxId;type:int;not null" json:"trx_id" validate:"required"`
	TrxDate   time.Time  `gorm:"column:TrxDate;type:date;not null" json:"trx_date" validate:"required"`
	IdPeserta int32      `gorm:"column:IdPeserta;type:int;not null" json:"id_peserta" validate:"required"`
	Status    *bool      `gorm:"column:Status;type:bit;not null" json:"status" validate:"required"`
	ModAct    *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy     *int32     `gorm:"column:ModBy;type:int" json:"mod_by" validate:"omitempty"`
	ModDate   *time.Time `gorm:"column:ModDate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasAbsensi) TableName() string {
	return "T_TRX_KELAS_ABSENSI"
}

type DonasiSxy struct {
	ID              int32     `gorm:"primaryKey;column:id" json:"id"`
	NoKwitansi      string    `gorm:"column:nokwitansi" json:"no_kwitansi" validate:"required,max=50"`
	Tanggal         time.Time `gorm:"column:tanggal" json:"tanggal"`
	Donatur         int32     `gorm:"column:donatur" json:"donatur_id"`
	Penggalang      int32     `gorm:"column:penggalang" json:"penggalang_id"`
	Jumlah          float64   `gorm:"column:jumlah" json:"jumlah" validate:"omitempty,gte=0"` // Maps NUMERIC(18,0) cleanly
	TipeSumbangan   int32     `gorm:"column:tipesumbangan" json:"tipe_sumbangan"`
	NoKupon         string    `gorm:"column:nokupon" json:"no_kupon" validate:"omitempty,max=50"`
	Keterangan      string    `gorm:"column:keterangan" json:"keterangan" validate:"omitempty,max=500"`
	Status          bool      `gorm:"column:STATUS" json:"status"`
	CreatedBy       int32     `gorm:"column:createdby" json:"created_by"`
	CreatedDate     time.Time `gorm:"column:createddate" json:"created_date"`
	UpdatedBy       int32     `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate     time.Time `gorm:"column:updateddate" json:"updated_date"`
	TanggalTransfer time.Time `gorm:"column:tanggaltransfer" json:"tanggal_transfer"`
	AtasNama        string    `gorm:"column:atasnama" json:"atas_nama" validate:"omitempty,max=500"`
	TtkSent         bool      `gorm:"column:ttksent" json:"ttk_sent"`
}

type DonasiSxyResponse struct {
	ID                 int32      `gorm:"primaryKey;column:id" json:"id"`
	NoKwitansi         string     `gorm:"column:nokwitansi" json:"no_kwitansi"`
	NoKupon            *string    `gorm:"column:nokupon" json:"no_kupon,omitempty"`
	Tanggal            DateOnly   `gorm:"column:tanggal" json:"tanggal"`
	Keterangan         *string    `gorm:"column:keterangan" json:"keterangan,omitempty"`
	Penggalang         int32      `gorm:"column:penggalang" json:"penggalang_id"`
	TipeSumbangan      int32      `gorm:"column:tipesumbangan" json:"tipe_sumbangan"`
	Jumlah             float64    `gorm:"column:jumlah" json:"jumlah"` // Maps NUMERIC(18,0) cleanly
	TipeSummbanganDesc *string    `gorm:"column:tipesumbangandesc" json:"tipe_sumbangan_desc,omitempty"`
	NamaPenggalang     *string    `gorm:"column:namapenggalang" json:"nama_penggalang,omitempty"`
	Donatur            int32      `gorm:"column:donatur" json:"donatur_id"`
	NamaDonatur        *string    `gorm:"column:namadonatur" json:"nama_donatur,omitempty"`
	TanggalTransfer    *time.Time `gorm:"column:tanggaltransfer" json:"tanggal_transfer,omitempty"`
	AtasNama           *string    `gorm:"column:atasnama" json:"atas_nama,omitempty"`
	EmailPenggalang    *string    `gorm:"column:emailpenggalang" json:"email_penggalang,omitempty"`
}

func (DonasiSxy) TableName() string { return "T_SXY_TRANSAKSI" }

type WorkMapping struct {
	ID        int64  `gorm:"primaryKey;column:Id;autoIncrement" json:"id"`
	Divisi    string `gorm:"column:Divisi;type:varchar(50)" json:"divisi"`
	SubDivisi string `gorm:"column:SubDivisi;type:varchar(50)" json:"sub_divisi"`
	Status    bool   `gorm:"column:Status;type:bit;default:1" json:"status"`
}

func (WorkMapping) TableName() string { return "T_BUS_WORK_MAPPING" }
