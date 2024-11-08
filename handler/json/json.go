package jsonhandler

import (
  "encoding/json"
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

type JsonHandler struct {
  Handler *hdl.Handler
  RouteKey string
  RouteAdvanceHandlers map[string]*AdvanceMapHandler
  RouteInspectHandlers map[string]*InspectMapHandler
}

func NewJsonHandler(routeKey string) *JsonHandler {
  return AddJsonHandler(routeKey, hdl.NewSimpleHandler())
}

func AddJsonHandler(routeKey string, handler *hdl.Handler) *JsonHandler {
	if routeKey == "" {
		panic("json handler: invalid route key")
	}

  h := JsonHandler{RouteKey: routeKey, Handler: handler}
  h.Handler.HandleAdvanceRoutes(h.jsonAdvanceHandler)
  h.Handler.HandleInspectRoutes(h.jsonInspectHandler)
  return &h
}

func (h *JsonHandler) HandleAdvanceRoute(route string, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("json handler: nil handler")
	}
	if route == "" {
		panic("json handler: invalid route")
	}
  if h.RouteAdvanceHandlers == nil {
    h.RouteAdvanceHandlers = make(map[string]*AdvanceMapHandler)
  }
	if h.RouteAdvanceHandlers[route] != nil {
		panic("json handler: route already added")
	}
  fnHandler := AdvanceMapHandler{fnHandle}
  h.RouteAdvanceHandlers[route] = &fnHandler
  if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Created JSON Advance route for",route) }
}


func (h *JsonHandler) HandleInspectRoute(route string, fnHandle InspectMapHandlerFunc) {
	if fnHandle == nil {
		panic("json handler: nil handler")
	}
	if route == "" {
		panic("json handler: invalid route")
	}
  if h.RouteInspectHandlers == nil {
    h.RouteInspectHandlers = make(map[string]*InspectMapHandler)
  }
	if h.RouteInspectHandlers[route] != nil {
		panic("json handler: route already added")
	}
  fnHandler := InspectMapHandler{fnHandle}
  h.RouteInspectHandlers[route] = &fnHandler
  if h.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created JSON Inspect route for",route) }
}

func (h *JsonHandler) getRoute(payloadHex string) (string,map[string]interface{},bool) {
  var result map[string]interface{}
  if h.RouteKey != "" {
    // decode json and get route key
    if payload, err := rollups.Hex2Str(payloadHex); err == nil {
      if err = json.Unmarshal([]byte(payload), &result); err == nil {
        if route, ok := result[h.RouteKey].(string); ok {
          return route,result, true
        }
      }
    }
  }
  return "",result, false
}

func (h *JsonHandler) jsonAdvanceHandler(metadata *rollups.Metadata, payloadHex string) (error,bool) {
  if route,result, ok := h.getRoute(payloadHex); ok {
    if h.RouteAdvanceHandlers[route] != nil {
      if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received JSON route",route,"Advance Request:",result) }
      return h.RouteAdvanceHandlers[route].Handler.Handle(metadata,result),true
    }
  }
  return nil,false
}

func (h *JsonHandler) jsonInspectHandler(payloadHex string) (error,bool) {
  if route,result, ok := h.getRoute(payloadHex); ok {
    if h.RouteInspectHandlers[route] != nil {
      if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received JSON route",route,"Inspect Request:",result) }
      return h.RouteInspectHandlers[route].Handler.Handle(result),true
    }
  }
  return nil,false
}


func (h *JsonHandler) SetDebug() {h.Handler.SetDebug()}
func (h *JsonHandler) SetLogLevel(logLevel hdl.LogLevel) {h.Handler.SetLogLevel(logLevel)}
func (h *JsonHandler) HandleDefault(fnHandle hdl.InspectHandlerFunc) {h.Handler.HandleDefault(fnHandle)}
func (h *JsonHandler) HandleInspect(fnHandle hdl.InspectHandlerFunc) {h.Handler.HandleInspect(fnHandle)}
func (h *JsonHandler) HandleAdvance(fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleAdvance(fnHandle)}
func (h *JsonHandler) HandleRollupsFixedAddresses(fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleRollupsFixedAddresses(fnHandle)}
func (h *JsonHandler) HandleFixedAddress(address string, fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleFixedAddress(address,fnHandle)}
func (h *JsonHandler) SendNotice(payloadHex string) (uint64,error) {return h.Handler.SendNotice(payloadHex)}
func (h *JsonHandler) SendVoucher(destination string, payloadHex string, value *big.Int) (uint64,error) {return h.Handler.SendVoucher(destination,payloadHex,value)}
func (h *JsonHandler) SendReport(payloadHex string) error {return h.Handler.SendReport(payloadHex)}
func (h *JsonHandler) SendException(payloadHex string) error {return h.Handler.SendException(payloadHex)}
func (h *JsonHandler) Run() error {return h.Handler.Run()}
func (h *JsonHandler) InitializeRollupsAddresses(currentNetwork string) error {return h.Handler.InitializeRollupsAddresses(currentNetwork)}
