package hsds_types

import (
	"time"
)

// // -- HSDS Definitions -- ////
type Organization struct {
    // Foreign Key Relationships
    ParentOrganizationID string `json:"parent_organization_id,omitempty" db:"parent_organization_id"`
    
    // Organization Data
    ID              string `json:"id" validate:"required" db:"id"`
    Name            string `json:"name" validate:"required" db:"name"`
    AlternateName   string `json:"alternate_name,omitempty" db:"alternate_name"`
    Description     string `json:"description" validate:"required" db:"description"`
    Email           string `json:"email,omitempty" db:"email"`
    LegalStatus     string `json:"legal_status,omitempty" db:"legal_status"`
    Logo            string `json:"logo,omitempty" db:"logo"`
    TaxID           string `json:"tax_id,omitempty" db:"tax_id"`
    TaxStatus       string `json:"tax_status,omitempty" db:"tax_status"`
    URI             string `json:"uri,omitempty" db:"uri"`
    Website         string `json:"website,omitempty" db:"website"`
    YearIncorporated int   `json:"year_incorporated,omitempty" db:"year_incorporated"`
}

type OrganizationIdentifier struct {
    //Foreign Key Relationships
    OrganizationID string `json:"organization_id" validate:"required" db:"organization_id"`
    //OrganizationIdentifier Data
    ID               string `json:"id" validate:"required" db:"id"`
    IdentifierScheme string `json:"identifier_scheme,omitempty" db:"identifier_scheme"`
    IdentifierType   string `json:"identifier_type" validate:"required" db:"identifier_type"`
    Identifier       string `json:"identifier" validate:"required" db:"identifier"`
}

type URL struct {
	// Foreign Key Relationships
	OrganizationID string `json:"organization_id,omitempty" db:"organization_id"`
	ServiceID      string `json:"service_id,omitempty" db:"service_id"`
	// URL Data
	ID    string `json:"id" validate:"required" db:"id"`
	Label string `json:"label,omitempty" db:"label"`
	URL   string `json:"url" validate:"required" db:"url"`
}

type Funding struct {
	// Foreign Key Relationships
	OrganizationID string `json:"organization_id,omitempty" db:"organization_id"`
	ServiceID      string `json:"service_id,omitempty" db:"service_id"`

	//Funding Data
	ID     string `json:"id" validate:"required" db:"id"`
	Source string `json:"source,omitempty" db:"source"`
}

type Unit struct {
    ID         string `json:"id" validate:"required" db:"id"`
    Name       string `json:"name" validate:"required" db:"name"`
    Scheme     string `json:"scheme,omitempty" db:"scheme"`
    Identifier string `json:"identifier,omitempty" db:"identifier"`
    URI        string `json:"uri,omitempty" db:"uri"`
}

type Program struct {
	// Foreign Key Relationships
	OrganizationID string `json:"organization_id" validate:"required" db:"organization_id"`
	// Program Data
	ID            string `json:"id" validate:"required" db:"id"`
	Name          string `json:"name" validate:"required" db:"name"`
	AlternateName string `json:"alternate_name,omitempty" db:"alternate_name"`
	Description   string `json:"description" validate:"required" db:"description"`
}

type Service struct {
    // Foreign Key Relationships
    OrganizationID string `json:"organization_id" validate:"required" db:"organization_id"`
    ProgramID string `json:"program_id,omitempty" db:"program_id"`
    // Service Data
    ID string `json:"id" validate:"required" db:"id"`
    Name string `json:"name" validate:"required" db:"name"`
    AlternateName string `json:"alternate_name,omitempty" db:"alternate_name"`
    Description string `json:"description,omitempty" db:"description"`
    URL string `json:"url,omitempty" db:"url"`
    Email string `json:"email,omitempty" db:"email"`
    Status ServiceStatusEnum `json:"status" validate:"required" db:"status"`
    InterpretationServices string `json:"interpretation_services,omitempty" db:"interpretation_services"`
    ApplicationProcess string `json:"application_process,omitempty" db:"application_process"`
    FeesDescription string `json:"fees_description,omitempty" db:"fees_description"`
    WaitTime string `json:"wait_time,omitempty" db:"wait_time"` // Deprecated
    Fees string `json:"fees,omitempty" db:"fees"` // Deprecated
    Accreditations string `json:"accreditations,omitempty" db:"accreditations"`
    EligibilityDescription string `json:"eligibility_description,omitempty" db:"eligibility_description"`
    MinimumAge float64 `json:"minimum_age,omitempty" db:"minimum_age"`
    MaximumAge float64 `json:"maximum_age,omitempty" db:"maximum_age"`
    AssuredDate time.Time `json:"assured_date,omitempty" db:"assured_date"`
    AssurerEmail string `json:"assurer_email,omitempty" db:"assurer_email"`
    Licenses string `json:"licenses,omitempty" db:"licenses"` // Deprecated
    Alert string `json:"alert,omitempty" db:"alert"`
    LastModified time.Time `json:"last_modified,omitempty" db:"last_modified"`
}

