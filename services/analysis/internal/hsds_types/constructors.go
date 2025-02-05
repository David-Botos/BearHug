package hsds_types

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ValidateUUID ensures the string is a valid UUIDv4
func ValidateUUID(u string) bool {
	id, err := uuid.Parse(u)
	if err != nil {
		return false
	}
	// Verify it's a v4 UUID
	return id.Version() == 4
}

// newUUIDV4 generates a new UUIDv4 string
func newUUIDV4() string {
	return uuid.New().String()
}

// getICalTime returns a time.Time formatted according to iCal specs (RFC5545)
// iCal format example: 20240328T150000Z
func getICalTime() time.Time {
	return time.Now().UTC().Round(time.Second)
}

// OrganizationOptions contains all the optional fields for creating an Organization
type OrganizationOptions struct {
	ParentOrganizationID *string
	AlternateName        *string
	Email                *string
	LegalStatus          *string
	Logo                 *string
	TaxID                *string
	TaxStatus            *string
	URI                  *string
	Website              *string
	YearIncorporated     *int
}

// NewOrganization creates a new Organization with required fields and optional fields via OrganizationOptions
func NewOrganization(name, description string, opts *OrganizationOptions) (*Organization, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	org := &Organization{
		CreatedAt:   now,
		ID:          id,
		Name:        name,
		Description: description,
	}

	if opts != nil {
		if opts.ParentOrganizationID != nil && !ValidateUUID(*opts.ParentOrganizationID) {
			return nil, fmt.Errorf("invalid parent organization ID format: must be UUIDv4")
		}
		org.ParentOrganizationID = opts.ParentOrganizationID
		org.AlternateName = opts.AlternateName
		org.Email = opts.Email
		org.LegalStatus = opts.LegalStatus
		org.Logo = opts.Logo
		org.TaxID = opts.TaxID
		org.TaxStatus = opts.TaxStatus
		org.URI = opts.URI
		org.Website = opts.Website
		org.YearIncorporated = opts.YearIncorporated
	}

	return org, nil
}

// OrganizationIdentifierOptions contains optional fields for creating an OrganizationIdentifier
type OrganizationIdentifierOptions struct {
	IdentifierScheme *string
}

// NewOrganizationIdentifier creates a new OrganizationIdentifier with required fields and optional fields
func NewOrganizationIdentifier(organizationID, identifierType, identifier string, opts *OrganizationIdentifierOptions) (*OrganizationIdentifier, error) {
	if !ValidateUUID(organizationID) {
		return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
	}

	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	orgIdentifier := &OrganizationIdentifier{
		CreatedAt:      now,
		ID:             id,
		OrganizationID: organizationID,
		IdentifierType: identifierType,
		Identifier:     identifier,
	}

	if opts != nil {
		orgIdentifier.IdentifierScheme = opts.IdentifierScheme
	}

	return orgIdentifier, nil
}

// URLOptions contains optional fields for creating a URL
type URLOptions struct {
	OrganizationID *string
	ServiceID      *string
	Label          *string
}

// NewURL creates a new URL with required fields and optional fields
func NewURL(url string, opts *URLOptions) (*URL, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	urlObj := &URL{
		CreatedAt: now,
		ID:        id,
		URL:       url,
	}

	if opts != nil {
		if opts.OrganizationID != nil && !ValidateUUID(*opts.OrganizationID) {
			return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
		}
		if opts.ServiceID != nil && !ValidateUUID(*opts.ServiceID) {
			return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
		}
		urlObj.OrganizationID = opts.OrganizationID
		urlObj.ServiceID = opts.ServiceID
		urlObj.Label = opts.Label
	}

	return urlObj, nil
}

// FundingOptions contains optional fields for creating a Funding
type FundingOptions struct {
	OrganizationID *string
	ServiceID      *string
	Source         *string
}

func NewFunding(opts *FundingOptions) (*Funding, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	funding := &Funding{
		CreatedAt: now,
		ID:        id,
	}

	if opts != nil {
		if opts.OrganizationID != nil && !ValidateUUID(*opts.OrganizationID) {
			return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
		}
		if opts.ServiceID != nil && !ValidateUUID(*opts.ServiceID) {
			return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
		}
		funding.OrganizationID = opts.OrganizationID
		funding.ServiceID = opts.ServiceID
		funding.Source = opts.Source
	}

	return funding, nil
}

