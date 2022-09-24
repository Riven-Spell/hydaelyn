package rolereact

import (
	"github.com/Riven-Spell/hydaelyn/database/queries"
	"strings"
)

func ParseRoleReact(s string) queries.RoleReact {
	sepIdx := strings.LastIndex(s, ":") // roles don't contain colons.
	roles := strings.Split(s[sepIdx+1:], ",")
	for k, v := range roles {
		roles[k] = strings.TrimSpace(v)
	}
	emoji := strings.TrimSpace(s[:sepIdx])

	return queries.RoleReact{Roles: roles, Emoji: emoji}
}