type ServiceArea struct {
    // Foreign Key Relationships
    ServiceID string `json:"service_id,omitempty" db:"service_id"`
    ServiceAtLocationID string `json:"service_at_location_id,omitempty" db:"service_at_location_id"`
    // Service Area Data
    ID string `json:"id" validate:"required" db:"id"`
    Name string `json:"name,omitempty" db:"name"`
    Description string `json:"description,omitempty" db:"description"`
    Extent string `json:"extent,omitempty" db:"extent"`
    ExtentType ExtentTypeEnum `json:"extent_type,omitempty" db:"extent_type"`
    URI string `json:"uri,omitempty" db:"uri"`
}

type ServiceAtLocation struct {
    // Foreign Key Relationships
    ServiceID string `json:"service_id" validate:"required" db:"service_id"`
    LocationID string `json:"location_id" validate:"required" db:"location_id"`
    // ServiceAtLocation Data
    ID string `json:"id" validate:"required" db:"id"`
    Description string `json:"description,omitempty" db:"description"`
}

type Location struct {
    // Foreign Key Relationships
    OrganizationID string `json:"organization_id,omitempty" db:"organization_id"`
    // Location Data
    ID string `json:"id" validate:"required" db:"id"`
    LocationType LocationLocationTypeEnum `json:"location_type" validate:"required" db:"location_type"`
    URL string `json:"url,omitempty" db:"url"`
    Name string `json:"name,omitempty" db:"name"`
    AlternateName string `json:"alternate_name,omitempty" db:"alternate_name"`
    Description string `json:"description,omitempty" db:"description"`
    Transportation string `json:"transportation,omitempty" db:"transportation"`
    Latitude float64 `json:"latitude,omitempty" db:"latitude"`
    Longitude float64 `json:"longitude,omitempty" db:"longitude"`
    ExternalIdentifier string `json:"external_identifier,omitempty" db:"external_identifier"`
    ExternalIdentifierType string `json:"external_identifier_type,omitempty" db:"external_identifier_type"`
}

