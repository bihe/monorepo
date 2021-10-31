package store

import (
	"fmt"
	"time"
)

// --------------------------------------------------------------------------
// Entity definitions
// --------------------------------------------------------------------------

type Logintype uint

const (
	// DIRECT login to the site
	DIRECT Logintype = iota
	// FLOW or indirect login via login-flow
	FLOW
)

// LoginsEntity stores user logins
type LoginsEntity struct {
	ID        int       `gorm:"primaryKey;TYPE:int;autoIncrement;COLUMN:id"`
	User      string    `gorm:"TYPE:varchar(128);COLUMN:user;NOT NULL;INDEX:IX_LOGINS_USER;check:user <> ''"`
	CreatedAt time.Time `gorm:"COLUMN:created;NOT NULL"`
	Type      Logintype `gorm:"TYPE:uint;COLUMN:type;NOT NULL;INDEX:IX_LOGINS_TYPE"`
}

func (b LoginsEntity) String() string {
	return fmt.Sprintf("LoginsEntity: '%d, User: %s'", b.ID, b.User)
}

// TableName specifies the name of the Table used
func (LoginsEntity) TableName() string {
	return "LOGINS"
}

// UserSiteEntity holds the defintions of user-access, sites and roles
type UserSiteEntity struct {
	Name      string    `gorm:"primaryKey;TYPE:varchar(128);COLUMN:name;NOT NULL;INDEX:IX_USERSITE_PK"`
	User      string    `gorm:"primaryKey;TYPE:varchar(128);COLUMN:user;NOT NULL;INDEX:IX_USERSITE_PK"`
	URL       string    `gorm:"TYPE:varchar(256);COLUMN:url;NOT NULL"`
	PermList  string    `gorm:"TYPE:varchar(256);COLUMN:permission_list;NOT NULL"`
	CreatedAt time.Time `gorm:"COLUMN:created;NOT NULL"`
}

func (b UserSiteEntity) String() string {
	return fmt.Sprintf("UserSiteEntity: '%s,%s'", b.Name, b.User)
}

// TableName specifies the name of the Table used
func (UserSiteEntity) TableName() string {
	return "USERSITE"
}
