// Copyright 2017 Google Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package layers

import (
	"encoding/binary"
	"fmt"

	"github.com/google/gopacket"
)

// PPI represents the decoded header and fields
type PPI struct {
	BaseLayer
	Version uint8  // Currently 0 in spec
	Flags   uint8  // Flags. Bit 0 (LSB) specifies 32-bit alignment
	Length  uint16 // Length of whole message including this header and TLV payload
	DLT     uint32 // Data Link Type of enclosed packet data
	Fields  []PPIField
}

// PPIField represents a decoded field including data
type PPIField struct {
	Type   uint16 // Type
	Length uint16 // Length of Data
	Data   []byte
}

func (ppi *PPI) DecodeFromBytes(data []byte, p gopacket.DecodeFeedback) error {
	ppi.Version = data[0]
	ppi.Flags = data[1]
	ppi.Length = binary.LittleEndian.Uint16(data[2:4])
	ppi.DLT = binary.LittleEndian.Uint32(data[4:8])

	ppi.BaseLayer = BaseLayer{Contents: data[:ppi.Length], Payload: data[ppi.Length:]}

	// Todo: Parse fields

	return nil
}

func (p *PPI) LayerType() gopacket.LayerType { return LayerTypePPI }

func decodePPI(data []byte, p gopacket.PacketBuilder) error {
	if len(data) < 8 {
		return fmt.Errorf("Not a valid PPI Packet. Packet length too small.")
	}
	ppi := &PPI{}
	return decodingLayerDecoder(ppi, data, p)
}

func (p *PPI) NextLayerType() gopacket.LayerType {
	// PPI can nominally contain anything. In practice, it's
	// almost always 802.11, but it might contain other kinds of
	// raw radio capture packets.

	// This is deeply suboptimal. I'm not sure how else to go
	// from a numerical layer type to the correct type without
	// adding a lookup table to enums.go. Is that the right
	// solution?
	if p.DLT == 105 {
		return LayerTypeDot11
	} else {
		panic(fmt.Sprintf("%v is an unknown PPI next layer", p.DLT))
	}
}
