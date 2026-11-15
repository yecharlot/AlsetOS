package zyrion

// Z = {F, V, I} — ATS-001 Parte III.
// Contrato numérico canónico: 0=False, 1=True, 2=Indeterminado.
type Z int

const (
	F Z = 0
	V Z = 1
	I Z = 2
)

func (z Z) String() string {
	switch z {
	case V:
		return "V"
	case F:
		return "F"
	default:
		return "I"
	}
}

func Eval(v interface{}) Z {
	switch t := v.(type) {
	case bool:
		if t {
			return V
		}
		return F
	case nil:
		return I
	case string:
		switch t {
		case "true", "TRUE", "V", "yes":
			return V
		case "false", "FALSE", "F", "no":
			return F
		default:
			return I
		}
	default:
		return I
	}
}

func And(a, b Z) Z {
	if a == F || b == F {
		return F
	}
	if a == I || b == I {
		return I
	}
	return V
}

func Or(a, b Z) Z {
	if a == V || b == V {
		return V
	}
	if a == I || b == I {
		return I
	}
	return F
}

func Not(a Z) Z {
	if a == V {
		return F
	}
	if a == F {
		return V
	}
	return I
}
