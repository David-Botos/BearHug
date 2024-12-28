// TODO: Enable ENUM types with GORM

package hsds_types

import (
	"time"
)

// // -- HSDS Definitions -- ////
type Organization struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	ParentOrganizationID *string `json:"parent_organization_id,omitempty" gorm:"type:varchar(250);column:parent_organization_id"`

	// Organization Data
	ID               string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name             string  `json:"name" gorm:"type:text;not null" validate:"required"`
	AlternateName    *string `json:"alternate_name,omitempty" gorm:"type:text"`
	Description      string  `json:"description" gorm:"type:text;not null" validate:"required"`
	Email            *string `json:"email,omitempty" gorm:"type:text"`
	LegalStatus      *string `json:"legal_status,omitempty" gorm:"type:text"`
	Logo             *string `json:"logo,omitempty" gorm:"type:text"`
	TaxID            *string `json:"tax_id,omitempty" gorm:"type:text;column:tax_id"`
	TaxStatus        *string `json:"tax_status,omitempty" gorm:"type:text"`
	URI              *string `json:"uri,omitempty" gorm:"type:text"`
	Website          *string `json:"website,omitempty" gorm:"type:text"`
	YearIncorporated *int    `json:"year_incorporated,omitempty" gorm:"type:numeric"`
}

type OrganizationIdentifier struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	OrganizationID string       `json:"organization_id" gorm:"type:varchar(250);not null;foreignKey:OrganizationID;references:ID" validate:"required"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID" json:"-"`

	// OrganizationIdentifier Data
	ID               string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	IdentifierScheme *string `json:"identifier_scheme,omitempty" gorm:"type:text"`
	IdentifierType   string  `json:"identifier_type" gorm:"type:text;not null" validate:"required"`
	Identifier       string  `json:"identifier" gorm:"type:text;not null" validate:"required"`
}

type URL struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	OrganizationID *string      `json:"organization_id,omitempty" gorm:"type:varchar(250);foreignKey:OrganizationID;references:ID"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID" json:"-"`
	ServiceID      *string      `json:"service_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceID;references:ID"`
	Service        Service      `gorm:"foreignKey:ServiceID;references:ID" json:"-"`

	// URL Data
	ID    string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Label *string `json:"label,omitempty" gorm:"type:text"`
	URL   string  `json:"url" gorm:"type:text;not null" validate:"required"`
}

type Funding struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	OrganizationID *string      `json:"organization_id,omitempty" gorm:"type:varchar(250);foreignKey:OrganizationID;references:ID"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID" json:"-"`
	ServiceID      *string      `json:"service_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceID;references:ID"`
	Service        Service      `gorm:"foreignKey:ServiceID;references:ID" json:"-"`

	// Funding Data
	ID     string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Source *string `json:"source,omitempty" gorm:"type:text"`
}

type Unit struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Unit Data
	ID         string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name       string  `json:"name" gorm:"type:text;not null" validate:"required"`
	Scheme     *string `json:"scheme,omitempty" gorm:"type:text"`
	Identifier *string `json:"identifier,omitempty" gorm:"type:text"`
	URI        *string `json:"uri,omitempty" gorm:"type:text"`
}

type Program struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	OrganizationID string       `json:"organization_id" gorm:"type:varchar(250);not null;uniqueIndex;foreignKey:OrganizationID;references:ID" validate:"required"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID" json:"-"`

	// Program Data
	ID            string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name          string  `json:"name" gorm:"type:text;not null" validate:"required"`
	AlternateName *string `json:"alternate_name,omitempty" gorm:"type:text"`
	Description   string  `json:"description" gorm:"type:text;not null" validate:"required"`
}

