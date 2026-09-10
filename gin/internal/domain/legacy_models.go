package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ==========================================
// FIELD TRANSLATIONS DICTIONARY & HELPER
// ==========================================

// FieldTranslations maps struct namespace (e.g. "Umat.NamaIndonesia"),
// struct field name (e.g. "NamaIndonesia"), or json tag name (e.g. "nama_indonesia")
// to custom user-friendly label translations.
var FieldTranslations = map[string]string{
	"Umat.NamaIndonesia":        "Nama Indonesia",
	"Umat.NamaMandarin":         "Nama Mandarin",
	"Umat.TanggalChiutaoInt":    "Tanggal Ciu Tao (Masehi)",
	"Umat.TanggalChiutaoMan":    "Tanggal Ciu Tao (Imlek)",
	"Umat.TahunChiutaoMandarin": "Tahun Ciu Tao (Mandarin)",
	"Umat.WaktuChiutaoMandarin": "Waktu Ciu Tao (Mandarin)",
	"Umat.TanggalAncuo":         "Tanggal An cuo",
	"Umat.NamaCetyaRumah":       "Nama Cetya Rumah",
	"Kelas.StartDate":           "Tanggal Mulai",
	"Kelas.EndDate":             "Tanggal Selesai",
	"StartDate":                 "Tanggal Mulai",
	"EndDate":                   "Tanggal Selesai",
	"start_date":                "Tanggal Mulai",
	"end_date":                  "Tanggal Selesai",
	"tanggal_chiutao_int":       "Tanggal Ciu Tao",
	"waktu_chiutao_mandarin":    "Waktu Ciu Tao",
}

// GetFieldLabel translates a field identifier (struct namespace like "Umat.NamaIndonesia",
// struct field name like "NamaIndonesia", or json key like "nama_indonesia") to a human-readable label.
// If not found in FieldTranslations map, it defaults to splitting the model field name into space-separated words.
func GetFieldLabel(identifier string) string {
	if identifier == "" {
		return ""
	}

	// 1. Check exact identifier in FieldTranslations map
	if translated, ok := FieldTranslations[identifier]; ok && translated != "" {
		return translated
	}

	// If identifier has a struct namespace prefix (e.g. "Umat.NamaIndonesia"), check field name part
	fieldName := identifier
	if idx := strings.LastIndex(identifier, "."); idx != -1 {
		fieldName = identifier[idx+1:]
		if translated, ok := FieldTranslations[fieldName]; ok && translated != "" {
			return translated
		}
	}

	// 2. Default fallback: space separate model field name
	return FormatFieldName(fieldName)
}

