package types

import (
	"time"
)

// Database represents the main database structure
type Database struct {
	Public PublicSchema
}

// PublicSchema represents the public schema of the database
type PublicSchema struct {
	Tables map[string]TableContent
	Views  map[string]TableContent
}

// TableContent represents the content of a table or view
type TableContent struct {
	Row interface{}
}

// Common types for database relationships
type Relationship struct {
	ForeignKeyName     string   `json:"foreignKeyName"`
	Columns            []string `json:"columns"`
	IsOneToOne         bool     `json:"isOneToOne"`
	ReferencedRelation string   `json:"referencedRelation"`
	ReferencedColumns  []string `json:"referencedColumns"`
}

//// -- HSDS Definitions -- ////

type Organization struct {
	Row struct {
		ID                   string  `json:"id" validate:"required"`
		Name                 string  `json:"name" validate:"required"`
		AlternateName        *string `json:"alternate_name,omitempty"`
		Description          string  `json:"description" validate:"required"`
		Email                *string `json:"email,omitempty"`
		LegalStatus          *string `json:"legal_status,omitempty"`
		Logo                 *string `json:"logo,omitempty"`
		ParentOrganizationID *string `json:"parent_organization_id,omitempty"`
		TaxID                *string `json:"tax_id,omitempty"`
		TaxStatus            *string `json:"tax_status,omitempty"`
		URI                  *string `json:"uri,omitempty"`
		Website              *string `json:"website,omitempty"`
		YearIncorporated     *int    `json:"year_incorporated,omitempty"`
	}
	Insert struct {
		ID                   string  `json:"id" validate:"required"`
		Name                 string  `json:"name" validate:"required"`
		AlternateName        *string `json:"alternate_name,omitempty"`
		Description          string  `json:"description" validate:"required"`
		Email                *string `json:"email,omitempty"`
		LegalStatus          *string `json:"legal_status,omitempty"`
		Logo                 *string `json:"logo,omitempty"`
		ParentOrganizationID *string `json:"parent_organization_id,omitempty"`
		TaxID                *string `json:"tax_id,omitempty"`
		TaxStatus            *string `json:"tax_status,omitempty"`
		URI                  *string `json:"uri,omitempty"`
		Website              *string `json:"website,omitempty"`
		YearIncorporated     *int    `json:"year_incorporated,omitempty"`
	}
	Update struct {
		ID                   *string `json:"id,omitempty"`
		Name                 *string `json:"name,omitempty"`
		AlternateName        *string `json:"alternate_name,omitempty"`
		Description          *string `json:"description,omitempty"`
		Email                *string `json:"email,omitempty"`
		LegalStatus          *string `json:"legal_status,omitempty"`
		Logo                 *string `json:"logo,omitempty"`
		ParentOrganizationID *string `json:"parent_organization_id,omitempty"`
		TaxID                *string `json:"tax_id,omitempty"`
		TaxStatus            *string `json:"tax_status,omitempty"`
		URI                  *string `json:"uri,omitempty"`
		Website              *string `json:"website,omitempty"`
		YearIncorporated     *int    `json:"year_incorporated,omitempty"`
	}
}

type Program struct {
	Row struct {
		ID             string  `json:"id" validate:"required"`
		OrganizationID string  `json:"organization_id" validate:"required"`
		Name           string  `json:"name" validate:"required"`
		AlternateName  *string `json:"alternate_name,omitempty"`
		Description    string  `json:"description" validate:"required"`
	}
	Insert struct {
		ID             string  `json:"id" validate:"required"`
		OrganizationID string  `json:"organization_id" validate:"required"`
		Name           string  `json:"name" validate:"required"`
		AlternateName  *string `json:"alternate_name,omitempty"`
		Description    string  `json:"description" validate:"required"`
	}
	Update struct {
		ID             *string `json:"id,omitempty"`
		OrganizationID *string `json:"organization_id,omitempty"`
		Name           *string `json:"name,omitempty"`
		AlternateName  *string `json:"alternate_name,omitempty"`
		Description    *string `json:"description,omitempty"`
	}
}