// UnitOptions contains optional fields for creating a Unit
type UnitOptions struct {
	Scheme     *string
	Identifier *string
	URI        *string
}

func NewUnit(name string, opts *UnitOptions) (*Unit, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	unit := &Unit{
		CreatedAt: now,
		ID:        id,
		Name:      name,
	}

	if opts != nil {
		unit.Scheme = opts.Scheme
		unit.Identifier = opts.Identifier
		unit.URI = opts.URI
	}

	return unit, nil
}

// ProgramOptions contains optional fields for creating a Program
type ProgramOptions struct {
	AlternateName *string
}

func NewProgram(organizationID, name, description string, opts *ProgramOptions) (*Program, error) {
	if !ValidateUUID(organizationID) {
		return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
	}

	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	program := &Program{
		CreatedAt:      now,
		ID:             id,
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
	}

	if opts != nil {
		program.AlternateName = opts.AlternateName
	}

	return program, nil
}

// ServiceOptions contains optional fields for creating a Service
type ServiceOptions struct {
	ProgramID              *string
	AlternateName          *string
	Description            *string
	URL                    *string
	Email                  *string
	InterpretationServices *string
	ApplicationProcess     *string
	FeesDescription        *string
	WaitTime               *string
	Fees                   *string
	Accreditations         *string
	EligibilityDescription *string
	MinimumAge             *float64
	MaximumAge             *float64
	AssuredDate            *time.Time
	AssurerEmail           *string
	Licenses               *string
	Alert                  *string
}

func NewService(organizationID, name string, status ServiceStatusEnum, opts *ServiceOptions) (*Service, error) {
	if !ValidateUUID(organizationID) {
		return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
	}

	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	service := &Service{
		CreatedAt:      now,
		ID:             id,
		OrganizationID: organizationID,
		Name:           name,
		Status:         status,
	}

	if opts != nil {
		if opts.ProgramID != nil && !ValidateUUID(*opts.ProgramID) {
			return nil, fmt.Errorf("invalid program ID format: must be UUIDv4")
		}
		service.ProgramID = opts.ProgramID
		service.AlternateName = opts.AlternateName
		service.Description = opts.Description
		service.URL = opts.URL
		service.Email = opts.Email
		service.InterpretationServices = opts.InterpretationServices
		service.ApplicationProcess = opts.ApplicationProcess
		service.FeesDescription = opts.FeesDescription
		service.WaitTime = opts.WaitTime
		service.Fees = opts.Fees
		service.Accreditations = opts.Accreditations
		service.EligibilityDescription = opts.EligibilityDescription
		service.MinimumAge = opts.MinimumAge
		service.MaximumAge = opts.MaximumAge
		service.AssuredDate = opts.AssuredDate
		service.AssurerEmail = opts.AssurerEmail
		service.Licenses = opts.Licenses
		service.Alert = opts.Alert
	}

	return service, nil
}

// ServiceAreaOptions contains optional fields for creating a ServiceArea
type ServiceAreaOptions struct {
	ServiceID           *string
	ServiceAtLocationID *string
	Name                *string
	Description         *string
	Extent              *string
	ExtentType          *ExtentTypeEnum
	URI                 *string
}

func NewServiceArea(opts *ServiceAreaOptions) (*ServiceArea, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	serviceArea := &ServiceArea{
		CreatedAt: now,
		ID:        id,
	}

	if opts != nil {
		if opts.ServiceID != nil && !ValidateUUID(*opts.ServiceID) {
			return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
		}
		if opts.ServiceAtLocationID != nil && !ValidateUUID(*opts.ServiceAtLocationID) {
			return nil, fmt.Errorf("invalid service at location ID format: must be UUIDv4")
		}
		serviceArea.ServiceID = opts.ServiceID
		serviceArea.ServiceAtLocationID = opts.ServiceAtLocationID
		serviceArea.Name = opts.Name
		serviceArea.Description = opts.Description
		serviceArea.Extent = opts.Extent
		serviceArea.ExtentType = opts.ExtentType
		serviceArea.URI = opts.URI
	}

	return serviceArea, nil
}

// ServiceAtLocationOptions contains optional fields for creating a ServiceAtLocation
type ServiceAtLocationOptions struct {
	Description *string
}

