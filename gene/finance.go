package gene

import (
	"math"
)

// Zyrion codes
const (
	ZyrionNo        = 0 // falso / no
	ZyrionYes       = 1 // verdadero / sí
	ZyrionUncertain = 2 // incierto
)

// RatioResult contiene el valor numérico y la evaluación ternaria.
type RatioResult struct {
	Value   float64
	Zyrion  int // 0 / 1 / 2
	Message string
}

// LiquidezCorriente = Activo Corriente / Pasivo Corriente
func LiquidezCorriente(activoCorriente, pasivoCorriente float64) RatioResult {
	if pasivoCorriente == 0 {
		return RatioResult{0, ZyrionUncertain, "pasivo corriente = 0"}
	}
	v := activoCorriente / pasivoCorriente
	z := ZyrionYes
	msg := "liquidez aceptable"
	if v < 1.0 {
		z = ZyrionNo
		msg = "liquidez crítica (< 1.0)"
	} else if v < 1.2 {
		z = ZyrionUncertain
		msg = "liquidez justa"
	}
	return RatioResult{v, z, msg}
}

// PruebaAcida = (Activo Corriente - Inventario) / Pasivo Corriente
func PruebaAcida(activoCorriente, inventario, pasivoCorriente float64) RatioResult {
	if pasivoCorriente == 0 {
		return RatioResult{0, ZyrionUncertain, "pasivo corriente = 0"}
	}
	v := (activoCorriente - inventario) / pasivoCorriente
	z := ZyrionYes
	msg := "prueba ácida aceptable"
	if v < 0.8 {
		z = ZyrionNo
		msg = "prueba ácida baja"
	} else if v < 1.0 {
		z = ZyrionUncertain
		msg = "prueba ácida justa"
	}
	return RatioResult{v, z, msg}
}

// Endeudamiento = Pasivo Total / Activo Total
func Endeudamiento(pasivoTotal, activoTotal float64) RatioResult {
	if activoTotal == 0 {
		return RatioResult{0, ZyrionUncertain, "activo total = 0"}
	}
	v := pasivoTotal / activoTotal
	z := ZyrionYes
	msg := "endeudamiento controlado"
	if v > 0.7 {
		z = ZyrionNo
		msg = "endeudamiento alto (> 0.70)"
	} else if v > 0.55 {
		z = ZyrionUncertain
		msg = "endeudamiento moderado-alto"
	}
	return RatioResult{v, z, msg}
}

// ROE = Utilidad Neta / Patrimonio
func ROE(utilidadNeta, patrimonio float64) RatioResult {
	if patrimonio == 0 {
		return RatioResult{0, ZyrionUncertain, "patrimonio = 0"}
	}
	v := utilidadNeta / patrimonio
	z := ZyrionYes
	msg := "ROE aceptable"
	if v < 0.05 {
		z = ZyrionNo
		msg = "ROE bajo"
	} else if v < 0.10 {
		z = ZyrionUncertain
		msg = "ROE moderado"
	}
	return RatioResult{v, z, msg}
}

// ROA = Utilidad Neta / Activo Total
func ROA(utilidadNeta, activoTotal float64) RatioResult {
	if activoTotal == 0 {
		return RatioResult{0, ZyrionUncertain, "activo total = 0"}
	}
	v := utilidadNeta / activoTotal
	z := ZyrionYes
	msg := "ROA aceptable"
	if v < 0.03 {
		z = ZyrionNo
		msg = "ROA bajo"
	} else if v < 0.06 {
		z = ZyrionUncertain
		msg = "ROA moderado"
	}
	return RatioResult{v, z, msg}
}

// MargenNeto = Utilidad Neta / Ingresos
func MargenNeto(utilidadNeta, ingresos float64) RatioResult {
	if ingresos == 0 {
		return RatioResult{0, ZyrionUncertain, "ingresos = 0"}
	}
	v := utilidadNeta / ingresos
	z := ZyrionYes
	msg := "margen aceptable"
	if v < 0.05 {
		z = ZyrionNo
		msg = "margen bajo"
	} else if v < 0.10 {
		z = ZyrionUncertain
		msg = "margen moderado"
	}
	return RatioResult{v, z, msg}
}

// VPN calcula el Valor Presente Neto.
// flujos[0] suele ser la inversión inicial (negativa).
func VPN(flujos []float64, tasa float64) RatioResult {
	if len(flujos) == 0 {
		return RatioResult{0, ZyrionUncertain, "sin flujos"}
	}
	var npv float64
	for t, cf := range flujos {
		npv += cf / math.Pow(1+tasa, float64(t))
	}
	z := ZyrionYes
	msg := "VPN positivo"
	if npv < 0 {
		z = ZyrionNo
		msg = "VPN negativo"
	} else if npv < 1000 { // umbral configurable
		z = ZyrionUncertain
		msg = "VPN cercano a cero"
	}
	return RatioResult{npv, z, msg}
}

// TIR calcula la Tasa Interna de Retorno por método de Newton-Raphson.
// Devuelve la tasa (decimal, ej. 0.15 = 15%) y evaluación.
func TIR(flujos []float64) RatioResult {
	if len(flujos) < 2 {
		return RatioResult{0, ZyrionUncertain, "se necesitan al menos 2 flujos"}
	}
	// Newton-Raphson
	r := 0.1
	for i := 0; i < 50; i++ {
		var npv, dnpv float64
		for t, cf := range flujos {
			ft := float64(t)
			den := math.Pow(1+r, ft)
			npv += cf / den
			if t > 0 {
				dnpv -= ft * cf / (den * (1 + r))
			}
		}
		if math.Abs(dnpv) < 1e-12 {
			break
		}
		r2 := r - npv/dnpv
		if math.Abs(r2-r) < 1e-8 {
			r = r2
			break
		}
		r = r2
	}
	z := ZyrionYes
	msg := "TIR calculada"
	if math.IsNaN(r) || math.IsInf(r, 0) {
		return RatioResult{0, ZyrionUncertain, "TIR no converge"}
	}
	if r < 0 {
		z = ZyrionNo
		msg = "TIR negativa"
	} else if r < 0.08 {
		z = ZyrionUncertain
		msg = "TIR baja"
	}
	return RatioResult{r, z, msg}
}

// AlertaFinanciera combina ratios y devuelve si hay alerta.
func AlertaFinanciera(liq, acida, end, roe, roa, margen float64) RatioResult {
	alertas := 0
	if liq < 1.0 {
		alertas++
	}
	if acida < 0.8 {
		alertas++
	}
	if end > 0.7 {
		alertas++
	}
	if roe < 0.05 {
		alertas++
	}
	if roa < 0.03 {
		alertas++
	}
	if margen < 0.05 {
		alertas++
	}
	if alertas >= 2 {
		return RatioResult{float64(alertas), ZyrionNo, "múltiples alertas financieras"}
	}
	if alertas == 1 {
		return RatioResult{1, ZyrionUncertain, "una alerta financiera"}
	}
	return RatioResult{0, ZyrionYes, "sin alertas críticas"}
}
