// Copyright © 2019 Erin Shepherd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package n76

import (
	"errors"

	"github.com/GabrielCRZ12/ms51prog/protocol"
	"github.com/GabrielCRZ12/ms51prog/target"
)

//go:generate enumer -type=BootSelect -trimprefix=BootFrom -transform=snake -json -text

type MS51FB9AELDROMSize byte

const (
	MS51FB9AELDROM0KB MS51FB9AELDROMSize = iota
	MS51FB9AELDROM1KB
	MS51FB9AELDROM2KB
	MS51FB9AELDROM3KB
	MS51FB9AELDROM4KB
)

type MS51FB9AEConfig struct {
	// CONFIG0.CBS[7]
	BootSelect BootSelect `json:"boot_select"`
	// CONFIG0.OCDPWM[5]
	PWMEnabledDuringOCD bool `json:"pwm_enabled_during_ocd"`
	// CONFIG0.OCDEN[4]
	OCDEnabled bool `json:"ocd_enabled"`
	// CONFIG0.RPD[2]
	ResetPinDisabled bool `json:"reset_pin_disabled"`

	// CONFIG0.LOCK[1]
	Locked bool `json:"locked"`

	// CONFIG1.LDSIZE[2:0]
	LDROMSize MS51FB9AELDROMSize `json:"ldrom_size"`

	// CONFIG2.CBODEN[7]
	BODDisabled bool `json:"bod_disabled"`

	// CONFIG2.COV[5:4]
	BODVoltage BODVoltage `json:"bod_voltage"`

	// CONFIG2.BOIAP[3]
	IAPEnabledInBrownout bool `json:"iap_enabled_in_brownout"`

	// CONFIG2.CBORST[2]
	BODResetDisabled bool `json:"bod_reset_disabled"`

	// CONFIG3.WDTEN[7:4]
	WDT WDTMode `json:"wdt"`
}

func (cfg *MS51FB9AEConfig) UnmarshalBinary(buf []byte) error {
	if len(buf) < 4 {
		return errors.New("Too short for config bytes")
	}

	cfg.BootSelect = BootFromAPROM
	if buf[0]&0x80 == 0 {
		cfg.BootSelect = BootFromLDROM
	}

	cfg.PWMEnabledDuringOCD = buf[0]&0x20 == 0
	cfg.OCDEnabled = buf[0]&0x10 == 0
	cfg.ResetPinDisabled = buf[0]&0x04 == 0
	cfg.Locked = buf[0]&0x02 == 0

	switch buf[1] & 0x7 {
	case 7:
		cfg.LDROMSize = MS51FB9AELDROM0KB
	case 6:
		cfg.LDROMSize = MS51FB9AELDROM1KB
	case 5:
		cfg.LDROMSize = MS51FB9AELDROM2KB
	case 4:
		cfg.LDROMSize = MS51FB9AELDROM3KB
	default:
		cfg.LDROMSize = MS51FB9AELDROM4KB
	}

	cfg.BODDisabled = buf[2]&0x80 == 0
	switch (buf[2] >> 4) & 0x3 {
	case 0:
		cfg.BODVoltage = BODVoltage4v4
	case 1:
		cfg.BODVoltage = BODVoltage3v7
	case 2:
		cfg.BODVoltage = BODVoltage2v7
	default:
		cfg.BODVoltage = BODVoltage2v2
	}

	cfg.IAPEnabledInBrownout = buf[2]&0x08 == 0
	cfg.BODResetDisabled = buf[2]&0x04 == 0
	switch buf[3] >> 4 {
	case 0xF:
		cfg.WDT = WDTDisabled
	case 0x5:
		cfg.WDT = WDTEnabled
	default:
		cfg.WDT = WDTEnabledAlways
	}

	return nil
}

func (cfg *MS51FB9AEConfig) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 8)
	for i := range buf {
		buf[i] = 0xFF
	}

	if cfg.BootSelect == BootFromLDROM {
		buf[0] &= 0x7F
	}

	if cfg.PWMEnabledDuringOCD {
		buf[0] &= 0xDF
	}

	if cfg.OCDEnabled {
		buf[0] &= 0xEF
	}

	if cfg.ResetPinDisabled {
		buf[0] &= 0xFB
	}

	if cfg.Locked {
		buf[0] &= 0xFD
	}

	switch cfg.LDROMSize {
	case MS51FB9AELDROM0KB:
		buf[1] = 0xFF
	case MS51FB9AELDROM1KB:
		buf[1] = 0xFE
	case MS51FB9AELDROM2KB:
		buf[1] = 0xFD
	case MS51FB9AELDROM3KB:
		buf[1] = 0xFC
	case MS51FB9AELDROM4KB:
		buf[1] = 0xFB
	}

	if cfg.BODDisabled {
		buf[2] &= 0x7F
	}

	switch cfg.BODVoltage {
	case BODVoltage4v4:
		buf[2] &= 0xCF
	case BODVoltage3v7:
		buf[2] &= 0xDF
	case BODVoltage2v7:
		buf[2] &= 0xEF
	case BODVoltage2v2:
		buf[2] &= 0xFF
	}

	if cfg.IAPEnabledInBrownout {
		buf[2] &= 0xF7
	}

	if cfg.BODResetDisabled {
		buf[2] &= 0xFB
	}

	switch cfg.WDT {
	case WDTDisabled:
		buf[3] = 0xFF
	case WDTEnabled:
		buf[3] = 0x5F
	case WDTEnabledAlways:
		buf[3] = 0x0F
	}

	// Sense checking: We should unmarshal to the same values
	var newCfg MS51FB9AEConfig
	if err := newCfg.UnmarshalBinary(buf); err != nil {
		return nil, err
	}

	if newCfg != *cfg {
		panic("Roundtrip error")
	}

	return buf, nil
}

func (c *MS51FB9AEConfig) GetLDROMSize() uint {
	switch c.LDROMSize {
	case MS51FB9AELDROM0KB:
		return 0
	case MS51FB9AELDROM1KB:
		return 1024
	case MS51FB9AELDROM2KB:
		return 2048
	case MS51FB9AELDROM3KB:
		return 3072
	case MS51FB9AELDROM4KB:
		return 4096
	default:
		panic("Invalid size")
	}
}

var MS51FB9AE = &target.Definition{
	Name:        "MS51FB9AE",
	Family:      protocol.ChipFamilyN76E003,
	DeviceID:    protocol.DeviceMS51FB9AE,
	ProgMemSize: 12 * 1024,
	LDROMOffset: 0x3800,
	Config: target.ConfigSpace{
		IHexOffset: 0x30000,
		MinSize:    4,
		ReadSize:   8,
		WriteSize:  32,
		NewConfig:  func() target.Config { return new(MS51FB9AEConfig) },
	},
}

func init() {
	target.Register(MS51FB9AE)
}