type Service struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	OrganizationID string       `json:"organization_id" gorm:"type:varchar(250);not null;foreignKey:OrganizationID;references:ID" validate:"required"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID" json:"-"`
	ProgramID      *string      `json:"program_id,omitempty" gorm:"type:varchar(250);foreignKey:ProgramID;references:ID"`
	Program        Program      `gorm:"foreignKey:ProgramID;references:ID" json:"-"`

	// Service Data
	ID                     string            `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name                   string            `json:"name" gorm:"type:text;not null" validate:"required"`
	AlternateName          *string           `json:"alternate_name,omitempty" gorm:"type:text"`
	Description            *string           `json:"description,omitempty" gorm:"type:text"`
	URL                    *string           `json:"url,omitempty" gorm:"type:text"`
	Email                  *string           `json:"email,omitempty" gorm:"type:text"`
	Status                 ServiceStatusEnum `json:"status" gorm:"type:service_status_enum;not null" validate:"required"`
	InterpretationServices *string           `json:"interpretation_services,omitempty" gorm:"type:text"`
	ApplicationProcess     *string           `json:"application_process,omitempty" gorm:"type:text"`
	FeesDescription        *string           `json:"fees_description,omitempty" gorm:"type:text"`
	WaitTime               *string           `json:"wait_time,omitempty" gorm:"type:text"` // Deprecated
	Fees                   *string           `json:"fees,omitempty" gorm:"type:text"`      // Deprecated
	Accreditations         *string           `json:"accreditations,omitempty" gorm:"type:text"`
	EligibilityDescription *string           `json:"eligibility_description,omitempty" gorm:"type:text"`
	MinimumAge             *float64          `json:"minimum_age,omitempty" gorm:"type:numeric"`
	MaximumAge             *float64          `json:"maximum_age,omitempty" gorm:"type:numeric"`
	AssuredDate            *time.Time        `json:"assured_date,omitempty" gorm:"type:date"`
	AssurerEmail           *string           `json:"assurer_email,omitempty" gorm:"type:text"`
	Licenses               *string           `json:"licenses,omitempty" gorm:"type:text"` // Deprecated
	Alert                  *string           `json:"alert,omitempty" gorm:"type:text"`
	LastModified           *time.Time        `json:"last_modified,omitempty" gorm:"type:timestamp without time zone"`
}

type ServiceArea struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	ServiceID           *string           `json:"service_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceID;references:ID"`
	Service             Service           `gorm:"foreignKey:ServiceID;references:ID" json:"-"`
	ServiceAtLocationID *string           `json:"service_at_location_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceAtLocationID;references:ID"`
	ServiceAtLocation   ServiceAtLocation `gorm:"foreignKey:ServiceAtLocationID;references:ID" json:"-"`

	// Service Area Data
	ID          string          `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name        *string         `json:"name,omitempty" gorm:"type:text"`
	Description *string         `json:"description,omitempty" gorm:"type:text"`
	Extent      *string         `json:"extent,omitempty" gorm:"type:text"`
	ExtentType  *ExtentTypeEnum `json:"extent_type,omitempty" gorm:"type:text"`
	URI         *string         `json:"uri,omitempty" gorm:"type:text"`
}

type ServiceAtLocation struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	ServiceID  string   `json:"service_id" gorm:"type:varchar(250);not null;foreignKey:ServiceID;references:ID" validate:"required"`
	Service    Service  `gorm:"foreignKey:ServiceID;references:ID" json:"-"`
	LocationID string   `json:"location_id" gorm:"type:varchar(250);not null;foreignKey:LocationID;references:ID" validate:"required"`
	Location   Location `gorm:"foreignKey:LocationID;references:ID" json:"-"`

	// ServiceAtLocation Data
	ID          string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Description *string `json:"description,omitempty" gorm:"type:text"`
}

type Location struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`
	// Foreign Key Relationships
	OrganizationID *string       `json:"organization_id,omitempty" gorm:"type:varchar(250);column:organization_id;foreignKey:id"`
	Organization   *Organization `json:"-" gorm:"foreignKey:OrganizationID;references:ID"`
	// Location Data
	ID                     string                   `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	LocationType           LocationLocationTypeEnum `json:"location_type" gorm:"type:location_location_type_enum;not null" validate:"required"`
	URL                    *string                  `json:"url,omitempty" gorm:"type:text"`
	Name                   *string                  `json:"name,omitempty" gorm:"type:text"`
	AlternateName          *string                  `json:"alternate_name,omitempty" gorm:"type:text"`
	Description            *string                  `json:"description,omitempty" gorm:"type:text"`
	Transportation         *string                  `json:"transportation,omitempty" gorm:"type:text"`
	Latitude               *float64                 `json:"latitude,omitempty" gorm:"type:numeric"`
	Longitude              *float64                 `json:"longitude,omitempty" gorm:"type:numeric"`
	ExternalIdentifier     *string                  `json:"external_identifier,omitempty" gorm:"type:text"`
	ExternalIdentifierType *string                  `json:"external_identifier_type,omitempty" gorm:"type:text"`
}

