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

// Turnier represents a tournament from the portal64_svw database
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
	ID         string    `json:"id"`         // Format: C0101-1014
	Name       string    `json:"name"`
	Firstname  string    `json:"firstname"`
	Club       string    `json:"club"`
	ClubID     string    `json:"club_id"`    // Format: C0101
	Birth      *time.Time `json:"birth"`
	Gender     string    `json:"gender"`
	Nation     string    `json:"nation"`
	FideID     uint      `json:"fide_id"`
	CurrentDWZ int       `json:"current_dwz"`
	DWZIndex   int       `json:"dwz_index"`
	Status     string    `json:"status"`
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
