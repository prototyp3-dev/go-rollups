package urihandler

import (
  "regexp"
  hdl "github.com/prototyp3-dev/go-rollups/handler"
  "github.com/prototyp3-dev/go-rollups/rollups"
)

type AdvanceMapHandlerFunc func(*rollups.Metadata,map[string]interface{}) error
func (f AdvanceMapHandlerFunc) Handle(m *rollups.Metadata,p map[string]interface{}) error {
	return f(m,p)
}
type AdvanceMapHandler struct {
  Handler AdvanceMapHandlerFunc 
}

type InspectMapHandlerFunc func(map[string]interface{}) error
func (f InspectMapHandlerFunc) Handle(p map[string]interface{}) error {
	return f(p)
}
type InspectMapHandler struct {
  Handler InspectMapHandlerFunc 
}

type UriHandler struct {
  hdl.Handler
  RouteAdvanceHandlers map[string]*AdvanceMapHandler
  RouteInspectHandlers map[string]*InspectMapHandler
}

func NewUriHandler() *UriHandler {
  return AddUriHandler(hdl.NewSimpleHandler())
}

func AddUriHandler(handler *hdl.Handler) *UriHandler {
  h := UriHandler{Handler: *handler}
  h.HandleAdvanceRoutes(h.uriAdvanceHandler)
  h.HandleInspectRoutes(h.uriInspectHandler)
  return &h
}

func (this *UriHandler) HandleAdvanceRoute(route string, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("uri handler: nil handler")
	}
	if route == "" {
		panic("uri handler: invalid route")
	}
  if this.RouteAdvanceHandlers == nil {
    this.RouteAdvanceHandlers = make(map[string]*AdvanceMapHandler)
  }
	if this.RouteAdvanceHandlers[route] != nil {
		panic("uri handler: route already added")
	}
  fnHandler := AdvanceMapHandler{fnHandle}
  this.RouteAdvanceHandlers[route] = &fnHandler
}


func (this *UriHandler) HandleInspectRoute(route string, fnHandle InspectMapHandlerFunc) {
	if fnHandle == nil {
		panic("uri handler: nil handler")
	}
	if route == "" {
		panic("uri handler: invalid route")
	}
  if this.RouteInspectHandlers == nil {
    this.RouteInspectHandlers = make(map[string]*InspectMapHandler)
  }
	if this.RouteInspectHandlers[route] != nil {
		panic("uri handler: route already added")
	}
  fnHandler := InspectMapHandler{fnHandle}
  this.RouteInspectHandlers[route] = &fnHandler
}

func (this *UriHandler) uriAdvanceHandler(metadata *rollups.Metadata, payloadHex string) (error,bool) {
  if payloadStr, err := rollups.Hex2Str(payloadHex); err == nil {
    for route, handler := range this.RouteAdvanceHandlers {
      if result, ok := tryUri(route,payloadStr); ok {
        return handler.Handler.Handle(metadata,result),true
      }
    }
  }
  return nil,false
}

func (this *UriHandler) uriInspectHandler(payloadHex string) (error,bool) {
  if payloadStr, err := rollups.Hex2Str(payloadHex); err == nil {
    for route, handler := range this.RouteInspectHandlers {
      if result, ok := tryUri(route,payloadStr); ok {
        return handler.Handler.Handle(result),true
      }
    }
  }
  return nil,false
}

func tryUri(patern string,path string) (map[string]interface{}, bool) {
	params := make(map[string]interface{})
	var i, j int
	for i < len(path) {
		switch {
		case j >= len(patern):
			if len(patern) > 0 && patern != "/" && patern[len(patern)-1] == '/' {
				return params, true
			}
			return nil, false
		case patern[j] == ':':
			var name, val string
			var nextc byte
			name, nextc, j = match(patern, isAlnum, j+1)
			val, _, i = match(path, matchPart(nextc), i)
			params[name] = val
		case path[i] == patern[j]:
			i++
			j++
		default:
			return nil, false
		}
	}
	if j != len(patern) {
		return nil, false
	}
	return params, true
}


func matchPart(b byte) func(byte) bool {
	return func(c byte) bool {
		return c != b && c != '/'
	}
}

func match(s string, f func(byte) bool, i int) (matched string, next byte, j int) {
	j = i
	for j < len(s) && f(s[j]) {
		j++
	}
	if j < len(s) {
		next = s[j]
	}
	return s[i:j], next, j
}

func isAlnum(ch byte) bool {
  match, _ := regexp.MatchString("[_a-zA-Z0-9]",string(ch))
	return match
}