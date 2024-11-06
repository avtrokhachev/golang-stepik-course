package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// Структуры для ручек
type HttpServer struct {
	Name     string
	Handlers []ApiHandler
}

type ApiHandler struct {
	ServerName        string
	HandlerName       string
	HandlerMethodType string
	Url               string
	RequestClassName  string
	IsAuthEnabled     bool
}

// Структуры для валидации
type ValidatedRequest struct {
	Name   string
	Fields []ValidatedField
}

type ValidatedField struct {
	Name         string
	Type         string
	ParamName    string
	IsRequired   bool
	DefaultValue string
	EnumValues   []string
	MinCheck     bool
	MaxCheck     bool
	MinValue     int
	MaxValue     int
}

// Структуры для парсинга json
type ApiHandlerInfo struct {
	Url    string
	Auth   bool
	Method string
}

// Шаблоны
var (
	validatorRequestTpl = template.Must(template.New("validatorRequest").Parse(`
func validate{{.Name}}(query url.Values) ({{.Name}}, error) {
	validatedRequest := {{.Name}}{}
	{{range .Fields}}
	{{if eq .Type "Int"}}var err error
	validatedRequest.{{.Name}}, err = strconv.Atoi(query.Get("{{.ParamName}}"))
	if err != nil {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on convertation")}
	}
	{{else}}validatedRequest.{{.Name}} = query.Get("{{.ParamName }}")
	{{end}}
	{{if .IsRequired}}
	if validatedRequest.{{.Name}} == "" {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("Required filed is empty")}
	}{{end}}

	{{if .DefaultValue}}
	if validatedRequest.{{.Name}} == "" {
		validatedRequest.{{.Name}} = "{{.DefaultValue}}"
	}{{end}}

	{{if and .MaxCheck (eq .Type "Int")}}
	if validatedRequest.{{.Name}} > {{.MaxValue}} {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation maxValue")}
	}{{end}}

	{{if and .MaxCheck (eq .Type "String")}}
	if len(validatedRequest.{{.Name}}) > {{.MaxValue}} {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation maxValue")}
	}{{end}}

	{{if and .MinCheck (eq .Type "Int")}}
	if validatedRequest.{{.Name}} < {{.MinValue}} {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation minValue")}
	}{{end}}

	{{if and .MinCheck (eq .Type "String")}}
	if len(validatedRequest.{{.Name}}) < {{.MinValue}} {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("An error occured on validation maxValue")}
	}{{end}}

	{{if .EnumValues}}
	enum{{.Name}}Values := []string{ {{range $i, $v := .EnumValues}} {{if $i}}, {{end}}"{{$v}}"{{end}} }
	isValid := false
	for _, val := range enum{{.Name}}Values {
		if val == validatedRequest.{{.Name}} {
			isValid = true
			break
		}
	}
	if isValid {
		return validatedRequest, ApiError{http.StatusBadRequest, fmt.Errorf("{{.ParamName}} must be one of [%s]", strings.Join(enum{{.Name}}Values, ", "))}
	}{{end}}
	{{end}}

	return validatedRequest, nil
}
`))

	wrapperTpl = template.Must(template.New("wrapperTpl").Parse(`
func (h *{{ .ServerName }}) wrapper{{ .HandlerName }}(w http.ResponseWriter, r *http.Request) {
	{{if .IsAuthEnabled}}
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		return
	}{{end}}

	{{if .HandlerMethodType}}
	if r.Method != "{{.HandlerMethodType}}" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}{{end}}

	var query url.Values
	if r.Method == "GET" {
		query = r.URL.Query()
	} else {
		_ = r.ParseForm()
		query = r.Form
	}

	request, err := validate{{.RequestClassName}}(query)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	var response interface{}
	response, err = h.{{.HandlerName}}(r.Context(), request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]interface{}{
		"response": response,
		"error":    "",
	})
	w.Write(data)
}
`))

	serveHTTPTpl = template.Must(template.New("serveHTTP").Parse(`
func (h *{{ .Name }}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	{{ range .Handlers }}case "{{ .Url }}":
		h.wrapper{{.HandlerName}}(w, r)
	{{ end }}default:
		w.WriteHeader(http.StatusNotFound)
	}
}
`))
)

