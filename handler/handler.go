package handler

import (
  "encoding/json"
  "io"
  "strconv"
  "strings"
  "fmt"
  "log"
  "os"

  "github.com/prototyp3-dev/go-rollups/rollups"
)

type LogLevel uint8
const (
  None LogLevel = iota
  Critical
  Error
  Warning
  Info
  Debug
  Trace
)

type NetworkAddresses struct {
  DappAddressRelay string           `json:"DAPP_RELAY_ADDRESS"`
  EtherPortalAddress string         `json:"ETHER_PORTAL_ADDRESS"`
  Erc20PortalAddress string         `json:"ERC20_PORTAL_ADDRESS"`
  Erc721PortalAddress string        `json:"ERC721_PORTAL_ADDRESS"`
  Erc1155SinglePortalAddress string `json:"ERC1155_SINGLE_PORTAL_ADDRESS"`
  Erc1155BatchPortalAddress string  `json:"ERC1155_BATCH_PORTAL_ADDRESS"`
}

type AdvanceHandlerFunc func(*rollups.Metadata,string) error
func (f AdvanceHandlerFunc) handle(m *rollups.Metadata,p string) error {
	return f(m,p)
}
type AdvanceHandler struct {
  Handler AdvanceHandlerFunc 
}

type InspectHandlerFunc func(string) error
func (f InspectHandlerFunc) handle(p string) error {
	return f(p)
}
type InspectHandler struct {
  Handler InspectHandlerFunc 
}

type RoutesAdvanceHandlerFunc func(*rollups.Metadata,string) (error,bool)
func (f RoutesAdvanceHandlerFunc) handle(m *rollups.Metadata,p string) (error,bool) {
	return f(m,p)
}
type RoutesAdvanceHandler struct {
  Handler RoutesAdvanceHandlerFunc 
}

type RoutesInspectHandlerFunc func(string) (error,bool)
func (f RoutesInspectHandlerFunc) handle(p string) (error,bool) {
	return f(p)
}
type RoutesInspectHandler struct {
  Handler RoutesInspectHandlerFunc 
}

type Handler struct {
  LogLevel LogLevel
  DefaultHandler *InspectHandler
  AdvanceHandler *AdvanceHandler
  InspectHandler *InspectHandler
  RoutesAdvanceHandlers []*RoutesAdvanceHandler
  RoutesInspectHandlers []*RoutesInspectHandler
  RollupsFixedAddressHandler *AdvanceHandler
  FixedAddressHandlers map[string]*RoutesAdvanceHandler
}

var ErrorLogger *log.Logger
var TraceLogger *log.Logger
var DebugLogger *log.Logger

var LocalHandler = NewSimpleHandler()

func (h *Handler) SetDebug() {
  h.LogLevel = Debug
}

func (h *Handler) SetLogLevel(logLevel LogLevel) {
  h.LogLevel = logLevel
}

func HandleDefault(fnHandle InspectHandlerFunc) {
  LocalHandler.HandleDefault(fnHandle)
}

func (h *Handler) HandleDefault(fnHandle InspectHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  fnHandler := InspectHandler{fnHandle}
  h.DefaultHandler = &fnHandler
}

func HandleInspect(fnHandle InspectHandlerFunc) {
  LocalHandler.HandleInspect(fnHandle)
}

func (h *Handler) HandleInspect(fnHandle InspectHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  fnHandler := InspectHandler{fnHandle}
  h.InspectHandler = &fnHandler
}

func HandleAdvance(fnHandle AdvanceHandlerFunc) {
  LocalHandler.HandleAdvance(fnHandle)
}
func (h *Handler) HandleAdvance(fnHandle AdvanceHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  fnHandler := AdvanceHandler{fnHandle}
  h.AdvanceHandler = &fnHandler
}

