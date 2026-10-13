package gene

import "testing"

func TestLiquidezCorriente(t *testing.T) {
	r := LiquidezCorriente(150, 80)
	if r.Value < 1.8 || r.Value > 1.9 {
		t.Fatalf("liq=%v", r.Value)
	}
	if r.Zyrion != ZyrionYes {
		t.Fatalf("zyrion=%d msg=%s", r.Zyrion, r.Message)
	}
	r2 := LiquidezCorriente(50, 80)
	if r2.Zyrion != ZyrionNo {
		t.Fatalf("expected no, got %d", r2.Zyrion)
	}
}

func TestTIRVPN(t *testing.T) {
	flujos := []float64{-100000, 40000, 40000, 40000, 40000}
	tir := TIR(flujos)
	if tir.Zyrion == ZyrionUncertain && tir.Message == "TIR no converge" {
		t.Fatal(tir.Message)
	}
	vpn := VPN(flujos, 0.1)
	if vpn.Value == 0 && vpn.Zyrion == ZyrionUncertain {
		t.Fatal("vpn empty")
	}
}

func TestAlertaFinanciera(t *testing.T) {
	a := AlertaFinanciera(0.5, 0.5, 0.8, 0.02, 0.01, 0.02)
	if a.Zyrion != ZyrionNo {
		t.Fatalf("expected alertas, got %d %s", a.Zyrion, a.Message)
	}
	ok := AlertaFinanciera(1.5, 1.2, 0.4, 0.15, 0.08, 0.12)
	if ok.Zyrion != ZyrionYes {
		t.Fatalf("expected clean, got %d", ok.Zyrion)
	}
}
