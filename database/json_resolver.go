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

func (j *JsonResolveTarget) BuildArg() interface{} {
	buf, err := json.Marshal(j.Target)

	if err != nil {
		panic(err)
	}

	return string(buf)
}
