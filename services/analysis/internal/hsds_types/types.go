package hsds_types

import (
	"time"
)

// // -- HSDS Definitions -- ////
type Organization struct {
	// Foreign Key Relationships
	ParentOrganizationID string `json:"parent_organization_id,omitempty"`
	// Organization Data
	ID               string `json:"id" validate:"required"`
	Name             string `json:"name" validate:"required"`
	AlternateName    string `json:"alternate_name,omitempty"`
	Description      string `json:"description" validate:"required"`
	Email            string `json:"email,omitempty"`
	LegalStatus      string `json:"legal_status,omitempty"`
	Logo             string `json:"logo,omitempty"`
	TaxID            string `json:"tax_id,omitempty"`
	TaxStatus        string `json:"tax_status,omitempty"`
	URI              string `json:"uri,omitempty"`
	Website          string `json:"website,omitempty"`
	YearIncorporated int    `json:"year_incorporated,omitempty"`
}

type OrganizationIdentifier struct {
	//Foreign Key Relationships
	OrganizationID string `json:"organization_id" validate:"required"`
	//OrganizationIdentifier Data
	ID               string `json:"id" validate:"required"`
	IdentifierScheme string `json:"identifier_scheme,omitempty"`
	IdentifierType   string `json:"identifier_type" validate:"required"`
	Identifier       string `json:"identifier" validate:"required"`
}

type URL struct {
	// Foreign Key Relationships
	OrganizationID string `json:"organization_id,omitempty"`
	ServiceID      string `json:"service_id,omitempty"`
	// URL Data
	ID    string `json:"id" validate:"required"`
	Label string `json:"label,omitempty"`
	URL   string `json:"url" validate:"required"`
}

type Funding struct {
	// Foreign Key Relationships
	OrganizationID string `json:"organization_id,omitempty"`
	ServiceID      string `json:"service_id,omitempty"`

	//Funding Data
	ID     string `json:"id" validate:"required"`
	Source string `json:"source,omitempty"`
}

type Unit struct {
	ID         string `json:"id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Scheme     string `json:"scheme,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	URI        string `json:"uri,omitempty"`
}

type Program struct {
	// Foreign Key Relationships
	OrganizationID string `json:"organization_id" validate:"required"`
	// Program Data
	ID            string `json:"id" validate:"required"`
	Name          string `json:"name" validate:"required"`
	AlternateName string `json:"alternate_name,omitempty"`
	Description   string `json:"description" validate:"required"`
}

type Service struct {
	// Foreign Key Relationships
	OrganizationID string `json:"organization_id" validate:"required"`
	ProgramID      string `json:"program_id,omitempty"`
	// Service Data
	ID                     string            `json:"id" validate:"required"`
	Name                   string            `json:"name" validate:"required"`
	AlternateName          string            `json:"alternate_name,omitempty"`
	Description            string            `json:"description,omitempty"`
	URL                    string            `json:"url,omitempty"`
	Email                  string            `json:"email,omitempty"`
	Status                 ServiceStatusEnum `json:"status" validate:"required"`
	InterpretationServices string            `json:"interpretation_services,omitempty"`
	ApplicationProcess     string            `json:"application_process,omitempty"`
	FeesDescription        string            `json:"fees_description,omitempty"`
	WaitTime               string            `json:"wait_time,omitempty"` // Deprecated
	Fees                   string            `json:"fees,omitempty"`      // Deprecated
	Accreditations         string            `json:"accreditations,omitempty"`
	EligibilityDescription string            `json:"eligibility_description,omitempty"`
	MinimumAge             float64           `json:"minimum_age,omitempty"`
	MaximumAge             float64           `json:"maximum_age,omitempty"`
	AssuredDate            time.Time         `json:"assured_date,omitempty"`
	AssurerEmail           string            `json:"assurer_email,omitempty"`
	Licenses               string            `json:"licenses,omitempty"` // Deprecated
	Alert                  string            `json:"alert,omitempty"`
	LastModified           time.Time         `json:"last_modified,omitempty"`
}

