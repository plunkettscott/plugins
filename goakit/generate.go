package goakit

import (
	"path"
	"regexp"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

// Register the plugin Generator functions.
func init() {
	codegen.RegisterPluginFirst("goakit", "gen", nil, Generate)
	codegen.RegisterPluginLast("goakit-goakitify", "gen", nil, Goakitify)
	codegen.RegisterPluginLast("goakit-goakitify-example", "example", nil, GoakitifyExample)
}

// Generate generates go-kit specific decoders and encoders.
func Generate(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, root := range roots {
		if r, ok := root.(*expr.RootExpr); ok {
			files = append(files, EncodeDecodeFiles(genpkg, r)...)
			files = append(files, MountFiles(r)...)
		}
	}
	return files, nil
}

// Goakitify modifies all the previously generated files by adding go-kit
// imports and replacing the following instances
// * "goa.Endpoint" with "github.com/go-kit/kit/endpoint".Endpoint
// * "log.Logger" with "github.com/go-kit/kit/log".Logger
//
// Goakitify also wraps instances of endpoint.Endpoint into instances of
// goa.Endpoint when used as argument of either goagrpc.NewStreamHandler or
// goagrpc.NewUnaryHandler.
func Goakitify(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, f := range files {
		goakitify(f)
	}
	return files, nil
}
func goakitify(f *codegen.File) {
	var hasEndpoint bool
	for _, s := range f.SectionTemplates {
		if !hasEndpoint {
			hasEndpoint = goaEndpointRegexp.MatchString(s.Source)
		}
		s.Source = goaEndpointRegexp.ReplaceAllString(s.Source, "${1}endpoint.Endpoint${2}")
		if s.Name == "grpc-handler-init" {
			s.Source = strings.Replace(s.Source, "Handler(endpoint, ", "Handler(goa.Endpoint(endpoint), ", 1)
		}
	}
	if hasEndpoint {
		codegen.AddImport(
			f.SectionTemplates[0],
			&codegen.ImportSpec{Path: "github.com/go-kit/kit/endpoint"},
		)
	}
}

// GoakitifyExample  modifies all the previously generated example files by
// adding go-kit imports.
func GoakitifyExample(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	for _, f := range files {
		gokitifyExampleServer(genpkg, f)
	}
	return files, nil
}

// goaEndpointRegexp matches occurrences of the "goa.Endpoint" type in Go code.
var goaEndpointRegexp = regexp.MustCompile(`([^\p{L}_])goa\.Endpoint([^\p{L}_])`)

// goaLoggerRegexp matches occurrences of "logger.<function>" in Go code.
var goaLoggerRegexp = regexp.MustCompile(`logger\.\w+\((.*)\)`)

// gokitifyExampleServer imports gokit endpoint, logger, and transport
// packages in the example server implementation. It also replaces every stdlib
// logger with gokit logger.
func gokitifyExampleServer(genpkg string, file *codegen.File) {
	goakitify(file)
	var hasLogger bool
	for _, s := range file.SectionTemplates {
		if !hasLogger {
			hasLogger = strings.Contains(s.Source, "*log.Logger")
		}
		s.Source = strings.Replace(s.Source, "*log.Logger", "log.Logger", -1)
		s.Source = strings.Replace(s.Source, "adapter = middleware.NewLogger(logger)", "adapter = logger", 1)
		codegen.AddImport(file.SectionTemplates[0], &codegen.ImportSpec{Path: "fmt"})
		s.Source = goaLoggerRegexp.ReplaceAllString(s.Source, "logger.Log(\"info\", fmt.Sprintf(${1}))")

		switch s.Name {
		case "server-main-logger":
			codegen.AddImport(file.SectionTemplates[0], &codegen.ImportSpec{Path: "github.com/go-kit/kit/log"})
			s.Source = gokitLoggerT
		case "server-http-logger":
			s.Source = ""
		case "server-http-middleware":
			s.Source = strings.Replace(s.Source, "adapter", "logger", -1)
		case "server-http-init":
			codegen.AddImport(file.SectionTemplates[0], &codegen.ImportSpec{Path: "github.com/go-kit/kit/transport/http", Name: "kithttp"})
			codegen.AddImport(file.SectionTemplates[0], &codegen.ImportSpec{Path: "github.com/go-kit/kit/endpoint"})
			data := s.Data.(map[string]interface{})
			svcs := data["Services"].([]*httpcodegen.ServiceData)
			for _, svc := range svcs {
				svcData := httpcodegen.HTTPServices.Get(svc.Service.Name).Service
				codegen.AddImport(file.SectionTemplates[0], &codegen.ImportSpec{
					Path: path.Join(genpkg, "http", svcData.PathName, "kitserver"),
					Name: svcData.PkgName + "kitsvr",
				})
			}
			s.Source = gokitServerInitT
		}
	}
	if hasLogger {
		// Replace existing stdlib logger with gokit logger in imports
		if data, ok := file.SectionTemplates[0].Data.(map[string]interface{}); ok {
			if imports, ok := data["Imports"]; ok {
				specs := imports.([]*codegen.ImportSpec)
				for _, s := range specs {
					if s.Path == "log" {
						s.Path = "github.com/go-kit/kit/log"
					}
				}
			}
		}
	}
}

const gokitLoggerT = `
  // Setup gokit logger.
  var (
    logger log.Logger
  )
  {
    logger = log.NewLogfmtLogger(os.Stderr)
    logger = log.With(logger, "ts", log.DefaultTimestampUTC)
    logger = log.With(logger, "caller", log.DefaultCaller)
  }
`

const gokitServerInitT = `
  // Wrap the endpoints with the transport specific layers. The generated
  // server packages contains code generated from the design which maps
  // the service input and output data structures to HTTP requests and
  // responses.
  var (
  {{- range .Services }}
    {{- range .Endpoints }}
      {{ .ServiceVarName }}{{ .Method.VarName }}Handler *kithttp.Server
    {{- end }}
    {{ .Service.VarName }}Server *{{.Service.PkgName}}svr.Server
  {{- end }}
  )
  {
    eh := errorHandler(logger)
    {{- if needStream .Services }}
      upgrader := &websocket.Upgrader{}
    {{- end }}
  {{- range .Services }}
    {{- if .Endpoints }}
      {{- range .Endpoints }}
        {{ .ServiceVarName }}{{ .Method.VarName }}Handler = kithttp.NewServer(
          endpoint.Endpoint({{ .ServiceVarName }}Endpoints.{{ .Method.VarName }}),
          {{- if .Payload.Ref }}
            {{ .ServicePkgName}}kitsvr.{{ .RequestDecoder }}(mux, dec),
          {{- else }}
            func(context.Context, *http.Request) (request interface{}, err error) { return nil, nil },
          {{- end }}
          {{ .ServicePkgName}}kitsvr.{{ .ResponseEncoder }}(enc),
          {{- if .Errors }}
            kithttp.ServerErrorEncoder({{ .ServicePkgName}}kitsvr.{{ .ErrorEncoder }}(enc, nil)),
          {{- end }}
        )
      {{- end }}
      {{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New({{ .Service.VarName }}Endpoints, mux, dec, enc, eh, nil{{ if needStream $.Services }}, upgrader, nil{{ end }}{{ range .Endpoints }}{{ if .MultipartRequestDecoder }}, {{ $.APIPkg }}.{{ .MultipartRequestDecoder.FuncName }}{{ end }}{{ end }})
    {{-  else }}
      {{ .Service.VarName }}Server = {{ .Service.PkgName }}svr.New(nil, mux, dec, enc, eh, nil)
    {{-  end }}
  {{- end }}
  }

  // Configure the mux.
  {{- range .Services }}{{ $service := . }}
    {{- range .Endpoints }}
  {{ .ServicePkgName}}kitsvr.{{ .MountHandler }}(mux, {{ .ServiceVarName }}{{ .Method.VarName }}Handler)
    {{- end }}
    {{- range .FileServers }}
  {{ $service.Service.PkgName}}kitsvr.{{ .MountHandler }}(mux)
    {{- end }}
  {{- end }}
`