type RequiredDocument struct {
	Row struct {
		ID        string     `db:"id" json:"id" validate:"required"`
		ServiceID *string    `db:"service_id" json:"service_id,omitempty"`
		Document  *string    `db:"document" json:"document,omitempty"`
		URI       *string    `db:"uri" json:"uri,omitempty"`
		CreatedAt time.Time  `db:"created_at" json:"created_at"`
		UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	}
	Insert struct {
		ID        string  `db:"id" json:"id" validate:"required"`
		ServiceID *string `db:"service_id" json:"service_id,omitempty"`
		Document  *string `db:"document" json:"document,omitempty"`
		URI       *string `db:"uri" json:"uri,omitempty"`
	}
	Update struct {
		ID        *string `db:"id" json:"id,omitempty"`
		ServiceID *string `db:"service_id" json:"service_id,omitempty"`
		Document  *string `db:"document" json:"document,omitempty"`
		URI       *string `db:"uri" json:"uri,omitempty"`
	}
}

type Attribute struct {
	Row struct {
		ID             string     `db:"id" json:"id" validate:"required"`
		LinkID         string     `db:"link_id" json:"link_id" validate:"required"`
		TaxonomyTermID string     `db:"taxonomy_term_id" json:"taxonomy_term_id" validate:"required"`
		LinkType       *string    `db:"link_type" json:"link_type,omitempty"`
		LinkEntity     string     `db:"link_entity" json:"link_entity" validate:"required"`
		Value          *string    `db:"value" json:"value,omitempty"`
		Label          *string    `db:"label" json:"label,omitempty"`
		CreatedAt      time.Time  `db:"created_at" json:"created_at"`
		UpdatedAt      *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	}
	Insert struct {
		ID             string  `db:"id" json:"id" validate:"required"`
		LinkID         string  `db:"link_id" json:"link_id" validate:"required"`
		TaxonomyTermID string  `db:"taxonomy_term_id" json:"taxonomy_term_id" validate:"required"`
		LinkType       *string `db:"link_type" json:"link_type,omitempty"`
		LinkEntity     string  `db:"link_entity" json:"link_entity" validate:"required"`
		Value          *string `db:"value" json:"value,omitempty"`
		Label          *string `db:"label" json:"label,omitempty"`
	}
	Update struct {
		ID             *string `db:"id" json:"id,omitempty"`
		LinkID         *string `db:"link_id" json:"link_id,omitempty"`
		TaxonomyTermID *string `db:"taxonomy_term_id" json:"taxonomy_term_id,omitempty"`
		LinkType       *string `db:"link_type" json:"link_type,omitempty"`
		LinkEntity     *string `db:"link_entity" json:"link_entity,omitempty"`
		Value          *string `db:"value" json:"value,omitempty"`
		Label          *string `db:"label" json:"label,omitempty"`
	}
}

type Contact struct {
	Row struct {
		ID                  string     `db:"id" json:"id" validate:"required"`
		OrganizationID      *string    `db:"organization_id" json:"organization_id,omitempty"`
		ServiceID           *string    `db:"service_id" json:"service_id,omitempty"`
		ServiceAtLocationID *string    `db:"service_at_location_id" json:"service_at_location_id,omitempty"`
		LocationID          *string    `db:"location_id" json:"location_id,omitempty"`
		Name                *string    `db:"name" json:"name,omitempty"`
		Title               *string    `db:"title" json:"title,omitempty"`
		Department          *string    `db:"department" json:"department,omitempty"`
		Email               *string    `db:"email" json:"email,omitempty"`
		CreatedAt           time.Time  `db:"created_at" json:"created_at"`
		UpdatedAt           *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	}
	Insert struct {
		ID                  string  `db:"id" json:"id" validate:"required"`
		OrganizationID      *string `db:"organization_id" json:"organization_id,omitempty"`
		ServiceID           *string `db:"service_id" json:"service_id,omitempty"`
		ServiceAtLocationID *string `db:"service_at_location_id" json:"service_at_location_id,omitempty"`
		LocationID          *string `db:"location_id" json:"location_id,omitempty"`
		Name                *string `db:"name" json:"name,omitempty"`
		Title               *string `db:"title" json:"title,omitempty"`
		Department          *string `db:"department" json:"department,omitempty"`
		Email               *string `db:"email" json:"email,omitempty"`
	}
	Update struct {
		ID                  *string `db:"id" json:"id,omitempty"`
		OrganizationID      *string `db:"organization_id" json:"organization_id,omitempty"`
		ServiceID           *string `db:"service_id" json:"service_id,omitempty"`
		ServiceAtLocationID *string `db:"service_at_location_id" json:"service_at_location_id,omitempty"`
		LocationID          *string `db:"location_id" json:"location_id,omitempty"`
		Name                *string `db:"name" json:"name,omitempty"`
		Title               *string `db:"title" json:"title,omitempty"`
		Department          *string `db:"department" json:"department,omitempty"`
		Email               *string `db:"email" json:"email,omitempty"`
	}
}