func HandleRollupsFixedAddresses(fnHandle AdvanceHandlerFunc) {
  LocalHandler.HandleRollupsFixedAddresses(fnHandle)
}
func (h *Handler) HandleRollupsFixedAddresses(fnHandle AdvanceHandlerFunc) {
  if RollupsAddresses == (NetworkAddresses{}) {
		panic("rollups handler: uninitialized RollupsAddresses")
  }
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  fnHandler := AdvanceHandler{fnHandle}
  h.RollupsFixedAddressHandler = &fnHandler
}

func HandleFixedAddress(address string, fnHandle AdvanceHandlerFunc) {
  LocalHandler.HandleFixedAddress(address,fnHandle)
}
func (h *Handler) HandleFixedAddress(address string, fnHandle AdvanceHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
	if address == "" || address[:2] != "0x" || len(address) != 42 {
		panic("rollups handler: invalid address")
	}
  if h.FixedAddressHandlers == nil {
    h.FixedAddressHandlers = make(map[string]*RoutesAdvanceHandler)
  }
  fnHandler := RoutesAdvanceHandler{func(metadata *rollups.Metadata,payloadHex string) (error,bool) {
    return fnHandle(metadata,payloadHex),true
  }}
  h.FixedAddressHandlers[strings.ToLower(address)] = &fnHandler
}

func HandleFixedAddressRoutes(address string, fnHandle RoutesAdvanceHandlerFunc) {
  LocalHandler.HandleFixedAddressRoutes(address,fnHandle)
}
func (h *Handler) HandleFixedAddressRoutes(address string, fnHandle RoutesAdvanceHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
	if address == "" || address[:2] != "0x" || len(address) != 42 {
		panic("rollups handler: invalid address")
	}
  if h.FixedAddressHandlers == nil {
    h.FixedAddressHandlers = make(map[string]*RoutesAdvanceHandler)
  }
  fnHandler := RoutesAdvanceHandler{fnHandle}
  h.FixedAddressHandlers[strings.ToLower(address)] = &fnHandler
}

func (h *Handler) HandleAdvanceRoutes(fnHandle RoutesAdvanceHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  if h.RoutesAdvanceHandlers == nil {
    h.RoutesAdvanceHandlers = []*RoutesAdvanceHandler{}
  }
  fnHandler := RoutesAdvanceHandler{fnHandle}
  h.RoutesAdvanceHandlers = append(h.RoutesAdvanceHandlers,&fnHandler)
}

func (h *Handler) HandleInspectRoutes(fnHandle RoutesInspectHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  if h.RoutesInspectHandlers == nil {
    h.RoutesInspectHandlers = []*RoutesInspectHandler{}
  }
  fnHandler := RoutesInspectHandler{fnHandle}
  h.RoutesInspectHandlers = append(h.RoutesInspectHandlers,&fnHandler)
}

func (h *Handler) SendNotice(payloadHex string) (uint64,error) {
  notice := &rollups.Notice{Payload:payloadHex}
  if h.LogLevel >= Trace {TraceLogger.Println("Sending notice status",notice)}
  res, err := rollups.SendNotice(notice)
  if err != nil {
    return 0,fmt.Errorf("SendNotice: error making http request: %s", err)
  }
 
  body, err := io.ReadAll(res.Body)
  if err != nil {
    return 0,fmt.Errorf("SendNotice: could not read response body: %s", err)
  }
  
  var indexRes rollups.IndexResponse
  err = json.Unmarshal(body, &indexRes)
  if err != nil {
    return 0,fmt.Errorf("SendNotice: Error unmarshaling body: %s", err)
  }
  if h.LogLevel >= Debug {DebugLogger.Println("Received notice status", strconv.Itoa(res.StatusCode), "body", string(body), "index", strconv.FormatUint(indexRes.Index,10))}

  return indexRes.Index,nil
}

