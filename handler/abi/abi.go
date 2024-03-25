package abihandler

import (
  "strings"
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

func (h *AbiHandler) HandleAdvanceRoute(routeCodec *Codec, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("abi handler: nil handler")
	}
  if len(routeCodec.Header) > 0 && len(routeCodec.Header) != 66 {
    panic("abi handler: codec header format")
  }
  if len(routeCodec.Fields) != 0 && len(routeCodec.PackedFields) != 0 {
    panic("abi handler: ambiguous codec fields")
  }
  if h.RouteAdvanceHandlers == nil {
    h.RouteAdvanceHandlers = make(map[string]*AdvanceMapHandler)
  }
  if h.AdvanceCodecs == nil {
    h.AdvanceCodecs = make(map[string]*Codec)
  }
	if h.RouteAdvanceHandlers[routeCodec.Header] != nil {
		panic("abi handler: route already added")
	}
  if (len(h.RouteAdvanceHandlers) > 0 && routeCodec.Header == "") || h.AdvanceCodecs[""] != nil {
    panic("abi handler: multiple codecs with no header-codec ")
  }
  fnHandler := AdvanceMapHandler{fnHandle}
  h.RouteAdvanceHandlers[routeCodec.Header] = &fnHandler
  h.AdvanceCodecs[routeCodec.Header] = routeCodec
  if h.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created ABI Advance route for",routeCodec) }
}

func (h *AbiHandler) HandleFixedAddressAdvance(address string, routeCodec *Codec, fnHandle AdvanceMapHandlerFunc) {
	if fnHandle == nil {
		panic("abi handler: nil handler")
	}
  if len(routeCodec.Header) > 0 && len(routeCodec.Header) != 66 {
    panic("abi handler: codec header format")
  }
  if len(routeCodec.Fields) != 0 && len(routeCodec.PackedFields) != 0 {
    panic("abi handler: ambiguous codec fields")
  }
  if h.FixedAddressAdvanceHandlers == nil {
    h.FixedAddressAdvanceHandlers = make(map[string]map[string]*AdvanceMapHandler)
  }
  if h.FixedAdvanceCodecs == nil {
    h.FixedAdvanceCodecs = make(map[string]map[string]*Codec)
  }
  address = strings.ToLower(address)
  if h.FixedAddressAdvanceHandlers[address] == nil {
    h.FixedAddressAdvanceHandlers[address] = make(map[string]*AdvanceMapHandler)
  }
  if h.FixedAdvanceCodecs[address] == nil {
    h.FixedAdvanceCodecs[address] = make(map[string]*Codec)
  }

	if h.FixedAddressAdvanceHandlers[address][routeCodec.Header] != nil {
		panic("abi handler: route already added")
	}
  if (len(h.FixedAddressAdvanceHandlers[address]) > 0 && routeCodec.Header == "") || h.FixedAdvanceCodecs[address][""] != nil {
    panic("abi handler: multiple codecs with no header-codec ")
  }
  fnHandler := AdvanceMapHandler{fnHandle}
  h.FixedAddressAdvanceHandlers[address][routeCodec.Header] = &fnHandler
  h.FixedAdvanceCodecs[address][routeCodec.Header] = routeCodec

  if h.Handler.FixedAddressHandlers[address] == nil {
    h.Handler.HandleFixedAddressRoutes(address, h.abiFixedAdvanceHandler(address))
  }

  if h.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created Fixed ABI Advance route for",address,routeCodec) }
}

func (h *AbiHandler) HandleInspectRoute(routeCodec *Codec, fnHandle InspectMapHandlerFunc) {
	if fnHandle == nil {
		panic("abi handler: nil handler")
	}
  if len(routeCodec.Header) > 0 && len(routeCodec.Header) != 66 {
    panic("abi handler: codec header format")
  }
  if len(routeCodec.Fields) != 0 && len(routeCodec.PackedFields) != 0 {
    panic("abi handler: ambiguous codec fields")
  }
  if h.RouteInspectHandlers == nil {
    h.RouteInspectHandlers = make(map[string]*InspectMapHandler)
  }
  if h.InspectCodecs == nil {
    h.InspectCodecs = make(map[string]*Codec)
  }
	if h.RouteInspectHandlers[routeCodec.Header] != nil {
		panic("abi handler: route already added")
	}
  if (len(h.RouteInspectHandlers) > 0 && routeCodec.Header == "") || h.InspectCodecs[""] != nil {
    panic("abi handler: multiple codecs with no header-codec ")
  }
  fnHandler := InspectMapHandler{fnHandle}
  h.RouteInspectHandlers[routeCodec.Header] = &fnHandler
  h.InspectCodecs[routeCodec.Header] = routeCodec
  if h.Handler.LogLevel >= hdl.Debug {hdl.DebugLogger.Println("Created ABI Inspect route for",routeCodec) }
}