// FormatFieldName converts PascalCase/camelCase or snake_case string into space-separated words.
// E.g., "NamaIndonesia" -> "Nama Indonesia", "nama_indonesia" -> "Nama Indonesia".
func FormatFieldName(s string) string {
	if s == "" {
		return ""
	}

	// Handle snake_case if present
	if strings.Contains(s, "_") {
		parts := strings.Split(s, "_")
		var b strings.Builder
		b.Grow(len(s))
		for i, part := range parts {
			if part == "" {
				continue
			}
			if i > 0 && b.Len() > 0 {
				b.WriteByte(' ')
			}
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			b.WriteString(string(runes))
		}
		return b.String()
	}

	// Handle PascalCase / camelCase
	runes := []rune(s)
	n := len(runes)
	var b strings.Builder
	b.Grow(n + 5) // Pre-allocate capacity to minimize GC pressure

	for i := 0; i < n; i++ {
		if i > 0 {
			prev := runes[i-1]
			curr := runes[i]

			if unicode.IsUpper(curr) {
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					b.WriteByte(' ')
				} else if unicode.IsUpper(prev) && i+1 < n && unicode.IsLower(runes[i+1]) {
					b.WriteByte(' ')
				}
			} else if unicode.IsDigit(curr) && !unicode.IsDigit(prev) {
				b.WriteByte(' ')
			}
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}

// ==========================================
// 1. ADMIN & AUTHORIZATION MODELS
// ==========================================

// Admin represents table [dbo].[T_Login_Mst]
type Admin struct {
	ID           int32    `gorm:"primaryKey;column:LoginId" json:"id"`
	Username     string   `gorm:"column:Username" json:"username" validate:"required,max=50"`
	Password     string   `gorm:"column:Password" json:"-" validate:"omitempty,max=50"`
	GroupId      int32    `gorm:"column:GroupId" json:"group_id,omitempty"`
	Email        *string  `gorm:"column:email" json:"email,omitempty" validate:"omitempty,max=50,email"`
	PhoneNumber  *string  `gorm:"column:PhoneNumber" json:"phone_number,omitempty" validate:"omitempty,max=50"`
	ImgUrl       string   `gorm:"column:ImgUrl" json:"img_url" validate:"omitempty,max=50"`
	FlagUse      bool     `gorm:"column:FlagUse" json:"flag_use"`
	DateStart    DateOnly `gorm:"column:DateStart" json:"date_start,omitempty"`
	DateEnd      DateOnly `gorm:"column:DateEnd" json:"date_end,omitempty"`
	LoginDesc    *string  `gorm:"column:LoginDesc" json:"login_desc,omitempty" validate:"omitempty,max=350"`
	LastLogin    DateTime `gorm:"column:LastLogin" json:"last_login"`
	DepartmentId int32    `gorm:"column:DepartmentId" json:"department_id"`
	IsWarehouse  bool     `gorm:"column:IsWarehouse" json:"is_warehouse"`

	AdminGroup AdminGroup     `gorm:"foreignKey:GroupId;references:GroupId;constraint:false" json:"admin_group" validate:"-"`
	Department *DepartmentMst `gorm:"foreignKey:DepartmentId;references:DepartmentId;constraint:false" json:"department,omitempty" validate:"-"`
}

func (Admin) TableName() string { return "T_Login_Mst" }

type AdminMatrix struct {
	CruID       int64  `gorm:"primaryKey;autoIncrement:false;column:CRUID;type:bigint;not null" json:"cru_id"`
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
	DepartmentId   int16     `gorm:"primaryKey;autoIncrement:false;column:DepartmentId;type:smallint;not null" json:"department_id"`
	DepartmentCode *string   `gorm:"column:DepartmentCode;type:varchar(25)" json:"department_code" validate:"omitempty,max=25"`
	Departmentname *string   `gorm:"column:Departmentname;type:varchar(100)" json:"department_name" validate:"omitempty,max=100"`
	Status         bool      `gorm:"column:Status;type:bit;not null" json:"status"`
	ModAct         *string   `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,max=1"`
	ModBy          *string   `gorm:"column:ModBy;type:varchar(25)" json:"mod_by" validate:"omitempty,max=25"`
	ModDate        *DateTime `gorm:"column:ModDate;type:datetime" json:"mod_date"`
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
	SubWhId         int32    `gorm:"primaryKey;autoIncrement:false;column:SUBWHID" json:"sub_wh_id"`
	WhId            int64    `gorm:"column:WHID" json:"wh_id"`
	FullName        string   `gorm:"column:FULL_NAME" json:"full_name" validate:"required,max=100"`
	Pic             string   `gorm:"column:PIC" json:"pic" validate:"omitempty,max=75"`
	DocCode         string   `gorm:"column:DOCCODE" json:"doc_code" validate:"omitempty,max=50"`
	CruId           int64    `gorm:"column:CRUID" json:"cru_id"`
	UpdateUId       int64    `gorm:"column:UPDATEUID" json:"update_uid"`
	LstUpdate       DateTime `gorm:"column:LSTUPDATE" json:"lst_update"`
	FlagProductions bool     `gorm:"column:FlagProductions" json:"flag_productions"`
	SubWhType       string   `gorm:"column:SubWhType" json:"sub_wh_type" validate:"omitempty,max=3"`
}

func (AdminSubWarehouse) TableName() string { return "T_WH_SUBWH_MST" }

// ==========================================
// 2. CORE SYSTEM & MASTER RECORD MODELS
// ==========================================

// Umat represents table [dbo].[T_BUS_UMAT]
type Umat struct {
	ID                   int32       `gorm:"primaryKey;autoIncrement:false;column:id;autoIncrement:false" json:"id"`
	Kode                 *string     `gorm:"column:kode" json:"kode" validate:"omitempty,max=50"`
	Alias                *string     `gorm:"column:alias" json:"alias" validate:"omitempty,max=50"`
	NamaIndonesia        string      `gorm:"column:namaindonesia" json:"nama_indonesia" validate:"required,max=50"`
	Marga                *string     `gorm:"column:marga" json:"marga" validate:"omitempty,max=10"`
	NamaMandarin         *string     `gorm:"column:namamandarin" json:"nama_mandarin" validate:"omitempty,max=50"`
	Alamat               string      `gorm:"column:alamat" json:"alamat" validate:"omitempty,max=200"`
	Alamat2              *string     `gorm:"column:alamat2" json:"alamat2" validate:"omitempty,max=100"`
	Telepon              *string     `gorm:"column:telepon" json:"telepon" validate:"omitempty,max=50"`
	Mobile               *string     `gorm:"column:mobile" json:"mobile" validate:"omitempty,max=50"`
	TempatLahir          string      `gorm:"column:tempatlahir" json:"tempat_lahir" validate:"omitempty,max=50"`
	TanggalLahir         DateOnly    `gorm:"column:tanggallahir" json:"tanggal_lahir"`
	Usia                 *int32      `gorm:"column:usia" json:"usia" validate:"omitempty,gte=0,lte=150"`
	Wilayah              *string     `gorm:"column:wilayah" json:"wilayah" validate:"omitempty,max=100"`
	JenisKelamin         string      `gorm:"column:jeniskelamin" json:"jenis_kelamin" validate:"required,omitempty,max=3"`
	JenisKelaminInfo     *AppLookup  `gorm:"foreignKey:JenisKelamin;references:LookupValue;constraint:false" json:"jenis_kelamin_info,omitempty" validate:"-"`
	Pekerjaan            *string     `gorm:"column:pekerjaan" json:"pekerjaan" validate:"omitempty,max=3"`
	Pendidikan           *string     `gorm:"column:pendidikan" json:"pendidikan" validate:"omitempty,max=3"`
	TanggalChiutaoInt    DateOnly    `gorm:"column:tanggalchiutaoint" json:"tanggal_chiutao_int"`
	TanggalChiutaoMan    *string     `gorm:"column:tanggalchiutaoman" json:"tanggal_chiutao_man" validate:"omitempty,max=6"`
	TahunChiutaoMandarin string      `gorm:"column:tahunchiutaomandarin" json:"tahun_chiutao_mandarin" validate:"omitempty,max=20"`
	WaktuChiutaoMandarin string      `gorm:"column:waktuchiutaomandarin" json:"waktu_chiutao_mandarin" validate:"omitempty,max=3"`
	Pengajak             *FlexString `gorm:"column:pengajak" json:"pengajak" validate:"omitempty,max=20"`
	// PengajakUmat         *Umat      `gorm:"foreignKey:Pengajak;references:Kode" json:"pengajak_umat,omitempty"`
	PengajakManual *string     `gorm:"column:pengajakmanual" json:"pengajak_manual" validate:"omitempty,max=50"`
	Penanggung     *FlexString `gorm:"column:penanggung" json:"penanggung" validate:"omitempty,max=20"`
	// PenanggungUmat       *Umat      `gorm:"foreignKey:Penanggung;references:Kode" json:"penanggung_umat,omitempty"`
	PenanggungManual    *string   `gorm:"column:penanggungmanual" json:"penanggung_manual" validate:"omitempty,max=50"`
	Tcs                 string    `gorm:"column:tcs" json:"tcs" validate:"omitempty,max=3"`
	UangPahala          float64   `gorm:"column:uangpahala" json:"uang_pahala" validate:"omitempty,gte=0"`
	FotangChiutao       string    `gorm:"column:fotangciutao" json:"fotang_chiutao" validate:"omitempty,max=3"`
	FotangAktif         string    `gorm:"column:fotangaktif" json:"fotang_aktif" validate:"omitempty,max=3"`
	Sd2                 bool      `gorm:"column:sd2" json:"sd2"`
	TempatSd2           *string   `gorm:"column:tempatsd2" json:"tempat_sd2" validate:"omitempty,max=3"`
	TanggalSd2          *DateOnly `gorm:"column:tanggalsd2" json:"tanggal_sd2"`
	Sd3                 bool      `gorm:"column:sd3" json:"sd3"`
	TempatSd3           *string   `gorm:"column:tempatsd3" json:"tempat_sd3" validate:"omitempty,max=3"`
	TanggalSd3          *DateOnly `gorm:"column:tanggalsd3" json:"tanggal_sd3"`
	KelasUmum           *string   `gorm:"column:kelasumum" json:"kelas_umum" validate:"omitempty,max=3"`
	KelasKhusus         *string   `gorm:"column:kelaskhusus" json:"kelas_khusus" validate:"omitempty,max=3"`
	ChingKhou           bool      `gorm:"column:chingkhou" json:"ching_khou"`
	TanggalChingKhou    *DateOnly `gorm:"column:tanggalchingkhou" json:"tanggal_ching_khou"`
	TanggalAncuo        *DateOnly `gorm:"column:tanggalancuo" json:"tanggal_ancuo"`
	NamaCetyaRumah      *string   `gorm:"column:namacetyarumah" json:"nama_cetya_rumah" validate:"omitempty,max=50"`
	Meninggal           bool      `gorm:"column:meninggal" json:"meninggal"`
	TanggalMeninggal    *DateOnly `gorm:"column:tanggalmeninggal" json:"tanggal_meninggal"`
	TimKerja            *string   `gorm:"column:timkerja" json:"tim_kerja" validate:"omitempty,max=3"`
	Posisi              *string   `gorm:"column:posisi" json:"posisi" validate:"omitempty,max=3"`
	StatusUmat          *string   `gorm:"column:statusumat" json:"status_umat" validate:"omitempty,max=3"`
	Keterangan          *string   `gorm:"column:keterangan" json:"keterangan" validate:"omitempty,max=200"`
	Email               *string   `gorm:"column:email" json:"email" validate:"omitempty,max=50,email"`
	ImagePath           *string   `gorm:"column:imagepath" json:"image_path" validate:"omitempty,max=50"`
	Status              bool      `gorm:"column:status" json:"status"`
	ModAct              string    `gorm:"column:modact" json:"mod_act" validate:"omitempty,max=1"`
	ModBy               int32     `gorm:"column:modby" json:"mod_by"`
	ModDate             DateTime  `gorm:"column:moddate" json:"mod_date"`
	Ikrar1              bool      `gorm:"column:ikrar1" json:"ikrar_1"`
	Ikrar2              bool      `gorm:"column:ikrar2" json:"ikrar_2"`
	Ikrar3              bool      `gorm:"column:ikrar3" json:"ikrar_3"`
	Ikrar4              bool      `gorm:"column:ikrar4" json:"ikrar_4"`
	Ikrar5              bool      `gorm:"column:ikrar5" json:"ikrar_5"`
	Ikrar6              bool      `gorm:"column:ikrar6" json:"ikrar_6"`
	RenChaiPan          bool      `gorm:"column:RenChaiPan" json:"ren_chai_pan"`
	TanggalRenChaiPan   *DateOnly `gorm:"column:TanggalRenChaiPan" json:"tanggal_ren_chai_pan"`
	LienCiangPan        bool      `gorm:"column:LienCiangPan" json:"lien_ciang_pan"`
	TanggalLienCiangPan *DateOnly `gorm:"column:TanggalLienCiangPan" json:"tanggal_lien_ciang_pan"`
	CiangYenPan         bool      `gorm:"column:CiangYenPan" json:"ciang_yen_pan"`
	TanggalCiangYenPan  *DateOnly `gorm:"column:TanggalCiangYenPan" json:"tanggal_ciang_yen_pan"`
	ActiveStatus        *bool     `gorm:"column:ActiveStatus" json:"active_status"`
	NamaFotangLain      *string   `gorm:"column:NamaFotangLain" json:"nama_fotang_lain" validate:"omitempty,max=50"`
	NamaTcsLain         *string   `gorm:"column:NamaTcsLain" json:"nama_tcs_lain" validate:"omitempty,max=50"`
	KodeBuku            *string   `gorm:"column:KodeBuku" json:"kode_buku" validate:"omitempty,max=50"`
}

func (u *Umat) UnmarshalJSON(data []byte) error {
	type Alias Umat
	aux := &struct {
		Ikrar1Alt *bool `json:"ikrar1"`
		Ikrar2Alt *bool `json:"ikrar2"`
		Ikrar3Alt *bool `json:"ikrar3"`
		Ikrar4Alt *bool `json:"ikrar4"`
		Ikrar5Alt *bool `json:"ikrar5"`
		Ikrar6Alt *bool `json:"ikrar6"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Ikrar1Alt != nil {
		u.Ikrar1 = *aux.Ikrar1Alt
	}
	if aux.Ikrar2Alt != nil {
		u.Ikrar2 = *aux.Ikrar2Alt
	}
	if aux.Ikrar3Alt != nil {
		u.Ikrar3 = *aux.Ikrar3Alt
	}
	if aux.Ikrar4Alt != nil {
		u.Ikrar4 = *aux.Ikrar4Alt
	}
	if aux.Ikrar5Alt != nil {
		u.Ikrar5 = *aux.Ikrar5Alt
	}
	if aux.Ikrar6Alt != nil {
		u.Ikrar6 = *aux.Ikrar6Alt
	}
	return nil
}

func (Umat) TableName() string { return "T_BUS_UMAT" }

type UmatFoto struct {
	FileID      int32      `gorm:"column:FileID;primaryKey;autoIncrement:false" json:"file_id"`
	Id          *int32     `gorm:"column:Id" json:"id"`
	DocPath     *string    `gorm:"column:DocPath;type:varchar(1000)" json:"doc_path" validate:"omitempty,max=1000"`
	DocFile     *string    `gorm:"column:DocFile;type:varchar(1000)" json:"doc_file" validate:"omitempty,max=1000"`
	DocFileName *string    `gorm:"column:DocFileName;type:varchar(150)" json:"doc_file_name" validate:"omitempty,max=150"`
	DocDesc     *string    `gorm:"column:DocDesc;type:varchar(2000)" json:"doc_desc" validate:"omitempty,max=2000"`
	Status      *bool      `gorm:"column:Status" json:"status"`
	ModBy       *int32     `gorm:"column:ModBy" json:"mod_by"`
	ModAct      *string    `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModDate     *time.Time `gorm:"column:ModDate" json:"mod_date"`
}

// TableName menentukan nama tabel secara eksplisit di SQL Server
func (UmatFoto) TableName() string {
	return "T_BUS_UMAT_FOTO"
}

type AppLookup struct {
	LookupId          string    `gorm:"primaryKey;autoIncrement:false;column:LookupId;type:varchar(25);not null" json:"lookup_id" validate:"omitempty,max=25"`
	CategoryId        *string   `gorm:"column:CategoryId;type:varchar(25)" json:"category_id,omitempty" validate:"omitempty,max=25"`
	LookupValue       *string   `gorm:"column:LookupValue;type:varchar(50)" json:"lookup_value,omitempty" validate:"omitempty,max=50"`
	LookupDescription *string   `gorm:"column:LookupDescription;type:nvarchar(150)" json:"lookup_description,omitempty" validate:"omitempty,max=150"`
	Status            *bool     `gorm:"column:Status;type:bit" json:"status,omitempty"`
	ModAct            *string   `gorm:"column:ModAct;type:char(1)" json:"mod_act,omitempty" validate:"omitempty,max=1"`
	ModBy             *string   `gorm:"column:ModBy;type:varchar(25)" json:"mod_by,omitempty" validate:"omitempty,max=25"`
	ModDate           *DateTime `gorm:"column:ModDate;type:datetime" json:"mod_date,omitempty"`

	Category *AppLookupCategory `gorm:"foreignKey:CategoryId;references:CategoryId;constraint:false" json:"category,omitempty" validate:"-"`
}

// TableName menentukan nama tabel secara eksplisit di database
func (AppLookup) TableName() string {
	return "T_APP_LOOKUP"
}

type AppLookupCategory struct {
	CategoryId          string    `gorm:"primaryKey;autoIncrement:false;column:CategoryId;type:varchar(25);not null" json:"category_id" validate:"omitempty,max=25"`
	CategoryType        string    `gorm:"column:CategoryType;type:char(1);not null" json:"category_type" validate:"omitempty,max=1"`
	CategoryDescription string    `gorm:"column:CategoryDescription;type:varchar(150);not null" json:"category_description" validate:"omitempty,max=150"`
	Status              bool      `gorm:"column:Status;type:bit;not null" json:"status"`
	ModAct              *string   `gorm:"column:ModAct;type:char(1)" json:"mod_act,omitempty" validate:"omitempty,max=1"`
	ModBy               *string   `gorm:"column:ModBy;type:varchar(25)" json:"mod_by,omitempty" validate:"omitempty,max=25"`
	ModDate             *DateTime `gorm:"column:ModDate;type:datetime" json:"mod_date,omitempty"`
}

// TableName memaksa GORM menggunakan nama tabel spesifik tanpa pluralisasi otomatis
func (AppLookupCategory) TableName() string {
	return "T_APP_LOOKUPCATEGORY"
}

// Topic represents table [dbo].[T_BUS_TOPIC]
type Topic struct {
	TopicCode     string   `gorm:"primaryKey;autoIncrement:false;column:TopicCode" json:"topic_code" validate:"required,max=20"`
	TopicName     string   `gorm:"column:TopicName" json:"topic_name" validate:"required,max=300"`
	TopicCategory string   `gorm:"column:TopicCategory" json:"topic_category" validate:"omitempty,max=3"`
	Description   string   `gorm:"column:Description" json:"description" validate:"omitempty,max=200"`
	Status        bool     `gorm:"column:Status" json:"status"`
	ModAct        string   `gorm:"column:ModAct" json:"mod_act" validate:"omitempty,max=1"`
	ModBy         string   `gorm:"column:ModBy" json:"mod_by" validate:"omitempty,max=25"`
	ModDate       DateTime `gorm:"column:ModDate" json:"mod_date"`

	TopicCategoryInfo *AppLookup `gorm:"foreignKey:TopicCategory;references:LookupValue;constraint:false" json:"topic_category_info,omitempty" validate:"-"`
}

func (Topic) TableName() string { return "T_BUS_TOPIC" }

// Activity represents table [dbo].[T_BUS_EVENT]
type Activity struct {
	EventCode     string   `gorm:"primaryKey;autoIncrement:false;column:EventCode" json:"event_code" validate:"required,max=10"`
	EventName     string   `gorm:"column:EventName" json:"event_name" validate:"required,max=100"`
	EventCategory string   `gorm:"column:EventCategory" json:"event_category" validate:"omitempty,max=3"`
	Description   string   `gorm:"column:Description" json:"description" validate:"omitempty,max=200"`
	Status        bool     `gorm:"column:Status" json:"status"`
	ModAct        string   `gorm:"column:ModAct" json:"mod_act" validate:"omitempty,max=1"`
	ModBy         string   `gorm:"column:ModBy" json:"mod_by" validate:"omitempty,max=25"`
	ModDate       DateTime `gorm:"column:ModDate" json:"mod_date"`

	EventCategoryInfo *AppLookup `gorm:"foreignKey:EventCategory;references:LookupValue;constraint:false" json:"event_category_info,omitempty" validate:"-"`
}

func (Activity) TableName() string { return "T_BUS_EVENT" } // TimKerja represents records inside [dbo].[T_APP_LOOKUP] filtered by CategoryId = 'B_POSISI'

type TimKerja struct {
	LookupId          string   `gorm:"primaryKey;autoIncrement:false;column:LookupId" json:"lookup_id" validate:"omitempty,max=25"`
	CategoryId        string   `gorm:"column:CategoryId;default:B_POSISI" json:"category_id" validate:"omitempty,max=25"`
	LookupValue       string   `gorm:"column:LookupValue" json:"lookup_value" validate:"required,max=50"`
	LookupDescription string   `gorm:"column:LookupDescription" json:"lookup_description" validate:"required,max=150"`
	Status            bool     `gorm:"column:Status" json:"status"`
	ModAct            string   `gorm:"column:ModAct" json:"mod_act" validate:"omitempty,max=1"`
	ModBy             string   `gorm:"column:ModBy" json:"mod_by" validate:"omitempty,max=25"`
	ModDate           DateTime `gorm:"column:ModDate" json:"mod_date"`
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
	if d.IsZero() {
		return nil, nil
	}
	return d.Format("2006-01-02"), nil
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

// Buat custom type untuk string fleksibel
type FlexString string

// UnmarshalJSON menangani konversi otomatis dari int/float/string ke FlexString
func (fs *FlexString) UnmarshalJSON(data []byte) error {
	// 1. Coba unmarshal sebagai string biasa
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fs = FlexString(s)
		return nil
	}

	// 2. Jika gagal, coba unmarshal sebagai integer (int64)
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*fs = FlexString(strconv.FormatInt(i, 10))
		return nil
	}

	// 3. Jika gagal, coba unmarshal sebagai float
	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*fs = FlexString(strconv.FormatFloat(f, 'f', -1, 64))
		return nil
	}

	// Jika bernilai null atau tipe data lain yang tidak dikenali
	*fs = ""
	return nil
}