func (h *Handler) SendVoucher(destination string, payloadHex string) (uint64,error) {
  voucher := &rollups.Voucher{Destination: destination, Payload: payloadHex}
  if h.LogLevel >= Trace {TraceLogger.Println("Sending voucher status",voucher)}
  res, err := rollups.SendVoucher(voucher)
  if err != nil {
    return 0,fmt.Errorf("SendVoucher: error making http request: %s", err)
  }
 
  body, err := io.ReadAll(res.Body)
  if err != nil {
    return 0,fmt.Errorf("SendVoucher: could not read response body: %s", err)
  }
  
  var indexRes rollups.IndexResponse
  err = json.Unmarshal(body, &indexRes)
  if err != nil {
    return 0,fmt.Errorf("SendVoucher: Error unmarshaling body: %s", err)
  }
  if h.LogLevel >= Debug {DebugLogger.Println("Received voucher status", strconv.Itoa(res.StatusCode), "body", string(body), "index", strconv.FormatUint(indexRes.Index,10))}

  return indexRes.Index,nil
}

func (h *Handler) SendReport(payloadHex string) error {
  report := &rollups.Report{Payload:payloadHex}
  if h.LogLevel >= Trace {TraceLogger.Println("Sending report status",report)}
  res, err := rollups.SendReport(report)
  if err != nil {
    return fmt.Errorf("SendReport: error making http request: %s", err)
  }

  body, err := io.ReadAll(res.Body)
  if err != nil {
    return fmt.Errorf("SendReport: could not read response body: %s", err)
  }
  if h.LogLevel >= Debug {DebugLogger.Println("Received report status", strconv.Itoa(res.StatusCode), "body", string(body))}

  return nil
}

func (h *Handler) SendException(payloadHex string) error {
  exception := &rollups.Exception{Payload:payloadHex}
  if h.LogLevel >= Trace {TraceLogger.Println("Sending exception status",exception)}
  res, err := rollups.SendException(exception)
  if err != nil {
    return fmt.Errorf("SendException: error making http request: %s", err)
  }

  body, err := io.ReadAll(res.Body)
  if err != nil {
    return fmt.Errorf("SendException: could not read response body: %s", err)
  }
  if h.LogLevel >= Debug {DebugLogger.Println("Received exception status", strconv.Itoa(res.StatusCode), "body", string(body))}

  return nil
}



func NewSimpleHandler() *Handler {
  TraceLogger = log.New(os.Stderr, "[ trace ] ", log.Lshortfile)
  DebugLogger = log.New(os.Stderr, "[ debug ] ", log.Lshortfile)
  ErrorLogger = log.New(os.Stderr, "[ error ] ", log.Lshortfile)

  h := Handler{}
  h.LogLevel = Error
  return &h
}


func RunDebug() error {
  LocalHandler.SetDebug()
  return LocalHandler.Run()
}

func Run() error {
  LocalHandler.SetLogLevel(Error)
  return LocalHandler.Run()
}

func (h *Handler) Run() error {
  if rollups.GetRollupServer() == "" {
    rollups.SetRollupServer(os.Getenv("ROLLUP_HTTP_SERVER_URL"))
  }
  if rollups.GetRollupServer() == "" {
    return fmt.Errorf("rollup server not defined")
  }

  finish := rollups.Finish{Status:"accept"}

  for {
    if h.LogLevel >= Trace {TraceLogger.Println("Sending finish")}
    res, err := rollups.SendFinish(&finish)
    if err != nil {
      return fmt.Errorf("error making http request: %s", err)
    }
    if h.LogLevel >= Trace {TraceLogger.Println("Received finish status", strconv.Itoa(res.StatusCode))}
    
    if (res.StatusCode == 202){
      if h.LogLevel >= Trace {TraceLogger.Println("No pending rollup request, trying again")}
    } else {

      resBody, err := io.ReadAll(res.Body)
      if err != nil {
        return fmt.Errorf("error: could not read response body: %s", err)
      }
      if h.LogLevel >= Debug {DebugLogger.Println("Received request",string(resBody))}
      
      var response rollups.FinishResponse
      err = json.Unmarshal(resBody, &response)
      if err != nil {
        return fmt.Errorf("error: unmarshaling body: %s", err)
      }

      finish.Status = "accept"
      err = h.internalHandleFinish(&response)
      if err != nil {
        if h.LogLevel >= Error {ErrorLogger.Println("Error:", err)}
        finish.Status = "reject"
      }
    }
  }
}

