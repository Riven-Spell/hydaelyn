package database

import "encoding/json"

type JsonResolveTarget struct {
	Target         any
	internalTarget string
}

func (j *JsonResolveTarget) Substitute() any {
	return &j.internalTarget
}

func (j *JsonResolveTarget) Resolve() error {
	return json.Unmarshal([]byte(j.internalTarget), j.Target)
}
