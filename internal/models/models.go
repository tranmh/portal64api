package models

import (
	"time"
)

// Person represents a chess player from the mvdsb database
type Person struct {
	ID            uint      `json:"id" gorm:"primaryKey;column:id"`
	UUID          string    `json:"uuid" gorm:"column:uuid"`
	PKZ           string    `json:"pkz" gorm:"column:pkz"`
	Name          string    `json:"name" gorm:"column:name"`
	Vorname       string    `json:"vorname" gorm:"column:vorname"`
	Geburtsort    string    `json:"geburtsort" gorm:"column:geburtsort"`
	Geburtsdatum  *time.Time `json:"geburtsdatum" gorm:"column:geburtsdatum"`
	Titel         uint      `json:"titel" gorm:"column:titel"`
	Geschlecht    int       `json:"geschlecht" gorm:"column:geschlecht"`
	Nation        string    `json:"nation" gorm:"column:nation"`
	NationFide    string    `json:"nation_fide" gorm:"column:nationfide"`
	IDFide        uint      `json:"id_fide" gorm:"column:idfide"`
	Gleichstellung string   `json:"gleichstellung" gorm:"column:gleichstellung"`
	Datenschutz   string    `json:"datenschutz" gorm:"column:datenschutz"`
	Verstorben    string    `json:"verstorben" gorm:"column:verstorben"`
	Status        uint      `json:"status" gorm:"column:status"`
	Beatimestamp  time.Time `json:"beatimestamp" gorm:"column:beatimestamp"`
}

// TableName returns the table name for Person
func (Person) TableName() string {
	return "person"
}

// Organisation represents a chess club or organization from the mvdsb database
type Organisation struct {
	ID               uint      `json:"id" gorm:"primaryKey;column:id"`
	UUID             string    `json:"uuid" gorm:"column:uuid"`
	Name             string    `json:"name" gorm:"column:name"`
	Kurzname         string    `json:"kurzname" gorm:"column:kurzname"`
	Verband          string    `json:"verband" gorm:"column:verband"`
	Unterverband     string    `json:"unterverband" gorm:"column:unterverband"`
	Bezirk           string    `json:"bezirk" gorm:"column:bezirk"`
	Verein           string    `json:"verein" gorm:"column:verein"`
	VKZ              string    `json:"vkz" gorm:"column:vkz"`
	Grundungsdatum   *time.Time `json:"grundungsdatum" gorm:"column:grundungsdatum"`
	Organisationsart int       `json:"organisationsart" gorm:"column:organisationsart"`
	Status           uint      `json:"status" gorm:"column:status"`
	Beatimestamp     time.Time `json:"beatimestamp" gorm:"column:beatimestamp"`
}

// TableName returns the table name for Organisation
func (Organisation) TableName() string {
	return "organisation"
}

// Mitgliedschaft represents club membership from the mvdsb database
type Mitgliedschaft struct {
	ID           uint      `json:"id" gorm:"primaryKey;column:id"`
	UUID         string    `json:"uuid" gorm:"column:uuid"`
	Person       uint      `json:"person" gorm:"column:person"`
	Organisation uint      `json:"organisation" gorm:"column:organisation"`
	Spielernummer uint     `json:"spielernummer" gorm:"column:spielernummer"`
	Von          *time.Time `json:"von" gorm:"column:von"`
	Bis          *time.Time `json:"bis" gorm:"column:bis"`
	Spielberechtigung uint `json:"spielberechtigung" gorm:"column:spielberechtigung"`
	Stat1        uint      `json:"stat1" gorm:"column:stat1"`
	Stat2        uint      `json:"stat2" gorm:"column:stat2"`
	Status       uint      `json:"status" gorm:"column:status"`
	Beatimestamp time.Time `json:"beatimestamp" gorm:"column:beatimestamp"`
}

// TableName returns the table name for Mitgliedschaft
func (Mitgliedschaft) TableName() string {
	return "mitgliedschaft"
}

