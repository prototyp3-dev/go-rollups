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
  hdl.Handler
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

  h := JsonHandler{RouteKey: routeKey, Handler: *handler}
  h.HandleAdvanceRoutes(h.jsonAdvanceHandler)
  h.HandleInspectRoutes(h.jsonInspectHandler)
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
      if this.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Received JSON route",route,"Advance Request:",result) }
      return this.RouteAdvanceHandlers[route].Handler.Handle(metadata,result),true
    }
  }
  return nil,false
}

func (this *JsonHandler) jsonInspectHandler(payloadHex string) (error,bool) {
  if route,result, ok := this.getRoute(payloadHex); ok {
    if this.RouteInspectHandlers[route] != nil {
      if this.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Received JSON route",route,"Inspect Request:",result) }
      return this.RouteInspectHandlers[route].Handler.Handle(result),true
    }
  }
  return nil,false
}