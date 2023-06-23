package rollups

import (
  "encoding/json"
  "log"
  "os"
)

type AdvanceMapHandlerFunc func(*Metadata,map[string]interface{}) error
func (f AdvanceMapHandlerFunc) Handle(m *Metadata,p map[string]interface{}) error {
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
  Handler
  RouteKey string
  RouteAdvanceHandlers map[string]*AdvanceMapHandler
  RouteInspectHandlers map[string]*InspectMapHandler
}

func NewJsonHandler(routeKey string) *JsonHandler {
	if routeKey == "" {
		panic("rollups handler: invalid route key")
	}

  traceLogger = log.New(os.Stderr, "[ trace ] ", log.Lshortfile)
  debugLogger = log.New(os.Stderr, "[ debug ] ", log.Lshortfile)
  errorLogger = log.New(os.Stderr, "[ error ] ", log.Lshortfile)

  handler := JsonHandler{RouteKey: routeKey}
  handler.logLevel = Error
  handler.HandleAdvanceRoutes(handler.jsonAdvanceHandler)
  handler.HandleInspectRoutes(handler.jsonInspectHandler)
  return &handler
}

func (this *JsonHandler) HandleAdvanceRoute(route string, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
	if route == "" {
		panic("rollups handler: invalid route")
	}
  if this.RouteAdvanceHandlers == nil {
    this.RouteAdvanceHandlers = make(map[string]*AdvanceMapHandler)
  }
  fnHandler := AdvanceMapHandler{fnHandle}
  this.RouteAdvanceHandlers[route] = &fnHandler
}


func (this *JsonHandler) HandleInspectRoute(route string, fnHandle InspectMapHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
	if route == "" {
		panic("rollups handler: invalid route")
	}
  if this.RouteInspectHandlers == nil {
    this.RouteInspectHandlers = make(map[string]*InspectMapHandler)
  }
  fnHandler := InspectMapHandler{fnHandle}
  this.RouteInspectHandlers[route] = &fnHandler
}

func (this *JsonHandler) getRoute(payloadHex string) (string,map[string]interface{},bool) {
  var result map[string]interface{}
  if this.RouteKey != "" && this.RouteAdvanceHandlers != nil {
    // decode json and get route key
    if payload, err := Hex2Str(payloadHex); err == nil {
      if err = json.Unmarshal([]byte(payload), &result); err == nil {
        if route, ok := result[this.RouteKey].(string); ok {
          return route,result, true
        }
      }
    }
  }
  return "",result, false
}

func (this *JsonHandler) jsonAdvanceHandler(metadata *Metadata, payloadHex string) (error,bool) {
  if route,result, ok := this.getRoute(payloadHex); ok {
    if this.RouteAdvanceHandlers[route] != nil {
      if this.logLevel >= Debug {debugLogger.Println("Received JSON route",route,"Advance Request:",result) }
      return this.RouteAdvanceHandlers[route].Handler.Handle(metadata,result),true
    }
  }
  return nil,false
}

func (this *JsonHandler) jsonInspectHandler(payloadHex string) (error,bool) {
  if route,result, ok := this.getRoute(payloadHex); ok {
    if this.RouteInspectHandlers[route] != nil {
      if this.logLevel >= Debug {debugLogger.Println("Received JSON route",route,"Inspect Request:",result) }
      return this.RouteInspectHandlers[route].Handler.Handle(result),true
    }
  }
  return nil,false
}
