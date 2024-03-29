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

package cmd

import (
	"github.com/GabrielCRZ12/ms51prog/protocol"
	"github.com/spf13/cobra"
	"fmt"
)

// programCmd represents the program command
var programCmd = &cobra.Command{
	Use:   "program",
	Short: "Program a target device",
	Long:  `Program a target device`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("connecting to target\r\n")

		dev, td, err := connectToTarget()
		if err != nil {
			return err
		}
		defer resetAndCloseDevice(dev)

		config, _ := cmd.Flags().GetString("config")
		image, _ := cmd.Flags().GetString("image")
		aprom, _ := cmd.Flags().GetString("aprom")
		ldrom, _ := cmd.Flags().GetString("ldrom")

		fmt.Printf("reading target\r\n")
		data, err := ReadTargetData(config, image, aprom, ldrom, td, true)
		if err != nil {
			return err
		}

		fmt.Printf("erase chip\r\n")
		if err := dev.EraseFlashChip(); err != nil {
			return err
		}

		fmt.Printf("writing config\r\n")
		if len(data.Config) != 0 {
			for len(data.Config) < int(td.Config.WriteSize) {
				data.Config = append(data.Config, 0xFF)
			}

			if err := dev.WriteMemory(protocol.ConfigSpace, 0, data.Config[:td.Config.WriteSize]); err != nil {
				return err
			}
		}

		apromB, err := data.APROM()
		if err != nil {
			return err
		}
		ldromB, err := data.LDROM()
		if err != nil {
			return err
		}

		fmt.Printf("writing APROM\r\n")
		for i := 0; i < len(apromB); i += 32 {
			if err := dev.WriteMemory(protocol.ProgramSpace, uint16(i), apromB[i:i+32]); err != nil {
				return err
			}
			fmt.Printf(".")
		}

		fmt.Printf("\r\nwriting LDROM (fix len 1kB)\r\n")
		for i := 0; i < len(ldromB); i += 32 {
			offs := uint16(td.LDROMOffset) + uint16(i)
			if err := dev.WriteMemory(protocol.ProgramSpace, offs, ldromB[i:i+32]); err != nil {
				return err
			}
			fmt.Printf(".")
		}
		
		fmt.Printf("\r\n\r\n		Flash Success!!!		\r\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(programCmd)
	programCmd.Flags().StringP("image", "i", "", "Image file, e.g. image.ihx")
	programCmd.Flags().StringP("config", "c", "", "Configuration, e.g. 6FFBFFFF or @config.json")
	programCmd.Flags().StringP("aprom", "a", "", "APROM file e.g. aprom.ihx")
	programCmd.Flags().StringP("ldrom", "l", "", "LDROM file e.g. ldrom.ihx")
	programCmd.Flags().BoolP("verify", "V", true, "Verify memory contents")
}