type Service struct {
	Row struct {
		ID                     string            `json:"id" validate:"required"`
		OrganizationID         string            `json:"organization_id" validate:"required"`
		ProgramID              *string           `json:"program_id,omitempty"`
		Name                   string            `json:"name" validate:"required"`
		AlternateName          *string           `json:"alternate_name,omitempty"`
		Description            *string           `json:"description,omitempty"`
		URL                    *string           `json:"url,omitempty"`
		Email                  *string           `json:"email,omitempty"`
		Status                 ServiceStatusEnum `json:"status" validate:"required"`
		InterpretationServices *string           `json:"interpretation_services,omitempty"`
		ApplicationProcess     *string           `json:"application_process,omitempty"`
		FeesDescription        *string           `json:"fees_description,omitempty"`
		WaitTime               *string           `json:"wait_time,omitempty"` // Deprecated
		Fees                   *string           `json:"fees,omitempty"`      // Deprecated
		Accreditations         *string           `json:"accreditations,omitempty"`
		EligibilityDescription *string           `json:"eligibility_description,omitempty"`
		MinimumAge             *float64          `json:"minimum_age,omitempty"`
		MaximumAge             *float64          `json:"maximum_age,omitempty"`
		AssuredDate            *time.Time        `json:"assured_date,omitempty"`
		AssurerEmail           *string           `json:"assurer_email,omitempty"`
		Licenses               *string           `json:"licenses,omitempty"` // Deprecated
		Alert                  *string           `json:"alert,omitempty"`
		LastModified           *time.Time        `json:"last_modified,omitempty"`

		// Foreign key relationships
		Organization *Organization `json:"organization,omitempty" gorm:"foreignkey:OrganizationID"`
		Program      *Program      `json:"program,omitempty" gorm:"foreignkey:ProgramID"`
	}

	Insert struct {
		ID                     string            `json:"id" validate:"required"`
		OrganizationID         string            `json:"organization_id" validate:"required"`
		ProgramID              *string           `json:"program_id,omitempty"`
		Name                   string            `json:"name" validate:"required"`
		AlternateName          *string           `json:"alternate_name,omitempty"`
		Description            *string           `json:"description,omitempty"`
		URL                    *string           `json:"url,omitempty"`
		Email                  *string           `json:"email,omitempty"`
		Status                 ServiceStatusEnum `json:"status" validate:"required"`
		InterpretationServices *string           `json:"interpretation_services,omitempty"`
		ApplicationProcess     *string           `json:"application_process,omitempty"`
		FeesDescription        *string           `json:"fees_description,omitempty"`
		WaitTime               *string           `json:"wait_time,omitempty"`
		Fees                   *string           `json:"fees,omitempty"`
		Accreditations         *string           `json:"accreditations,omitempty"`
		EligibilityDescription *string           `json:"eligibility_description,omitempty"`
		MinimumAge             *float64          `json:"minimum_age,omitempty"`
		MaximumAge             *float64          `json:"maximum_age,omitempty"`
		AssuredDate            *time.Time        `json:"assured_date,omitempty"`
		AssurerEmail           *string           `json:"assurer_email,omitempty"`
		Licenses               *string           `json:"licenses,omitempty"`
		Alert                  *string           `json:"alert,omitempty"`
		LastModified           *time.Time        `json:"last_modified,omitempty"`
	}

	Update struct {
		ID                     *string            `json:"id,omitempty"`
		OrganizationID         *string            `json:"organization_id,omitempty"`
		ProgramID              *string            `json:"program_id,omitempty"`
		Name                   *string            `json:"name,omitempty"`
		AlternateName          *string            `json:"alternate_name,omitempty"`
		Description            *string            `json:"description,omitempty"`
		URL                    *string            `json:"url,omitempty"`
		Email                  *string            `json:"email,omitempty"`
		Status                 *ServiceStatusEnum `json:"status,omitempty"`
		InterpretationServices *string            `json:"interpretation_services,omitempty"`
		ApplicationProcess     *string            `json:"application_process,omitempty"`
		FeesDescription        *string            `json:"fees_description,omitempty"`
		WaitTime               *string            `json:"wait_time,omitempty"`
		Fees                   *string            `json:"fees,omitempty"`
		Accreditations         *string            `json:"accreditations,omitempty"`
		EligibilityDescription *string            `json:"eligibility_description,omitempty"`
		MinimumAge             *float64           `json:"minimum_age,omitempty"`
		MaximumAge             *float64           `json:"maximum_age,omitempty"`
		AssuredDate            *time.Time         `json:"assured_date,omitempty"`
		AssurerEmail           *string            `json:"assurer_email,omitempty"`
		Licenses               *string            `json:"licenses,omitempty"`
		Alert                  *string            `json:"alert,omitempty"`
		LastModified           *time.Time         `json:"last_modified,omitempty"`
	}
}

