package models

// DashboardStats representa as estatísticas do dashboard
type DashboardStats struct {
	TotalEquinos        int64 `json:"total_equinos"`
	EquinosAtivos       int64 `json:"equinos_ativos"`
	GestacoesAtivas     int64 `json:"gestacoes_ativas"`
	EventosProximos     int64 `json:"eventos_proximos"`
	ValorizacaoMedia    int64 `json:"valorizacao_media"`
	TotalLeiloes        int64 `json:"total_leiloes"`
	EquinosTokenizados  int64 `json:"equinos_tokenizados"`
}
