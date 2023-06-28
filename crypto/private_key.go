package crypto

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"io"
	"time"
)

type OpenpgpPrivateKey struct {
	privateKey *packet.PrivateKey
}

func (k *OpenpgpPrivateKey) PublicKey() *OpenpgpPubKey {
	return newOpenPGPPubKey(&k.privateKey.PublicKey)
}

func (k *OpenpgpPrivateKey) keyEncode(w io.Writer) error {
	return k.privateKey.Serialize(w)
}

var openpgpConfig = &packet.Config{
	DefaultHash: crypto.SHA512,
}

func (k *OpenpgpPrivateKey) SignContent(content []byte) ([]byte, error) {
	sig, err := k.sign(content)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = sig.Serialize(buf)
	if err != nil {
		return nil, err
	}

	return EncodeV1(buf.Bytes()), nil
}

func (k *OpenpgpPrivateKey) sign(content []byte) (*packet.Signature, error) {
	privk := k.privateKey
	sig := new(packet.Signature)
	sig.PubKeyAlgo = privk.PubKeyAlgo
	sig.Hash = openpgpConfig.Hash()
	sig.CreationTime = time.Now()

	h := openpgpConfig.Hash().New()
	h.Write(content)

	err := sig.Sign(h, privk, openpgpConfig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func ReadPrivateKey(pk string) (OpenpgpPrivateKey, *rsa.PrivateKey) {
	rd := bytes.NewReader([]byte(pk))
	blk, err := armor.Decode(rd)
	var body io.Reader
	if err == nil {
		body = blk.Body
	} else {
		rd.Seek(0, 0)
		// try unarmored
		body = base64.NewDecoder(base64.StdEncoding, rd)
	}
	pkt, err := packet.Read(body)
	if err != nil {
		panic(err)
	}

	pkPkt := pkt.(*packet.PrivateKey)
	rsaPrivateKey, ok := pkPkt.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		panic("not a RSA key")
	}

	return RSAPrivateKey(rsaPrivateKey), rsaPrivateKey
}
