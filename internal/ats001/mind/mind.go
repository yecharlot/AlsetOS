package mind

type Goal struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

type Mind struct {
	STM       []map[string]interface{} `json:"stm"`
	LTM       []map[string]interface{} `json:"ltm"`
	Goals     []Goal                   `json:"goals"`
	Knowledge map[string]interface{}   `json:"knowledge"`
}

func New() *Mind {
	return &Mind{Knowledge: map[string]interface{}{}}
}

func (m *Mind) AddGoal(name string, p int) {
	m.Goals = append(m.Goals, Goal{Name: name, Priority: p})
}

func (m *Mind) Remember(item map[string]interface{}, long bool) {
	if long {
		m.LTM = append(m.LTM, item)
		if len(m.LTM) > 200 {
			m.LTM = m.LTM[len(m.LTM)-200:]
		}
	} else {
		m.STM = append(m.STM, item)
		if len(m.STM) > 40 {
			m.STM = m.STM[len(m.STM)-40:]
		}
	}
}
