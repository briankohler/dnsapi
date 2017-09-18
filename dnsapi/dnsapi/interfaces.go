package main

type RecordDeletor interface {
	DeleteRecord() error
}
type RecordEditor interface {
	EditRecord() error
}
type RecordAdder interface {
	AddRecord() error
}

type ConsulVerifier interface {
	VerifyInConsul() (bool, error)
}

type ConsulRecGenerator interface {
	GenConsulRecord() Consul_Record
}

type DataHandler interface {
	Check() error
	GenConsulRecord() Consul_Record
}

type Consul_Record struct {
	key    string
	value  string
	tiny   string
	record string // This is the actual entity for which a record is created
	rtype  string // TYPE A,CNAME etc.
}
