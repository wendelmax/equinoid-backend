package models

type SearchCriteria struct {
	Nome     string `json:"nome"`
	Raca     string `json:"raca"`
	Sexo     string `json:"sexo"`
	IdadeMin *int   `json:"idade_min"`
	IdadeMax *int   `json:"idade_max"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}
