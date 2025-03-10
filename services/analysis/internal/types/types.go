package types

type ServiceCategory string

const (
	ServiceCategoryDisabilities ServiceCategory = "DISABILITIES"
	ServiceCategoryEmployment   ServiceCategory = "EMPLOYMENT"
	ServiceCategoryFood         ServiceCategory = "FOOD"
	ServiceCategoryPersonal     ServiceCategory = "PERSONAL"
	ServiceCategoryTransport    ServiceCategory = "TRANSPORT"
	ServiceCategoryMental       ServiceCategory = "MENTAL"
	ServiceCategoryEducation    ServiceCategory = "EDUCATION"
	ServiceCategoryFinancial    ServiceCategory = "FINANCIAL"
	ServiceCategoryHealthcare   ServiceCategory = "HEALTHCARE"
	ServiceCategoryShelter      ServiceCategory = "SHELTER"
	ServiceCategoryBrainTrauma  ServiceCategory = "BRAIN_TRAUMA"
)

// IsValid checks if the ServiceCategory is one of the defined constants
func (sc ServiceCategory) IsValid() bool {
	switch sc {
	case ServiceCategoryDisabilities,
		ServiceCategoryEmployment,
		ServiceCategoryFood,
		ServiceCategoryPersonal,
		ServiceCategoryTransport,
		ServiceCategoryMental,
		ServiceCategoryEducation,
		ServiceCategoryFinancial,
		ServiceCategoryHealthcare,
		ServiceCategoryShelter,
		ServiceCategoryBrainTrauma:
		return true
	}
	return false
}

type TranscriptsReqBody struct {
	OrganizationID string `json:"organization_id"`
	RoomURL        string `json:"room_url"`
	Transcript     string `json:"transcript"`
}

type ProcTranscriptParams struct {
	OrganizationID string `json:"organization_id"`
	RoomURL        string `json:"room_url"`
	Transcript     string `json:"transcript"`
	CallID         string `json:"call_fk"`
}