func NewServiceAtLocation(serviceID, locationID string, opts *ServiceAtLocationOptions) (*ServiceAtLocation, error) {
	if !ValidateUUID(serviceID) {
		return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
	}
	if !ValidateUUID(locationID) {
		return nil, fmt.Errorf("invalid location ID format: must be UUIDv4")
	}

	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	serviceAtLocation := &ServiceAtLocation{
		CreatedAt:  now,
		ID:         id,
		ServiceID:  serviceID,
		LocationID: locationID,
	}

	if opts != nil {
		serviceAtLocation.Description = opts.Description
	}

	return serviceAtLocation, nil
}

// LocationOptions contains optional fields for creating a Location
type LocationOptions struct {
	OrganizationID         *string
	URL                    *string
	Name                   *string
	AlternateName          *string
	Description            *string
	Transportation         *string
	Latitude               *float64
	Longitude              *float64
	ExternalIdentifier     *string
	ExternalIdentifierType *string
}

func NewLocation(locationType LocationLocationTypeEnum, opts *LocationOptions) (*Location, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	location := &Location{
		CreatedAt:    now,
		ID:           id,
		LocationType: locationType,
	}

	if opts != nil {
		if opts.OrganizationID != nil && !ValidateUUID(*opts.OrganizationID) {
			return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
		}
		location.OrganizationID = opts.OrganizationID
		location.URL = opts.URL
		location.Name = opts.Name
		location.AlternateName = opts.AlternateName
		location.Description = opts.Description
		location.Transportation = opts.Transportation
		location.Latitude = opts.Latitude
		location.Longitude = opts.Longitude
		location.ExternalIdentifier = opts.ExternalIdentifier
		location.ExternalIdentifierType = opts.ExternalIdentifierType
	}

	return location, nil
}

// AddressOptions contains optional fields for creating an Address
type AddressOptions struct {
	LocationID *string
	Attention  *string
	Address2   *string
	Region     *string
}

func NewAddress(
	address1, city, stateProvince, postalCode, country string,
	addressType LocationLocationTypeEnum,
	opts *AddressOptions,
) (*Address, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	if len(country) != 2 {
		return nil, fmt.Errorf("country must be a 2-letter code")
	}

	address := &Address{
		CreatedAt:     now,
		ID:            id,
		Address1:      address1,
		City:          city,
		StateProvince: stateProvince,
		PostalCode:    postalCode,
		Country:       country,
		AddressType:   addressType,
	}

	if opts != nil {
		if opts.LocationID != nil && !ValidateUUID(*opts.LocationID) {
			return nil, fmt.Errorf("invalid location ID format: must be UUIDv4")
		}
		address.LocationID = opts.LocationID
		address.Attention = opts.Attention
		address.Address2 = opts.Address2
		address.Region = opts.Region
	}

	return address, nil
}

// RequiredDocumentOptions contains optional fields for creating a RequiredDocument
type RequiredDocumentOptions struct {
	ServiceID *string
	Document  *string
	URI       *string
}

func NewRequiredDocument(opts *RequiredDocumentOptions) (*RequiredDocument, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	requiredDocument := &RequiredDocument{
		CreatedAt: now,
		ID:        id,
	}

	if opts != nil {
		if opts.ServiceID != nil && !ValidateUUID(*opts.ServiceID) {
			return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
		}
		requiredDocument.ServiceID = opts.ServiceID
		requiredDocument.Document = opts.Document
		requiredDocument.URI = opts.URI
	}

	return requiredDocument, nil
}

// LanguageOptions contains optional fields for creating a Language
type LanguageOptions struct {
	ServiceID  *string
	LocationID *string
	PhoneID    *string
	Name       *string
	Code       *string
	Note       *string
}

func NewLanguage(opts *LanguageOptions) (*Language, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	language := &Language{
		CreatedAt: now,
		ID:        id,
	}

	if opts != nil {
		if opts.ServiceID != nil && !ValidateUUID(*opts.ServiceID) {
			return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
		}
		if opts.LocationID != nil && !ValidateUUID(*opts.LocationID) {
			return nil, fmt.Errorf("invalid location ID format: must be UUIDv4")
		}
		if opts.PhoneID != nil && !ValidateUUID(*opts.PhoneID) {
			return nil, fmt.Errorf("invalid phone ID format: must be UUIDv4")
		}
		language.ServiceID = opts.ServiceID
		language.LocationID = opts.LocationID
		language.PhoneID = opts.PhoneID
		language.Name = opts.Name
		language.Code = opts.Code
		language.Note = opts.Note
	}

	return language, nil
}