// DateTime adalah tipe kustom untuk menangani serialisasi/deserialisasi waktu dengan format "YYYY-MM-DD HH:mm:ss"
type DateTime struct {
	time.Time
}

func NewDateTime(t time.Time) DateTime {
	return DateTime{Time: t}
}

func NowDateTime() DateTime {
	return DateTime{Time: time.Now()}
}

// Scan mendeteksi data dari database SQL Server
func (dt *DateTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		dt.Time = v
		return nil
	case string:
		return dt.parseString(v)
	case []byte:
		return dt.parseString(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into DateTime", value)
	}
}

func (dt *DateTime) parseString(s string) error {
	s = strings.TrimSpace(s)
	if s == "" || s == "null" {
		return nil
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999",
		"2006-01-02 15:04:05.999999",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05.999Z07:00",
		"2006-01-02T15:04:05.999999Z07:00",
		"2006-01-02",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			dt.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse date string %q into DateTime", s)
}

// Value untuk menyimpan kembali ke database jika diperlukan
func (dt DateTime) Value() (driver.Value, error) {
	if dt.IsZero() {
		return nil, nil
	}
	return dt.Format("2006-01-02 15:04:05"), nil
}

// UnmarshalJSON mendeteksi format string JSON (misal "2022-10-08 09:12:39" atau "2022-10-08T09:12:39.423Z")
func (dt *DateTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	return dt.parseString(s)
}

// MarshalJSON menampilkan format YYYY-MM-DD HH:mm:ss
func (dt DateTime) MarshalJSON() ([]byte, error) {
	if dt.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(dt.Format("2006-01-02 15:04:05"))
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
	TahunMandarin string   `gorm:"primaryKey;autoIncrement:false;column:TahunMandarin" json:"tahun_mandarin" validate:"required,max=20"`
	StartDate     DateOnly `gorm:"column:StartDate" json:"start_date"`
	EndDate       DateOnly `gorm:"column:EndDate" json:"end_date"`
	ModAct        string   `gorm:"column:ModAct" json:"mod_act" validate:"omitempty,max=1"`
	ModBy         string   `gorm:"column:ModBy" json:"mod_by" validate:"omitempty,max=25"`
	ModDate       DateTime `gorm:"column:ModDate" json:"mod_date"`
	Status        bool     `gorm:"column:status" json:"status"`
	Description   string   `gorm:"column:description" json:"description" validate:"omitempty,max=50"`
}

func (TahunCiuTao) TableName() string { return "T_BUS_TAHUN_CIUTAO" } // PenggalangDana represents table [dbo].[T_SXY_MST_PENGGALANG]

type PenggalangDana struct {
	ID            int32    `gorm:"primaryKey;autoIncrement:false;column:id" json:"id"`
	No            string   `gorm:"column:no" json:"no" validate:"required,max=50"`
	Nama          string   `gorm:"column:nama" json:"nama" validate:"required,max=50"`
	Mandarin      string   `gorm:"column:mandarin" json:"mandarin" validate:"omitempty,max=50"`
	Keterangan    string   `gorm:"column:keterangan" json:"keterangan" validate:"omitempty,max=255"`
	LookupFothang int32    `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Alamat        string   `gorm:"column:alamat" json:"alamat" validate:"omitempty,max=255"`
	Telepon       string   `gorm:"column:telepon" json:"telepon" validate:"omitempty,max=50"`
	Mobile        string   `gorm:"column:mobile" json:"mobile" validate:"omitempty,max=50"`
	Email         string   `gorm:"column:email" json:"email" validate:"omitempty,max=100,email"`
	Status        bool     `gorm:"column:STATUS" json:"status"`
	CreatedBy     int32    `gorm:"column:createdby" json:"created_by"`
	CreatedDate   DateTime `gorm:"column:createddate" json:"created_date"`
	UpdatedBy     int32    `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate   DateTime `gorm:"column:updateddate" json:"updated_date"`
}
type PenggalangDanaResponse struct {
	ID            int32  `gorm:"primaryKey;autoIncrement:false;column:id" json:"id"`
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
	ID            int32    `gorm:"primaryKey;autoIncrement:false;column:id" json:"id"`
	No            string   `gorm:"column:no" json:"no" validate:"required,max=10"`
	Nama          string   `gorm:"column:nama" json:"nama" validate:"required,max=50"`
	Mandarin      string   `gorm:"column:mandarin" json:"mandarin" validate:"omitempty,max=50"`
	Keterangan    string   `gorm:"column:keterangan" json:"keterangan" validate:"omitempty,max=255"`
	LookupFothang int32    `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Alamat        string   `gorm:"column:alamat" json:"alamat" validate:"omitempty,max=255"`
	Telepon       string   `gorm:"column:telepon" json:"telepon" validate:"omitempty,max=50"`
	Mobile        string   `gorm:"column:mobile" json:"mobile" validate:"omitempty,max=50"`
	Email         string   `gorm:"column:email" json:"email" validate:"omitempty,max=100,email"`
	Status        bool     `gorm:"column:STATUS" json:"status"`
	CreatedBy     int32    `gorm:"column:createdby" json:"created_by"`
	CreatedDate   DateTime `gorm:"column:createddate" json:"created_date"`
	UpdatedBy     int32    `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate   DateTime `gorm:"column:updateddate" json:"updated_date"`
}

type SxyDonaturResponse struct {
	ID            int32    `gorm:"primaryKey;autoIncrement:false;column:id" json:"id"`
	No            string   `gorm:"column:no" json:"no"`
	Nama          string   `gorm:"column:nama" json:"nama"`
	Mandarin      string   `gorm:"column:mandarin" json:"mandarin"`
	Keterangan    string   `gorm:"column:keterangan" json:"keterangan"`
	LookupFothang int32    `gorm:"column:lookup_fothang" json:"lookup_fothang"`
	Fotang        string   `gorm:"column:fothang" json:"fotang"`
	Alamat        string   `gorm:"column:alamat" json:"alamat"`
	Telepon       string   `gorm:"column:telepon" json:"telepon"`
	Mobile        string   `gorm:"column:mobile" json:"mobile"`
	Email         string   `gorm:"column:email" json:"email"`
	Status        bool     `gorm:"column:STATUS" json:"status"`
	CreatedBy     int32    `gorm:"column:createdby" json:"created_by"`
	CreatedDate   DateTime `gorm:"column:createddate" json:"created_date"`
	UpdatedBy     int32    `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate   DateTime `gorm:"column:updateddate" json:"updated_date"`
}

type SxyDonasiReport struct {
	NoKwitansi      string   `gorm:"column:nokwitansi" json:"no_kwitansi"`
	Tanggal         DateOnly `gorm:"column:tanggal" json:"tanggal"`
	TanggalTransfer DateOnly `gorm:"column:tanggaltransfer" json:"tanggal_transfer"`

	AtasNama       string  `gorm:"column:atasnama" json:"atas_nama"`
	Donatur        string  `gorm:"column:donatur" json:"donatur"`
	PenggalangDana string  `gorm:"column:penggalang_dana" json:"penggalang_dana"`
	TipeSumbangan  string  `gorm:"column:tipe_sumbangan" json:"tipe_sumbangan"`
	Jumlah         float64 `gorm:"column:jumlah" json:"jumlah"`
	NoKupon        string  `gorm:"column:nokupon" json:"nokupon"`
	Keterangan     string  `gorm:"column:keterangan" json:"keterangan"`
	Fotang         string  `gorm:"column:Fotang" json:"fotang"`
}

func (SxyDonatur) TableName() string { return "T_SXY_MST_DONATUR" }

// ==========================================// 3. SPECIAL CUSTOM TYPE & DONATION MODELS// ==========================================// Kelas represents records inside [dbo].[T_APP_LOOKUP] filtered by CategoryId = 'B_KELASKHUSUS'

type Kelas struct {
	TrxId      int32     `gorm:"primaryKey;autoIncrement:false;column:trxid;type:int;not null" json:"trx_id,omitempty"`
	KodeKelas  *string   `gorm:"column:kodekelas;type:varchar(3)" json:"kode_kelas,omitempty" validate:"required,max=3"`
	StartDate  *DateOnly `gorm:"column:startdate;type:date" json:"start_date,omitempty" validate:"omitempty"`
	EndDate    *DateOnly `gorm:"column:enddate;type:date" json:"end_date,omitempty" validate:"omitempty,gtefield=StartDate"`
	KodeFotang *string   `gorm:"column:kodefotang;type:varchar(3)" json:"kode_fotang,omitempty" validate:"omitempty,max=3"`
	Lokasi     *string   `gorm:"column:lokasi;type:varchar(50)" json:"lokasi,omitempty" validate:"omitempty,max=50"`
	PIC        *string   `gorm:"column:PIC;type:nvarchar(100)" json:"pic,omitempty" validate:"omitempty,max=100"`
	Keterangan *string   `gorm:"column:keterangan;type:varchar(200)" json:"keterangan,omitempty" validate:"omitempty,max=200"`
	Status     *bool     `gorm:"column:status;type:bit" json:"status,omitempty"`
	ModAct     *string   `gorm:"column:modact;type:char(1)" json:"mod_act,omitempty" validate:"omitempty,max=1"`
	ModBy      *int32    `gorm:"column:modby;type:int" json:"mod_by,omitempty"`
	ModDate    *DateTime `gorm:"column:moddate;type:datetime" json:"mod_date,omitempty"`
	Level      *string   `gorm:"column:Level;type:varchar(3)" json:"level,omitempty" validate:"omitempty,max=3"`
	Mc1        *string   `gorm:"column:Mc1;type:nvarchar(100)" json:"mc1,omitempty" validate:"omitempty,max=100"`
	Mc2        *string   `gorm:"column:Mc2;type:nvarchar(100)" json:"mc2,omitempty" validate:"omitempty,max=100"`
	Mc3        *string   `gorm:"column:Mc3;type:nvarchar(100)" json:"mc3,omitempty" validate:"omitempty,max=100"`
	Mc4        *string   `gorm:"column:Mc4;type:nvarchar(100)" json:"mc4,omitempty" validate:"omitempty,max=100"`
	Mc5        *string   `gorm:"column:Mc5;type:nvarchar(100)" json:"mc5,omitempty" validate:"omitempty,max=100"`
	Deadline   *DateOnly `gorm:"column:deadline;type:date" json:"deadline,omitempty"`

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
	DetailId        int32     `gorm:"primaryKey;autoIncrement:false;column:detailid;type:int;not null" json:"detail_id"`
	TrxId           int32     `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	IdPeserta       *int32    `gorm:"column:idpeserta;type:int" json:"id_peserta" validate:"required"`
	Sumbangan       *float64  `gorm:"column:sumbangan;type:decimal(13,2)" json:"sumbangan" validate:"omitempty"`
	Barang          *string   `gorm:"column:barang;type:varchar(100)" json:"barang" validate:"omitempty,max=100"`
	TimKerja        *string   `gorm:"column:timkerja;type:varchar(3)" json:"tim_kerja" validate:"omitempty,max=3"`
	Keterangan      *string   `gorm:"column:keterangan;type:nvarchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Status          *bool     `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct          *string   `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy           *int32    `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate         *DateTime `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`
	Lulus           *bool     `gorm:"column:lulus;type:bit" json:"lulus" validate:"omitempty"`
	KeteranganLulus *string   `gorm:"column:keteranganlulus;type:varchar(200)" json:"keterangan_lulus" validate:"omitempty,max=200"`
	Anak            *string   `gorm:"column:Anak;type:varchar(30)" json:"anak" validate:"omitempty,max=30"`
	Suster          *string   `gorm:"column:Suster;type:varchar(30)" json:"suster" validate:"omitempty,max=30"`
	Menginap        *string   `gorm:"column:Menginap;type:varchar(30)" json:"menginap" validate:"omitempty,max=30"`
	MakananPagi     *string   `gorm:"column:MakananPagi;type:varchar(30)" json:"makanan_pagi" validate:"omitempty,max=30"`
	MakananSiang    *string   `gorm:"column:MakananSiang;type:varchar(30)" json:"makanan_siang" validate:"omitempty,max=30"`
	MakananMalam    *string   `gorm:"column:MakananMalam;type:varchar(30)" json:"makanan_malam" validate:"omitempty,max=30"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
	Umat  *Umat  `gorm:"foreignKey:IdPeserta;references:ID;constraint:false" json:"umat,omitempty" validate:"-"`
}

// TableName menentukan nama tabel secara eksplisit di database
func (KelasPeserta) TableName() string {
	return "T_TRX_KELAS_PESERTA"
}

// UnmarshalJSON custom unmarshaler for KelasPeserta to enforce single umat ID
func (p *KelasPeserta) UnmarshalJSON(data []byte) error {
	type auxKelasPeserta struct {
		DetailId        int32           `json:"detail_id"`
		TrxId           int32           `json:"trx_id"`
		IdPeserta       json.RawMessage `json:"id_peserta"`
		IdPesertaAlt    json.RawMessage `json:"idpeserta"`
		Sumbangan       *float64        `json:"sumbangan"`
		Barang          *string         `json:"barang"`
		TimKerja        *string         `json:"tim_kerja"`
		Keterangan      *string         `json:"keterangan"`
		Status          *bool           `json:"status"`
		ModAct          *string         `json:"mod_act"`
		ModBy           *int32          `json:"mod_by"`
		ModDate         *DateTime       `json:"mod_date"`
		Lulus           *bool           `json:"lulus"`
		KeteranganLulus *string         `json:"keterangan_lulus"`
		Anak            *string         `json:"anak"`
		Suster          *string         `json:"suster"`
		Menginap        *string         `json:"menginap"`
		MakananPagi     *string         `json:"makanan_pagi"`
		MakananSiang    *string         `json:"makanan_siang"`
		MakananMalam    *string         `json:"makanan_malam"`
	}

	var aux auxKelasPeserta
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	rawID := aux.IdPeserta
	if len(rawID) == 0 {
		rawID = aux.IdPesertaAlt
	}

	if len(rawID) > 0 && string(rawID) != "null" {
		trimmedID := strings.TrimSpace(string(rawID))
		if strings.HasPrefix(trimmedID, "[") {
			return fmt.Errorf("idpeserta: idpeserta should be single value of umat")
		}

		var singleInt int32
		if err := json.Unmarshal(rawID, &singleInt); err == nil {
			p.IdPeserta = &singleInt
		} else {
			var singleStr string
			if err := json.Unmarshal(rawID, &singleStr); err == nil {
				if val, err := strconv.ParseInt(strings.TrimSpace(singleStr), 10, 32); err == nil {
					v32 := int32(val)
					p.IdPeserta = &v32
				} else {
					return fmt.Errorf("idpeserta: idpeserta should be single value of umat")
				}
			} else {
				return fmt.Errorf("idpeserta: idpeserta should be single value of umat")
			}
		}
	}

	p.DetailId = aux.DetailId
	p.TrxId = aux.TrxId
	p.Sumbangan = aux.Sumbangan
	p.Barang = aux.Barang
	p.TimKerja = aux.TimKerja
	p.Keterangan = aux.Keterangan
	p.Status = aux.Status
	p.ModAct = aux.ModAct
	p.ModBy = aux.ModBy
	p.ModDate = aux.ModDate
	p.Lulus = aux.Lulus
	p.KeteranganLulus = aux.KeteranganLulus
	p.Anak = aux.Anak
	p.Suster = aux.Suster
	p.Menginap = aux.Menginap
	p.MakananPagi = aux.MakananPagi
	p.MakananSiang = aux.MakananSiang
	p.MakananMalam = aux.MakananMalam

	return nil
}

type KelasPesertaBulkRequest struct {
	TrxId           int32    `json:"trx_id" validate:"required"`
	IdPeserta       []int32  `json:"id_peserta" validate:"required,min=1"`
	Sumbangan       *float64 `json:"sumbangan" validate:"omitempty"`
	Barang          *string  `json:"barang" validate:"omitempty,max=100"`
	TimKerja        *string  `json:"tim_kerja" validate:"omitempty,max=3"`
	Keterangan      *string  `json:"keterangan" validate:"omitempty,max=200"`
	Status          *bool    `json:"status" validate:"omitempty"`
	Lulus           *bool    `json:"lulus" validate:"omitempty"`
	KeteranganLulus *string  `json:"keterangan_lulus" validate:"omitempty,max=200"`
	Anak            *string  `json:"anak" validate:"omitempty,max=30"`
	Suster          *string  `json:"suster" validate:"omitempty,max=30"`
	Menginap        *string  `json:"menginap" validate:"omitempty,max=30"`
	MakananPagi     *string  `json:"makanan_pagi" validate:"omitempty,max=30"`
	MakananSiang    *string  `json:"makanan_siang" validate:"omitempty,max=30"`
	MakananMalam    *string  `json:"makanan_malam" validate:"omitempty,max=30"`
}

// UnmarshalJSON custom unmarshaler for KelasPesertaBulkRequest to enforce array of umat IDs
func (p *KelasPesertaBulkRequest) UnmarshalJSON(data []byte) error {
	type auxBulkRequest struct {
		TrxId           int32           `json:"trx_id"`
		IdPeserta       json.RawMessage `json:"id_peserta"`
		IdPesertaAlt    json.RawMessage `json:"idpeserta"`
		Sumbangan       *float64        `json:"sumbangan"`
		Barang          *string         `json:"barang"`
		TimKerja        *string         `json:"tim_kerja"`
		Keterangan      *string         `json:"keterangan"`
		Status          *bool           `json:"status"`
		Lulus           *bool           `json:"lulus"`
		KeteranganLulus *string         `json:"keterangan_lulus"`
		Anak            *string         `json:"anak"`
		Suster          *string         `json:"suster"`
		Menginap        *string         `json:"menginap"`
		MakananPagi     *string         `json:"makanan_pagi"`
		MakananSiang    *string         `json:"makanan_siang"`
		MakananMalam    *string         `json:"makanan_malam"`
	}

	var aux auxBulkRequest
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	rawID := aux.IdPeserta
	if len(rawID) == 0 {
		rawID = aux.IdPesertaAlt
	}

	if len(rawID) > 0 && string(rawID) != "null" {
		trimmedID := strings.TrimSpace(string(rawID))
		if !strings.HasPrefix(trimmedID, "[") {
			return fmt.Errorf("idpeserta: idpeserta must array")
		}

		var sliceInt []int32
		if err := json.Unmarshal(rawID, &sliceInt); err == nil {
			p.IdPeserta = sliceInt
		} else {
			var sliceStr []string
			if err := json.Unmarshal(rawID, &sliceStr); err == nil {
				parsedList := make([]int32, 0, len(sliceStr))
				for _, s := range sliceStr {
					if val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 32); err == nil {
						parsedList = append(parsedList, int32(val))
					} else {
						return fmt.Errorf("idpeserta: idpeserta format is invalid")
					}
				}
				p.IdPeserta = parsedList
			} else {
				return fmt.Errorf("idpeserta: idpeserta format is invalid")
			}
		}
	}

	p.TrxId = aux.TrxId
	p.Sumbangan = aux.Sumbangan
	p.Barang = aux.Barang
	p.TimKerja = aux.TimKerja
	p.Keterangan = aux.Keterangan
	p.Status = aux.Status
	p.Lulus = aux.Lulus
	p.KeteranganLulus = aux.KeteranganLulus
	p.Anak = aux.Anak
	p.Suster = aux.Suster
	p.Menginap = aux.Menginap
	p.MakananPagi = aux.MakananPagi
	p.MakananSiang = aux.MakananSiang
	p.MakananMalam = aux.MakananMalam

	return nil
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
	DetailId       int32     `gorm:"primaryKey;autoIncrement:false;column:detailid;type:int;not null" json:"detail_id"`
	TrxId          int32     `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	IdPengabdi     *int32    `gorm:"column:idpengabdi;type:int" json:"id_pengabdi" validate:"required"`
	Sumbangan      *float64  `gorm:"column:sumbangan;type:decimal(13,2)" json:"sumbangan" validate:"omitempty"`
	Barang         *string   `gorm:"column:barang;type:varchar(100)" json:"barang" validate:"omitempty,max=100"`
	TimKerja       *string   `gorm:"column:timkerja;type:varchar(3)" json:"tim_kerja" validate:"omitempty,max=3"`
	TimKerjaReport *string   `gorm:"column:timkerjareport;type:varchar(3)" json:"tim_kerja_report" validate:"omitempty,max=3"`
	Keterangan     *string   `gorm:"column:keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=500"`
	Status         *bool     `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct         *string   `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy          *int32    `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate        *DateTime `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`
	Hari           *string   `gorm:"column:hari;type:varchar(30)" json:"hari" validate:"omitempty,max=30"`
	SubKerja       *string   `gorm:"column:SubKerja;type:varchar(3)" json:"sub_kerja" validate:"omitempty,max=3"`
	Anak           *string   `gorm:"column:Anak;type:varchar(30)" json:"anak" validate:"omitempty,max=30"`
	Suster         *string   `gorm:"column:Suster;type:varchar(30)" json:"suster" validate:"omitempty,max=30"`
	Menginap       *string   `gorm:"column:Menginap;type:varchar(30)" json:"menginap" validate:"omitempty,max=30"`
	MakananPagi    *string   `gorm:"column:MakananPagi;type:varchar(30)" json:"makanan_pagi" validate:"omitempty,max=30"`
	MakananSiang   *string   `gorm:"column:MakananSiang;type:varchar(30)" json:"makanan_siang" validate:"omitempty,max=30"`
	MakananMalam   *string   `gorm:"column:MakananMalam;type:varchar(30)" json:"makanan_malam" validate:"omitempty,max=30"`

	NamaIndonesia   *string `gorm:"column:nama_indonesia;->" json:"nama_indonesia,omitempty"`
	NamaMandarin    *string `gorm:"column:nama_mandarin;->" json:"nama_mandarin,omitempty"`
	FotangAktifDesc *string `gorm:"column:fotang_aktif_desc;->" json:"fotang_aktif_desc,omitempty"`
	TimKerjaDesc    *string `gorm:"column:tim_kerja_desc;->" json:"tim_kerja_desc,omitempty"`
	SubKerjaDesc    *string `gorm:"column:sub_kerja_desc;->" json:"sub_kerja_desc,omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

// TableName menentukan nama tabel secara eksplisit di database
func (KelasPengabdi) TableName() string {
	return "T_TRX_KELAS_PENGABDI"
}

type KelasTopik struct {
	DetailId      int32     `gorm:"primaryKey;autoIncrement:false;column:detailid;type:int;not null" json:"detail_id"`
	TrxId         int32     `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	KodeTopik     *string   `gorm:"column:kodetopik;type:varchar(20)" json:"kode_topik" validate:"omitempty,max=20"`
	Urutan        *int32    `gorm:"column:urutan;type:int" json:"urutan" validate:"omitempty"`
	TopikDate     *DateOnly `gorm:"column:topikdate;type:date" json:"topik_date" validate:"omitempty"`
	Penceramah    *int32    `gorm:"column:penceramah;type:int" json:"penceramah" validate:"omitempty"`
	PenceramahExt *string   `gorm:"column:penceramahext;type:nvarchar(100)" json:"penceramah_ext" validate:"omitempty,max=100"`
	Keterangan    *string   `gorm:"column:keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Status        *bool     `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct        *string   `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy         *int32    `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate       *DateTime `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`
	Durasi        *int32    `gorm:"column:Durasi;type:int" json:"durasi" validate:"omitempty"`
	Penterjemah   *string   `gorm:"column:Penterjemah;type:nvarchar(200)" json:"penterjemah" validate:"omitempty,max=200"`

	NamaTopik     *string `gorm:"column:nama_topik;->" json:"nama_topik,omitempty"`
	TopicCategory *string `gorm:"column:topic_category;->" json:"topic_category,omitempty"`
	TopicDesc     *string `gorm:"column:topic_desc;->" json:"topic_desc,omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasTopik) TableName() string {
	return "T_TRX_KELAS_TOPIK"
}

type KelasKendaraan struct {
	DetailId      int32     `gorm:"primaryKey;autoIncrement:false;column:detailid;type:int;not null" json:"detail_id"`
	TrxId         int32     `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	NoPolisi      *string   `gorm:"column:nopolisi;type:varchar(15)" json:"no_polisi" validate:"omitempty,max=15"`
	Pengendara    *string   `gorm:"column:pengendara;type:nvarchar(100)" json:"pengendara" validate:"omitempty,max=100"`
	TipeKendaraan *string   `gorm:"column:tipekendaraan;type:varchar(50)" json:"tipe_kendaraan" validate:"omitempty,max=50"`
	Fotang        *string   `gorm:"column:fotang;type:char(3)" json:"fotang" validate:"omitempty,len=3"`
	Hari          *string   `gorm:"column:hari;type:varchar(30)" json:"hari" validate:"omitempty,max=30"`
	Keterangan    *string   `gorm:"column:keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Status        *bool     `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct        *string   `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy         *int32    `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate       *DateTime `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasKendaraan) TableName() string {
	return "T_TRX_KELAS_KENDARAAN"
}

type KelasDonasi struct {
	DetailId int32     `gorm:"primaryKey;autoIncrement:false;column:DetailId;type:int;not null" json:"detail_id"`
	TrxId    int32     `gorm:"column:TrxId;type:int;not null" json:"trx_id" validate:"required"`
	Donatur  string    `gorm:"column:Donatur;type:nvarchar(100);not null" json:"donatur" validate:"required,max=100"`
	Donasi   float64   `gorm:"column:Donasi;type:decimal(15,2);not null" json:"donasi" validate:"required"`
	Status   *bool     `gorm:"column:Status;type:bit" json:"status" validate:"omitempty"`
	ModAct   *string   `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy    *string   `gorm:"column:ModBy;type:varchar(25)" json:"mod_by" validate:"omitempty,max=25"`
	ModDate  *DateTime `gorm:"column:ModDate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasDonasi) TableName() string {
	return "T_TRX_KELAS_DONASI"
}

type KelasDonasiBarang struct {
	DetailId int32     `gorm:"primaryKey;autoIncrement:false;column:DetailId;type:int;not null" json:"detail_id"`
	TrxId    int32     `gorm:"column:TrxId;type:int;not null" json:"trx_id" validate:"required"`
	Donatur  string    `gorm:"column:Donatur;type:nvarchar(100);not null" json:"donatur" validate:"required,max=100"`
	Barang   string    `gorm:"column:Barang;type:varchar(100);not null" json:"barang" validate:"required,max=100"`
	Status   *bool     `gorm:"column:Status;type:bit" json:"status" validate:"omitempty"`
	ModAct   *string   `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy    *string   `gorm:"column:ModBy;type:varchar(25)" json:"mod_by" validate:"omitempty,max=25"`
	ModDate  *DateTime `gorm:"column:ModDate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasDonasiBarang) TableName() string {
	return "T_TRX_KELAS_DONASI_BARANG"
}

type KelasPengeluaran struct {
	DetailId   int32     `gorm:"primaryKey;autoIncrement:false;column:detailid;type:int;not null" json:"detail_id"`
	TrxId      int32     `gorm:"column:trxid;type:int;not null" json:"trx_id" validate:"required"`
	TimKerja   *string   `gorm:"column:timkerja;type:varchar(3)" json:"tim_kerja" validate:"omitempty,max=3"`
	Keterangan *string   `gorm:"column:keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Biaya      *float64  `gorm:"column:biaya;type:decimal(13,2)" json:"biaya" validate:"omitempty"`
	Status     *bool     `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct     *string   `gorm:"column:modact;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy      *int32    `gorm:"column:modby;type:int" json:"mod_by" validate:"omitempty"`
	ModDate    *DateTime `gorm:"column:moddate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasPengeluaran) TableName() string {
	return "T_TRX_KELAS_PENGELUARAN"
}

type KelasMusik struct {
	DetailId   int32     `gorm:"primaryKey;autoIncrement:false;column:DetailId;type:int;not null" json:"detail_id"`
	TrxId      int32     `gorm:"column:TrxId;type:int;not null" json:"trx_id" validate:"required"`
	MusicId    int32     `gorm:"column:MusicId;type:int;not null" json:"music_id" validate:"required"`
	Urutan     *int32    `gorm:"column:Urutan;type:int" json:"urutan" validate:"omitempty"`
	MusicDate  *DateOnly `gorm:"column:MusicDate;type:date" json:"music_date" validate:"omitempty"`
	Keterangan *string   `gorm:"column:Keterangan;type:varchar(200)" json:"keterangan" validate:"omitempty,max=200"`
	Status     *bool     `gorm:"column:status;type:bit" json:"status" validate:"omitempty"`
	ModAct     *string   `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy      *int32    `gorm:"column:ModBy;type:int" json:"mod_by" validate:"omitempty"`
	ModDate    *DateTime `gorm:"column:ModDate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasMusik) TableName() string {
	return "T_TRX_MUSIK"
}

type KelasAbsensi struct {
	Id        int32     `gorm:"primaryKey;autoIncrement:false;column:Id;type:int;not null" json:"id"`
	TrxId     int32     `gorm:"column:TrxId;type:int;not null" json:"trx_id" validate:"required"`
	TrxDate   DateTime  `gorm:"column:TrxDate;type:date;not null" json:"trx_date" validate:"required"`
	IdPeserta int32     `gorm:"column:IdPeserta;type:int;not null" json:"id_peserta" validate:"required"`
	Status    *bool     `gorm:"column:Status;type:bit;not null" json:"status" validate:"required"`
	ModAct    *string   `gorm:"column:ModAct;type:char(1)" json:"mod_act" validate:"omitempty,len=1"`
	ModBy     *int32    `gorm:"column:ModBy;type:int" json:"mod_by" validate:"omitempty"`
	ModDate   *DateTime `gorm:"column:ModDate;type:datetime" json:"mod_date" validate:"omitempty"`

	Kelas *Kelas `gorm:"foreignKey:TrxId;references:TrxId;constraint:false" json:"kelas,omitempty" validate:"-"`
}

func (KelasAbsensi) TableName() string {
	return "T_TRX_KELAS_ABSENSI"
}

type DonasiSxy struct {
	ID              int32    `gorm:"primaryKey;autoIncrement:false;column:id" json:"id"`
	NoKwitansi      string   `gorm:"column:nokwitansi" json:"no_kwitansi" validate:"required,max=50"`
	Tanggal         DateOnly `gorm:"column:tanggal" json:"tanggal"`
	Donatur         int32    `gorm:"column:donatur" json:"donatur_id"`
	Penggalang      int32    `gorm:"column:penggalang" json:"penggalang_id"`
	Jumlah          float64  `gorm:"column:jumlah" json:"jumlah" validate:"omitempty,gte=0"` // Maps NUMERIC(18,0) cleanly
	TipeSumbangan   int32    `gorm:"column:tipesumbangan" json:"tipe_sumbangan"`
	NoKupon         string   `gorm:"column:nokupon" json:"no_kupon" validate:"omitempty,max=50"`
	Keterangan      string   `gorm:"column:keterangan" json:"keterangan" validate:"omitempty,max=500"`
	Status          bool     `gorm:"column:STATUS" json:"status"`
	CreatedBy       int32    `gorm:"column:createdby" json:"created_by"`
	CreatedDate     DateTime `gorm:"column:createddate" json:"created_date"`
	UpdatedBy       int32    `gorm:"column:updatedby" json:"updated_by"`
	UpdatedDate     DateTime `gorm:"column:updateddate" json:"updated_date"`
	TanggalTransfer DateOnly `gorm:"column:tanggaltransfer" json:"tanggal_transfer"`
	AtasNama        string   `gorm:"column:atasnama" json:"atas_nama" validate:"omitempty,max=500"`
	TtkSent         bool     `gorm:"column:ttksent" json:"ttk_sent"`
}

func (DonasiSxy) TableName() string { return "T_SXY_TRANSAKSI" }

type DonasiSxyResponse struct {
	ID                int32     `gorm:"primaryKey;autoIncrement:false;column:id" json:"id"`
	NoKwitansi        string    `gorm:"column:nokwitansi" json:"no_kwitansi"`
	NoKupon           *string   `gorm:"column:nokupon" json:"no_kupon"`
	Tanggal           DateOnly  `gorm:"column:tanggal" json:"tanggal"`
	Keterangan        *string   `gorm:"column:keterangan" json:"keterangan"`
	Penggalang        int32     `gorm:"column:penggalang" json:"penggalang_id"`
	TipeSumbangan     int32     `gorm:"column:tipesumbangan" json:"tipe_sumbangan"`
	Jumlah            float64   `gorm:"column:jumlah" json:"jumlah"` // Maps NUMERIC(18,0) cleanly
	TipeSumbanganDesc *string   `gorm:"column:tipesumbangandesc" json:"tipe_sumbangan_desc"`
	NamaPenggalang    *string   `gorm:"column:namapenggalang" json:"nama_penggalang"`
	Donatur           int32     `gorm:"column:donatur" json:"donatur_id"`
	NamaDonatur       *string   `gorm:"column:namadonatur" json:"nama_donatur"`
	TanggalTransfer   *DateOnly `gorm:"column:tanggaltransfer" json:"tanggal_transfer"`
	AtasNama          *string   `gorm:"column:atasnama" json:"atas_nama"`
	EmailPenggalang   *string   `gorm:"column:emailpenggalang" json:"email_penggalang"`
}

type WorkMapping struct {
	ID        int64  `gorm:"primaryKey;autoIncrement:false;column:Id;autoIncrement" json:"id"`
	Divisi    string `gorm:"column:Divisi;type:varchar(50)" json:"divisi"`
	SubDivisi string `gorm:"column:SubDivisi;type:varchar(50)" json:"sub_divisi"`
	Status    bool   `gorm:"column:Status;type:bit;default:1" json:"status"`
}

func (WorkMapping) TableName() string { return "T_BUS_WORK_MAPPING" }

type UmatReport struct {
	// RowNo                int64    `gorm:"column:RowNo" json:"row_no"`
	Id                   int32    `gorm:"column:id" json:"id"`
	Kode                 string   `gorm:"column:Kode" json:"kode"`
	TanggalChiuTaoInt    DateOnly `gorm:"column:tanggalchiutaoint" json:"tanggal_chiu_tao_int"`
	TanggalChiuTaoMan    string   `gorm:"column:tanggalchiutaoman" json:"tanggal_chiu_tao_man"`
	TahunChiuTaoMandarin string   `gorm:"column:tahunchiutaomandarin" json:"tahun_chiu_tao_mandarin"`
	WaktuChiuTaoMandarin string   `gorm:"column:waktuchiutaomandarin" json:"waktu_chiu_tao_mandarin"`
	NamaIndonesia        string   `gorm:"column:namaindonesia" json:"nama_indonesia"`
	NamaMandarin         string   `gorm:"column:namamandarin" json:"nama_mandarin"`
	Alias                string   `gorm:"column:alias" json:"alias"`
	Alamat               string   `gorm:"column:alamat" json:"alamat"`
	Alamat2              string   `gorm:"column:alamat2" json:"alamat2"`
	UangPahala           float64  `gorm:"column:uangpahala" json:"uang_pahala"`
	UsiaThn              int32    `gorm:"column:usiathn" json:"usia_thn"`
	TanggalLahir         DateOnly `gorm:"column:tanggallahir" json:"tanggal_lahir"`
	TempatLahir          string   `gorm:"column:tempatlahir" json:"tempat_lahir"`
	PengajakManual       string   `gorm:"column:pengajakmanual" json:"pengajak_manual"`
	PenanggungManual     string   `gorm:"column:penanggungmanual" json:"penanggung_manual"`
	Tcs                  string   `gorm:"column:Tcs" json:"tcs"`
	Telepon              string   `gorm:"column:telepon" json:"telepon"`
	Mobile               string   `gorm:"column:mobile" json:"mobile"`
	Email                string   `gorm:"column:email" json:"email"`
	FotangCiuTaoDesc     string   `gorm:"column:FotangCiuTaoDesc" json:"fotang_ciu_tao_desc"`
	FotangAktifDesc      string   `gorm:"column:FotangAktifDesc" json:"fotang_aktif_desc"`
	JenisKelamin         string   `gorm:"column:JenisKelamin" json:"jenis_kelamin"`
	Wilayah              string   `gorm:"column:Wilayah" json:"wilayah"`
	PekerjaanDesc        string   `gorm:"column:PekerjaanDesc" json:"pekerjaan_desc"`
	PendidikanDesc       string   `gorm:"column:PendidikanDesc" json:"pendidikan_desc"`
	TanggalSd3           DateOnly `gorm:"column:tanggalsd3" json:"tanggal_sd3"`
	TempatSd3Desc        string   `gorm:"column:TempatSd3Desc" json:"tempat_sd3_desc"`
	TanggalChingKhou     DateOnly `gorm:"column:tanggalchingkhou" json:"tanggal_ching_khou"`
	Keterangan           string   `gorm:"column:keterangan" json:"keterangan"`
	KelasUmumDesc        string   `gorm:"column:KelasUmumDesc" json:"kelas_umum_desc"`
	KelasKhususDesc      string   `gorm:"column:KelasKhususDesc" json:"kelas_khusus_desc"`
	TanggalAnCuo         DateOnly `gorm:"column:tanggalancuo" json:"tanggal_an_cuo"`
	NamaCetyaRumah       string   `gorm:"column:namacetyarumah" json:"nama_cetya_rumah"`
	StatusUmatDesc       string   `gorm:"column:StatusUmatDesc" json:"status_umat_desc"`
	Ikrar1               bool     `gorm:"column:ikrar1" json:"ikrar1"`
	Ikrar2               bool     `gorm:"column:ikrar2" json:"ikrar2"`
	Ikrar3               bool     `gorm:"column:ikrar3" json:"ikrar3"`
	Ikrar4               bool     `gorm:"column:ikrar4" json:"ikrar4"`
	Ikrar5               bool     `gorm:"column:ikrar5" json:"ikrar5"`
	Ikrar6               bool     `gorm:"column:ikrar6" json:"ikrar6"`
	TotalRow             int64    `gorm:"column:TotalRow" json:"total_row"`
}

type KelasPesertaPrevious struct {
	IdPeserta        int32  `gorm:"column:idpeserta" json:"id_peserta"`
	Id               int32  `gorm:"column:id" json:"id"`
	Kode             string `gorm:"column:kode" json:"kode"`
	NamaIndonesia    string `gorm:"column:namaindonesia" json:"nama_indonesia"`
	NamaMandarin     string `gorm:"column:namamandarin" json:"nama_mandarin"`
	Alias            string `gorm:"column:alias" json:"alias"`
	Marga            string `gorm:"column:marga" json:"marga"`
	Alamat           string `gorm:"column:alamat" json:"alamat"`
	FotangAktif      string `gorm:"column:fotangaktif" json:"fotang_aktif"`
	FotangAktifDesc  string `gorm:"column:FotangAktifDesc" json:"fotang_aktif_desc"`
	FotangCiuTao     string `gorm:"column:fotangciutao" json:"fotang_ciu_tao"`
	FotangCiuTaoDesc string `gorm:"column:FotangCiuTaoDesc" json:"fotang_ciu_tao_desc"`
	Pengajak         string `gorm:"column:pengajak" json:"pengajak"`
	Penanggung       string `gorm:"column:penanggung" json:"penanggung"`
}

