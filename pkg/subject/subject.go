package subject

import "github.com/Mattilsynet/map-types/gen/go/query/v1"

type QuerySubject struct {
	Kind    string
	Id      string
	Session string
	prefix  string
}

func NewQuerySubject(qry *query.Query) *QuerySubject {
	id := qry.GetStatus().GetId()
	session := qry.GetSpec().GetSession()
	kind := qry.GetSpec().GetType().GetKind()
	prefix := "map"
	return &QuerySubject{
		Kind:    kind,
		Session: session,
		Id:      id,
		prefix:  prefix,
	}
}

func (qs *QuerySubject) ToQuery() string {
	subject := qs.prefix + "." + qs.Kind + "." + qs.Session + "." + qs.Id + ".GET"
	return subject
}
