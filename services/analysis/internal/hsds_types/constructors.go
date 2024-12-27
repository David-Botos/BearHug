// TODO: uuidv4 for ids
// TODO: check to make sure date time is created to the correct iCal schema
// TODO: break into its own package
// TODO: fork openreferral/specification and write a PR

package hsds_types

import (
	"errors"
	"time"
)

// Organization Constructor

type OrganizationOption func(*Organization)

func NewOrganization(id, name, description string, opts ...OrganizationOption) (*Organization, error) {
	if id == "" || name == "" || description == "" {
		return nil, errors.New("id, name, and description are required fields")
	}

	org := &Organization{
		ID:          id,
		Name:        name,
		Description: description,
	}

	for _, opt := range opts {
		opt(org)
	}

	return org, nil
}

func FK_ParentOrganizationID(parentID string) OrganizationOption {
	return func(o *Organization) {
		o.ParentOrganizationID = parentID
	}
}

func WithAlternateName(name string) OrganizationOption {
	return func(o *Organization) {
		o.AlternateName = name
	}
}

func WithEmail(email string) OrganizationOption {
	return func(o *Organization) {
		o.Email = email
	}
}

func WithLegalStatus(status string) OrganizationOption {
	return func(o *Organization) {
		o.LegalStatus = status
	}
}

func WithLogo(logo string) OrganizationOption {
	return func(o *Organization) {
		o.Logo = logo
	}
}

func FK_TaxInfo(taxID, taxStatus string) OrganizationOption {
	return func(o *Organization) {
		o.TaxID = taxID
		o.TaxStatus = taxStatus
	}
}

func WithURI(uri string) OrganizationOption {
	return func(o *Organization) {
		o.URI = uri
	}
}

func WithWebsite(website string) OrganizationOption {
	return func(o *Organization) {
		o.Website = website
	}
}

func WithYearIncorporated(year int) OrganizationOption {
	return func(o *Organization) {
		o.YearIncorporated = year
	}
}

// Organization Identifier Constructor

type OrganizationIdentifierOption func(*OrganizationIdentifier)

func NewOrganizationIdentifier(id, organizationID, identifierType, identifier string, opts ...OrganizationIdentifierOption) (*OrganizationIdentifier, error) {
	if id == "" || organizationID == "" || identifierType == "" || identifier == "" {
		return nil, errors.New("id, organization_id, identifier_type, and identifier are required fields")
	}

	orgIdentifier := &OrganizationIdentifier{
		ID:             id,
		OrganizationID: organizationID,
		IdentifierType: identifierType,
		Identifier:     identifier,
	}

	for _, opt := range opts {
		opt(orgIdentifier)
	}

	return orgIdentifier, nil
}

func WithIdentifierScheme(scheme string) OrganizationIdentifierOption {
	return func(o *OrganizationIdentifier) {
		o.IdentifierScheme = scheme
	}
}

// URL Constructor

type URLOption func(*URL)

func NewURL(id, url string, opts ...URLOption) (*URL, error) {
	if id == "" || url == "" {
		return nil, errors.New("id and url are required fields")
	}

	u := &URL{
		ID:  id,
		URL: url,
	}

	for _, opt := range opts {
		opt(u)
	}

	return u, nil
}

func WithURLLabel(label string) URLOption {
	return func(u *URL) {
		u.Label = label
	}
}

func FK_URLOrganizationID(orgID string) URLOption {
	return func(u *URL) {
		u.OrganizationID = orgID
	}
}

func FK_URLServiceID(serviceID string) URLOption {
	return func(u *URL) {
		u.ServiceID = serviceID
	}
}

// Funding Constructor

type FundingOption func(*Funding)

func NewFunding(id string, opts ...FundingOption) (*Funding, error) {
	if id == "" {
		return nil, errors.New("id is a required field")
	}

	funding := &Funding{
		ID: id,
	}

	for _, opt := range opts {
		opt(funding)
	}

	return funding, nil
}

func FK_FundingOrganizationID(organizationID string) FundingOption {
	return func(f *Funding) {
		f.OrganizationID = organizationID
	}
}