type Address struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	LocationID *string  `json:"location_id,omitempty" gorm:"type:varchar(250);foreignKey:LocationID;references:ID"`
	Location   Location `gorm:"foreignKey:LocationID;references:ID" json:"-"`

	// Address Data
	ID            string                   `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Attention     *string                  `json:"attention,omitempty" gorm:"type:text"`
	Address1      string                   `json:"address_1" gorm:"type:text;column:address_1;not null" validate:"required"`
	Address2      *string                  `json:"address_2,omitempty" gorm:"type:text;column:address_2"`
	City          string                   `json:"city" gorm:"type:text;not null" validate:"required"`
	Region        *string                  `json:"region,omitempty" gorm:"type:text"`
	StateProvince string                   `json:"state_province" gorm:"type:text;column:state_province;not null" validate:"required"`
	PostalCode    string                   `json:"postal_code" gorm:"type:text;column:postal_code;not null" validate:"required"`
	Country       string                   `json:"country" gorm:"type:text;not null" validate:"required,len=2"`
	AddressType   LocationLocationTypeEnum `json:"address_type" gorm:"type:address_address_type_enum;column:address_type;not null" validate:"required,oneof=physical postal virtual"`
}

type RequiredDocument struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"updated_at,omitempty"`

	// Foreign Key Relationships
	ServiceID *string `json:"service_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceID;references:ID"`
	Service   Service `gorm:"foreignKey:ServiceID;references:ID" json:"-"`

	// Required Document Data
	ID       string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Document *string `json:"document,omitempty" gorm:"type:text"`
	URI      *string `json:"uri,omitempty" gorm:"type:text"`
}

type Language struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"updated_at"`

	// Foreign Key Relationships
	ServiceID  *string  `json:"service_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceID;references:ID"`
	Service    Service  `gorm:"foreignKey:ServiceID;references:ID" json:"-"`
	LocationID *string  `json:"location_id,omitempty" gorm:"type:varchar(250);foreignKey:LocationID;references:ID"`
	Location   Location `gorm:"foreignKey:LocationID;references:ID" json:"-"`
	PhoneID    *string  `json:"phone_id,omitempty" gorm:"type:varchar(250);foreignKey:PhoneID;references:ID"`
	Phone      Phone    `gorm:"foreignKey:PhoneID;references:ID" json:"-"`

	// Language Data
	ID   string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name *string `json:"name,omitempty" gorm:"type:text"`
	Code *string `json:"code,omitempty" gorm:"type:text"`
	Note *string `json:"note,omitempty" gorm:"type:text"`
}

type Accessibility struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationship
	LocationID *string  `json:"location_id,omitempty" gorm:"type:varchar(250);foreignKey:LocationID;references:ID"`
	Location   Location `gorm:"foreignKey:LocationID;references:ID" json:"-"`

	// Accessibility Data
	ID          string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Description *string `json:"description,omitempty" gorm:"type:text"`
	Details     *string `json:"details,omitempty" gorm:"type:text"`
	URL         *string `json:"url,omitempty" gorm:"type:text"`
}

type Attribute struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"updated_at,omitempty"`

	// Foreign Key Relationship
	TaxonomyTermID string       `json:"taxonomy_term_id" gorm:"type:varchar(250);not null;foreignKey:TaxonomyTermID;references:ID" validate:"required"`
	TaxonomyTerm   TaxonomyTerm `gorm:"foreignKey:TaxonomyTermID;references:ID" json:"-"`

	// Attribute Data
	ID         string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	LinkID     string  `json:"link_id" gorm:"type:text;not null" validate:"required"`
	LinkType   *string `json:"link_type,omitempty" gorm:"type:text;column:link_type"`
	LinkEntity string  `json:"link_entity" gorm:"type:text;not null" validate:"required"`
	Value      *string `json:"value,omitempty" gorm:"type:text"`
	Label      *string `json:"label,omitempty" gorm:"type:text"`
}

