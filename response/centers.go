package response

type Center struct {
	CenterID     int64     `json:"center_id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	StateName    string    `json:"state_name"`
	DistrictName string    `json:"district_name"`
	BlockName    string    `json:"block_name"`
	Pincode      int64     `json:"pincode"`
	Lat          int64     `json:"lat"`
	Long         int64     `json:"long"`
	From         string    `json:"from"`
	To           string    `json:"to"`
	FeeType      FeeType   `json:"fee_type"`
	Sessions     []Session `json:"sessions"`
}

type Session struct {
	SessionID         string   `json:"session_id"`
	Date              string   `json:"date"`
	AvailableCapacity int64    `json:"available_capacity"`
	MinAgeLimit       int64    `json:"min_age_limit"`
	Vaccine           Vaccine  `json:"vaccine"`
	Slots             []string `json:"slots"`
}

type Vaccine string

const (
	Covaxin    Vaccine = "COVAXIN"
	Covishield Vaccine = "COVISHIELD"
)

type FeeType string

const (
	Free FeeType = "Free"
	Paid FeeType = "Paid"
)
