package abihandler

import (
  "strings"
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
  Handler *hdl.Handler
  RouteAdvanceHandlers map[string]*AdvanceMapHandler
  RouteInspectHandlers map[string]*InspectMapHandler
  FixedAddressAdvanceHandlers map[string]map[string]*AdvanceMapHandler
  FixedAdvanceCodecs map[string]map[string]*Codec
  AdvanceCodecs map[string]*Codec
  InspectCodecs map[string]*Codec
}

func NewAbiHandler() *AbiHandler {
  return AddAbiHandler(hdl.NewSimpleHandler())
}

func AddAbiHandler(handler *hdl.Handler) *AbiHandler {
  h := AbiHandler{Handler: handler}
  h.Handler.HandleAdvanceRoutes(h.abiAdvanceHandler)
  h.Handler.HandleInspectRoutes(h.abiInspectHandler)
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
  if this.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created ABI Advance route for",routeCodec) }
}

func (this *AbiHandler) HandleFixedAddressAdvance(address string, routeCodec *Codec, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("abi handler: nil handler")
	}
  if len(routeCodec.Header) > 0 && len(routeCodec.Header) != 66 {
    panic("abi handler: codec header format")
  }
  if len(routeCodec.Fields) != 0 && len(routeCodec.PackedFields) != 0 {
    panic("abi handler: ambiguous codec fields")
  }
  if this.FixedAddressAdvanceHandlers == nil {
    this.FixedAddressAdvanceHandlers = make(map[string]map[string]*AdvanceMapHandler)
  }
  if this.FixedAdvanceCodecs == nil {
    this.FixedAdvanceCodecs = make(map[string]map[string]*Codec)
  }
  address = strings.ToLower(address)
  if this.FixedAddressAdvanceHandlers[address] == nil {
    this.FixedAddressAdvanceHandlers[address] = make(map[string]*AdvanceMapHandler)
  }
  if this.FixedAdvanceCodecs[address] == nil {
    this.FixedAdvanceCodecs[address] = make(map[string]*Codec)
  }

	if this.FixedAddressAdvanceHandlers[address][routeCodec.Header] != nil {
		panic("abi handler: route already added")
	}
  if (len(this.FixedAddressAdvanceHandlers[address]) > 0 && routeCodec.Header == "") || this.FixedAdvanceCodecs[address][""] != nil {
    panic("abi handler: multiple codecs with no header-codec ")
  }
  fnHandler := AdvanceMapHandler{fnHandle}
  this.FixedAddressAdvanceHandlers[address][routeCodec.Header] = &fnHandler
  this.FixedAdvanceCodecs[address][routeCodec.Header] = routeCodec

  if this.Handler.FixedAddressHandlers[address] == nil {
    this.Handler.HandleFixedAddressRoutes(address, this.abiFixedAdvanceHandler(address))
  }

  if this.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created Fixed ABI Advance route for",address,routeCodec) }
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
  if this.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created ABI Inspect route for",routeCodec) }
}

func (this *AbiHandler) abiAdvanceHandler(metadata *rollups.Metadata, payloadHex string) (error,bool) {
  if this.AdvanceCodecs[""] != nil {
    codec := this.AdvanceCodecs[""]
    result,err := codec.Decode(payloadHex)
    if err != nil {
      return err,true
    }
    if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI no header Advance Request:",result) }
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
      if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI route",header,"Advance Request:",result) }
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
    if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI no header Inspect Request:",result) }
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
      if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI route",header,"Inspect Request:",result) }
      return this.RouteInspectHandlers[header].Handler.Handle(result),true
    }
  }
  return nil,false
}

func (this *AbiHandler) abiFixedAdvanceHandler(address string) (func(*rollups.Metadata,string) (error,bool)) {
  address = strings.ToLower(address)
  return func(metadata *rollups.Metadata, payloadHex string) (error,bool) {
    if this.FixedAdvanceCodecs[address][""] != nil {
      codec := this.FixedAdvanceCodecs[address][""]
      result,err := codec.Decode(payloadHex)
      if err != nil {
        return err,true
      }
      if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI no header Fixed Advance Request:",result) }
      return this.FixedAddressAdvanceHandlers[address][""].Handler.Handle(metadata,result),true
    }
    if len(payloadHex) >= 66 {
      header := payloadHex[:66]
      if this.FixedAddressAdvanceHandlers[address][header] != nil {
        codec := this.FixedAdvanceCodecs[address][header]
        result,err := codec.Decode(payloadHex)
        if err != nil {
          return err,true
        }
        if this.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI route",header,"Advance Request:",result) }
        return this.FixedAddressAdvanceHandlers[address][header].Handler.Handle(metadata,result),true
      }
    }
    return nil,false
  }
}

func (this *AbiHandler) SetDebug() {this.Handler.SetDebug()}
func (this *AbiHandler) SetLogLevel(logLevel hdl.LogLevel) {this.Handler.SetLogLevel(logLevel)}
func (this *AbiHandler) HandleDefault(fnHandle hdl.InspectHandlerFunc) {this.Handler.HandleDefault(fnHandle)}
func (this *AbiHandler) HandleInspect(fnHandle hdl.InspectHandlerFunc) {this.Handler.HandleInspect(fnHandle)}
func (this *AbiHandler) HandleAdvance(fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleAdvance(fnHandle)}
func (this *AbiHandler) HandleRollupsFixedAddresses(fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleRollupsFixedAddresses(fnHandle)}
func (this *AbiHandler) HandleFixedAddress(address string, fnHandle hdl.AdvanceHandlerFunc) {this.Handler.HandleFixedAddress(address,fnHandle)}
func (this *AbiHandler) SendNotice(payloadHex string) (uint64,error) {return this.Handler.SendNotice(payloadHex)}
func (this *AbiHandler) SendVoucher(destination string, payloadHex string) (uint64,error) {return this.Handler.SendVoucher(destination,payloadHex)}
func (this *AbiHandler) SendReport(payloadHex string) error {return this.Handler.SendReport(payloadHex)}
func (this *AbiHandler) SendException(payloadHex string) error {return this.Handler.SendException(payloadHex)}
func (this *AbiHandler) Run() error {return this.Handler.Run()}
func (this *AbiHandler) InitializeRollupsAddresses(currentNetwork string) error {return this.Handler.InitializeRollupsAddresses(currentNetwork)}