type Taxonomy struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Taxonomy Data
	ID          string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name        string  `json:"name" gorm:"type:text;not null" validate:"required"`
	Description string  `json:"description" gorm:"type:text;not null" validate:"required"`
	URI         *string `json:"uri,omitempty" gorm:"type:text"`
	Version     *string `json:"version,omitempty" gorm:"type:text"`
}

type TaxonomyTerm struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	TaxonomyID *string        `json:"taxonomy_id,omitempty" gorm:"type:varchar(250);foreignKey:TaxonomyID;references:ID"`
	Taxonomy   Taxonomy       `gorm:"foreignKey:TaxonomyID;references:ID" json:"-"`
	ParentID   *string        `json:"parent_id,omitempty" gorm:"type:text"`
	Parent     *TaxonomyTerm  `gorm:"foreignKey:ParentID;references:ID" json:"-"`
	Children   []TaxonomyTerm `gorm:"foreignKey:ParentID" json:"-"`

	// TaxonomyTerm Data
	ID          string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Code        *string `json:"code,omitempty" gorm:"type:text;uniqueIndex"`
	Name        string  `json:"name" gorm:"type:text;not null" validate:"required"`
	Description string  `json:"description" gorm:"type:text;not null" validate:"required"`
	TaxonomyStr *string `json:"taxonomy,omitempty" gorm:"type:text;column:taxonomy"` // Renamed to avoid conflict
	Language    *string `json:"language,omitempty" gorm:"type:text"`
	TermURI     *string `json:"term_uri,omitempty" gorm:"type:text;column:term_uri"`
}

type Contact struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"updated_at,omitempty"`

	// Foreign Key Relationships
	OrganizationID *string      `json:"organization_id,omitempty" gorm:"type:varchar(250);foreignKey:OrganizationID;references:ID"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID" json:"-"`

	ServiceID *string `json:"service_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceID;references:ID"`
	Service   Service `gorm:"foreignKey:ServiceID;references:ID" json:"-"`

	ServiceAtLocationID *string           `json:"service_at_location_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceAtLocationID;references:ID"`
	ServiceAtLocation   ServiceAtLocation `gorm:"foreignKey:ServiceAtLocationID;references:ID" json:"-"`

	LocationID *string  `json:"location_id,omitempty" gorm:"type:varchar(250);foreignKey:LocationID;references:ID"`
	Location   Location `gorm:"foreignKey:LocationID;references:ID" json:"-"`

	// Contact Data
	ID         string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name       *string `json:"name,omitempty" gorm:"type:text"`
	Title      *string `json:"title,omitempty" gorm:"type:text"`
	Department *string `json:"department,omitempty" gorm:"type:text"`
	Email      *string `json:"email,omitempty" gorm:"type:text"`
}