// Tournament represents a tournament from the portal64_bdw database
type Tournament struct {
	ID             uint      `json:"id" gorm:"primaryKey;column:id"`
	IDOrganisation uint      `json:"id_organisation" gorm:"column:idOrganisation"`
	TName          string    `json:"tname" gorm:"column:tname"`
	TCode          string    `json:"tcode" gorm:"column:tcode"`
	Acron          string    `json:"acron" gorm:"column:acron"`
	Type           string    `json:"type" gorm:"column:type"`
	Note           string    `json:"note" gorm:"column:note"`
	Rounds         int       `json:"rounds" gorm:"column:rounds"`
	Filename       string    `json:"filename" gorm:"column:filename"`
	FinishedOn     *time.Time `json:"finished_on" gorm:"column:finishedOn"`
	ComputedOn     *time.Time `json:"computed_on" gorm:"column:computedOn"`
	RecomputedOn   *time.Time `json:"recomputed_on" gorm:"column:recomputedOn"`
	Assessor1      uint      `json:"assessor1" gorm:"column:assessor1"`
	Assessor2      uint      `json:"assessor2" gorm:"column:assessor2"`
	LockedOn       *time.Time `json:"locked_on" gorm:"column:lockedOn"`
	LockedBy       uint      `json:"locked_by" gorm:"column:lockedBy"`
	IDMaster       uint      `json:"id_master" gorm:"column:idMaster"`
	Import         string    `json:"import" gorm:"column:import"`
}

// TableName returns the table name for Tournament
func (Tournament) TableName() string {
	return "tournament"
}

// Evaluation represents a DWZ rating evaluation from the portal64_bdw database
type Evaluation struct {
	ID           uint    `json:"id" gorm:"primaryKey;column:id"`
	IDMaster     uint    `json:"id_master" gorm:"column:idMaster"`
	IDPerson     uint    `json:"id_person" gorm:"column:idPerson"`
	ECoefficient int     `json:"e_coefficient" gorm:"column:eCoefficient"`
	We           float64 `json:"we" gorm:"column:we"`
	Achievement  int     `json:"achievement" gorm:"column:achievement"`
	Level        int     `json:"level" gorm:"column:level"`
	Games        int     `json:"games" gorm:"column:games"`
	UnratedGames int     `json:"unrated_games" gorm:"column:unratedGames"`
	Points       float64 `json:"points" gorm:"column:points"`
	DWZOld       int     `json:"dwz_old" gorm:"column:dwzOld"`
	DWZOldIndex  int     `json:"dwz_old_index" gorm:"column:dwzOldIndex"`
	DWZNew       int     `json:"dwz_new" gorm:"column:dwzNew"`
	DWZNewIndex  int     `json:"dwz_new_index" gorm:"column:dwzNewIndex"`
}

// TableName returns the table name for Evaluation
func (Evaluation) TableName() string {
	return "evaluation"
}

// Participant represents a tournament participant from the portal64_bdw database
type Participant struct {
	ID           uint `json:"id" gorm:"primaryKey;column:id"`
	IDTournament uint `json:"id_tournament" gorm:"column:idTournament"`
	IDPerson     uint `json:"id_person" gorm:"column:idPerson"`
	No           int  `json:"no" gorm:"column:no"`
	IDMembership uint `json:"id_membership" gorm:"column:idMembership"`
	UseRating    *int `json:"use_rating" gorm:"column:useRating"`
	UseRatingIndex *int `json:"use_rating_index" gorm:"column:useRatingIndex"`
}

// TableName returns the table name for Participant
func (Participant) TableName() string {
	return "participant"
}

