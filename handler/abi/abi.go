package abihandler

import (
  hdl "github.com/prototyp3-dev/go-rollups/handler"
  "github.com/prototyp3-dev/go-rollups/rollups"
)

type AdvanceMapHandlerFunc func(*rollups.Metadata,[]interface{}) error
func (f AdvanceMapHandlerFunc) Handle(m *rollups.Metadata,p []interface{}) error {
	return f(m,p)
}
type AdvanceMapHandler struct {
  Handler AdvanceMapHandlerFunc 
}

type InspectMapHandlerFunc func([]interface{}) error
func (f InspectMapHandlerFunc) Handle(p []interface{}) error {
	return f(p)
}
type InspectMapHandler struct {
  Handler InspectMapHandlerFunc 
}

type AbiHandler struct {
  hdl.Handler
  RouteAdvanceHandlers map[string]*AdvanceMapHandler
  RouteInspectHandlers map[string]*InspectMapHandler
  AdvanceCodecs map[string]*Codec
  InspectCodecs map[string]*Codec
}

func NewAbiHandler() *AbiHandler {
  return AddAbiHandler(hdl.NewSimpleHandler())
}

func AddAbiHandler(handler *hdl.Handler) *AbiHandler {
  h := AbiHandler{Handler: *handler}
  h.HandleAdvanceRoutes(h.abiAdvanceHandler)
  h.HandleInspectRoutes(h.abiInspectHandler)
  return &h
}

func (this *AbiHandler) HandleAdvanceRoute(routeCodec *Codec, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("abi handler: nil handler")
	}
  if len(routeCodec.Header) > 0 && len(routeCodec.Header) != 66 {
    panic("abi handler: codec header format")
  }
  if len(routeCodec.Fields) != 0 && len(routeCodec.PackedFields) != 0 {
    panic("abi handler: ambiguous codec fields")
  }
  if this.RouteAdvanceHandlers == nil {
    this.RouteAdvanceHandlers = make(map[string]*AdvanceMapHandler)
  }
  if this.AdvanceCodecs == nil {
    this.AdvanceCodecs = make(map[string]*Codec)
  }
	if this.RouteAdvanceHandlers[routeCodec.Header] != nil {
		panic("abi handler: route already added")
	}
  if (len(this.RouteAdvanceHandlers) > 0 && routeCodec.Header == "") || this.AdvanceCodecs[""] != nil {
    panic("abi handler: multiple codecs with no header-codec ")
  }
  fnHandler := AdvanceMapHandler{fnHandle}
  this.RouteAdvanceHandlers[routeCodec.Header] = &fnHandler
  this.AdvanceCodecs[routeCodec.Header] = routeCodec
  if this.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created Advance route for",routeCodec) }
}

func (this *AbiHandler) HandleInspectRoute(routeCodec *Codec, fnHandle InspectMapHandlerFunc) {
	if fnHandle == nil {
		panic("abi handler: nil handler")
	}
  if len(routeCodec.Header) > 0 && len(routeCodec.Header) != 66 {
    panic("abi handler: codec header format")
  }
  if len(routeCodec.Fields) != 0 && len(routeCodec.PackedFields) != 0 {
    panic("abi handler: ambiguous codec fields")
  }
  if this.RouteInspectHandlers == nil {
    this.RouteInspectHandlers = make(map[string]*InspectMapHandler)
  }
  if this.InspectCodecs == nil {
    this.InspectCodecs = make(map[string]*Codec)
  }
	if this.RouteInspectHandlers[routeCodec.Header] != nil {
		panic("abi handler: route already added")
	}
  if (len(this.RouteInspectHandlers) > 0 && routeCodec.Header == "") || this.InspectCodecs[""] != nil {
    panic("abi handler: multiple codecs with no header-codec ")
  }
  fnHandler := InspectMapHandler{fnHandle}
  this.RouteInspectHandlers[routeCodec.Header] = &fnHandler
  this.InspectCodecs[routeCodec.Header] = routeCodec
  if this.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created Inspect route for",routeCodec) }
}

func (this *AbiHandler) abiAdvanceHandler(metadata *rollups.Metadata, payloadHex string) (error,bool) {
  if this.AdvanceCodecs[""] != nil {
    codec := this.AdvanceCodecs[""]
    result,err := codec.Decode(payloadHex)
    if err != nil {
      return err,true
    }
    if this.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Received Abi no header Advance Request:",result) }
    return this.RouteAdvanceHandlers[""].Handler.Handle(metadata,result),true
  }
  if len(payloadHex) >= 66 {
    header := payloadHex[:66]
    if this.RouteAdvanceHandlers[header] != nil {
      codec := this.AdvanceCodecs[header]
      result,err := codec.Decode(payloadHex)
      if err != nil {
        return err,true
      }
      if this.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Received Abi route",header,"Advance Request:",result) }
      return this.RouteAdvanceHandlers[header].Handler.Handle(metadata,result),true
    }
  }
  return nil,false
}

func (this *AbiHandler) abiInspectHandler(payloadHex string) (error,bool) {
  if this.InspectCodecs[""] != nil {
    codec := this.InspectCodecs[""]
    result,err := codec.Decode(payloadHex)
    if err != nil {
      return err,true
    }
    if this.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Received Abi no header Inspect Request:",result) }
    return this.RouteInspectHandlers[""].Handler.Handle(result),true
  }
  if len(payloadHex) >= 66 {
    header := payloadHex[:66]
    if this.RouteInspectHandlers[header] != nil {
      codec := this.InspectCodecs[header]
      result,err := codec.Decode(payloadHex)
      if err != nil {
        return err,true
      }
      if this.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Received Abi route",header,"Inspect Request:",result) }
      return this.RouteInspectHandlers[header].Handler.Handle(result),true
    }
  }
  return nil,false
}