// AccessibilityOptions contains optional fields for creating an Accessibility
type AccessibilityOptions struct {
	LocationID  *string
	Description *string
	Details     *string
	URL         *string
}

func NewAccessibility(opts *AccessibilityOptions) (*Accessibility, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	accessibility := &Accessibility{
		CreatedAt: now,
		ID:        id,
	}

	if opts != nil {
		if opts.LocationID != nil && !ValidateUUID(*opts.LocationID) {
			return nil, fmt.Errorf("invalid location ID format: must be UUIDv4")
		}
		accessibility.LocationID = opts.LocationID
		accessibility.Description = opts.Description
		accessibility.Details = opts.Details
		accessibility.URL = opts.URL
	}

	return accessibility, nil
}

// AttributeOptions contains optional fields for creating an Attribute
type AttributeOptions struct {
	LinkType *string
	Value    *string
	Label    *string
}

func NewAttribute(taxonomyTermID, linkID, linkEntity string, opts *AttributeOptions) (*Attribute, error) {
	if !ValidateUUID(taxonomyTermID) {
		return nil, fmt.Errorf("invalid taxonomy term ID format: must be UUIDv4")
	}

	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	attribute := &Attribute{
		CreatedAt:      now,
		ID:             id,
		TaxonomyTermID: taxonomyTermID,
		LinkID:         linkID,
		LinkEntity:     linkEntity,
	}

	if opts != nil {
		attribute.LinkType = opts.LinkType
		attribute.Value = opts.Value
		attribute.Label = opts.Label
	}

	return attribute, nil
}

// TaxonomyOptions contains optional fields for creating a Taxonomy
type TaxonomyOptions struct {
	URI     *string
	Version *string
}

func NewTaxonomy(name, description string, opts *TaxonomyOptions) (*Taxonomy, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	taxonomy := &Taxonomy{
		CreatedAt:   now,
		ID:          id,
		Name:        name,
		Description: description,
	}

	if opts != nil {
		taxonomy.URI = opts.URI
		taxonomy.Version = opts.Version
	}

	return taxonomy, nil
}

// TaxonomyTermOptions contains optional fields for creating a TaxonomyTerm
type TaxonomyTermOptions struct {
	TaxonomyID  *string
	ParentID    *string
	Code        *string
	TaxonomyStr *string
	Language    *string
	TermURI     *string
}

func NewTaxonomyTerm(name, description string, opts *TaxonomyTermOptions) (*TaxonomyTerm, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	taxonomyTerm := &TaxonomyTerm{
		CreatedAt:   now,
		ID:          id,
		Name:        name,
		Description: description,
	}

	if opts != nil {
		if opts.TaxonomyID != nil && !ValidateUUID(*opts.TaxonomyID) {
			return nil, fmt.Errorf("invalid taxonomy ID format: must be UUIDv4")
		}
		if opts.ParentID != nil && !ValidateUUID(*opts.ParentID) {
			return nil, fmt.Errorf("invalid parent ID format: must be UUIDv4")
		}
		taxonomyTerm.TaxonomyID = opts.TaxonomyID
		taxonomyTerm.ParentID = opts.ParentID
		taxonomyTerm.Code = opts.Code
		taxonomyTerm.TaxonomyStr = opts.TaxonomyStr
		taxonomyTerm.Language = opts.Language
		taxonomyTerm.TermURI = opts.TermURI
	}

	return taxonomyTerm, nil
}

// ContactOptions contains optional fields for creating a Contact
type ContactOptions struct {
	OrganizationID      *string
	ServiceID           *string
	ServiceAtLocationID *string
	LocationID          *string
	Name                *string
	Title               *string
	Department          *string
	Email               *string
}

