package handler

import (
  "encoding/json"
  "io/ioutil"
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
  FixedAddressHandlers map[string]*AdvanceHandler
}

var ErrorLogger *log.Logger
var TraceLogger *log.Logger
var DebugLogger *log.Logger

var LocalHandler = NewSimpleHandler()

func (this *Handler) SetDebug() {
  this.LogLevel = Debug
}

func (this *Handler) SetLogLevel(LogLevel LogLevel) {
  this.LogLevel = LogLevel
}

func HandleDefault(fnHandle InspectHandlerFunc) {
  LocalHandler.HandleDefault(fnHandle)
}

func (this *Handler) HandleDefault(fnHandle InspectHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  fnHandler := InspectHandler{fnHandle}
  this.DefaultHandler = &fnHandler
}

func HandleInspect(fnHandle InspectHandlerFunc) {
  LocalHandler.HandleInspect(fnHandle)
}

func (this *Handler) HandleInspect(fnHandle InspectHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  fnHandler := InspectHandler{fnHandle}
  this.InspectHandler = &fnHandler
}

func HandleAdvance(fnHandle AdvanceHandlerFunc) {
  LocalHandler.HandleAdvance(fnHandle)
}
func (this *Handler) HandleAdvance(fnHandle AdvanceHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  fnHandler := AdvanceHandler{fnHandle}
  this.AdvanceHandler = &fnHandler
}

func HandleRollupsFixedAddresses(fnHandle AdvanceHandlerFunc) {
  LocalHandler.HandleRollupsFixedAddresses(fnHandle)
}
func (this *Handler) HandleRollupsFixedAddresses(fnHandle AdvanceHandlerFunc) {
  if RollupsAddresses == (NetworkAddresses{}) {
		panic("rollups handler: uninitialized RollupsAddresses")
  }
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  fnHandler := AdvanceHandler{fnHandle}
  this.RollupsFixedAddressHandler = &fnHandler
}

func HandleFixedAddress(address string, fnHandle AdvanceHandlerFunc) {
  LocalHandler.HandleFixedAddress(address,fnHandle)
}
func (this *Handler) HandleFixedAddress(address string, fnHandle AdvanceHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
	if address == "" || address[:2] != "0x" || len(address) != 42 {
		panic("rollups handler: invalid address")
	}
  if this.FixedAddressHandlers == nil {
    this.FixedAddressHandlers = make(map[string]*AdvanceHandler)
  }
  fnHandler := AdvanceHandler{fnHandle}
  this.FixedAddressHandlers[strings.ToLower(address)] = &fnHandler
}

func (this *Handler) HandleAdvanceRoutes(fnHandle RoutesAdvanceHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  if this.RoutesAdvanceHandlers == nil {
    this.RoutesAdvanceHandlers = []*RoutesAdvanceHandler{}
  }
  fnHandler := RoutesAdvanceHandler{fnHandle}
  this.RoutesAdvanceHandlers = append(this.RoutesAdvanceHandlers,&fnHandler)
}

func (this *Handler) HandleInspectRoutes(fnHandle RoutesInspectHandlerFunc) {
	if fnHandle == nil {
		panic("rollups handler: nil handler")
	}
  if this.RoutesInspectHandlers == nil {
    this.RoutesInspectHandlers = []*RoutesInspectHandler{}
  }
  fnHandler := RoutesInspectHandler{fnHandle}
  this.RoutesInspectHandlers = append(this.RoutesInspectHandlers,&fnHandler)
}

func (this *Handler) SendNotice(notice *rollups.Notice) (uint64,error) {
  if this.LogLevel >= Trace {TraceLogger.Println("Sending notice status",notice)}
  res, err := rollups.SendNotice(notice)
  if err != nil {
    return 0,fmt.Errorf("SendNotice: error making http request: %s", err)
  }
 
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return 0,fmt.Errorf("SendNotice: could not read response body: %s", err)
  }
  
  var indexRes rollups.IndexResponse
  err = json.Unmarshal(body, &indexRes)
  if err != nil {
    return 0,fmt.Errorf("SendNotice: Error unmarshaling body: %s", err)
  }
  if this.LogLevel >= Debug {DebugLogger.Println("Received notice status", strconv.Itoa(res.StatusCode), "body", string(body), "index", strconv.FormatUint(indexRes.Index,10))}

  return indexRes.Index,nil
}

func (this *Handler) SendVoucher(voucher *rollups.Voucher) (uint64,error) {
  if this.LogLevel >= Trace {TraceLogger.Println("Sending voucher status",voucher)}
  res, err := rollups.SendVoucher(voucher)
  if err != nil {
    return 0,fmt.Errorf("SendVoucher: error making http request: %s", err)
  }
 
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return 0,fmt.Errorf("SendVoucher: could not read response body: %s", err)
  }
  
  var indexRes rollups.IndexResponse
  err = json.Unmarshal(body, &indexRes)
  if err != nil {
    return 0,fmt.Errorf("SendVoucher: Error unmarshaling body: %s", err)
  }
  if this.LogLevel >= Debug {DebugLogger.Println("Received voucher status", strconv.Itoa(res.StatusCode), "body", string(body), "index", strconv.FormatUint(indexRes.Index,10))}

  return indexRes.Index,nil
}