// Turnier represents a tournament (legacy model - no longer in use)
type Turnier struct {
	TID                int        `json:"tid" gorm:"primaryKey;column:TID"`
	TName              string     `json:"tname" gorm:"column:TName"`
	LoginID            uint       `json:"login_id" gorm:"column:LoginID"`
	Staffelleiter      *int       `json:"staffelleiter" gorm:"column:staffelleiter"`
	Saison             *int       `json:"saison" gorm:"column:Saison"`
	SaisonAnzeige      string     `json:"saison_anzeige" gorm:"column:SaisonAnzeige"`
	MID                uint       `json:"mid" gorm:"column:MID"`
	Brettertausch      int        `json:"brettertausch" gorm:"column:Brettertausch"`
	AnzStammspieler    int        `json:"anz_stammspieler" gorm:"column:AnzStammspieler"`
	AnzErsatzspieler   int        `json:"anz_ersatzspieler" gorm:"column:AnzErsatzspieler"`
	AnzGastspieler     int        `json:"anz_gastspieler" gorm:"column:AnzGastspieler"`
	Teilnahmeschluss   *time.Time `json:"teilnahmeschluss" gorm:"column:Teilnahmeschluss"`
	Meldeschluss       *time.Time `json:"meldeschluss" gorm:"column:Meldeschluss"`
	Wettbewerb         uint       `json:"wettbewerb" gorm:"column:Wettbewerb"`
	IsOffen            bool       `json:"is_offen" gorm:"column:isOffen"`
	IsFreigegeben      bool       `json:"is_freigegeben" gorm:"column:isFreigegeben"`
	IsSpielebene       bool       `json:"is_spielebene" gorm:"column:isSpielebene"`
	Organisation       *uint      `json:"organisation" gorm:"column:Organisation"`
}

// TableName returns the table name for Turnier
func (Turnier) TableName() string {
	return "Turnier"
}

// API Response Models

// PlayerResponse represents a player in API responses
type PlayerResponse struct {
	ID         string    `json:"id"`         // Format: C0101-123 (3-digit membership number)
	Name       string    `json:"name"`
	Firstname  string    `json:"firstname"`
	Club       string    `json:"club"`
	ClubID     string    `json:"club_id"`    // Format: C0101
	BirthYear  *int      `json:"birth_year"` // GDPR compliant: only birth year, not full date
	Gender     string    `json:"gender"`
	Nation     string    `json:"nation"`
	FideID     uint      `json:"fide_id"`
	CurrentDWZ int       `json:"current_dwz"`
	DWZIndex   int       `json:"dwz_index"`
	Status     string    `json:"status"`
}

// RatingHistoryResponse represents a rating history entry in API responses
type RatingHistoryResponse struct {
	ID           uint    `json:"id"`
	TournamentID string  `json:"tournament_id"`   // Format: C531-634-S25 (tournament code)
	ECoefficient int     `json:"e_coefficient"`
	We           float64 `json:"we"`
	Achievement  int     `json:"achievement"`
	Level        int     `json:"level"`
	Games        int     `json:"games"`
	UnratedGames int     `json:"unrated_games"`
	Points       float64 `json:"points"`
	DWZOld       int     `json:"dwz_old"`
	DWZOldIndex  int     `json:"dwz_old_index"`
	DWZNew       int     `json:"dwz_new"`
	DWZNewIndex  int     `json:"dwz_new_index"`
}

// ClubResponse represents a club in API responses
type ClubResponse struct {
	ID               string    `json:"id"`               // Format: C0101
	Name             string    `json:"name"`
	ShortName        string    `json:"short_name"`
	Region           string    `json:"region"`
	District         string    `json:"district"`
	FoundingDate     *time.Time `json:"founding_date"`
	MemberCount      int       `json:"member_count"`
	AverageDWZ       float64   `json:"average_dwz"`
	Status           string    `json:"status"`
}