func NewContact(opts *ContactOptions) (*Contact, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	contact := &Contact{
		CreatedAt: now,
		ID:        id,
	}

	if opts != nil {
		if opts.OrganizationID != nil && !ValidateUUID(*opts.OrganizationID) {
			return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
		}
		if opts.ServiceID != nil && !ValidateUUID(*opts.ServiceID) {
			return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
		}
		if opts.ServiceAtLocationID != nil && !ValidateUUID(*opts.ServiceAtLocationID) {
			return nil, fmt.Errorf("invalid service at location ID format: must be UUIDv4")
		}
		if opts.LocationID != nil && !ValidateUUID(*opts.LocationID) {
			return nil, fmt.Errorf("invalid location ID format: must be UUIDv4")
		}
		contact.OrganizationID = opts.OrganizationID
		contact.ServiceID = opts.ServiceID
		contact.ServiceAtLocationID = opts.ServiceAtLocationID
		contact.LocationID = opts.LocationID
		contact.Name = opts.Name
		contact.Title = opts.Title
		contact.Department = opts.Department
		contact.Email = opts.Email
	}

	return contact, nil
}

// PhoneOptions contains optional fields for creating a Phone
type PhoneOptions struct {
	LocationID          *string
	ServiceID           *string
	OrganizationID      *string
	ContactID           *string
	ServiceAtLocationID *string
	Extension           *float64
	Type                *string
	Description         *string
}

func NewPhone(number string, opts *PhoneOptions) (*Phone, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	phone := &Phone{
		CreatedAt: now,
		ID:        id,
		Number:    number,
	}

	if opts != nil {
		if opts.LocationID != nil && !ValidateUUID(*opts.LocationID) {
			return nil, fmt.Errorf("invalid location ID format: must be UUIDv4")
		}
		if opts.ServiceID != nil && !ValidateUUID(*opts.ServiceID) {
			return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
		}
		if opts.OrganizationID != nil && !ValidateUUID(*opts.OrganizationID) {
			return nil, fmt.Errorf("invalid organization ID format: must be UUIDv4")
		}
		if opts.ContactID != nil && !ValidateUUID(*opts.ContactID) {
			return nil, fmt.Errorf("invalid contact ID format: must be UUIDv4")
		}
		if opts.ServiceAtLocationID != nil && !ValidateUUID(*opts.ServiceAtLocationID) {
			return nil, fmt.Errorf("invalid service at location ID format: must be UUIDv4")
		}
		phone.LocationID = opts.LocationID
		phone.ServiceID = opts.ServiceID
		phone.OrganizationID = opts.OrganizationID
		phone.ContactID = opts.ContactID
		phone.ServiceAtLocationID = opts.ServiceAtLocationID
		phone.Extension = opts.Extension
		phone.Type = opts.Type
		phone.Description = opts.Description
	}

	return phone, nil
}

// ScheduleOptions contains optional fields for creating a Schedule
type ScheduleOptions struct {
	ServiceID           *string
	LocationID          *string
	ServiceAtLocationID *string
	ValidFrom           *time.Time
	ValidTo             *time.Time
	DTStart             *time.Time
	Timezone            *float64
	Until               *time.Time
	Count               *int
	Wkst                *ScheduleWkstEnum
	Freq                *ScheduleFreqEnum
	Interval            *int
	Byday               *string
	Byweekno            *string
	Bymonthday          *string
	Byyearday           *string
	Description         *string
	OpensAt             *time.Time
	ClosesAt            *time.Time
	ScheduleLink        *string
	AttendingType       *string
	Notes               *string
}

func NewSchedule(opts *ScheduleOptions) (*Schedule, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	schedule := &Schedule{
		CreatedAt: now,
		ID:        id,
	}

	if opts != nil {
		if opts.ServiceID != nil && !ValidateUUID(*opts.ServiceID) {
			return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
		}
		if opts.LocationID != nil && !ValidateUUID(*opts.LocationID) {
			return nil, fmt.Errorf("invalid location ID format: must be UUIDv4")
		}
		if opts.ServiceAtLocationID != nil && !ValidateUUID(*opts.ServiceAtLocationID) {
			return nil, fmt.Errorf("invalid service at location ID format: must be UUIDv4")
		}
		schedule.ServiceID = opts.ServiceID
		schedule.LocationID = opts.LocationID
		schedule.ServiceAtLocationID = opts.ServiceAtLocationID
		schedule.ValidFrom = opts.ValidFrom
		schedule.ValidTo = opts.ValidTo
		schedule.DTStart = opts.DTStart
		schedule.Timezone = opts.Timezone
		schedule.Until = opts.Until
		schedule.Count = opts.Count
		schedule.Wkst = opts.Wkst
		schedule.Freq = opts.Freq
		schedule.Interval = opts.Interval
		schedule.Byday = opts.Byday
		schedule.Byweekno = opts.Byweekno
		schedule.Bymonthday = opts.Bymonthday
		schedule.Byyearday = opts.Byyearday
		schedule.Description = opts.Description
		schedule.OpensAt = opts.OpensAt
		schedule.ClosesAt = opts.ClosesAt
		schedule.ScheduleLink = opts.ScheduleLink
		schedule.AttendingType = opts.AttendingType
		schedule.Notes = opts.Notes
	}

	return schedule, nil
}

