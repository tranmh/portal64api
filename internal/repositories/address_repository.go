package repositories

import (
	"fmt"
	"strings"

	"portal64api/internal/database"
	"portal64api/internal/models"
	"portal64api/pkg/errors"

	"gorm.io/gorm"
)

// AddressRepository handles address-related database operations
type AddressRepository struct {
	mvdsb *gorm.DB
}

// NewAddressRepository creates a new address repository
func NewAddressRepository(dbs *database.Databases) *AddressRepository {
	return &AddressRepository{
		mvdsb: dbs.MVDSB,
	}
}

// GetRegionAddresses retrieves addresses for officials/functionaries in a specific region
func (r *AddressRepository) GetRegionAddresses(region string, addressType string) ([]models.RegionAddressResponse, error) {
	results := make([]models.RegionAddressResponse, 0)

	// Build the query to get addresses for a specific region and type
	query := `
		SELECT DISTINCT
			a.id as address_id,
			a.uuid as address_uuid,
			COALESCE(p.name, '') as person_name,
			COALESCE(p.vorname, '') as person_firstname,
			COALESCE(o.name, '') as organisation_name,
			COALESCE(o.kurzname, '') as organisation_shortname,
			o.verband as region,
			COALESCE(fa.bezeichnung, f.funktionsalias, '') as function_name,
			f.funktion as function_id,
			a.organisation as organisation_id,
			a.person as person_id
		FROM adressen a
		LEFT JOIN person p ON a.person = p.id AND a.istperson = 1
		LEFT JOIN organisation o ON a.organisation = o.id
		LEFT JOIN funktion f ON a.funktion = f.id
		LEFT JOIN funktionsart fa ON f.funktion = fa.id
		WHERE a.status = 1
	`

	args := []interface{}{}

	// Add region filter if specified
	if region != "" {
		query += ` AND o.verband = ?`
		args = append(args, region)
	}

	// Add address type filter if specified (e.g., "praesidium")
	if addressType != "" {
		// Map address types to function IDs or names
		switch strings.ToLower(addressType) {
		case "praesidium", "präsidium":
			query += ` AND (fa.bezeichnung LIKE '%präsident%' OR fa.bezeichnung LIKE '%praesidium%' OR fa.bezeichnung LIKE '%präsidium%' OR f.funktionsalias LIKE '%präsident%')`
		case "vorstand":
			query += ` AND (fa.bezeichnung LIKE '%vorstand%' OR f.funktionsalias LIKE '%vorstand%')`
		case "schriftfuehrer", "schriftführer":
			query += ` AND (fa.bezeichnung LIKE '%schriftführer%' OR fa.bezeichnung LIKE '%schriftfuehrer%' OR f.funktionsalias LIKE '%schriftführer%')`
		case "kassenwart":
			query += ` AND (fa.bezeichnung LIKE '%kasse%' OR fa.bezeichnung LIKE '%kassenwart%' OR f.funktionsalias LIKE '%kasse%')`
		default:
			query += ` AND (fa.bezeichnung LIKE ? OR f.funktionsalias LIKE ?)`
			args = append(args, "%"+addressType+"%", "%"+addressType+"%")
		}
	}

	query += ` ORDER BY o.name, fa.bezeichnung, f.funktionsalias, p.name`

	// Execute the query
	rows, err := r.mvdsb.Raw(query, args...).Rows()
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to query region addresses: " + err.Error())
	}
	defer rows.Close()

	addressIDs := []uint{}
	addressMap := make(map[uint]*models.RegionAddressResponse)

	// Scan the results
	for rows.Next() {
		var result struct {
			AddressID             uint    `json:"address_id"`
			AddressUUID           *string `json:"address_uuid"`  // Made nullable to handle NULL values
			PersonName            string  `json:"person_name"`
			PersonFirstname       string  `json:"person_firstname"`
			OrganisationName      string  `json:"organisation_name"`
			OrganisationShortname string  `json:"organisation_shortname"`
			Region                *string `json:"region"`       // Made nullable to handle NULL values
			FunctionName          string  `json:"function_name"`
			FunctionID            *uint   `json:"function_id"`  // Made nullable to handle NULL values
			OrganisationID        *uint   `json:"organisation_id"` // Made nullable to handle NULL values
			PersonID              *uint   `json:"person_id"`    // Made nullable to handle NULL values
		}

		if err := rows.Scan(
			&result.AddressID,
			&result.AddressUUID,
			&result.PersonName,
			&result.PersonFirstname,
			&result.OrganisationName,
			&result.OrganisationShortname,
			&result.Region,
			&result.FunctionName,
			&result.FunctionID,
			&result.OrganisationID,
			&result.PersonID,
		); err != nil {
			return nil, errors.NewInternalServerError("Failed to scan address row: " + err.Error())
		}

		// Create full name
		fullName := strings.TrimSpace(result.PersonFirstname + " " + result.PersonName)
		if fullName == "" {
			fullName = result.OrganisationName
		}

		// Create address response
		functionID := uint(0)
		if result.FunctionID != nil {
			functionID = *result.FunctionID
		}
		
		organisationID := uint(0)
		if result.OrganisationID != nil {
			organisationID = *result.OrganisationID
		}
		
		personID := uint(0)
		if result.PersonID != nil {
			personID = *result.PersonID
		}

		region := ""
		if result.Region != nil {
			region = *result.Region
		}

		uuid := ""
		if result.AddressUUID != nil {
			uuid = *result.AddressUUID
		}

		addressResp := &models.RegionAddressResponse{
			ID:               result.AddressID,
			UUID:             uuid,
			Name:             fullName,
			PersonName:       result.PersonName,
			PersonFirstname:  result.PersonFirstname,
			OrganisationName: result.OrganisationName,
			OrganisationShortname: result.OrganisationShortname,
			Region:           region,
			FunctionName:     result.FunctionName,
			FunctionID:       functionID,
			OrganisationID:   organisationID,
			PersonID:         personID,
			ContactDetails:   []models.ContactDetail{},
		}

		addressMap[result.AddressID] = addressResp
		addressIDs = append(addressIDs, result.AddressID)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewInternalServerError("Error iterating over address rows: " + err.Error())
	}

	// If we have addresses, get their contact details
	if len(addressIDs) > 0 {
		contactDetails, err := r.getContactDetailsForAddresses(addressIDs)
		if err != nil {
			return nil, err
		}

		// Assign contact details to addresses
		for addressID, details := range contactDetails {
			if addr, exists := addressMap[addressID]; exists {
				addr.ContactDetails = details
			}
		}
	}

	// Convert map to slice
	for _, addr := range addressMap {
		results = append(results, *addr)
	}

	return results, nil
}