type Phone struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	LocationID *string  `json:"location_id,omitempty" gorm:"type:varchar(250);foreignKey:LocationID;references:ID"`
	Location   Location `gorm:"foreignKey:LocationID;references:ID" json:"-"`

	ServiceID *string `json:"service_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceID;references:ID"`
	Service   Service `gorm:"foreignKey:ServiceID;references:ID" json:"-"`

	OrganizationID *string      `json:"organization_id,omitempty" gorm:"type:varchar(250);foreignKey:OrganizationID;references:ID"`
	Organization   Organization `gorm:"foreignKey:OrganizationID;references:ID" json:"-"`

	ContactID *string `json:"contact_id,omitempty" gorm:"type:varchar(250);foreignKey:ContactID;references:ID"`
	Contact   Contact `gorm:"foreignKey:ContactID;references:ID" json:"-"`

	ServiceAtLocationID *string           `json:"service_at_location_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceAtLocationID;references:ID"`
	ServiceAtLocation   ServiceAtLocation `gorm:"foreignKey:ServiceAtLocationID;references:ID" json:"-"`

	// Phone Data
	ID          string   `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Number      string   `json:"number" gorm:"type:text;not null" validate:"required"`
	Extension   *float64 `json:"extension,omitempty" gorm:"type:numeric"`
	Type        *string  `json:"type,omitempty" gorm:"type:text"`
	Description *string  `json:"description,omitempty" gorm:"type:text"`
}

type Schedule struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	ServiceID *string `json:"service_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceID;references:ID"`
	Service   Service `gorm:"foreignKey:ServiceID;references:ID" json:"-"`

	LocationID *string  `json:"location_id,omitempty" gorm:"type:varchar(250);foreignKey:LocationID;references:ID"`
	Location   Location `gorm:"foreignKey:LocationID;references:ID" json:"-"`

	ServiceAtLocationID *string           `json:"service_at_location_id,omitempty" gorm:"type:varchar(250);foreignKey:ServiceAtLocationID;references:ID"`
	ServiceAtLocation   ServiceAtLocation `gorm:"foreignKey:ServiceAtLocationID;references:ID" json:"-"`

	// Schedule Data
	ID            string            `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	ValidFrom     *time.Time        `json:"valid_from,omitempty" gorm:"type:date"`
	ValidTo       *time.Time        `json:"valid_to,omitempty" gorm:"type:date"`
	DTStart       *time.Time        `json:"dtstart,omitempty" gorm:"type:date;column:dtstart"`
	Timezone      *float64          `json:"timezone,omitempty" gorm:"type:numeric"`
	Until         *time.Time        `json:"until,omitempty" gorm:"type:date"`
	Count         *int              `json:"count,omitempty" gorm:"type:numeric"`
	Wkst          *ScheduleWkstEnum `json:"wkst,omitempty" gorm:"type:schedule_wkst_enum"`
	Freq          *ScheduleFreqEnum `json:"freq,omitempty" gorm:"type:schedule_freq_enum"`
	Interval      *int              `json:"interval,omitempty" gorm:"type:numeric;column:interval"`
	Byday         *string           `json:"byday,omitempty" gorm:"type:text"`
	Byweekno      *string           `json:"byweekno,omitempty" gorm:"type:text"`
	Bymonthday    *string           `json:"bymonthday,omitempty" gorm:"type:text"`
	Byyearday     *string           `json:"byyearday,omitempty" gorm:"type:text"`
	Description   *string           `json:"description,omitempty" gorm:"type:text"`
	OpensAt       *time.Time        `json:"opens_at,omitempty" gorm:"type:time without time zone"`
	ClosesAt      *time.Time        `json:"closes_at,omitempty" gorm:"type:time without time zone"`
	ScheduleLink  *string           `json:"schedule_link,omitempty" gorm:"type:text"`
	AttendingType *string           `json:"attending_type,omitempty" gorm:"type:text"`
	Notes         *string           `json:"notes,omitempty" gorm:"type:text"`
}

type ServiceCapacity struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	ServiceID string  `json:"service_id" gorm:"type:varchar(250);not null;foreignKey:ServiceID;references:ID" validate:"required"`
	Service   Service `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	UnitID string `json:"unit_id" gorm:"type:varchar(250);not null;foreignKey:UnitID;references:ID" validate:"required"`
	Unit   Unit   `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	// Service Capacity Data
	ID          string    `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Available   float64   `json:"available" gorm:"type:numeric;not null" validate:"required"`
	Maximum     *float64  `json:"maximum,omitempty" gorm:"type:numeric"`
	Description *string   `json:"description,omitempty" gorm:"type:text"`
	Updated     time.Time `json:"updated" gorm:"type:timestamp;not null" validate:"required"`
}

type CostOption struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Foreign Key Relationships
	ServiceID string  `json:"service_id" gorm:"type:varchar(250);not null;foreignKey:ServiceID;references:ID" validate:"required"`
	Service   Service `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	// CostOption Data
	ID                string     `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	ValidFrom         *time.Time `json:"valid_from,omitempty" gorm:"type:date"`
	ValidTo           *time.Time `json:"valid_to,omitempty" gorm:"type:date"`
	Option            *string    `json:"option,omitempty" gorm:"type:text"`
	Currency          *string    `json:"currency,omitempty" gorm:"type:text"`
	Amount            *float64   `json:"amount,omitempty" gorm:"type:numeric"`
	AmountDescription *string    `json:"amount_description,omitempty" gorm:"type:text"`
}

type Metadata struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// Resource Reference
	ResourceID string `json:"resource_id" gorm:"type:text;not null" validate:"required"`

	// Metadata Data
	ID               string    `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	ResourceType     string    `json:"resource_type" gorm:"type:text;not null" validate:"required"`
	LastActionDate   time.Time `json:"last_action_date" gorm:"type:date;not null" validate:"required"`
	LastActionType   string    `json:"last_action_type" gorm:"type:text;not null" validate:"required"`
	FieldName        string    `json:"field_name" gorm:"type:text;not null" validate:"required"`
	PreviousValue    string    `json:"previous_value" gorm:"type:text;not null" validate:"required"`
	ReplacementValue string    `json:"replacement_value" gorm:"type:text;not null" validate:"required"`
	UpdatedBy        string    `json:"updated_by" gorm:"type:text;not null" validate:"required"`
}