type ServiceArea struct {
	// Foreign Key Relationships
	ServiceID           string `json:"service_id,omitempty"`
	ServiceAtLocationID string `json:"service_at_location_id,omitempty"`
	// Service Area Data
	ID          string         `json:"id" validate:"required"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Extent      string         `json:"extent,omitempty"`
	ExtentType  ExtentTypeEnum `json:"extent_type,omitempty"`
	URI         string         `json:"uri,omitempty"`
}

type ServiceAtLocation struct {
	// Foreign Key Relationships
	ServiceID  string `json:"service_id" validate:"required"`
	LocationID string `json:"location_id" validate:"required"`
	// ServiceAtLocation Data
	ID          string `json:"id" validate:"required"`
	Description string `json:"description,omitempty"`
}

type Location struct {
	// Foreign Key Relationships
	OrganizationID string `json:"organization_id,omitempty"`
	// Location Data
	ID                     string                   `json:"id" validate:"required"`
	LocationType           LocationLocationTypeEnum `json:"location_type" validate:"required"`
	URL                    string                   `json:"url,omitempty"`
	Name                   string                   `json:"name,omitempty"`
	AlternateName          string                   `json:"alternate_name,omitempty"`
	Description            string                   `json:"description,omitempty"`
	Transportation         string                   `json:"transportation,omitempty"`
	Latitude               float64                  `json:"latitude,omitempty"`
	Longitude              float64                  `json:"longitude,omitempty"`
	ExternalIdentifier     string                   `json:"external_identifier,omitempty"`
	ExternalIdentifierType string                   `json:"external_identifier_type,omitempty"`
}

type Address struct {
	// Foreign Key Relationships
	LocationID string `json:"location_id,omitempty"`

	// Address Data
	ID            string                   `json:"id" validate:"required"`
	Attention     string                   `json:"attention,omitempty"`
	Address1      string                   `json:"address_1" validate:"required"`
	Address2      string                   `json:"address_2,omitempty"`
	City          string                   `json:"city" validate:"required"`
	Region        string                   `json:"region,omitempty"`
	StateProvince string                   `json:"state_province" validate:"required"`
	PostalCode    string                   `json:"postal_code" validate:"required"`
	Country       string                   `json:"country" validate:"required,len=2"`
	AddressType   LocationLocationTypeEnum `json:"address_type" validate:"required,oneof=physical postal virtual"`
}

type RequiredDocument struct {
	// Foreign Key Relationships
	ServiceID string `db:"service_id" json:"service_id,omitempty"`
	// Required Document Data
	ID        string    `db:"id" json:"id" validate:"required"`
	Document  string    `db:"document" json:"document,omitempty"`
	URI       string    `db:"uri" json:"uri,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type Language struct {
	// Foreign Key Relationships
	ServiceID  string `json:"service_id,omitempty"`
	LocationID string `json:"location_id,omitempty"`
	PhoneID    string `json:"phone_id,omitempty"`

	// Language Data
	ID        string    `json:"id" validate:"required"`
	Name      string    `json:"name,omitempty"`
	Code      string    `json:"code,omitempty"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Accessibility struct {
	// Foreign Key Relationship
	LocationID string `json:"location_id,omitempty"`
	// Accessibility Data
	ID          string `json:"id" validate:"required"`
	Description string `json:"description,omitempty"`
	Details     string `json:"details,omitempty"`
	URL         string `json:"url,omitempty"`
}

type Attribute struct {
	// Foreign Key Relationship
	TaxonomyTermID string `db:"taxonomy_term_id" json:"taxonomy_term_id" validate:"required"`
	LinkID         string `db:"link_id" json:"link_id" validate:"required"`
	// Attribute Data
	ID         string    `db:"id" json:"id" validate:"required"`
	LinkType   string    `db:"link_type" json:"link_type,omitempty"`
	LinkEntity string    `db:"link_entity" json:"link_entity" validate:"required"`
	Value      string    `db:"value" json:"value,omitempty"`
	Label      string    `db:"label" json:"label,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type Taxonomy struct {
	ID          string `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	URI         string `json:"uri,omitempty"`
	Version     string `json:"version,omitempty"`
}

type TaxonomyTerm struct {
	// Foreign Key Relationship
	TaxonomyID string `json:"taxonomy_id,omitempty"`
	ParentID   string `json:"parent_id,omitempty"`
	// TaxonomyTerm Data
	ID          string `json:"id" validate:"required"`
	Code        string `json:"code,omitempty"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Taxonomy    string `json:"taxonomy,omitempty"`
	Language    string `json:"language,omitempty"`
	TermURI     string `json:"term_uri,omitempty"`
}

