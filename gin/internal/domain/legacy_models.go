package domain

import "time"

// LoginGroup maps the legacy table T_Login_Group.
type LoginGroup struct {
	GroupID    int       `gorm:"column:GroupId;primaryKey;autoIncrement" json:"groupId"`
	GroupName  string    `gorm:"column:GroupName;size:50" json:"groupName"`
	GroupDesc  string    `gorm:"column:GroupDesc;size:350" json:"groupDesc"`
	FlagActive bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate    time.Time `gorm:"column:ModDate" json:"modDate"`
}

func (LoginGroup) TableName() string { return "T_Login_Group" }

// LoginMenu maps the legacy table T_Login_Menu.
type LoginMenu struct {
	MenuID     int    `gorm:"column:MenuId;primaryKey;autoIncrement" json:"menuId"`
	ParentID   int    `gorm:"column:ParentId" json:"parentId"`
	MenuName   string `gorm:"column:MenuName;size:200" json:"menuName"`
	PageURL    string `gorm:"column:PageUrl;size:250" json:"pageUrl"`
	Squence    int    `gorm:"column:Squence" json:"squence"`
	FlagActive bool   `gorm:"column:FlagActive" json:"flagActive"`
}

func (LoginMenu) TableName() string { return "T_Login_Menu" }

// LoginMenuGroup maps the legacy bridge table T_Login_Menu_Group.
type LoginMenuGroup struct {
	MenuGroupID int  `gorm:"column:MenuGroupId;primaryKey;autoIncrement" json:"menuGroupId"`
	GroupID     int  `gorm:"column:GroupId" json:"groupId"`
	MenuID      int  `gorm:"column:MenuId" json:"menuId"`
	RInsert     bool `gorm:"column:RInsert" json:"rInsert"`
	reporting   bool `gorm:"column:RReporting" json:"rReporting"`
	REdit       bool `gorm:"column:REdit" json:"rEdit"`
	RDelete     bool `gorm:"column:RDelete" json:"rDelete"`
	RReject     bool `gorm:"column:RReject" json:"rReject"`
	FlagUse     bool `gorm:"column:FlagUse" json:"flagUse"`
}

func (LoginMenuGroup) TableName() string { return "T_Login_Menu_Group" }

// LoginMst maps the legacy table T_Login_Mst.
type LoginMst struct {
	LoginID     int        `gorm:"column:LoginId;primaryKey;autoIncrement" json:"loginId"`
	UserName    string     `gorm:"column:UserName;size:100" json:"userName"`
	LoginName   string     `gorm:"column:LoginName;size:150" json:"loginName"`
	GroupID     int        `gorm:"column:GroupId" json:"groupId"`
	Group       LoginGroup `gorm:"foreignKey:GroupID;references:GroupId" json:"group"`
	FlagActive  bool       `gorm:"column:FlagActive" json:"flagActive"`
	IsWarehouse bool       `gorm:"column:IsWarehouse" json:"isWarehouse"`
	ModDate     time.Time  `gorm:"column:ModDate" json:"modDate"`
}

func (LoginMst) TableName() string { return "T_Login_Mst" }

// SubWarehouse maps the legacy table T_WH_SUBWH_MST.
type SubWarehouse struct {
	SUBWHID    int    `gorm:"column:SUBWHID;primaryKey;autoIncrement" json:"subWhId"`
	SubWhName  string `gorm:"column:SubWhName;size:150" json:"subWhName"`
	SubWhType  string `gorm:"column:SubWhType;size:3" json:"subWhType"`
	FlagActive bool   `gorm:"column:FlagActive" json:"flagActive"`
}

func (SubWarehouse) TableName() string { return "T_WH_SUBWH_MST" }

// Umat maps T_BUS_UMAT.
type Umat struct {
	ID         int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Kode       string `gorm:"column:kode;size:50" json:"kode"`
	Nama       string `gorm:"column:nama;size:150" json:"nama"`
	FlagActive bool   `gorm:"column:flagactive" json:"flagActive"`
}

func (Umat) TableName() string { return "T_BUS_UMAT" }

// Topic maps T_BUS_TOPIC.
type Topic struct {
	TopicCode  string    `gorm:"column:TopicCode;primaryKey;size:20" json:"topicCode"`
	TopicName  string    `gorm:"column:TopicName;size:150" json:"topicName"`
	FlagActive bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate    time.Time `gorm:"column:ModDate" json:"modDate"`
}

func (Topic) TableName() string { return "T_BUS_TOPIC" }

// Activity maps T_BUS_EVENT.
type Activity struct {
	EventCode  string    `gorm:"column:EventCode;primaryKey;size:10" json:"eventCode"`
	EventName  string    `gorm:"column:EventName;size:150" json:"eventName"`
	FlagActive bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate    time.Time `gorm:"column:ModDate" json:"modDate"`
}

func (Activity) TableName() string { return "T_BUS_EVENT" }

// TimKerja maps T_APP_LOOKUP for B_POSISI.
type TimKerja struct {
	LookupID          string    `gorm:"column:LookupId;primaryKey;size:25" json:"lookupId"`
	LookupCategory    string    `gorm:"column:LookupCategory;size:25" json:"lookupCategory"`
	LookupDescription string    `gorm:"column:LookupDescription;size:150" json:"lookupDescription"`
	FlagActive        bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate           time.Time `gorm:"column:ModDate" json:"modDate"`
}

func (TimKerja) TableName() string { return "T_APP_LOOKUP" }

// TahunCiuTao maps T_BUS_TAHUN_CIUTAO.
type TahunCiuTao struct {
	TahunMandarin string    `gorm:"column:TahunMandarin;primaryKey;size:20" json:"tahunMandarin"`
	Description   string    `gorm:"column:description;size:50" json:"description"`
	FlagActive    bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate       time.Time `gorm:"column:ModDate" json:"modDate"`
}

func (TahunCiuTao) TableName() string { return "T_BUS_TAHUN_CIUTAO" }

// PenggalangDana maps T_SXY_MST_PENGGALANG.
type PenggalangDana struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"column:Name;size:150" json:"name"`
	FlagActive bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate    time.Time `gorm:"column:updateddate" json:"updatedDate"`
}

func (PenggalangDana) TableName() string { return "T_SXY_MST_PENGGALANG" }

// SxyDonatur maps T_SXY_MST_DONATUR.
type SxyDonatur struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"column:Name;size:150" json:"name"`
	FlagActive bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate    time.Time `gorm:"column:updateddate" json:"updatedDate"`
}

func (SxyDonatur) TableName() string { return "T_SXY_MST_DONATUR" }

// Kelas maps T_APP_LOOKUP for B_KELASKHUSUS.
type Kelas struct {
	LookupID          string    `gorm:"column:LookupId;primaryKey;size:25" json:"lookupId"`
	LookupCategory    string    `gorm:"column:LookupCategory;size:25" json:"lookupCategory"`
	LookupDescription string    `gorm:"column:LookupDescription;size:150" json:"lookupDescription"`
	FlagActive        bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate           time.Time `gorm:"column:ModDate" json:"modDate"`
}

func (Kelas) TableName() string { return "T_APP_LOOKUP" }

// DonasiSxy maps T_SXY_TRANSAKSI.
type DonasiSxy struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"column:Name;size:150" json:"name"`
	FlagActive bool      `gorm:"column:FlagActive" json:"flagActive"`
	ModDate    time.Time `gorm:"column:ttksent" json:"ttkSent"`
}

func (DonasiSxy) TableName() string { return "T_SXY_TRANSAKSI" }