// getContactDetailsForAddresses retrieves contact details for a list of address IDs
func (r *AddressRepository) getContactDetailsForAddresses(addressIDs []uint) (map[uint][]models.ContactDetail, error) {
	contactMap := make(map[uint][]models.ContactDetail)

	if len(addressIDs) == 0 {
		return contactMap, nil
	}

	// Build placeholders for the IN clause
	placeholders := make([]string, len(addressIDs))
	args := make([]interface{}, len(addressIDs))
	for i, id := range addressIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT 
			adr.id_adressen as address_id,
			adr.id_art as contact_type_id,
			art.bezeichnung as contact_type_name,
			adr.wert as contact_value
		FROM adr
		LEFT JOIN adr_art art ON adr.id_art = art.id
		WHERE adr.id_adressen IN (%s) 
		AND adr.status = 1 
		AND art.status = 1
		AND adr.wert IS NOT NULL 
		AND adr.wert != ''
		ORDER BY adr.id_adressen, art.bezeichnung
	`, strings.Join(placeholders, ","))

	rows, err := r.mvdsb.Raw(query, args...).Rows()
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to query contact details: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var result struct {
			AddressID       uint   `json:"address_id"`
			ContactTypeID   uint   `json:"contact_type_id"`
			ContactTypeName string `json:"contact_type_name"`
			ContactValue    string `json:"contact_value"`
		}

		if err := rows.Scan(
			&result.AddressID,
			&result.ContactTypeID,
			&result.ContactTypeName,
			&result.ContactValue,
		); err != nil {
			return nil, errors.NewInternalServerError("Failed to scan contact detail row: " + err.Error())
		}

		// Create contact detail
		detail := models.ContactDetail{
			Type:  result.ContactTypeName,
			Value: result.ContactValue,
		}

		// Add to map
		contactMap[result.AddressID] = append(contactMap[result.AddressID], detail)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewInternalServerError("Error iterating over contact detail rows: " + err.Error())
	}

	return contactMap, nil
}

// GetAvailableRegions retrieves all available regions that have addresses
func (r *AddressRepository) GetAvailableRegions() ([]models.RegionInfo, error) {
	regions := make([]models.RegionInfo, 0)

	query := `
		SELECT DISTINCT 
			o.verband as region_code,
			COUNT(DISTINCT a.id) as address_count
		FROM adressen a
		INNER JOIN organisation o ON a.organisation = o.id
		WHERE a.status = 1 
		AND o.verband IS NOT NULL 
		AND o.verband != ''
		GROUP BY o.verband
		ORDER BY o.verband
	`

	rows, err := r.mvdsb.Raw(query).Rows()
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to query regions: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var region models.RegionInfo
		if err := rows.Scan(&region.Code, &region.AddressCount); err != nil {
			return nil, errors.NewInternalServerError("Failed to scan region row: " + err.Error())
		}

		// Set region name based on code
		region.Name = getRegionName(region.Code)
		regions = append(regions, region)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewInternalServerError("Error iterating over region rows: " + err.Error())
	}

	return regions, nil
}

// GetAddressTypes retrieves available address types for a region
func (r *AddressRepository) GetAddressTypes(region string) ([]models.AddressTypeInfo, error) {
	types := make([]models.AddressTypeInfo, 0)

	query := `
		SELECT DISTINCT
			COALESCE(fa.id, f.id, 0) as function_id,
			COALESCE(fa.bezeichnung, f.funktionsalias, 'Keine Funktion zugewiesen') as function_name,
			COUNT(DISTINCT a.id) as count
		FROM adressen a
		INNER JOIN organisation o ON a.organisation = o.id
		LEFT JOIN funktion f ON a.funktion = f.id
		LEFT JOIN funktionsart fa ON f.funktion = fa.id
		WHERE a.status = 1 
		AND o.verband IS NOT NULL 
		AND o.verband != ''
	`

	args := []interface{}{}

	if region != "" {
		query += ` AND o.verband = ?`
		args = append(args, region)
	}

	query += `
		GROUP BY COALESCE(fa.id, f.id, 0), COALESCE(fa.bezeichnung, f.funktionsalias, 'Keine Funktion zugewiesen')
		ORDER BY COALESCE(fa.bezeichnung, f.funktionsalias, 'Keine Funktion zugewiesen')
	`

	rows, err := r.mvdsb.Raw(query, args...).Rows()
	if err != nil {
		return nil, errors.NewInternalServerError("Failed to query address types: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var addressType models.AddressTypeInfo
		if err := rows.Scan(&addressType.ID, &addressType.Name, &addressType.Count); err != nil {
			return nil, errors.NewInternalServerError("Failed to scan address type row: " + err.Error())
		}

		types = append(types, addressType)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewInternalServerError("Error iterating over address type rows: " + err.Error())
	}

	return types, nil
}

// Helper function to get region name from code
func getRegionName(code string) string {
	regionNames := map[string]string{
		"B": "Baden",
		"W": "Württemberg", 
		"C": "Gesamtverband",
		"H": "Hessen",
		"P": "Pfalz",
		"R": "Rheinhessen",
		"N": "Nahe",
	}
	
	if name, exists := regionNames[code]; exists {
		return name
	}
	return code
}