func FK_FundingServiceID(serviceID string) FundingOption {
	return func(f *Funding) {
		f.ServiceID = serviceID
	}
}

func WithSource(source string) FundingOption {
	return func(f *Funding) {
		f.Source = source
	}
}

// Unit Constructor

type UnitOption func(*Unit)

func NewUnit(id, name string, opts ...UnitOption) (*Unit, error) {
	if id == "" || name == "" {
		return nil, errors.New("id and name are required fields")
	}

	unit := &Unit{
		ID:   id,
		Name: name,
	}

	for _, opt := range opts {
		opt(unit)
	}

	return unit, nil
}

func WithScheme(scheme string) UnitOption {
	return func(u *Unit) {
		u.Scheme = scheme
	}
}

func WithIdentifier(identifier string) UnitOption {
	return func(u *Unit) {
		u.Identifier = identifier
	}
}

func WithUnitURI(uri string) UnitOption {
	return func(u *Unit) {
		u.URI = uri
	}
}

// Program Option Constructor
type ProgramOption func(*Program)

func NewProgram(id, organizationID, name, description string, opts ...ProgramOption) (*Program, error) {
	if id == "" || organizationID == "" || name == "" || description == "" {
		return nil, errors.New("id, organization_id, name, and description are required fields")
	}

	prog := &Program{
		ID:             id,
		OrganizationID: organizationID,
		Name:           name,
		Description:    description,
	}

	for _, opt := range opts {
		opt(prog)
	}

	return prog, nil
}

func WithProgramAlternateName(name string) ProgramOption {
	return func(p *Program) {
		p.AlternateName = name
	}
}

// Service Constructor

type ServiceOption func(*Service)

