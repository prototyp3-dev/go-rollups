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
  Handler *hdl.Handler
  RouteAdvanceHandlers map[string]*AdvanceMapHandler
  RouteInspectHandlers map[string]*InspectMapHandler
}

func NewUriHandler() *UriHandler {
  return AddUriHandler(hdl.NewSimpleHandler())
}

func AddUriHandler(handler *hdl.Handler) *UriHandler {
  h := UriHandler{Handler: handler}
  h.Handler.HandleAdvanceRoutes(h.uriAdvanceHandler)
  h.Handler.HandleInspectRoutes(h.uriInspectHandler)
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
  if this.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created URI Advance route for",route) }
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
  if this.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created URI Inspect route for",route) }
}

func (this *UriHandler) uriAdvanceHandler(metadata *rollups.Metadata, payloadHex string) (error,bool) {
  if payloadStr, err := rollups.Hex2Str(payloadHex); err == nil {
    for route, handler := range this.RouteAdvanceHandlers {
      if result, ok := tryUri(route,payloadStr); ok {
        if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received URI route",route,"Advance Request:",result) }
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
        if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received URI route",route,"Inspect Request:",result) }
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

func (this *UriHandler) SetDebug() {this.Handler.SetDebug()}
func (this *UriHandler) SetLogLevel(logLevel hdl.LogLevel) {this.Handler.SetLogLevel(logLevel)}
func (this *UriHandler) HandleDefault(fnHandle hdl.InspectHandlerFunc) {this.Handler.HandleDefault(fnHandle)}
func (this *UriHandler) HandleInspect(fnHandle hdl.InspectHandlerFunc) {this.Handler.HandleInspect(fnHandle)}
func (this *UriHandler) HandleAdvance(fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleAdvance(fnHandle)}
func (this *UriHandler) HandleRollupsFixedAddresses(fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleRollupsFixedAddresses(fnHandle)}
func (this *UriHandler) HandleFixedAddress(address string, fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleFixedAddress(address,fnHandle)}
func (this *UriHandler) SendNotice(payloadHex string) (uint64,error) {return this.Handler.SendNotice(payloadHex)}
func (this *UriHandler) SendVoucher(destination string, payloadHex string) (uint64,error) {return this.Handler.SendVoucher(destination,payloadHex)}
func (this *UriHandler) SendReport(payloadHex string) error {return this.Handler.SendReport(payloadHex)}
func (this *UriHandler) SendException(payloadHex string) error {return this.Handler.SendException(payloadHex)}
func (this *UriHandler) Run() error {return this.Handler.Run()}
func (this *UriHandler) InitializeRollupsAddresses(currentNetwork string) error {return this.Handler.InitializeRollupsAddresses(currentNetwork)}