type Schedule struct {
	Row struct {
		ID                  string            `json:"id" validate:"required"`
		ServiceID           *string           `json:"service_id,omitempty"`
		LocationID          *string           `json:"location_id,omitempty"`
		ServiceAtLocationID *string           `json:"service_at_location_id,omitempty"`
		ValidFrom           *time.Time        `json:"valid_from,omitempty"`
		ValidTo             *time.Time        `json:"valid_to,omitempty"`
		DTStart             *time.Time        `json:"dtstart,omitempty"`
		Timezone            *float64          `json:"timezone,omitempty"`
		Until               *time.Time        `json:"until,omitempty"`
		Count               *int              `json:"count,omitempty"`
		Wkst                *ScheduleWkstEnum `json:"wkst,omitempty"`
		Freq                *ScheduleFreqEnum `json:"freq,omitempty"`
		Interval            *int              `json:"interval,omitempty"`
		Byday               *string           `json:"byday,omitempty"`
		Byweekno            *string           `json:"byweekno,omitempty"`
		Bymonthday          *string           `json:"bymonthday,omitempty"`
		Byyearday           *string           `json:"byyearday,omitempty"`
		Description         *string           `json:"description,omitempty"`
		OpensAt             *time.Time        `json:"opens_at,omitempty"`
		ClosesAt            *time.Time        `json:"closes_at,omitempty"`
		ScheduleLink        *string           `json:"schedule_link,omitempty"`
		AttendingType       *string           `json:"attending_type,omitempty"`
		Notes               *string           `json:"notes,omitempty"`
	}

	Insert struct {
		ID                  string            `json:"id" validate:"required"`
		ServiceID           *string           `json:"service_id,omitempty"`
		LocationID          *string           `json:"location_id,omitempty"`
		ServiceAtLocationID *string           `json:"service_at_location_id,omitempty"`
		ValidFrom           *time.Time        `json:"valid_from,omitempty"`
		ValidTo             *time.Time        `json:"valid_to,omitempty"`
		DTStart             *time.Time        `json:"dtstart,omitempty"`
		Timezone            *float64          `json:"timezone,omitempty"`
		Until               *time.Time        `json:"until,omitempty"`
		Count               *int              `json:"count,omitempty"`
		Wkst                *ScheduleWkstEnum `json:"wkst,omitempty"`
		Freq                *ScheduleFreqEnum `json:"freq,omitempty"`
		Interval            *int              `json:"interval,omitempty"`
		Byday               *string           `json:"byday,omitempty"`
		Byweekno            *string           `json:"byweekno,omitempty"`
		Bymonthday          *string           `json:"bymonthday,omitempty"`
		Byyearday           *string           `json:"byyearday,omitempty"`
		Description         *string           `json:"description,omitempty"`
		OpensAt             *time.Time        `json:"opens_at,omitempty"`
		ClosesAt            *time.Time        `json:"closes_at,omitempty"`
		ScheduleLink        *string           `json:"schedule_link,omitempty"`
		AttendingType       *string           `json:"attending_type,omitempty"`
		Notes               *string           `json:"notes,omitempty"`
	}

	Update struct {
		ID                  *string           `json:"id,omitempty"`
		ServiceID           *string           `json:"service_id,omitempty"`
		LocationID          *string           `json:"location_id,omitempty"`
		ServiceAtLocationID *string           `json:"service_at_location_id,omitempty"`
		ValidFrom           *time.Time        `json:"valid_from,omitempty"`
		ValidTo             *time.Time        `json:"valid_to,omitempty"`
		DTStart             *time.Time        `json:"dtstart,omitempty"`
		Timezone            *float64          `json:"timezone,omitempty"`
		Until               *time.Time        `json:"until,omitempty"`
		Count               *int              `json:"count,omitempty"`
		Wkst                *ScheduleWkstEnum `json:"wkst,omitempty"`
		Freq                *ScheduleFreqEnum `json:"freq,omitempty"`
		Interval            *int              `json:"interval,omitempty"`
		Byday               *string           `json:"byday,omitempty"`
		Byweekno            *string           `json:"byweekno,omitempty"`
		Bymonthday          *string           `json:"bymonthday,omitempty"`
		Byyearday           *string           `json:"byyearday,omitempty"`
		Description         *string           `json:"description,omitempty"`
		OpensAt             *time.Time        `json:"opens_at,omitempty"`
		ClosesAt            *time.Time        `json:"closes_at,omitempty"`
		ScheduleLink        *string           `json:"schedule_link,omitempty"`
		AttendingType       *string           `json:"attending_type,omitempty"`
		Notes               *string           `json:"notes,omitempty"`
	}
}