func NewService(id, organizationID, name string, status ServiceStatusEnum, opts ...ServiceOption) (*Service, error) {
	if id == "" || organizationID == "" || name == "" {
		return nil, errors.New("id, organizationID, and name are required fields")
	}

	svc := &Service{
		ID:             id,
		OrganizationID: organizationID,
		Name:           name,
		Status:         status,
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc, nil
}

func FK_ServiceProgramID(programID string) ServiceOption {
	return func(s *Service) {
		s.ProgramID = programID
	}
}

func WithServiceAlternateName(name string) ServiceOption {
	return func(s *Service) {
		s.AlternateName = name
	}
}

func WithServiceDescription(description string) ServiceOption {
	return func(s *Service) {
		s.Description = description
	}
}

func WithServiceURL(url string) ServiceOption {
	return func(s *Service) {
		s.URL = url
	}
}

func WithServiceEmail(email string) ServiceOption {
	return func(s *Service) {
		s.Email = email
	}
}

func WithInterpretationServices(services string) ServiceOption {
	return func(s *Service) {
		s.InterpretationServices = services
	}
}

func WithApplicationProcess(process string) ServiceOption {
	return func(s *Service) {
		s.ApplicationProcess = process
	}
}

func WithFeesDescription(description string) ServiceOption {
	return func(s *Service) {
		s.FeesDescription = description
	}
}

func WithAccreditations(accreditations string) ServiceOption {
	return func(s *Service) {
		s.Accreditations = accreditations
	}
}

func WithEligibilityDescription(description string) ServiceOption {
	return func(s *Service) {
		s.EligibilityDescription = description
	}
}

func WithAgeRange(min, max float64) ServiceOption {
	return func(s *Service) {
		s.MinimumAge = min
		s.MaximumAge = max
	}
}

func WithAssurance(assuredDate time.Time, assurerEmail string) ServiceOption {
	return func(s *Service) {
		s.AssuredDate = assuredDate
		s.AssurerEmail = assurerEmail
	}
}

func WithAlert(alert string) ServiceOption {
	return func(s *Service) {
		s.Alert = alert
	}
}

func WithLastModified(lastModified time.Time) ServiceOption {
	return func(s *Service) {
		s.LastModified = lastModified
	}
}

// Service Area Constructor

type ServiceAreaOption func(*ServiceArea)

func NewServiceArea(id string, opts ...ServiceAreaOption) (*ServiceArea, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	sa := &ServiceArea{
		ID: id,
	}

	for _, opt := range opts {
		opt(sa)
	}

	return sa, nil
}

func FK_AreaServiceID(serviceID string) ServiceAreaOption {
	return func(sa *ServiceArea) {
		sa.ServiceID = serviceID
	}
}

func FK_ServiceAtLocationID(locationID string) ServiceAreaOption {
	return func(sa *ServiceArea) {
		sa.ServiceAtLocationID = locationID
	}
}

func WithServiceAreaName(name string) ServiceAreaOption {
	return func(sa *ServiceArea) {
		sa.Name = name
	}
}

func WithServiceAreaDescription(description string) ServiceAreaOption {
	return func(sa *ServiceArea) {
		sa.Description = description
	}
}

func WithExtent(extent string) ServiceAreaOption {
	return func(sa *ServiceArea) {
		sa.Extent = extent
	}
}

func WithExtentType(extentType ExtentTypeEnum) ServiceAreaOption {
	return func(sa *ServiceArea) {
		sa.ExtentType = extentType
	}
}

func WithServiceAreaURI(uri string) ServiceAreaOption {
	return func(sa *ServiceArea) {
		sa.URI = uri
	}
}

// Service at Location Constructor

type ServiceAtLocationOption func(*ServiceAtLocation)

func NewServiceAtLocation(id, serviceID, locationID string, opts ...ServiceAtLocationOption) (*ServiceAtLocation, error) {
	if id == "" || serviceID == "" || locationID == "" {
		return nil, errors.New("id, service_id, and location_id are required fields")
	}

	sal := &ServiceAtLocation{
		ID:         id,
		ServiceID:  serviceID,
		LocationID: locationID,
	}

	for _, opt := range opts {
		opt(sal)
	}

	return sal, nil
}

func WithServiceAtLocationDescription(description string) ServiceAtLocationOption {
	return func(s *ServiceAtLocation) {
		s.Description = description
	}
}

// Location Constructor

type LocationOption func(*Location)

func NewLocation(id string, locationType LocationLocationTypeEnum, opts ...LocationOption) (*Location, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	loc := &Location{
		ID:           id,
		LocationType: locationType,
	}

	for _, opt := range opts {
		opt(loc)
	}

	return loc, nil
}

func FK_LocationOrganizationID(orgID string) LocationOption {
	return func(l *Location) {
		l.OrganizationID = orgID
	}
}

func WithLocationURL(url string) LocationOption {
	return func(l *Location) {
		l.URL = url
	}
}

func WithLocationName(name string) LocationOption {
	return func(l *Location) {
		l.Name = name
	}
}

func WithLocationAlternateName(name string) LocationOption {
	return func(l *Location) {
		l.AlternateName = name
	}
}

func WithLocationDescription(description string) LocationOption {
	return func(l *Location) {
		l.Description = description
	}
}

func WithTransportation(transportation string) LocationOption {
	return func(l *Location) {
		l.Transportation = transportation
	}
}

func WithCoordinates(latitude, longitude float64) LocationOption {
	return func(l *Location) {
		l.Latitude = latitude
		l.Longitude = longitude
	}
}

func WithExternalIdentifier(id string, idType string) LocationOption {
	return func(l *Location) {
		l.ExternalIdentifier = id
		l.ExternalIdentifierType = idType
	}
}

// Address Option Constructor

type AddressOption func(*Address)

func NewAddress(id, address1, city, stateProvince, postalCode, country string, addressType LocationLocationTypeEnum) (*Address, error) {
	if id == "" || address1 == "" || city == "" || stateProvince == "" || postalCode == "" || country == "" {
		return nil, errors.New("id, address1, city, stateProvince, postalCode, and country are required fields")
	}

	if len(country) != 2 {
		return nil, errors.New("country must be a 2-character code")
	}

	if addressType != "physical" && addressType != "postal" && addressType != "virtual" {
		return nil, errors.New("address type must be physical, postal, or virtual")
	}

	addr := &Address{
		ID:            id,
		Address1:      address1,
		City:          city,
		StateProvince: stateProvince,
		PostalCode:    postalCode,
		Country:       country,
		AddressType:   addressType,
	}

	return addr, nil
}

func FK_AddressLocationID(locationID string) AddressOption {
	return func(a *Address) {
		a.LocationID = locationID
	}
}

func WithAttention(attention string) AddressOption {
	return func(a *Address) {
		a.Attention = attention
	}
}

func WithAddress2(address2 string) AddressOption {
	return func(a *Address) {
		a.Address2 = address2
	}
}

func WithRegion(region string) AddressOption {
	return func(a *Address) {
		a.Region = region
	}
}

// Required Document Constructor

type RequiredDocumentOption func(*RequiredDocument)

func NewRequiredDocument(id string, opts ...RequiredDocumentOption) (*RequiredDocument, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	doc := &RequiredDocument{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(doc)
	}

	return doc, nil
}

func FK_DocumentServiceID(serviceID string) RequiredDocumentOption {
	return func(d *RequiredDocument) {
		d.ServiceID = serviceID
	}
}

func WithDocument(document string) RequiredDocumentOption {
	return func(d *RequiredDocument) {
		d.Document = document
	}
}

func WithDocumentURI(uri string) RequiredDocumentOption {
	return func(d *RequiredDocument) {
		d.URI = uri
	}
}

func WithCreatedAt(t time.Time) RequiredDocumentOption {
	return func(d *RequiredDocument) {
		d.CreatedAt = t
	}
}

func WithUpdatedAt(t time.Time) RequiredDocumentOption {
	return func(d *RequiredDocument) {
		d.UpdatedAt = t
	}
}

// Language Constructor

type LanguageOption func(*Language)

func NewLanguage(id string, opts ...LanguageOption) (*Language, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	now := time.Now()
	lang := &Language{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}

	for _, opt := range opts {
		opt(lang)
	}

	return lang, nil
}

func FK_LanguageServiceID(serviceID string) LanguageOption {
	return func(l *Language) {
		l.ServiceID = serviceID
	}
}

func FK_LanguageLocationID(locationID string) LanguageOption {
	return func(l *Language) {
		l.LocationID = locationID
	}
}

func FK_LanguagePhoneID(phoneID string) LanguageOption {
	return func(l *Language) {
		l.PhoneID = phoneID
	}
}

func WithLanguageName(name string) LanguageOption {
	return func(l *Language) {
		l.Name = name
	}
}

func WithLanguageCode(code string) LanguageOption {
	return func(l *Language) {
		l.Code = code
	}
}

func WithLanguageNote(note string) LanguageOption {
	return func(l *Language) {
		l.Note = note
	}
}

// Accessibility Constructor

type AccessibilityOption func(*Accessibility)

func NewAccessibility(id string, opts ...AccessibilityOption) (*Accessibility, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	acc := &Accessibility{
		ID: id,
	}

	for _, opt := range opts {
		opt(acc)
	}

	return acc, nil
}

func FK_AccessibilityLocationID(locationID string) AccessibilityOption {
	return func(a *Accessibility) {
		a.LocationID = locationID
	}
}

func WithAccessibilityDescription(description string) AccessibilityOption {
	return func(a *Accessibility) {
		a.Description = description
	}
}

func WithAccessibilityDetails(details string) AccessibilityOption {
	return func(a *Accessibility) {
		a.Details = details
	}
}

func WithAccessibilityURL(url string) AccessibilityOption {
	return func(a *Accessibility) {
		a.URL = url
	}
}

// Attribute Constructor

type AttributeOption func(*Attribute)

func NewAttribute(id, taxonomyTermID, linkID, linkEntity string, opts ...AttributeOption) (*Attribute, error) {
	if id == "" || taxonomyTermID == "" || linkID == "" || linkEntity == "" {
		return nil, errors.New("id, taxonomy_term_id, link_id, and link_entity are required fields")
	}

	attr := &Attribute{
		ID:             id,
		TaxonomyTermID: taxonomyTermID,
		LinkID:         linkID,
		LinkEntity:     linkEntity,
		CreatedAt:      time.Now(),
	}

	for _, opt := range opts {
		opt(attr)
	}

	return attr, nil
}

func WithLinkType(linkType string) AttributeOption {
	return func(a *Attribute) {
		a.LinkType = linkType
	}
}

func WithValue(value string) AttributeOption {
	return func(a *Attribute) {
		a.Value = value
	}
}

func WithLabel(label string) AttributeOption {
	return func(a *Attribute) {
		a.Label = label
	}
}

func WithAttributeCreatedAt(createdAt time.Time) AttributeOption {
	return func(a *Attribute) {
		a.CreatedAt = createdAt
	}
}

func WithAttributeUpdatedAt(updatedAt time.Time) AttributeOption {
	return func(a *Attribute) {
		a.UpdatedAt = updatedAt
	}
}

// Taxonomy Constructor

type TaxonomyOption func(*Taxonomy)

func NewTaxonomy(id, name, description string, opts ...TaxonomyOption) (*Taxonomy, error) {
	if id == "" || name == "" || description == "" {
		return nil, errors.New("id, name, and description are required fields")
	}

	taxonomy := &Taxonomy{
		ID:          id,
		Name:        name,
		Description: description,
	}

	for _, opt := range opts {
		opt(taxonomy)
	}

	return taxonomy, nil
}

func WithTaxonomyURI(uri string) TaxonomyOption {
	return func(t *Taxonomy) {
		t.URI = uri
	}
}

func WithVersion(version string) TaxonomyOption {
	return func(t *Taxonomy) {
		t.Version = version
	}
}

// Taxonomy Term Constructor

type TaxonomyTermOption func(*TaxonomyTerm)

func NewTaxonomyTerm(id, name, description string, opts ...TaxonomyTermOption) (*TaxonomyTerm, error) {
	if id == "" || name == "" || description == "" {
		return nil, errors.New("id, name, and description are required fields")
	}

	term := &TaxonomyTerm{
		ID:          id,
		Name:        name,
		Description: description,
	}

	for _, opt := range opts {
		opt(term)
	}

	return term, nil
}

func FK_TaxonomyID(taxonomyID string) TaxonomyTermOption {
	return func(t *TaxonomyTerm) {
		t.TaxonomyID = taxonomyID
	}
}

func FK_ParentID(parentID string) TaxonomyTermOption {
	return func(t *TaxonomyTerm) {
		t.ParentID = parentID
	}
}

func WithCode(code string) TaxonomyTermOption {
	return func(t *TaxonomyTerm) {
		t.Code = code
	}
}

func WithTaxonomy(taxonomy string) TaxonomyTermOption {
	return func(t *TaxonomyTerm) {
		t.Taxonomy = taxonomy
	}
}

func WithLanguage(language string) TaxonomyTermOption {
	return func(t *TaxonomyTerm) {
		t.Language = language
	}
}

func WithTermURI(uri string) TaxonomyTermOption {
	return func(t *TaxonomyTerm) {
		t.TermURI = uri
	}
}

// Contact Constructor

type ContactOption func(*Contact)

func NewContact(id string, opts ...ContactOption) (*Contact, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	contact := &Contact{
		ID:        id,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(contact)
	}

	return contact, nil
}

func WithContactOrganization(orgID string) ContactOption {
	return func(c *Contact) {
		c.OrganizationID = orgID
	}
}

func WithContactService(serviceID string) ContactOption {
	return func(c *Contact) {
		c.ServiceID = serviceID
	}
}

func WithServiceAtLocation(serviceAtLocationID string) ContactOption {
	return func(c *Contact) {
		c.ServiceAtLocationID = serviceAtLocationID
	}
}

func WithLocation(locationID string) ContactOption {
	return func(c *Contact) {
		c.LocationID = locationID
	}
}

func WithContactName(name string) ContactOption {
	return func(c *Contact) {
		c.Name = name
	}
}

func WithTitle(title string) ContactOption {
	return func(c *Contact) {
		c.Title = title
	}
}

func WithDepartment(department string) ContactOption {
	return func(c *Contact) {
		c.Department = department
	}
}

func WithContactEmail(email string) ContactOption {
	return func(c *Contact) {
		c.Email = email
	}
}

// Phone Constructor

type PhoneOption func(*Phone)

func NewPhone(id, number string, opts ...PhoneOption) (*Phone, error) {
	if id == "" || number == "" {
		return nil, errors.New("id and number are required fields")
	}

	phone := &Phone{
		ID:     id,
		Number: number,
	}

	for _, opt := range opts {
		opt(phone)
	}

	return phone, nil
}

func FK_LocationID(locationID string) PhoneOption {
	return func(p *Phone) {
		p.LocationID = locationID
	}
}

func FK_PhoneServiceID(serviceID string) PhoneOption {
	return func(p *Phone) {
		p.ServiceID = serviceID
	}
}

func FK_PhoneOrganizationID(organizationID string) PhoneOption {
	return func(p *Phone) {
		p.OrganizationID = organizationID
	}
}

func FK_ContactID(contactID string) PhoneOption {
	return func(p *Phone) {
		p.ContactID = contactID
	}
}

func FK_PhoneServiceAtLocationID(serviceAtLocationID string) PhoneOption {
	return func(p *Phone) {
		p.ServiceAtLocationID = serviceAtLocationID
	}
}

func WithExtension(extension float64) PhoneOption {
	return func(p *Phone) {
		p.Extension = extension
	}
}

func WithType(phoneType string) PhoneOption {
	return func(p *Phone) {
		p.Type = phoneType
	}
}

func WithPhoneDescription(description string) PhoneOption {
	return func(p *Phone) {
		p.Description = description
	}
}

// Schedule Constructor

type ScheduleOption func(*Schedule)

func NewSchedule(id string, opts ...ScheduleOption) (*Schedule, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	schedule := &Schedule{
		ID: id,
	}

	for _, opt := range opts {
		opt(schedule)
	}

	return schedule, nil
}

func FK_SchedServiceID(serviceID string) ScheduleOption {
	return func(s *Schedule) {
		s.ServiceID = serviceID
	}
}

func FK_SchedLocationID(locationID string) ScheduleOption {
	return func(s *Schedule) {
		s.LocationID = locationID
	}
}

func FK_SchedServiceAtLocationID(serviceAtLocationID string) ScheduleOption {
	return func(s *Schedule) {
		s.ServiceAtLocationID = serviceAtLocationID
	}
}

func WithValidityPeriod(from, to time.Time) ScheduleOption {
	return func(s *Schedule) {
		s.ValidFrom = from
		s.ValidTo = to
	}
}

func WithDTStart(dtstart time.Time) ScheduleOption {
	return func(s *Schedule) {
		s.DTStart = dtstart
	}
}

func WithTimezone(timezone float64) ScheduleOption {
	return func(s *Schedule) {
		s.Timezone = timezone
	}
}

func WithUntil(until time.Time) ScheduleOption {
	return func(s *Schedule) {
		s.Until = until
	}
}

func WithCount(count int) ScheduleOption {
	return func(s *Schedule) {
		s.Count = count
	}
}

func WithWkst(wkst ScheduleWkstEnum) ScheduleOption {
	return func(s *Schedule) {
		s.Wkst = wkst
	}
}

func WithFreq(freq ScheduleFreqEnum) ScheduleOption {
	return func(s *Schedule) {
		s.Freq = freq
	}
}

func WithInterval(interval int) ScheduleOption {
	return func(s *Schedule) {
		s.Interval = interval
	}
}

func WithByDay(byday string) ScheduleOption {
	return func(s *Schedule) {
		s.Byday = byday
	}
}

func WithByWeekNo(byweekno string) ScheduleOption {
	return func(s *Schedule) {
		s.Byweekno = byweekno
	}
}

func WithByMonthDay(bymonthday string) ScheduleOption {
	return func(s *Schedule) {
		s.Bymonthday = bymonthday
	}
}

func WithByYearDay(byyearday string) ScheduleOption {
	return func(s *Schedule) {
		s.Byyearday = byyearday
	}
}

func WithDescription(description string) ScheduleOption {
	return func(s *Schedule) {
		s.Description = description
	}
}

func WithOpeningHours(opensAt, closesAt time.Time) ScheduleOption {
	return func(s *Schedule) {
		s.OpensAt = opensAt
		s.ClosesAt = closesAt
	}
}

func WithScheduleLink(link string) ScheduleOption {
	return func(s *Schedule) {
		s.ScheduleLink = link
	}
}

func WithAttendingType(attendingType string) ScheduleOption {
	return func(s *Schedule) {
		s.AttendingType = attendingType
	}
}

func WithNotes(notes string) ScheduleOption {
	return func(s *Schedule) {
		s.Notes = notes
	}
}

// Service Capacity Constructor

type ServiceCapacityOption func(*ServiceCapacity)

func NewServiceCapacity(id, serviceID, unitID string, available float64, updated time.Time, opts ...ServiceCapacityOption) (*ServiceCapacity, error) {
	if id == "" || serviceID == "" || unitID == "" {
		return nil, errors.New("id, service_id, and unit_id are required fields")
	}

	sc := &ServiceCapacity{
		ID:        id,
		ServiceID: serviceID,
		UnitID:    unitID,
		Available: available,
		Updated:   updated,
	}

	for _, opt := range opts {
		opt(sc)
	}

	return sc, nil
}

func WithMaximum(max float64) ServiceCapacityOption {
	return func(sc *ServiceCapacity) {
		sc.Maximum = max
	}
}

func WithCapacityDescription(desc string) ServiceCapacityOption {
	return func(sc *ServiceCapacity) {
		sc.Description = desc
	}
}

// Cost Option Constructor

type CostOptionOption func(*CostOption)

func NewCostOption(id, serviceID string, opts ...CostOptionOption) (*CostOption, error) {
	if id == "" || serviceID == "" {
		return nil, errors.New("id and service_id are required fields")
	}

	costOption := &CostOption{
		ID:        id,
		ServiceID: serviceID,
	}

	for _, opt := range opts {
		opt(costOption)
	}

	return costOption, nil
}

func WithValidFrom(t time.Time) CostOptionOption {
	return func(c *CostOption) {
		c.ValidFrom = t
	}
}

func WithValidTo(t time.Time) CostOptionOption {
	return func(c *CostOption) {
		c.ValidTo = t
	}
}

func WithOption(opt string) CostOptionOption {
	return func(c *CostOption) {
		c.Option = opt
	}
}

func WithCurrency(currency string) CostOptionOption {
	return func(c *CostOption) {
		c.Currency = currency
	}
}

func WithAmount(amount float64) CostOptionOption {
	return func(c *CostOption) {
		c.Amount = amount
	}
}

func WithAmountDescription(description string) CostOptionOption {
	return func(c *CostOption) {
		c.AmountDescription = description
	}
}

// Metadata Constructor

type MetadataOption func(*Metadata)

func NewMetadata(id, resourceID, resourceType, lastActionType, fieldName, previousValue, replacementValue, updatedBy string, lastActionDate time.Time) (*Metadata, error) {
	if id == "" || resourceID == "" || resourceType == "" || lastActionType == "" || fieldName == "" ||
		previousValue == "" || replacementValue == "" || updatedBy == "" || lastActionDate.IsZero() {
		return nil, errors.New("all fields are required")
	}

	metadata := &Metadata{
		ID:               id,
		ResourceID:       resourceID,
		ResourceType:     resourceType,
		LastActionDate:   lastActionDate,
		LastActionType:   lastActionType,
		FieldName:        fieldName,
		PreviousValue:    previousValue,
		ReplacementValue: replacementValue,
		UpdatedBy:        updatedBy,
	}

	return metadata, nil
}

// Meta Table Description Constructor

type MetaTableDescriptionOption func(*MetaTableDescription)

func NewMetaTableDescription(id string, opts ...MetaTableDescriptionOption) (*MetaTableDescription, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	mtd := &MetaTableDescription{
		ID: id,
	}

	for _, opt := range opts {
		opt(mtd)
	}

	return mtd, nil
}

func WithName(name string) MetaTableDescriptionOption {
	return func(m *MetaTableDescription) {
		m.Name = name
	}
}

func WithMetaTableDescLanguage(language string) MetaTableDescriptionOption {
	return func(m *MetaTableDescription) {
		m.Language = language
	}
}

func WithCharacterSet(characterSet string) MetaTableDescriptionOption {
	return func(m *MetaTableDescription) {
		m.CharacterSet = characterSet
	}
}
