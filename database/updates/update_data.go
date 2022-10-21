package updates

import (
	"embed"
	_ "embed"
	"fmt"
	"github.com/Riven-Spell/hydaelyn/database"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//go:embed sql_files
var updates embed.FS

type DBUpdate struct {
	Version uint64
	Queries []database.TxOP
}

func ParseUpdate(in string, file string) DBUpdate {
	out := DBUpdate{}

	cCommand := ""
	for idx, line := range strings.Split(in, "\n") {
		{ // trim comment
			if idx == 0 {
				if !strings.HasPrefix(line, "#version:") {
					panic(file + ": update files must start with #version:number")
				}

				var err error
				out.Version, err = strconv.ParseUint(strings.TrimPrefix(line, "#version:"), 10, 64)
				if err != nil {
					panic(fmt.Errorf("%s: failed to parse version: %w", file, err))
				}
			}

			commentIdx := strings.Index(line, "#")
			if commentIdx != -1 {
				line = strings.TrimSpace(line[:commentIdx])
			}
		}

		cCommand += line
		if strings.HasSuffix(cCommand, ";") {
			out.Queries = append(out.Queries, database.TxOP{
				Op:    database.OpTypeManip,
				Query: cCommand,
			})
			cCommand = ""
		}
	}

	return out
}

var Init = func() DBUpdate {
	buf, err := updates.ReadFile("sql_files/init.sql")
	if err != nil {
		panic(err)
	}

	return ParseUpdate(string(buf), "sql_files/init.sql")
}()

var Validate = func() []database.TxOP {
	buf, err := updates.ReadFile("sql_files/validate.sql")
	if err != nil {
		panic(err)
	}

	return ParseUpdate(string(buf), "sql_files/validate.sql").Queries
}()

var updateFileRegex = regexp.MustCompile(`^\d+\.sql$`)

var Versions = func() []DBUpdate {
	entries, err := updates.ReadDir("sql_files")
	if err != nil {
		panic(err)
	}

	out := make([]DBUpdate, 0)
	path := "sql_files"

	for _, v := range entries {
		if v.IsDir() {
			continue
		}

		if !updateFileRegex.MatchString(filepath.Base(v.Name())) {
			continue
		}

		buf, err := updates.ReadFile(filepath.Join(path, v.Name()))
		if err != nil {
			panic(err)
		}

		out = append(out, ParseUpdate(string(buf), v.Name()))
	}

	sort.Sort(sortByVersion(out))

	return out
}()