func (this *Handler) SendReport(report *rollups.Report) error {
  if this.LogLevel >= Trace {TraceLogger.Println("Sending report status",report)}
  res, err := rollups.SendReport(report)
  if err != nil {
    return fmt.Errorf("SendReport: error making http request: %s", err)
  }

  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return fmt.Errorf("SendReport: could not read response body: %s", err)
  }
  if this.LogLevel >= Debug {DebugLogger.Println("Received report status", strconv.Itoa(res.StatusCode), "body", string(body))}

  return nil
}

func (this *Handler) SendException(exception *rollups.Exception) error {
  if this.LogLevel >= Trace {TraceLogger.Println("Sending exception status",exception)}
  res, err := rollups.SendException(exception)
  if err != nil {
    return fmt.Errorf("SendException: error making http request: %s", err)
  }

  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return fmt.Errorf("SendException: could not read response body: %s", err)
  }
  if this.LogLevel >= Debug {DebugLogger.Println("Received exception status", strconv.Itoa(res.StatusCode), "body", string(body))}

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

func (this *Handler) Run() error {
  finish := rollups.Finish{"accept"}

  for true {
    if this.LogLevel >= Trace == true {TraceLogger.Println("Sending finish")}
    res, err := rollups.SendFinish(&finish)
    if err != nil {
      return fmt.Errorf("Error making http request: %s", err)
    }
    if this.LogLevel >= Trace {TraceLogger.Println("Received finish status", strconv.Itoa(res.StatusCode))}
    
    if (res.StatusCode == 202){
      if this.LogLevel >= Trace {TraceLogger.Println("No pending rollup request, trying again")}
    } else {

      resBody, err := ioutil.ReadAll(res.Body)
      if err != nil {
        return fmt.Errorf("Error: could not read response body: %s", err)
      }
      if this.LogLevel >= Debug {DebugLogger.Println("Received request",string(resBody))}
      
      var response rollups.FinishResponse
      err = json.Unmarshal(resBody, &response)
      if err != nil {
        return fmt.Errorf("Error: unmarshaling body: %s", err)
      }

      finish.Status = "accept"
      err = this.internalHandleFinish(&response)
      if err != nil {
        if this.LogLevel >= Error {ErrorLogger.Println("Error:", err)}
        finish.Status = "reject"
      }
    }
  }

	return nil
}

func (this *Handler) internalHandleFinish(response *rollups.FinishResponse) error {
  var err error

  switch response.Type {
  case "advance_state":
    data := new(rollups.AdvanceResponse)
    if err = json.Unmarshal(response.Data, data); err != nil {
      return fmt.Errorf("Handler: Error unmarshaling advance: %s", err)
    }
    err = this.internalHandleAdvance(data)
  case "inspect_state":
    data := new(rollups.InspectResponse)
    if err = json.Unmarshal(response.Data, data); err != nil {
      return fmt.Errorf("Handler: Error unmarshaling inspect: %s", err)
    }
    err = this.internalHandleInspect(data)
  }
  return err
}

func (this *Handler) internalHandleAdvance(data *rollups.AdvanceResponse) error {
  if this.FixedAddressHandlers != nil {
    if this.FixedAddressHandlers[strings.ToLower(data.Metadata.MsgSender)] != nil {
      return this.FixedAddressHandlers[strings.ToLower(data.Metadata.MsgSender)].Handler.handle(&data.Metadata,data.Payload)
    }
  }
  if this.RollupsFixedAddressHandler != nil && KnownRollupsAddresses[strings.ToLower(data.Metadata.MsgSender)] {
    return this.RollupsFixedAddressHandler.Handler.handle(&data.Metadata,data.Payload)
  }
  if this.RoutesAdvanceHandlers != nil {
    for _, routeHandler := range this.RoutesAdvanceHandlers {
      if err,processed := routeHandler.Handler.handle(&data.Metadata,data.Payload); processed { 
        return err
      }
    }
  }
  if this.AdvanceHandler != nil {
    return this.AdvanceHandler.Handler.handle(&data.Metadata,data.Payload)
  }
  if this.DefaultHandler != nil {
    return this.DefaultHandler.Handler.handle(data.Payload)
  }
  return nil
}

func (this *Handler) internalHandleInspect(data *rollups.InspectResponse) error {
  if this.RoutesInspectHandlers != nil {
    for _, routeHandler := range this.RoutesInspectHandlers {
      if err,processed := routeHandler.Handler.handle(data.Payload); processed {
        return err
      }
    }
  }
  if this.InspectHandler != nil {
    return this.InspectHandler.Handler.handle(data.Payload)
  }
  if this.DefaultHandler != nil {
    return this.DefaultHandler.Handler.handle(data.Payload)
  }
  return nil
}



var RollupsAddresses NetworkAddresses
var KnownRollupsAddresses map[string]bool

func InitializeRollupsAddresses(currentNetwork string) error {
  if KnownRollupsAddresses != nil {
    return nil
  }
  var result map[string]interface{}
  json.Unmarshal([]byte(networks), &result)

  if result[currentNetwork] == nil {
    panic(fmt.Sprint("InitializeRollupsAddresses: Unknown network"))
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