type Address struct {
    // Foreign Key Relationships
    LocationID string `json:"location_id,omitempty" db:"location_id"`

    // Address Data
    ID            string                   `json:"id" validate:"required" db:"id"`
    Attention     string                   `json:"attention,omitempty" db:"attention"`
    Address1      string                   `json:"address_1" validate:"required" db:"address_1"`
    Address2      string                   `json:"address_2,omitempty" db:"address_2"`
    City          string                   `json:"city" validate:"required" db:"city"`
    Region        string                   `json:"region,omitempty" db:"region"`
    StateProvince string                   `json:"state_province" validate:"required" db:"state_province"`
    PostalCode    string                   `json:"postal_code" validate:"required" db:"postal_code"`
    Country       string                   `json:"country" validate:"required,len=2" db:"country"`
    AddressType   LocationLocationTypeEnum `json:"address_type" validate:"required,oneof=physical postal virtual" db:"address_type"`
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
    ServiceID  string `json:"service_id,omitempty" db:"service_id"`
    LocationID string `json:"location_id,omitempty" db:"location_id"`
    PhoneID    string `json:"phone_id,omitempty" db:"phone_id"`

    // Language Data
    ID        string    `json:"id" validate:"required" db:"id"`
    Name      string    `json:"name,omitempty" db:"name"`
    Code      string    `json:"code,omitempty" db:"code"`
    Note      string    `json:"note,omitempty" db:"note"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Accessibility struct {
    // Foreign Key Relationship
    LocationID string `json:"location_id,omitempty" db:"location_id"`

    // Accessibility Data
    ID          string `json:"id" validate:"required" db:"id"`
    Description string `json:"description,omitempty" db:"description"`
    Details     string `json:"details,omitempty" db:"details"`
    URL         string `json:"url,omitempty" db:"url"`
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
    ID          string `json:"id" validate:"required" db:"id"`
    Name        string `json:"name" validate:"required" db:"name"`
    Description string `json:"description" validate:"required" db:"description"`
    URI         string `json:"uri,omitempty" db:"uri"`
    Version     string `json:"version,omitempty" db:"version"`
}

type TaxonomyTerm struct {
    // Foreign Key Relationship
    TaxonomyID string `json:"taxonomy_id,omitempty" db:"taxonomy_id"`
    ParentID   string `json:"parent_id,omitempty" db:"parent_id"`

    // TaxonomyTerm Data
    ID          string `json:"id" validate:"required" db:"id"`
    Code        string `json:"code,omitempty" db:"code"`
    Name        string `json:"name" validate:"required" db:"name"`
    Description string `json:"description" validate:"required" db:"description"`
    Taxonomy    string `json:"taxonomy,omitempty" db:"taxonomy"`
    Language    string `json:"language,omitempty" db:"language"`
    TermURI     string `json:"term_uri,omitempty" db:"term_uri"`
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
    LocationID           string  `json:"location_id,omitempty" db:"location_id"`
    ServiceID           string  `json:"service_id,omitempty" db:"service_id"`
    OrganizationID      string  `json:"organization_id,omitempty" db:"organization_id"`
    ContactID           string  `json:"contact_id,omitempty" db:"contact_id"`
    ServiceAtLocationID string  `json:"service_at_location_id,omitempty" db:"service_at_location_id"`

    //Phone Data
    ID          string   `json:"id" validate:"required" db:"id"`
    Number      string   `json:"number" validate:"required" db:"number"`
    Extension   float64  `json:"extension,omitempty" db:"extension"`
    Type        string   `json:"type,omitempty" db:"type"`
    Description string   `json:"description,omitempty" db:"description"`
}

type Schedule struct {
    // Foreign Key Relationship
    ServiceID           string           `json:"service_id,omitempty" db:"service_id"`
    LocationID          string           `json:"location_id,omitempty" db:"location_id"`
    ServiceAtLocationID string           `json:"service_at_location_id,omitempty" db:"service_at_location_id"`

    // Schedule Data
    ID            string           `json:"id" validate:"required" db:"id"`
    ValidFrom     time.Time        `json:"valid_from,omitempty" db:"valid_from"`
    ValidTo       time.Time        `json:"valid_to,omitempty" db:"valid_to"`
    DTStart       time.Time        `json:"dtstart,omitempty" db:"dtstart"`
    Timezone      float64          `json:"timezone,omitempty" db:"timezone"`
    Until         time.Time        `json:"until,omitempty" db:"until"`
    Count         int              `json:"count,omitempty" db:"count"`
    Wkst          ScheduleWkstEnum `json:"wkst,omitempty" db:"wkst"`
    Freq          ScheduleFreqEnum `json:"freq,omitempty" db:"freq"`
    Interval      int              `json:"interval,omitempty" db:"interval"`
    Byday         string           `json:"byday,omitempty" db:"byday"`
    Byweekno      string           `json:"byweekno,omitempty" db:"byweekno"`
    Bymonthday    string           `json:"bymonthday,omitempty" db:"bymonthday"`
    Byyearday     string           `json:"byyearday,omitempty" db:"byyearday"`
    Description   string           `json:"description,omitempty" db:"description"`
    OpensAt       time.Time        `json:"opens_at,omitempty" db:"opens_at"`
    ClosesAt      time.Time        `json:"closes_at,omitempty" db:"closes_at"`
    ScheduleLink  string           `json:"schedule_link,omitempty" db:"schedule_link"`
    AttendingType string           `json:"attending_type,omitempty" db:"attending_type"`
    Notes         string           `json:"notes,omitempty" db:"notes"`
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
    ServiceID string `json:"service_id" validate:"required" db:"service_id"`

    // CostOption Data
    ID                string    `json:"id" validate:"required" db:"id"`
    ValidFrom         time.Time `json:"valid_from,omitempty" db:"valid_from"`
    ValidTo           time.Time `json:"valid_to,omitempty" db:"valid_to"`
    Option            string    `json:"option,omitempty" db:"option"`
    Currency          string    `json:"currency,omitempty" db:"currency"`
    Amount            float64   `json:"amount,omitempty" db:"amount"`
    AmountDescription string    `json:"amount_description,omitempty" db:"amount_description"`
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
    ID           string `json:"id" validate:"required" db:"id"`
    Name         string `json:"name,omitempty" db:"name"`
    Language     string `json:"language,omitempty" db:"language"`
    CharacterSet string `json:"character_set,omitempty" db:"character_set"`
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
