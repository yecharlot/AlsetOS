package mind

import (
	"testing"

	"github.com/yecharlot/AlsetOS/zyrion"
)

func TestZyrionNumericContract(t *testing.T) {
	if int(zyrion.No)!=0 || int(zyrion.Si)!=1 || int(zyrion.Incierto)!=2 {
		t.Fatalf("contrato incorrecto: No=%d Si=%d Incierto=%d",zyrion.No,zyrion.Si,zyrion.Incierto)
	}
}