func (h *AbiHandler) abiAdvanceHandler(metadata *rollups.Metadata, payloadHex string) (error,bool) {
  if h.AdvanceCodecs[""] != nil {
    codec := h.AdvanceCodecs[""]
    result,err := codec.Decode(payloadHex)
    if err != nil {
      return err,true
    }
    if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI no header Advance Request:",result) }
    return h.RouteAdvanceHandlers[""].Handler.Handle(metadata,result),true
  }
  if len(payloadHex) >= 66 {
    header := payloadHex[:66]
    if h.RouteAdvanceHandlers[header] != nil {
      codec := h.AdvanceCodecs[header]
      result,err := codec.Decode(payloadHex)
      if err != nil {
        return err,true
      }
      if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI route",header,"Advance Request:",result) }
      return h.RouteAdvanceHandlers[header].Handler.Handle(metadata,result),true
    }
  }
  return nil,false
}

func (h *AbiHandler) abiInspectHandler(payloadHex string) (error,bool) {
  if h.InspectCodecs[""] != nil {
    codec := h.InspectCodecs[""]
    result,err := codec.Decode(payloadHex)
    if err != nil {
      return err,true
    }
    if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI no header Inspect Request:",result) }
    return h.RouteInspectHandlers[""].Handler.Handle(result),true
  }
  if len(payloadHex) >= 66 {
    header := payloadHex[:66]
    if h.RouteInspectHandlers[header] != nil {
      codec := h.InspectCodecs[header]
      result,err := codec.Decode(payloadHex)
      if err != nil {
        return err,true
      }
      if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI route",header,"Inspect Request:",result) }
      return h.RouteInspectHandlers[header].Handler.Handle(result),true
    }
  }
  return nil,false
}

func (h *AbiHandler) abiFixedAdvanceHandler(address string) (func(*rollups.Metadata,string) (error,bool)) {
  address = strings.ToLower(address)
  return func(metadata *rollups.Metadata, payloadHex string) (error,bool) {
    if h.FixedAdvanceCodecs[address][""] != nil {
      codec := h.FixedAdvanceCodecs[address][""]
      result,err := codec.Decode(payloadHex)
      if err != nil {
        return err,true
      }
      if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI no header Fixed Advance Request:",result) }
      return h.FixedAddressAdvanceHandlers[address][""].Handler.Handle(metadata,result),true
    }
    if len(payloadHex) >= 66 {
      header := payloadHex[:66]
      if h.FixedAddressAdvanceHandlers[address][header] != nil {
        codec := h.FixedAdvanceCodecs[address][header]
        result,err := codec.Decode(payloadHex)
        if err != nil {
          return err,true
        }
        if h.Handler.LogLevel >= hdl.Trace {hdl.TraceLogger.Println("Received ABI route",header,"Advance Request:",result) }
        return h.FixedAddressAdvanceHandlers[address][header].Handler.Handle(metadata,result),true
      }
    }
    return nil,false
  }
}

func (h *AbiHandler) SetDebug() {h.Handler.SetDebug()}
func (h *AbiHandler) SetLogLevel(logLevel hdl.LogLevel) {h.Handler.SetLogLevel(logLevel)}
func (h *AbiHandler) HandleDefault(fnHandle hdl.InspectHandlerFunc) {h.Handler.HandleDefault(fnHandle)}
func (h *AbiHandler) HandleInspect(fnHandle hdl.InspectHandlerFunc) {h.Handler.HandleInspect(fnHandle)}
func (h *AbiHandler) HandleAdvance(fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleAdvance(fnHandle)}
func (h *AbiHandler) HandleRollupsFixedAddresses(fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleRollupsFixedAddresses(fnHandle)}
func (h *AbiHandler) HandleFixedAddress(address string, fnHandle hdl.AdvanceHandlerFunc) {h.Handler.HandleFixedAddress(address,fnHandle)}
func (h *AbiHandler) SendNotice(payloadHex string) (uint64,error) {return h.Handler.SendNotice(payloadHex)}
func (h *AbiHandler) SendVoucher(destination string, payloadHex string) (uint64,error) {return h.Handler.SendVoucher(destination,payloadHex)}
func (h *AbiHandler) SendReport(payloadHex string) error {return h.Handler.SendReport(payloadHex)}
func (h *AbiHandler) SendException(payloadHex string) error {return h.Handler.SendException(payloadHex)}
func (h *AbiHandler) Run() error {return h.Handler.Run()}
func (h *AbiHandler) InitializeRollupsAddresses(currentNetwork string) error {return h.Handler.InitializeRollupsAddresses(currentNetwork)}