// ClubProfileResponse represents a comprehensive club profile in API responses
type ClubProfileResponse struct {
	// Basic club information
	Club             ClubResponse          `json:"club"`
	
	// Players and statistics
	Players          []PlayerResponse      `json:"players"`
	PlayerCount      int                   `json:"player_count"`
	ActivePlayerCount int                  `json:"active_player_count"`
	
	// Rating statistics
	RatingStats      ClubRatingStats       `json:"rating_stats"`
	
	// Recent tournaments (if available)
	RecentTournaments []TournamentResponse  `json:"recent_tournaments,omitempty"`
	TournamentCount   int                  `json:"tournament_count"`
	
	// Additional club information
	Teams            []ClubTeam           `json:"teams,omitempty"`
	Contact          ClubContact          `json:"contact,omitempty"`
}

// ClubRatingStats represents rating statistics for a club
type ClubRatingStats struct {
	AverageDWZ       float64 `json:"average_dwz"`
	MedianDWZ        float64 `json:"median_dwz"`
	HighestDWZ       int     `json:"highest_dwz"`
	LowestDWZ        int     `json:"lowest_dwz"`
	PlayersWithDWZ   int     `json:"players_with_dwz"`
	RatingDistribution map[string]int `json:"rating_distribution"`
}

// ClubTeam represents a team within a club
type ClubTeam struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	League       string `json:"league"`
	Division     string `json:"division"`
	PlayerCount  int    `json:"player_count"`
}

// ClubContact represents contact information for a club
type ClubContact struct {
	Website      string `json:"website,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Address      string `json:"address,omitempty"`
	MeetingPlace string `json:"meeting_place,omitempty"`
	MeetingTime  string `json:"meeting_time,omitempty"`
}

// TournamentResponse represents a tournament in API responses
type TournamentResponse struct {
	ID              string     `json:"id"`              // Format: C529-K00-HT1
	Name            string     `json:"name"`
	Code            string     `json:"code"`
	Type            string     `json:"type"`
	Organization    string     `json:"organization"`
	Rounds          int        `json:"rounds"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	Status          string     `json:"status"`
	ParticipantCount int       `json:"participant_count"`
}