func (h *Handler) internalHandleFinish(response *rollups.FinishResponse) error {
  var err error

  switch response.Type {
  case "advance_state":
    data := new(rollups.AdvanceResponse)
    if err = json.Unmarshal(response.Data, data); err != nil {
      return fmt.Errorf("Handler: Error unmarshaling advance: %s", err)
    }
    err = h.internalHandleAdvance(data)
  case "inspect_state":
    data := new(rollups.InspectResponse)
    if err = json.Unmarshal(response.Data, data); err != nil {
      return fmt.Errorf("Handler: Error unmarshaling inspect: %s", err)
    }
    err = h.internalHandleInspect(data)
  }
  return err
}

func (h *Handler) internalHandleAdvance(data *rollups.AdvanceResponse) error {
  if h.FixedAddressHandlers != nil {
    if h.FixedAddressHandlers[strings.ToLower(data.Metadata.MsgSender)] != nil {
      if err,processed := h.FixedAddressHandlers[strings.ToLower(data.Metadata.MsgSender)].Handler.handle(&data.Metadata,data.Payload); processed { 
        return err
      }
    }
  }
  if h.RollupsFixedAddressHandler != nil && KnownRollupsAddresses[strings.ToLower(data.Metadata.MsgSender)] {
    return h.RollupsFixedAddressHandler.Handler.handle(&data.Metadata,data.Payload)
  }
  if h.RoutesAdvanceHandlers != nil {
    for _, routeHandler := range h.RoutesAdvanceHandlers {
      if err,processed := routeHandler.Handler.handle(&data.Metadata,data.Payload); processed { 
        return err
      }
    }
  }
  if h.AdvanceHandler != nil {
    return h.AdvanceHandler.Handler.handle(&data.Metadata,data.Payload)
  }
  if h.DefaultHandler != nil {
    return h.DefaultHandler.Handler.handle(data.Payload)
  }
  return nil
}

func (h *Handler) internalHandleInspect(data *rollups.InspectResponse) error {
  if h.RoutesInspectHandlers != nil {
    for _, routeHandler := range h.RoutesInspectHandlers {
      if err,processed := routeHandler.Handler.handle(data.Payload); processed {
        return err
      }
    }
  }
  if h.InspectHandler != nil {
    return h.InspectHandler.Handler.handle(data.Payload)
  }
  if h.DefaultHandler != nil {
    return h.DefaultHandler.Handler.handle(data.Payload)
  }
  return nil
}



var RollupsAddresses NetworkAddresses
var KnownRollupsAddresses map[string]bool

func (h *Handler) InitializeRollupsAddresses(currentNetwork string) error {
  return InitializeRollupsAddresses(currentNetwork)
}
func InitializeRollupsAddresses(currentNetwork string) error {
  if KnownRollupsAddresses != nil {
    return nil
  }
  var result map[string]interface{}
  json.Unmarshal([]byte(networks), &result)

  if result[currentNetwork] == nil {
    panic("InitializeRollupsAddresses: Unknown network")
  }

  jsonNetwork, _ := json.Marshal(result[currentNetwork])

  err := json.Unmarshal(jsonNetwork, &RollupsAddresses)
  if err != nil {
    panic(fmt.Sprint("InitializeRollupsAddresses: error unmarshaling network: ", err))
  }
  KnownRollupsAddresses = make(map[string]bool)
  KnownRollupsAddresses[strings.ToLower(RollupsAddresses.DappAddressRelay)] = true
  KnownRollupsAddresses[strings.ToLower(RollupsAddresses.EtherPortalAddress)] = true
  KnownRollupsAddresses[strings.ToLower(RollupsAddresses.Erc20PortalAddress)] = true
  KnownRollupsAddresses[strings.ToLower(RollupsAddresses.Erc721PortalAddress)] = true
  KnownRollupsAddresses[strings.ToLower(RollupsAddresses.Erc1155SinglePortalAddress)] = true
  KnownRollupsAddresses[strings.ToLower(RollupsAddresses.Erc1155BatchPortalAddress)] = true

  return nil
}

