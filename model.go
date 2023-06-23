package rollups

import (
	"encoding/json"
  "math/big"
)

type FinishResponse struct {
  Type string           `json:"request_type"`
  Data json.RawMessage  `json:"data"`
}

type InspectResponse struct {
  Payload string        `json:"payload"`
}

type AdvanceResponse struct {
  Metadata Metadata     `json:"metadata"`
  Payload string        `json:"payload"`
}

type Metadata struct {
  MsgSender string      `json:"msg_sender"`
  EpochIndex uint64     `json:"epoch_index"`
  InputIndex uint64     `json:"input_index"`
  BlockNumber uint64    `json:"block_number"`
  Timestamp uint64      `json:"timestamp"`
}

type Finish struct {
  Status string         `json:"status"`
}

type Report struct {
  Payload string        `json:"payload"`
}

type Notice struct {
  Payload string        `json:"payload"`
}

type Voucher struct {
  Destination string    `json:"destination"`
  Payload string        `json:"payload"`
}

type Exception struct {
  Payload string        `json:"payload"`
}

type IndexResponse struct {
  Index uint64        `json:"index"`
}

type NetworkAddresses struct {
  DappAddressRelay string           `json:"DAPP_RELAY_ADDRESS"`
  EtherPortalAddress string         `json:"ETHER_PORTAL_ADDRESS"`
  Erc20PortalAddress string         `json:"ERC20_PORTAL_ADDRESS"`
  Erc721PortalAddress string        `json:"ERC721_PORTAL_ADDRESS"`
  Erc1155SinglePortalAddress string `json:"ERC1155_SINGLE_PORTAL_ADDRESS"`
  Erc1155BatchPortalAddress string  `json:"ERC1155_BATCH_PORTAL_ADDRESS"`
}

type EtherDeposit struct {
  Depositor string
  Amount *big.Int
  Data []byte
}

type Erc20Deposit struct {
  Depositor string
  TokenAddress string
  Amount *big.Int
  Data []byte
}

type Erc721Deposit struct {
  Depositor string
  TokenAddress string
  TokenId *big.Int
  Data []byte
}

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