// Enhanced TournamentResponse with comprehensive tournament data
type EnhancedTournamentResponse struct {
	// Basic tournament info
	ID              string     `json:"id"`              // Format: C529-K00-HT1
	Name            string     `json:"name"`
	Code            string     `json:"code"`
	Type            string     `json:"type"`
	Organization    *OrganizationInfo `json:"organization"`
	Rounds          int        `json:"rounds"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	FinishedOn      *time.Time `json:"finished_on"`
	ComputedOn      *time.Time `json:"computed_on"`
	RecomputedOn    *time.Time `json:"recomputed_on"`
	Status          string     `json:"status"`
	Note            string     `json:"note,omitempty"`
	
	// Assessors/Officials
	Assessors       []PersonInfo `json:"assessors,omitempty"`
	
	// Participants
	Participants    []ParticipantInfo `json:"participants"`
	ParticipantCount int             `json:"participant_count"`
	
	// Games and Results  
	Games           []RoundInfo `json:"games,omitempty"`
	
	// Evaluation data (if computed)
	Evaluations     []EvaluationInfo `json:"evaluations,omitempty"`
}

// OrganizationInfo represents tournament organizing club/federation
type OrganizationInfo struct {
	ID          string `json:"id"`          // Format: C0101
	Name        string `json:"name"`
	ShortName   string `json:"short_name"`
	VKZ         string `json:"vkz"`         // Vereinskennzeichen
	Region      string `json:"region"`
	District    string `json:"district"`
}

// PersonInfo represents a person (assessor, player, etc.)
type PersonInfo struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Firstname string `json:"firstname"`
	FullName  string `json:"full_name"`
}

// ParticipantInfo represents a tournament participant with full details
type ParticipantInfo struct {
	ID            uint         `json:"id"`
	No            int          `json:"no"`           // Starting number
	PersonID      uint         `json:"person_id"`
	Name          string       `json:"name"`
	Firstname     string       `json:"firstname"`
	FullName      string       `json:"full_name"`
	BirthYear     *int         `json:"birth_year"`   // GDPR compliant: only birth year, not full date
	Gender        string       `json:"gender"`
	Nation        string       `json:"nation"`
	FideID        uint         `json:"fide_id"`
	Club          *ClubInfo    `json:"club"`
	Rating        *RatingInfo  `json:"rating"`
	State         int          `json:"state"`       // 0=blocked, 1=unknown, 2=ok
}

// ClubInfo represents club membership information
type ClubInfo struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	VKZ              string `json:"vkz"`
	MembershipNumber int    `json:"membership_number"`
}

// RatingInfo represents rating information for a participant
type RatingInfo struct {
	DWZOld       *int `json:"dwz_old"`
	DWZOldIndex  *int `json:"dwz_old_index"`
	DWZNew       *int `json:"dwz_new"`
	DWZNewIndex  *int `json:"dwz_new_index"`
	ELO          *int `json:"elo"`
	UseRating    *int `json:"use_rating"`
	UseRatingIndex *int `json:"use_rating_index"`
}

// RoundInfo represents a tournament round with games
type RoundInfo struct {
	Round       int        `json:"round"`
	Appointment string     `json:"appointment"`
	Games       []GameInfo `json:"games"`
}

// GameInfo represents a single game
type GameInfo struct {
	ID          uint       `json:"id"`
	Board       int        `json:"board"`
	White       PlayerRef  `json:"white"`
	Black       PlayerRef  `json:"black"`
	Result      string     `json:"result"`
	WhitePoints float64    `json:"white_points"`
	BlackPoints float64    `json:"black_points"`
}

// PlayerRef represents a player reference in a game
type PlayerRef struct {
	ID       uint   `json:"id"`
	No       int    `json:"no"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

// EvaluationInfo represents DWZ evaluation results
type EvaluationInfo struct {
	ID           uint    `json:"id"`
	PersonID     uint    `json:"person_id"`
	PlayerName   string  `json:"player_name"`
	ECoefficient int     `json:"e_coefficient"`
	We           float64 `json:"we"`
	Achievement  int     `json:"achievement"`
	Level        int     `json:"level"`
	Games        int     `json:"games"`
	UnratedGames int     `json:"unrated_games"`
	Points       float64 `json:"points"`
	DWZOld       int     `json:"dwz_old"`
	DWZOldIndex  int     `json:"dwz_old_index"`
	DWZNew       int     `json:"dwz_new"`
	DWZNewIndex  int     `json:"dwz_new_index"`
}

// SearchRequest represents search parameters
type SearchRequest struct {
	Query        string `json:"query" form:"query"`
	Limit        int    `json:"limit" form:"limit"`
	Offset       int    `json:"offset" form:"offset"`
	SortBy       string `json:"sort_by" form:"sort_by"`
	SortOrder    string `json:"sort_order" form:"sort_order"`
	FilterBy     string `json:"filter_by" form:"filter_by"`
	FilterValue  string `json:"filter_value" form:"filter_value"`
}

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta represents response metadata
type Meta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

// Database models for additional tournament data

// Game represents a game from the database
type Game struct {
	ID                      uint      `json:"id" gorm:"primaryKey;column:id"`
	IDTournament           uint      `json:"id_tournament" gorm:"column:idTournament"`
	IDAppointment          uint      `json:"id_appointment" gorm:"column:idAppointment"`
	IDResultsDisplayRating uint      `json:"id_results_display_rating" gorm:"column:idResultsDisplayRating"`
	Board                  int       `json:"board" gorm:"column:board"`
}

// TableName returns the table name for Game
func (Game) TableName() string {
	return "game"
}

// Result represents a game result from the database
type Result struct {
	ID           uint    `json:"id" gorm:"primaryKey;column:id"`
	IDPerson     uint    `json:"id_person" gorm:"column:idPerson"`
	IDTournament uint    `json:"id_tournament" gorm:"column:idTournament"`
	IDGame       uint    `json:"id_game" gorm:"column:idGame"`
	Color        string  `json:"color" gorm:"column:color"`
	Points       float64 `json:"points" gorm:"column:points"`
	Rating       int     `json:"rating" gorm:"column:rating"`
}

// TableName returns the table name for Result
func (Result) TableName() string {
	return "results"
}

// Appointment represents a tournament appointment/round from the database
type Appointment struct {
	ID           uint   `json:"id" gorm:"primaryKey;column:id"`
	IDTournament uint   `json:"id_tournament" gorm:"column:idTournament"`
	Round        int    `json:"round" gorm:"column:round"`
	Appointment  string `json:"appointment" gorm:"column:appointment"`
}

// TableName returns the table name for Appointment
func (Appointment) TableName() string {
	return "appointment"
}

// ResultsDisplay represents result display information
type ResultsDisplay struct {
	ID          uint    `json:"id" gorm:"primaryKey;column:id"`
	Display     string  `json:"display" gorm:"column:display"`
	PointsWhite float64 `json:"points_white" gorm:"column:pointsWhite"`
	PointsBlack float64 `json:"points_black" gorm:"column:pointsBlack"`
	RatingWhite int     `json:"rating_white" gorm:"column:ratingWhite"`
	RatingBlack int     `json:"rating_black" gorm:"column:ratingBlack"`
}

// TableName returns the table name for ResultsDisplay
func (ResultsDisplay) TableName() string {
	return "resultsDisplay"
}

// Address models for contact information

// Adressen represents the main address table linking organizations/persons to addresses
type Adressen struct {
	ID           uint `json:"id" gorm:"primaryKey;column:id"`
	UUID         string `json:"uuid" gorm:"column:uuid"`
	Organisation uint `json:"organisation" gorm:"column:organisation"`
	Person       uint `json:"person" gorm:"column:person"`
	Typ          uint `json:"typ" gorm:"column:typ"`
	IstPerson    uint `json:"ist_person" gorm:"column:istperson"`
	Funktion     uint `json:"funktion" gorm:"column:funktion"`
	Status       uint `json:"status" gorm:"column:status"`
	BeaPerson    uint `json:"bea_person" gorm:"column:beaperson"`
	Beatimestamp time.Time `json:"beatimestamp" gorm:"column:beatimestamp"`
}

// TableName returns the table name for Adressen
func (Adressen) TableName() string {
	return "adressen"
}

// Adr represents address detail records (phone, email, website, etc.)
type Adr struct {
	ID           uint   `json:"id" gorm:"primaryKey;column:id"`
	IDAdressen   uint   `json:"id_adressen" gorm:"column:id_adressen"`
	IDArt        uint   `json:"id_art" gorm:"column:id_art"`
	Wert         string `json:"wert" gorm:"column:wert"`
	Status       uint   `json:"status" gorm:"column:status"`
	Beatimestamp time.Time `json:"beatimestamp" gorm:"column:beatimestamp"`
	BeaPerson    uint   `json:"bea_person" gorm:"column:beaperson"`
}

// TableName returns the table name for Adr
func (Adr) TableName() string {
	return "adr"
}

// AdrArt represents address types (email, phone, website, etc.)
type AdrArt struct {
	ID           uint   `json:"id" gorm:"primaryKey;column:id"`
	Bezeichnung  string `json:"bezeichnung" gorm:"column:bezeichnung"`
	Test         string `json:"test" gorm:"column:test"`
	Status       uint   `json:"status" gorm:"column:status"`
	Beatimestamp time.Time `json:"beatimestamp" gorm:"column:beatimestamp"`
	BeaPerson    uint   `json:"bea_person" gorm:"column:beaperson"`
}

// TableName returns the table name for AdrArt
func (AdrArt) TableName() string {
	return "adr_art"
}

// Address type constants
const (
	AdrArtName         = 1   // Name
	AdrArtStrasse      = 2   // Straße (Street)
	AdrArtPLZ          = 3   // Postleitzahl (Postal Code)
	AdrArtOrt          = 4   // Ort (City)
	AdrArtLand         = 5   // Land (Country)
	AdrArtTelefon1     = 6   // Telefon 1 (Phone 1)
	AdrArtTelefon2     = 7   // Telefon 2 (Phone 2)
	AdrArtTelefon3     = 8   // Telefon 3 (Phone 3)
	AdrArtFax          = 9   // Fax
	AdrArtEmail1       = 10  // E-Mail 1
	AdrArtEmail2       = 11  // E-Mail 2
	AdrArtZusatz       = 12  // Zusatz (Additional)
	AdrArtBemerkung    = 13  // Bemerkung (Remarks)
	AdrArtHomepage     = 15  // Homepage (Website)
	AdrArtUebungsabend = 16  // Übungsabend (Practice Evening)
	AdrArtBreite       = 17  // geogr. Breite (Latitude)
	AdrArtLaenge       = 18  // geogr. Länge (Longitude)
	AdrArtVereinsitz   = 19  // Vereinssitz (Club Seat)
	AdrArtRegister     = 20  // Vereinsregister (Club Register)
	AdrArtUStNr        = 21  // Umsatzsteuernr (VAT Number)
	AdrArtGeburt       = 22  // Geburtsdatum (Birth Date)
	AdrArtBewirtschaftet = 23 // bewirtschaftet (Managed)
	AdrArtBehindertengerecht = 24 // behindertengerecht (Wheelchair Accessible)
)

// Address response models

// RegionAddressResponse represents an address entry for a region
type RegionAddressResponse struct {
	ID                    uint            `json:"id"`
	UUID                  string          `json:"uuid"`
	Name                  string          `json:"name"`                     // Full name (person or organisation)
	PersonName            string          `json:"person_name,omitempty"`
	PersonFirstname       string          `json:"person_firstname,omitempty"`
	OrganisationName      string          `json:"organisation_name,omitempty"`
	OrganisationShortname string          `json:"organisation_shortname,omitempty"`
	Region                string          `json:"region"`
	FunctionName          string          `json:"function_name"`
	FunctionID            uint            `json:"function_id"`
	OrganisationID        uint            `json:"organisation_id,omitempty"`
	PersonID              uint            `json:"person_id,omitempty"`
	ContactDetails        []ContactDetail `json:"contact_details"`
}

// ContactDetail represents a single contact detail (phone, email, etc.)
type ContactDetail struct {
	Type  string `json:"type"`   // e.g., "Email 1", "Telefon 1", "Homepage"
	Value string `json:"value"`  // The actual contact value
}

// RegionInfo represents information about a region
type RegionInfo struct {
	Code         string `json:"code"`          // e.g., "C", "B", "W"
	Name         string `json:"name"`          // e.g., "Gesamtverband", "Baden"
	AddressCount int    `json:"address_count"` // Number of addresses in this region
}

// AddressTypeInfo represents information about an address/function type
type AddressTypeInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`  // e.g., "Präsident", "Schriftführer"
	Count int    `json:"count"` // Number of entries of this type
}

// Funktion represents a function/role table
type Funktion struct {
	ID           uint      `json:"id" gorm:"primaryKey;column:id"`
	Bezeichnung  string    `json:"bezeichnung" gorm:"column:bezeichnung"`
	Status       uint      `json:"status" gorm:"column:status"`
	Beatimestamp time.Time `json:"beatimestamp" gorm:"column:beatimestamp"`
	BeaPerson    uint      `json:"bea_person" gorm:"column:beaperson"`
}

// TableName returns the table name for Funktion
func (Funktion) TableName() string {
	return "funktion"
}