// Конвертация из CamelCase в snake_case
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	var out *os.File
	out, err = os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import "fmt"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out, `import "strings"`)
	fmt.Fprintln(out, `import "encoding/json"`)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out, `import "net/url"`)
	fmt.Fprintln(out) // empty line

	servers := make(map[string]HttpServer)
	structs := make(map[string]ValidatedRequest)
	for _, decl := range node.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Doc == nil {
				fmt.Printf("SKIP struct %#v doesnt have comments\n", decl.Name.Name)
				continue
			}

			handlerInfo := ApiHandlerInfo{}
			for _, c := range decl.Doc.List {
				if strings.Contains(c.Text, "// apigen:api") {
					err = json.Unmarshal([]byte(c.Text[14:]), &handlerInfo)
					if err != nil {
						panic(err)
					}
					break
				}
			}

			var handlerName string
			for _, field := range decl.Recv.List {
				handlerName = field.Type.(*ast.StarExpr).X.(*ast.Ident).Name
			}
			method := ApiHandler{
				ServerName:        handlerName,
				HandlerName:       decl.Name.Name,
				HandlerMethodType: handlerInfo.Method,
				Url:               handlerInfo.Url,
				RequestClassName:  decl.Type.Params.List[1].Type.(*ast.Ident).Name,
				IsAuthEnabled:     handlerInfo.Auth,
			}

			handler := HttpServer{
				Name: handlerName,
			}
			if _, ok := servers[handlerName]; ok {
				handler = servers[handlerName]
			}

			handler.Handlers = append(handler.Handlers, method)
			servers[handlerName] = handler
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					fmt.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
					continue
				}

				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					fmt.Printf("SKIP %#T is not ast.StructType\n", currStruct)
					continue
				}

				for _, field := range currStruct.Fields.List {
					if field.Tag == nil {
						continue
					}

					if strings.Contains(field.Tag.Value, `apivalidator:`) {
						if _, ok := structs[currType.Name.Name]; !ok {
							structs[currType.Name.Name] = ValidatedRequest{
								Name: currType.Name.Name,
							}
						}

						fieldValidator := ValidatedField{
							Name:      field.Names[0].Name,
							Type:      strings.Title(field.Type.(*ast.Ident).Name),
							ParamName: ToSnakeCase(field.Names[0].Name),
						}
						var err error
						for _, rawExpression := range strings.Split(field.Tag.Value[15:len(field.Tag.Value)-2], ",") {
							validationExpression := strings.Split(rawExpression, "=")
							switch validationExpression[0] {
							case "required":
								fieldValidator.IsRequired = true
							case "default":
								fieldValidator.DefaultValue = validationExpression[1]
							case "paramname":
								fieldValidator.ParamName = validationExpression[1]
							case "enum":
								fieldValidator.EnumValues = strings.Split(validationExpression[1], "|")
							case "min":
								fieldValidator.MinCheck = true
								fieldValidator.MinValue, err = strconv.Atoi(validationExpression[1])
								if err != nil {
									panic(err)
								}
							case "max":
								fieldValidator.MaxCheck = true
								fieldValidator.MaxValue, err = strconv.Atoi(validationExpression[1])
								if err != nil {
									panic(err)
								}
							}
						}

						validatedReq := structs[currType.Name.Name]
						validatedReq.Fields = append(validatedReq.Fields, fieldValidator)
						structs[currType.Name.Name] = validatedReq
					}
				}
			}
		}
	}

	for _, server := range servers {
		serveHTTPTpl.Execute(out, server)
		for _, handler := range server.Handlers {
			wrapperTpl.Execute(out, handler)
		}
	}

	for _, validator := range structs {
		validatorRequestTpl.Execute(out, validator)
	}
}
