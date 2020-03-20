// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package secboot_test

import (
	"reflect"
	"testing"

	"github.com/chrisccoulson/go-tpm2"
	. "github.com/snapcore/secboot"
)

func TestAddSystemdEFIStubProfile(t *testing.T) {
	for _, data := range []struct {
		desc    string
		initial *PCRProtectionProfile
		params  SystemdEFIStubProfileParams
		values  []tpm2.PCRValues
	}{
		{
			desc: "UC20",
			params: SystemdEFIStubProfileParams{
				PCRAlgorithm: tpm2.HashAlgorithmSHA256,
				PCRIndex:     12,
				KernelCmdlines: []string{
					"console=ttyS0 console=tty1 panic=-1 systemd.gpt_auto=0 snapd_recovery_mode=run",
					"console=ttyS0 console=tty1 panic=-1 systemd.gpt_auto=0 snapd_recovery_mode=recover",
				},
			},
			values: []tpm2.PCRValues{
				{
					tpm2.HashAlgorithmSHA256: {
						12: decodeHexString(t, "fc433eaf039c6261f496a2a5bf2addfd8ff1104b0fc98af3fe951517e3bde824"),
					},
				},
				{
					tpm2.HashAlgorithmSHA256: {
						12: decodeHexString(t, "b3a29076eeeae197ae721c254da40480b76673038045305cfa78ec87421c4eea"),
					},
				},
			},
		},
		{
			desc: "SHA1",
			params: SystemdEFIStubProfileParams{
				PCRAlgorithm: tpm2.HashAlgorithmSHA1,
				PCRIndex:     12,
				KernelCmdlines: []string{
					"console=ttyS0 console=tty1 panic=-1 systemd.gpt_auto=0 snapd_recovery_mode=run",
					"console=ttyS0 console=tty1 panic=-1 systemd.gpt_auto=0 snapd_recovery_mode=recover",
				},
			},
			values: []tpm2.PCRValues{
				{
					tpm2.HashAlgorithmSHA1: {
						12: decodeHexString(t, "eb6312b7db70fe16206c162326e36b2fcda74b68"),
					},
				},
				{
					tpm2.HashAlgorithmSHA1: {
						12: decodeHexString(t, "bd612bea9efa582fcbfae97973c89b163756fe0b"),
					},
				},
			},
		},
		{
			desc: "Classic",
			params: SystemdEFIStubProfileParams{
				PCRAlgorithm: tpm2.HashAlgorithmSHA256,
				PCRIndex:     8,
				KernelCmdlines: []string{
					"root=/dev/mapper/vgubuntu-root ro quiet splash vt.handoff=7",
				},
			},
			values: []tpm2.PCRValues{
				{
					tpm2.HashAlgorithmSHA256: {
						8: decodeHexString(t, "74fe9080b798f9220c18d0fcdd0ccb82d50ce2a317bc6cdaa2d8715d02d0efbe"),
					},
				},
			},
		},
		{
			desc: "WithInitialProfile",
			initial: func() *PCRProtectionProfile {
				return NewPCRProtectionProfile().
					AddPCRValue(tpm2.HashAlgorithmSHA256, 7, makePCRDigestFromEvents(tpm2.HashAlgorithmSHA256, "foo")).
					AddPCRValue(tpm2.HashAlgorithmSHA256, 8, makePCRDigestFromEvents(tpm2.HashAlgorithmSHA256, "bar"))
			}(),
			params: SystemdEFIStubProfileParams{
				PCRAlgorithm: tpm2.HashAlgorithmSHA256,
				PCRIndex:     8,
				KernelCmdlines: []string{
					"root=/dev/mapper/vgubuntu-root ro quiet splash vt.handoff=7",
				},
			},
			values: []tpm2.PCRValues{
				{
					tpm2.HashAlgorithmSHA256: {
						7: makePCRDigestFromEvents(tpm2.HashAlgorithmSHA256, "foo"),
						8: decodeHexString(t, "3d39c0db757b47b484006003724d990403d533044ed06e8798ab374bd73f32dc"),
					},
				},
			},
		},
	} {
		t.Run(data.desc, func(t *testing.T) {
			profile := data.initial
			if profile == nil {
				profile = NewPCRProtectionProfile()
			}
			if err := AddSystemdEFIStubProfile(profile, &data.params); err != nil {
				t.Fatalf("AddSystemdEFIStubProfile failed: %v", err)
			}
			values, err := profile.ComputePCRValues(nil)
			if err != nil {
				t.Fatalf("ComputePCRValues failed: %v", err)
			}
			if !reflect.DeepEqual(values, data.values) {
				t.Errorf("ComputePCRValues returned unexpected values")
				for i, v := range values {
					t.Logf("Value %d:", i)
					for alg := range v {
						for pcr := range v[alg] {
							t.Logf(" PCR%d,%v: %x", pcr, alg, v[alg][pcr])
						}
					}
				}
			}
		})
	}
}