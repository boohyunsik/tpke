package tpke

import (
	"errors"
	"github.com/boohyunsik/tpke/bls"
)

type SecretKey struct {
	FR *bls.FR
}

func (s *SecretKey) PublicKey() *PublicKey {
	return &PublicKey {
		G1: bls.G1AffineOne.MulFR(s.FR.ToRepr()),
	}
}

func (s *SecretKey) Sign(msg []byte) *Signature {
	g2Hash := HashG2(msg).ToAffine()
	return &Signature {
		G2: g2Hash.MulFR(s.FR.ToRepr()),
	}
}

func (s *SecretKey) Decrypt(cipher *CipherText) []byte {
	if !cipher.Verify() {
		return nil
	}
	g := cipher.U.ToAffine().MulFR(s.FR.ToRepr())
	//d, _ := xorHash(*g, cipher.V)
	d := xorHash(*g, cipher.V)
	return d
}

func (s *SecretKey) Serialize() [32]byte {
	return s.FR.Bytes()
}

func NewSecretKeyFromBytes(bytes [32]byte) *SecretKey {
	frRepr := bls.FRReprFromBytes(bytes)
	return &SecretKey {
		FR: bls.FRReprToFR(frRepr),
	}
}

func (s *SecretKey) Equals(other *SecretKey) bool {
	return s.FR.Equals(other.FR)
}

type SecretKeySet struct {
	poly Poly
}

func RandomSecretKeySet(threshold int) *SecretKeySet {
	randomPoly := randomPoly(threshold + 1)
	return &SecretKeySet{
		poly: *randomPoly,
	}
}

func NewSecretKeySetFromBytes(bytes []byte) (*SecretKeySet, error) {
	l := len(bytes)
	if l % 32 != 0 {
		return nil, errors.New("length of byte must be a multiple of 32")
	}

	frs := make([]*bls.FR, 0)
	idx := 0
	for i := 0; i < l / 32; i++ {
		var b [32]byte
		for j := 0; j < 32; j++ {
			b[j] = bytes[idx]
			idx++
		}
		fr := bls.FRReprFromBytes(b)
		frs = append(frs, bls.FRReprToFR(fr))
	}

	return &SecretKeySet {
		poly: Poly {
			coeff: frs,
		},
	}, nil
}

func (sks *SecretKeySet) threshold() int {
	return sks.poly.degree()
}

func (sks *SecretKeySet) Serialize() []byte {
	ret := make([]byte, 0)
	for _, c := range sks.poly.coeff {
		b := c.Bytes()
		for i := 0; i < len(b); i++ {
			ret = append(ret, b[i])
		}
	}
	return ret
}

func (sks *SecretKeySet) PublicKeySet() *PublicKeySet {
	return &PublicKeySet{
		commitment: sks.poly.commitment(),
	}
}

func (sks *SecretKeySet) KeyShare(i int) *SecretKeyShare {
	//fr := bls.FRReprToFR(bls.NewFRRepr(uint64(i + 1)))
	fr := bls.FRReprToFR(bls.NewFRRepr(uint64(1)))
	x := bls.FRReprToFR(bls.NewFRRepr(uint64(i)))
	fr.AddAssign(x)
	eval := sks.poly.evaluate(*fr)
	return &SecretKeyShare {
		sk: &SecretKey {
			FR: eval,
		},
	}
}

func (sks *SecretKeySet) KeyShareUsingString(s string) *SecretKeyShare {
	frRepr, _ := bls.FRReprFromString(s, 10)
	fr := bls.FRReprToFR(frRepr)
	x := bls.FRReprToFR(frRepr)
	fr.AddAssign(x)
	eval := sks.poly.evaluate(*fr)
	return &SecretKeyShare {
		sk: &SecretKey {
			FR: eval,
		},
	}
}

func (sks *SecretKeySet) Equals (other *SecretKeySet) bool {
	if len(sks.poly.coeff) != len(other.poly.coeff) {
		return false
	}

	for i := range sks.poly.coeff {
		if !sks.poly.coeff[i].Equals(other.poly.coeff[i]) {
			return false
		}
	}

	return true
}

type SecretKeyShare struct {
	sk *SecretKey
}

func NewSecretKeyShare(sk *SecretKey) *SecretKeyShare {
	return &SecretKeyShare {
		sk: sk,
	}
}

func (sks *SecretKeyShare) DecryptShare(ct *CipherText) *DecryptionShare {
	// TODO : verify
	return &DecryptionShare {
		G1: ct.U.ToAffine().MulFR(sks.sk.FR.ToRepr()).Copy(),
	}
}

func (sks *SecretKeyShare) String() string {
	return sks.sk.FR.String()
}

func (sks *SecretKeyShare) Serialize() [32]byte {
	return sks.sk.Serialize()
}

func NewSecretKeyShareFromBytes(bytes [32]byte) *SecretKeyShare {
	return &SecretKeyShare {
		sk: NewSecretKeyFromBytes(bytes),
	}
}