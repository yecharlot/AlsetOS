package ingest

// FinanceStatement es el JSON canónico alset.finance.statement/1.0
// (duplicado aquí para que el paquete ingest sea autónomo;
// en AlsetOS se puede mover a un paquete compartido).
type FinanceStatement struct {
	Schema string `json:"schema"`
	Entity struct {
		Name     string `json:"name"`
		NIT      string `json:"nit"`
		Period   string `json:"period"`
		Currency string `json:"currency"`
	} `json:"entity"`
	Balance struct {
		Assets struct {
			Current    float64 `json:"current"`
			NonCurrent float64 `json:"non_current"`
			Total      float64 `json:"total"`
		} `json:"assets"`
		Liabilities struct {
			Current    float64 `json:"current"`
			NonCurrent float64 `json:"non_current"`
			Total      float64 `json:"total"`
		} `json:"liabilities"`
		Equity float64 `json:"equity"`
	} `json:"balance"`
	Income struct {
		Revenue  float64 `json:"revenue"`
		COGS     float64 `json:"cogs"`
		Gross    float64 `json:"gross"`
		Opex     float64 `json:"opex"`
		EBITDA   float64 `json:"ebitda"`
		EBIT     float64 `json:"ebit"`
		Interest float64 `json:"interest"`
		Tax      float64 `json:"tax"`
		Net      float64 `json:"net"`
	} `json:"income"`
	Cashflow struct {
		Operating float64 `json:"operating"`
		Investing float64 `json:"investing"`
		Financing float64 `json:"financing"`
		Net       float64 `json:"net"`
	} `json:"cashflow"`
	Source struct {
		Files       []string `json:"files"`
		ExtractedAt string   `json:"extracted_at"`
	} `json:"source"`
}