type ServiceCapacity struct {
	Row struct {
		ID          string    `db:"id" json:"id" validate:"required"`
		ServiceID   string    `db:"service_id" json:"service_id" validate:"required"`
		UnitID      string    `db:"unit_id" json:"unit_id" validate:"required"`
		Available   float64   `db:"available" json:"available" validate:"required"`
		Maximum     *float64  `db:"maximum" json:"maximum,omitempty"`
		Description *string   `db:"description" json:"description,omitempty"`
		Updated     time.Time `db:"updated" json:"updated" validate:"required"`
	}

	Insert struct {
		ID          string    `db:"id" json:"id" validate:"required"`
		ServiceID   string    `db:"service_id" json:"service_id" validate:"required"`
		UnitID      string    `db:"unit_id" json:"unit_id" validate:"required"`
		Available   float64   `db:"available" json:"available" validate:"required"`
		Maximum     *float64  `db:"maximum" json:"maximum,omitempty"`
		Description *string   `db:"description" json:"description,omitempty"`
		Updated     time.Time `db:"updated" json:"updated" validate:"required"`
	}

	Update struct {
		ID          *string    `db:"id" json:"id,omitempty"`
		ServiceID   *string    `db:"service_id" json:"service_id,omitempty"`
		UnitID      *string    `db:"unit_id" json:"unit_id,omitempty"`
		Available   *float64   `db:"available" json:"available,omitempty"`
		Maximum     *float64   `db:"maximum" json:"maximum,omitempty"`
		Description *string    `db:"description" json:"description,omitempty"`
		Updated     *time.Time `db:"updated" json:"updated,omitempty"`
	}
}

type Metadata struct {
	Row struct {
		ID               string    `db:"id" json:"id" validate:"required"`
		ResourceID       string    `db:"resource_id" json:"resource_id" validate:"required"`
		ResourceType     string    `db:"resource_type" json:"resource_type" validate:"required"`
		LastActionDate   time.Time `db:"last_action_date" json:"last_action_date" validate:"required"`
		LastActionType   string    `db:"last_action_type" json:"last_action_type" validate:"required"`
		FieldName        string    `db:"field_name" json:"field_name" validate:"required"`
		PreviousValue    string    `db:"previous_value" json:"previous_value" validate:"required"`
		ReplacementValue string    `db:"replacement_value" json:"replacement_value" validate:"required"`
		UpdatedBy        string    `db:"updated_by" json:"updated_by" validate:"required"`
	}
	Insert struct {
		ID               string    `db:"id" json:"id" validate:"required"`
		ResourceID       string    `db:"resource_id" json:"resource_id" validate:"required"`
		ResourceType     string    `db:"resource_type" json:"resource_type" validate:"required"`
		LastActionDate   time.Time `db:"last_action_date" json:"last_action_date" validate:"required"`
		LastActionType   string    `db:"last_action_type" json:"last_action_type" validate:"required"`
		FieldName        string    `db:"field_name" json:"field_name" validate:"required"`
		PreviousValue    string    `db:"previous_value" json:"previous_value" validate:"required"`
		ReplacementValue string    `db:"replacement_value" json:"replacement_value" validate:"required"`
		UpdatedBy        string    `db:"updated_by" json:"updated_by" validate:"required"`
	}
	Update struct {
		ID               *string    `db:"id" json:"id,omitempty"`
		ResourceID       *string    `db:"resource_id" json:"resource_id,omitempty"`
		ResourceType     *string    `db:"resource_type" json:"resource_type,omitempty"`
		LastActionDate   *time.Time `db:"last_action_date" json:"last_action_date,omitempty"`
		LastActionType   *string    `db:"last_action_type" json:"last_action_type,omitempty"`
		FieldName        *string    `db:"field_name" json:"field_name,omitempty"`
		PreviousValue    *string    `db:"previous_value" json:"previous_value,omitempty"`
		ReplacementValue *string    `db:"replacement_value" json:"replacement_value,omitempty"`
		UpdatedBy        *string    `db:"updated_by" json:"updated_by,omitempty"`
	}
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
