package postgrest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type FilterBuilder struct {
	client    *Client
	method    string
	body      []byte
	tableName string
	headers   map[string]string
	params    map[string]string
}

func (f *FilterBuilder) ExecuteString() (string, error) {
	return executeString(f.client, f.method, f.body, []string{f.tableName}, f.headers, f.params)
}

func (f *FilterBuilder) Execute() ([]byte, error) {
	return execute(f.client, f.method, f.body, []string{f.tableName}, f.headers, f.params)
}

func (f *FilterBuilder) ExecuteTo(to interface{}) error {
	return executeTo(f.client, f.method, f.body, to, []string{f.tableName}, f.headers, f.params)
}

var filterOperators = []string{"eq", "neq", "gt", "gte", "lt", "lte", "like", "ilike", "is", "in", "cs", "cd", "sl", "sr", "nxl", "nxr", "adj", "ov", "fts", "plfts", "phfts", "wfts"}

func isOperator(value string) bool {
	for _, operator := range filterOperators {
		if value == operator {
			return true
		}
	}
	return false
}

func (f *FilterBuilder) Filter(column, operator, value string) *FilterBuilder {
	if !isOperator(operator) {
		f.client.ClientError = fmt.Errorf("invalid filter operator")
		return f
	}
	f.params[column] = fmt.Sprintf("%s.%s", operator, value)
	return f
}

func (f *FilterBuilder) Or(filters, foreignTable string) *FilterBuilder {
	if foreignTable != "" {
		f.params[foreignTable+".or"] = fmt.Sprintf("(%s)", filters)
	} else {
		f.params[foreignTable+"or"] = fmt.Sprintf("(%s)", filters)
	}
	return f
}

func (f *FilterBuilder) Not(column, operator, value string) *FilterBuilder {
	if !isOperator(operator) {
		return f
	}
	f.params[column] = fmt.Sprintf("not.%s.%s", operator, value)
	return f
}

func (f *FilterBuilder) Match(userQuery map[string]string) *FilterBuilder {
	for key, value := range userQuery {
		f.params[key] = "eq." + value
	}
	return f
}

func (f *FilterBuilder) Eq(column, value string) *FilterBuilder {
	f.params[column] = "eq." + value
	return f
}

func (f *FilterBuilder) Neq(column, value string) *FilterBuilder {
	f.params[column] = "neq." + value
	return f
}

func (f *FilterBuilder) Gt(column, value string) *FilterBuilder {
	f.params[column] = "gt." + value
	return f
}

func (f *FilterBuilder) Gte(column, value string) *FilterBuilder {
	f.params[column] = "gte." + value
	return f
}

func (f *FilterBuilder) Lt(column, value string) *FilterBuilder {
	f.params[column] = "lt." + value
	return f
}

func (f *FilterBuilder) Lte(column, value string) *FilterBuilder {
	f.params[column] = "lte." + value
	return f
}

func (f *FilterBuilder) Like(column, value string) *FilterBuilder {
	f.params[column] = "like." + value
	return f
}

func (f *FilterBuilder) Ilike(column, value string) *FilterBuilder {
	f.params[column] = "ilike." + value
	return f
}

func (f *FilterBuilder) Is(column, value string) *FilterBuilder {
	f.params[column] = "is." + value
	return f
}

func (f *FilterBuilder) In(column string, values []string) *FilterBuilder {
	var cleanedValues []string
	illegalChars := regexp.MustCompile("[,()]")
	for _, value := range values {
		exp := illegalChars.MatchString(value)
		if exp {
			cleanedValues = append(cleanedValues, fmt.Sprintf("\"%s\"", value))
		} else {
			cleanedValues = append(cleanedValues, value)
		}
	}
	f.params[column] = fmt.Sprintf("in.(%s)", strings.Join(cleanedValues, ","))
	return f
}

func (f *FilterBuilder) Contains(column string, value []string) *FilterBuilder {
	f.params[column] = "cs." + strings.Join(value, ",")
	return f
}

func (f *FilterBuilder) ContainedBy(column string, value []string) *FilterBuilder {
	f.params[column] = "cd." + strings.Join(value, ",")
	return f
}

func (f *FilterBuilder) ContainsObject(column string, value interface{}) *FilterBuilder {
	sum, err := json.Marshal(value)
	if err != nil {
		f.client.ClientError = err
	}
	f.params[column] = "cs." + string(sum)
	return f
}

func (f *FilterBuilder) ContainedByObject(column string, value interface{}) *FilterBuilder {
	sum, err := json.Marshal(value)
	if err != nil {
		f.client.ClientError = err
	}
	f.params[column] = "cs." + string(sum)
	return f
}

func (f *FilterBuilder) RangeLt(column, value string) *FilterBuilder {
	f.params[column] = "sl." + value
	return f
}

func (f *FilterBuilder) RangeGt(column, value string) *FilterBuilder {
	f.params[column] = "sr." + value
	return f
}

func (f *FilterBuilder) RangeGte(column, value string) *FilterBuilder {
	f.params[column] = "nxl." + value
	return f
}

func (f *FilterBuilder) RangeLte(column, value string) *FilterBuilder {
	f.params[column] = "nxr." + value
	return f
}

func (f *FilterBuilder) RangeAdjacent(column, value string) *FilterBuilder {
	f.params[column] = "adj." + value
	return f
}

func (f *FilterBuilder) Overlaps(column string, value []string) *FilterBuilder {
	f.params[column] = "ov." + strings.Join(value, ",")
	return f
}

func (f *FilterBuilder) TextSearch(column, userQuery, config, tsType string) *FilterBuilder {
	var typePart, configPart string
	if tsType == "plain" {
		typePart = "pl"
	} else if tsType == "phrase" {
		typePart = "ph"
	} else if tsType == "websearch" {
		typePart = "w"
	} else if tsType == "" {
		typePart = ""
	} else {
		f.client.ClientError = fmt.Errorf("invalid text search type")
		return f
	}
	if config != "" {
		configPart = fmt.Sprintf("(%s)", config)
	}
	f.params[column] = typePart + "fts" + configPart + "." + userQuery
	return f
}
