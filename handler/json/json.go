package jsonhandler

import (
  "encoding/json"

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

func (this *JsonHandler) HandleAdvanceRoute(route string, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("json handler: nil handler")
	}
	if route == "" {
		panic("json handler: invalid route")
	}
  if this.RouteAdvanceHandlers == nil {
    this.RouteAdvanceHandlers = make(map[string]*AdvanceMapHandler)
  }
	if this.RouteAdvanceHandlers[route] != nil {
		panic("json handler: route already added")
	}
  fnHandler := AdvanceMapHandler{fnHandle}
  this.RouteAdvanceHandlers[route] = &fnHandler
  if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Created JSON Advance route for",route) }
}


func (this *JsonHandler) HandleInspectRoute(route string, fnHandle InspectMapHandlerFunc) {
	if fnHandle == nil {
		panic("json handler: nil handler")
	}
	if route == "" {
		panic("json handler: invalid route")
	}
  if this.RouteInspectHandlers == nil {
    this.RouteInspectHandlers = make(map[string]*InspectMapHandler)
  }
	if this.RouteInspectHandlers[route] != nil {
		panic("json handler: route already added")
	}
  fnHandler := InspectMapHandler{fnHandle}
  this.RouteInspectHandlers[route] = &fnHandler
  if this.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created JSON Inspect route for",route) }
}

func (this *JsonHandler) getRoute(payloadHex string) (string,map[string]interface{},bool) {
  var result map[string]interface{}
  if this.RouteKey != "" {
    // decode json and get route key
    if payload, err := rollups.Hex2Str(payloadHex); err == nil {
      if err = json.Unmarshal([]byte(payload), &result); err == nil {
        if route, ok := result[this.RouteKey].(string); ok {
          return route,result, true
        }
      }
    }
  }
  return "",result, false
}

func (this *JsonHandler) jsonAdvanceHandler(metadata *rollups.Metadata, payloadHex string) (error,bool) {
  if route,result, ok := this.getRoute(payloadHex); ok {
    if this.RouteAdvanceHandlers[route] != nil {
      if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received JSON route",route,"Advance Request:",result) }
      return this.RouteAdvanceHandlers[route].Handler.Handle(metadata,result),true
    }
  }
  return nil,false
}

func (this *JsonHandler) jsonInspectHandler(payloadHex string) (error,bool) {
  if route,result, ok := this.getRoute(payloadHex); ok {
    if this.RouteInspectHandlers[route] != nil {
      if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received JSON route",route,"Inspect Request:",result) }
      return this.RouteInspectHandlers[route].Handler.Handle(result),true
    }
  }
  return nil,false
}


func (this *JsonHandler) SetDebug() {this.Handler.SetDebug()}
func (this *JsonHandler) SetLogLevel(logLevel hdl.LogLevel) {this.Handler.SetLogLevel(logLevel)}
func (this *JsonHandler) HandleDefault(fnHandle hdl.InspectHandlerFunc) {this.Handler.HandleDefault(fnHandle)}
func (this *JsonHandler) HandleInspect(fnHandle hdl.InspectHandlerFunc) {this.Handler.HandleInspect(fnHandle)}
func (this *JsonHandler) HandleAdvance(fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleAdvance(fnHandle)}
func (this *JsonHandler) HandleRollupsFixedAddresses(fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleRollupsFixedAddresses(fnHandle)}
func (this *JsonHandler) HandleFixedAddress(address string, fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleFixedAddress(address,fnHandle)}
func (this *JsonHandler) SendNotice(payloadHex string) (uint64,error) {return this.Handler.SendNotice(payloadHex)}
func (this *JsonHandler) SendVoucher(destination string, payloadHex string) (uint64,error) {return this.Handler.SendVoucher(destination,payloadHex)}
func (this *JsonHandler) SendReport(payloadHex string) error {return this.Handler.SendReport(payloadHex)}
func (this *JsonHandler) SendException(payloadHex string) error {return this.Handler.SendException(payloadHex)}
func (this *JsonHandler) Run() error {return this.Handler.Run()}
func (this *JsonHandler) InitializeRollupsAddresses(currentNetwork string) error {return this.Handler.InitializeRollupsAddresses(currentNetwork)}
