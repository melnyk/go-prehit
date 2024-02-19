package prehit

// Metrics defines telemetry for Cache
type Metrics interface {
	Hit()
	Miss()
	Error()
	Add()
	Update()
	Evict()
	Delete()
}

// nometrics is a Metrics implementation that does nothing.
type nometrics struct{}

func (n *nometrics) Hit()    {}
func (n *nometrics) Miss()   {}
func (n *nometrics) Error()  {}
func (n *nometrics) Add()    {}
func (n *nometrics) Update() {}
func (n *nometrics) Evict()  {}
func (n *nometrics) Delete() {}
