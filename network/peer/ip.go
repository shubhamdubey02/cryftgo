// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package peer

import (
	"crypto"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/shubhamdubey02/cryftgo/staking"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/bls"
	"github.com/shubhamdubey02/cryftgo/utils/hashing"
	"github.com/shubhamdubey02/cryftgo/utils/ips"
	"github.com/shubhamdubey02/cryftgo/utils/wrappers"
)

var (
	errTimestampTooFarInFuture = errors.New("timestamp too far in the future")
	errInvalidTLSSignature     = errors.New("invalid TLS signature")
)

// UnsignedIP is used for a validator to claim an IP. The [Timestamp] is used to
// ensure that the most updated IP claim is tracked by peers for a given
// validator.
type UnsignedIP struct {
	ips.IPPort
	Timestamp uint64
}

// Sign this IP with the provided signer and return the signed IP.
func (ip *UnsignedIP) Sign(tlsSigner crypto.Signer, blsSigner *bls.SecretKey) (*SignedIP, error) {
	ipBytes := ip.bytes()
	tlsSignature, err := tlsSigner.Sign(
		rand.Reader,
		hashing.ComputeHash256(ipBytes),
		crypto.SHA256,
	)
	blsSignature := bls.SignProofOfPossession(blsSigner, ipBytes)
	return &SignedIP{
		UnsignedIP:        *ip,
		TLSSignature:      tlsSignature,
		BLSSignature:      blsSignature,
		BLSSignatureBytes: bls.SignatureToBytes(blsSignature),
	}, err
}

func (ip *UnsignedIP) bytes() []byte {
	p := wrappers.Packer{
		Bytes: make([]byte, ips.IPPortLen+wrappers.LongLen),
	}
	ips.PackIP(&p, ip.IPPort)
	p.PackLong(ip.Timestamp)
	return p.Bytes
}

// SignedIP is a wrapper of an UnsignedIP with the signature from a signer.
type SignedIP struct {
	UnsignedIP
	TLSSignature      []byte
	BLSSignature      *bls.Signature
	BLSSignatureBytes []byte
}

// Returns nil if:
// * [ip.Timestamp] is not after [maxTimestamp].
// * [ip.TLSSignature] is a valid signature over [ip.UnsignedIP] from [cert].
func (ip *SignedIP) Verify(
	cert *staking.Certificate,
	maxTimestamp time.Time,
) error {
	maxUnixTimestamp := uint64(maxTimestamp.Unix())
	if ip.Timestamp > maxUnixTimestamp {
		return fmt.Errorf("%w: timestamp %d > maxTimestamp %d", errTimestampTooFarInFuture, ip.Timestamp, maxUnixTimestamp)
	}

	if err := staking.CheckSignature(
		cert,
		ip.UnsignedIP.bytes(),
		ip.TLSSignature,
	); err != nil {
		return fmt.Errorf("%w: %w", errInvalidTLSSignature, err)
	}
	return nil
}
