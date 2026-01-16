package models

type ArvoreGenealogica struct {
	Equinoid   string      `json:"equinoid"`
	Nome       string      `json:"nome"`
	Geracoes   int         `json:"geracoes"`
	Ancestrais *Ancestrais `json:"ancestrais,omitempty"`
}

type Ancestrais struct {
	Pai *AncestralNode `json:"pai,omitempty"`
	Mae *AncestralNode `json:"mae,omitempty"`
}

type AncestralNode struct {
	Equinoid   string      `json:"equinoid"`
	Nome       string      `json:"nome"`
	Sexo       string      `json:"sexo"`
	Ancestrais *Ancestrais `json:"ancestrais,omitempty"`
}

type ResultadoValidacaoParentesco struct {
	Equino1                    string   `json:"equino1"`
	Equino2                    string   `json:"equino2"`
	SaoParentes                bool     `json:"sao_parentes"`
	GrauParentesco             string   `json:"grau_parentesco,omitempty"`
	AncestaisComuns            []string `json:"ancestrais_comuns,omitempty"`
	CoeficienteConsanguinidade float64  `json:"coeficiente_consanguinidade"`
}
