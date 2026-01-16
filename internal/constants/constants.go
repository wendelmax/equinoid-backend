package constants

const (
	MaxGeracoesArvoreGenealogica   = 10
	MaxGeracoesAncestrais          = 5
	CoeficienteConsanguinidadeBase = 0.0625

	MesesGestacaoEquino = 11

	MinTokenLength      = 10
	MinSecretLength     = 32
	EncryptionKeyLength = 32
	NonceLength         = 12

	PontosBaseCompeticao      = 100
	PontosBaseReproducao      = 80
	PontosBaseSaude           = 50
	PontosBaseTreinamento     = 40
	PontosBaseComercial       = 60
	PontosBaseMidia           = 30
	PontosBaseEducacao        = 40
	PontosBaseViagens         = 20
	PontosBaseReconhecimentos = 70
	PontosBaseParcerias       = 50
	PontosBaseAnalise         = 60

	MultiplicadorNivelBaixo       = 0.5
	MultiplicadorNivelMedio       = 1.0
	MultiplicadorNivelAlto        = 2.0
	MultiplicadorNivelCritico     = 3.0
	MultiplicadorNivelExcepcional = 5.0
)

var (
	PontosBasePorCategoria = map[string]int{
		"competicao":      PontosBaseCompeticao,
		"reproducao":      PontosBaseReproducao,
		"saude":           PontosBaseSaude,
		"treinamento":     PontosBaseTreinamento,
		"comercial":       PontosBaseComercial,
		"midia":           PontosBaseMidia,
		"educacao":        PontosBaseEducacao,
		"viagens":         PontosBaseViagens,
		"reconhecimentos": PontosBaseReconhecimentos,
		"parcerias":       PontosBaseParcerias,
		"analise":         PontosBaseAnalise,
	}

	MultiplicadoresPorNivel = map[string]float64{
		"baixo":       MultiplicadorNivelBaixo,
		"medio":       MultiplicadorNivelMedio,
		"alto":        MultiplicadorNivelAlto,
		"critico":     MultiplicadorNivelCritico,
		"excepcional": MultiplicadorNivelExcepcional,
	}
)