type MetaTableDescription struct {
	CreatedAt time.Time `gorm:"type:timestamp" json:"-"`
	UpdatedAt time.Time `gorm:"type:timestamp" json:"-"`

	// MetaTableDescription Data
	ID           string  `json:"id" gorm:"type:varchar(250);primaryKey;not null" validate:"required"`
	Name         *string `json:"name,omitempty" gorm:"type:text"`
	Language     *string `json:"language,omitempty" gorm:"type:text"`
	CharacterSet *string `json:"character_set,omitempty" gorm:"type:text;column:character_set"`
}

//// -- Enum Utilities -- ////

// Enums contains all enum definitions
type Enums struct {
	AddressAddressTypeEnum   []AddressAddressTypeEnum
	LocationLocationTypeEnum []LocationLocationTypeEnum
	ScheduleFreqEnum         []ScheduleFreqEnum
	ScheduleWkstEnum         []ScheduleWkstEnum
	ServiceStatusEnum        []ServiceStatusEnum
}

// Enum type definitions
type AddressAddressTypeEnum string
type LocationLocationTypeEnum string
type ScheduleFreqEnum string
type ScheduleWkstEnum string
type ServiceStatusEnum string
type ExtentTypeEnum string

// AddressAddressTypeEnum values
const (
	AddressTypePhysical AddressAddressTypeEnum = "physical"
	AddressTypePostal   AddressAddressTypeEnum = "postal"
	AddressTypeVirtual  AddressAddressTypeEnum = "virtual"
)

// LocationLocationTypeEnum values
const (
	LocationTypePhysical LocationLocationTypeEnum = "physical"
	LocationTypePostal   LocationLocationTypeEnum = "postal"
	LocationTypeVirtual  LocationLocationTypeEnum = "virtual"
)

// ScheduleFreqEnum values
const (
	ScheduleFreqWeekly  ScheduleFreqEnum = "WEEKLY"
	ScheduleFreqMonthly ScheduleFreqEnum = "MONTHLY"
)

// ScheduleWkstEnum values
const (
	ScheduleWkstMO ScheduleWkstEnum = "MO"
	ScheduleWkstTU ScheduleWkstEnum = "TU"
	ScheduleWkstWE ScheduleWkstEnum = "WE"
	ScheduleWkstTH ScheduleWkstEnum = "TH"
	ScheduleWkstFR ScheduleWkstEnum = "FR"
	ScheduleWkstSA ScheduleWkstEnum = "SA"
	ScheduleWkstSU ScheduleWkstEnum = "SU"
)

// ServiceStatusEnum values
const (
	ServiceStatusActive            ServiceStatusEnum = "active"
	ServiceStatusInactive          ServiceStatusEnum = "inactive"
	ServiceStatusDefunct           ServiceStatusEnum = "defunct"
	ServiceStatusTemporarilyClosed ServiceStatusEnum = "temporarily closed"
)

// ExtentTypeEnum values
const (
	ExtentTypeGeoJSON  ExtentTypeEnum = "geojson"
	ExtentTypeTopoJSON ExtentTypeEnum = "topojson"
	ExtentTypeKML      ExtentTypeEnum = "kml"
	ExtentTypeText     ExtentTypeEnum = "text"
)