type Contact struct {
	// Foreign Key Relationships
	OrganizationID      string `db:"organization_id" json:"organization_id,omitempty"`
	ServiceID           string `db:"service_id" json:"service_id,omitempty"`
	ServiceAtLocationID string `db:"service_at_location_id" json:"service_at_location_id,omitempty"`
	LocationID          string `db:"location_id" json:"location_id,omitempty"`
	// Contact Data
	ID         string    `db:"id" json:"id" validate:"required"`
	Name       string    `db:"name" json:"name,omitempty"`
	Title      string    `db:"title" json:"title,omitempty"`
	Department string    `db:"department" json:"department,omitempty"`
	Email      string    `db:"email" json:"email,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type Phone struct {
	//Foreign Key Relationships
	LocationID          string `json:"location_id,omitempty"`
	ServiceID           string `json:"service_id,omitempty"`
	OrganizationID      string `json:"organization_id,omitempty"`
	ContactID           string `json:"contact_id,omitempty"`
	ServiceAtLocationID string `json:"service_at_location_id,omitempty"`
	//Phone Data
	ID          string  `json:"id" validate:"required"`
	Number      string  `json:"number" validate:"required"`
	Extension   float64 `json:"extension,omitempty"`
	Type        string  `json:"type,omitempty"`
	Description string  `json:"description,omitempty"`
}

type Schedule struct {
	// Foreign Key Relationship
	ServiceID           string `json:"service_id,omitempty"`
	LocationID          string `json:"location_id,omitempty"`
	ServiceAtLocationID string `json:"service_at_location_id,omitempty"`
	// Schedule Data
	ID            string           `json:"id" validate:"required"`
	ValidFrom     time.Time        `json:"valid_from,omitempty"`
	ValidTo       time.Time        `json:"valid_to,omitempty"`
	DTStart       time.Time        `json:"dtstart,omitempty"`
	Timezone      float64          `json:"timezone,omitempty"`
	Until         time.Time        `json:"until,omitempty"`
	Count         int              `json:"count,omitempty"`
	Wkst          ScheduleWkstEnum `json:"wkst,omitempty"`
	Freq          ScheduleFreqEnum `json:"freq,omitempty"`
	Interval      int              `json:"interval,omitempty"`
	Byday         string           `json:"byday,omitempty"`
	Byweekno      string           `json:"byweekno,omitempty"`
	Bymonthday    string           `json:"bymonthday,omitempty"`
	Byyearday     string           `json:"byyearday,omitempty"`
	Description   string           `json:"description,omitempty"`
	OpensAt       time.Time        `json:"opens_at,omitempty"`
	ClosesAt      time.Time        `json:"closes_at,omitempty"`
	ScheduleLink  string           `json:"schedule_link,omitempty"`
	AttendingType string           `json:"attending_type,omitempty"`
	Notes         string           `json:"notes,omitempty"`
}

type ServiceCapacity struct {
	// Foreign Key Relationship
	ServiceID string `db:"service_id" json:"service_id" validate:"required"`
	UnitID    string `db:"unit_id" json:"unit_id" validate:"required"`
	// Service Capacity Data
	ID          string    `db:"id" json:"id" validate:"required"`
	Available   float64   `db:"available" json:"available" validate:"required"`
	Maximum     float64   `db:"maximum" json:"maximum,omitempty"`
	Description string    `db:"description" json:"description,omitempty"`
	Updated     time.Time `db:"updated" json:"updated" validate:"required"`
}

type CostOption struct {
	// Foreign Key Relationships
	ServiceID string `json:"service_id" validate:"required"`
	// CostOption Data
	ID                string    `json:"id" validate:"required"`
	ValidFrom         time.Time `json:"valid_from,omitempty"`
	ValidTo           time.Time `json:"valid_to,omitempty"`
	Option            string    `json:"option,omitempty"`
	Currency          string    `json:"currency,omitempty"`
	Amount            float64   `json:"amount,omitempty"`
	AmountDescription string    `json:"amount_description,omitempty"`
}

type Metadata struct {
	// Foreign Key Relationship
	ResourceID string `db:"resource_id" json:"resource_id" validate:"required"`
	// Metadata Data
	ID               string    `db:"id" json:"id" validate:"required"`
	ResourceType     string    `db:"resource_type" json:"resource_type" validate:"required"`
	LastActionDate   time.Time `db:"last_action_date" json:"last_action_date" validate:"required"`
	LastActionType   string    `db:"last_action_type" json:"last_action_type" validate:"required"`
	FieldName        string    `db:"field_name" json:"field_name" validate:"required"`
	PreviousValue    string    `db:"previous_value" json:"previous_value" validate:"required"`
	ReplacementValue string    `db:"replacement_value" json:"replacement_value" validate:"required"`
	UpdatedBy        string    `db:"updated_by" json:"updated_by" validate:"required"`
}

type MetaTableDescription struct {
	ID           string `json:"id" validate:"required"`
	Name         string `json:"name,omitempty"`
	Language     string `json:"language,omitempty"`
	CharacterSet string `json:"character_set,omitempty"`
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
