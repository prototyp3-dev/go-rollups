package urihandler

import (
  "regexp"
  "math/big"
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

func (h *UriHandler) HandleAdvanceRoute(route string, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("uri handler: nil handler")
	}
	if route == "" {
		panic("uri handler: invalid route")
	}
  if h.RouteAdvanceHandlers == nil {
    h.RouteAdvanceHandlers = make(map[string]*AdvanceMapHandler)
  }
	if h.RouteAdvanceHandlers[route] != nil {
		panic("uri handler: route already added")
	}
  fnHandler := AdvanceMapHandler{fnHandle}
  h.RouteAdvanceHandlers[route] = &fnHandler
  if h.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created URI Advance route for",route) }
}


func (h *UriHandler) HandleInspectRoute(route string, fnHandle InspectMapHandlerFunc) {
	if fnHandle == nil {
		panic("uri handler: nil handler")
	}
	if route == "" {
		panic("uri handler: invalid route")
	}
  if h.RouteInspectHandlers == nil {
    h.RouteInspectHandlers = make(map[string]*InspectMapHandler)
  }
	if h.RouteInspectHandlers[route] != nil {
		panic("uri handler: route already added")
	}
  fnHandler := InspectMapHandler{fnHandle}
  h.RouteInspectHandlers[route] = &fnHandler
  if h.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created URI Inspect route for",route) }
}

func (h *UriHandler) uriAdvanceHandler(metadata *rollups.Metadata, payloadHex string) (error,bool) {
  if payloadStr, err := rollups.Hex2Str(payloadHex); err == nil {
    for route, handler := range h.RouteAdvanceHandlers {
      if result, ok := tryUri(route,payloadStr); ok {
        if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received URI route",route,"Advance Request:",result) }
        return handler.Handler.Handle(metadata,result),true
      }
    }
  }
  return nil,false
}

func (h *UriHandler) uriInspectHandler(payloadHex string) (error,bool) {
  if payloadStr, err := rollups.Hex2Str(payloadHex); err == nil {
    for route, handler := range h.RouteInspectHandlers {
      if result, ok := tryUri(route,payloadStr); ok {
        if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received URI route",route,"Inspect Request:",result) }
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

func (h *UriHandler) SetDebug() {h.Handler.SetDebug()}
func (h *UriHandler) SetLogLevel(logLevel hdl.LogLevel) {h.Handler.SetLogLevel(logLevel)}
func (h *UriHandler) HandleDefault(fnHandle hdl.InspectHandlerFunc) {h.Handler.HandleDefault(fnHandle)}
func (h *UriHandler) HandleInspect(fnHandle hdl.InspectHandlerFunc) {h.Handler.HandleInspect(fnHandle)}
func (h *UriHandler) HandleAdvance(fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleAdvance(fnHandle)}
func (h *UriHandler) HandleRollupsFixedAddresses(fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleRollupsFixedAddresses(fnHandle)}
func (h *UriHandler) HandleFixedAddress(address string, fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleFixedAddress(address,fnHandle)}
func (h *UriHandler) SendNotice(payloadHex string) (uint64,error) {return h.Handler.SendNotice(payloadHex)}
func (h *UriHandler) SendVoucher(destination string, payloadHex string, value *big.Int) (uint64,error) {return h.Handler.SendVoucher(destination,payloadHex,value)}
func (h *UriHandler) SendReport(payloadHex string) error {return h.Handler.SendReport(payloadHex)}
func (h *UriHandler) SendException(payloadHex string) error {return h.Handler.SendException(payloadHex)}
func (h *UriHandler) Run() error {return h.Handler.Run()}
func (h *UriHandler) InitializeRollupsAddresses(currentNetwork string) error {return h.Handler.InitializeRollupsAddresses(currentNetwork)}