// ServiceCapacityOptions contains optional fields for creating a ServiceCapacity
type ServiceCapacityOptions struct {
	Maximum     *float64
	Description *string
}

func NewServiceCapacity(serviceID, unitID string, available float64, opts *ServiceCapacityOptions) (*ServiceCapacity, error) {
	if !ValidateUUID(serviceID) {
		return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
	}
	if !ValidateUUID(unitID) {
		return nil, fmt.Errorf("invalid unit ID format: must be UUIDv4")
	}

	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	serviceCapacity := &ServiceCapacity{
		CreatedAt: now,
		ID:        id,
		ServiceID: serviceID,
		UnitID:    unitID,
		Available: available,
	}

	if opts != nil {
		serviceCapacity.Maximum = opts.Maximum
		serviceCapacity.Description = opts.Description
	}

	return serviceCapacity, nil
}

// CostOptionOptions contains optional fields for creating a CostOption
type CostOptionOptions struct {
	ValidFrom         *time.Time
	ValidTo           *time.Time
	Option            *string
	Currency          *string
	Amount            *float64
	AmountDescription *string
}

func NewCostOption(serviceID string, opts *CostOptionOptions) (*CostOption, error) {
	if !ValidateUUID(serviceID) {
		return nil, fmt.Errorf("invalid service ID format: must be UUIDv4")
	}

	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	costOption := &CostOption{
		CreatedAt: now,
		ID:        id,
		ServiceID: serviceID,
	}

	if opts != nil {
		costOption.ValidFrom = opts.ValidFrom
		costOption.ValidTo = opts.ValidTo
		costOption.Option = opts.Option
		costOption.Currency = opts.Currency
		costOption.Amount = opts.Amount
		costOption.AmountDescription = opts.AmountDescription
	}

	return costOption, nil
}

// MetadataOptions contains all required fields for creating a Metadata
type MetadataOptions struct {
	ResourceType     string
	LastActionType   string
	FieldName        string
	PreviousValue    string
	ReplacementValue string
	UpdatedBy        string
}

// NewMetadata creates a new Metadata record with all required fields
func NewMetadata(
	resourceID string,
	callId string,
	resourceType string,
	lastActionType string,
	fieldName string,
	previousValue string,
	replacementValue string,
	updatedBy string,
) (*Metadata, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	metadata := &Metadata{
		CreatedAt:        now,
		ID:               id,
		CallID:           callId,
		ResourceID:       resourceID,
		ResourceType:     resourceType,
		LastActionDate:   now,
		LastActionType:   lastActionType,
		FieldName:        fieldName,
		PreviousValue:    previousValue,
		ReplacementValue: replacementValue,
		UpdatedBy:        updatedBy,
	}

	return metadata, nil
}

// MetaTableDescriptionOptions contains optional fields for creating a MetaTableDescription
type MetaTableDescriptionOptions struct {
	Name         *string
	Language     *string
	CharacterSet *string
}

// NewMetaTableDescription creates a new MetaTableDescription with optional fields
func NewMetaTableDescription(opts *MetaTableDescriptionOptions) (*MetaTableDescription, error) {
	now := getICalTime()
	id := newUUIDV4()

	if !ValidateUUID(id) {
		return nil, fmt.Errorf("failed to generate valid UUIDv4")
	}

	metaTableDesc := &MetaTableDescription{
		CreatedAt: now,
		ID:        id,
	}

	if opts != nil {
		metaTableDesc.Name = opts.Name
		metaTableDesc.Language = opts.Language
		metaTableDesc.CharacterSet = opts.CharacterSet
	}

	return metaTableDesc, nil
}
