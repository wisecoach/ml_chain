package crypto

import (
	"github.com/wisecoach/ml_chain/bccsp"
	"github.com/wisecoach/ml_chain/comm/comm"
)

type MessageCryptoService interface {
	// Sign sign the msg
	Sign(msg []byte) ([]byte, error)

	// Verify verify the signature
	Verify(pk, msg, signature []byte) (bool, error)
}

func New(csp bccsp.BCCSP, sk bccsp.Key, self *comm.RemotePeer,
	importOpts bccsp.KeyImportOpts, hashOpts bccsp.HashOpts, signerOpts bccsp.SignerOpts) MessageCryptoService {
	m := &messageCryptoServiceImpl{
		sk:            sk,
		bccsp:         csp,
		keyImportOpts: importOpts,
		hashOpts:      hashOpts,
		signerOpts:    signerOpts,
	}
	print(self.Endpoint + " " + string(self.PublicKey))
	m.pk, _ = csp.KeyImport(self.PublicKey, importOpts)
	return m
}

type messageCryptoServiceImpl struct {
	bccsp bccsp.BCCSP
	sk    bccsp.Key
	pk    bccsp.Key

	keyImportOpts bccsp.KeyImportOpts
	hashOpts      bccsp.HashOpts
	signerOpts    bccsp.SignerOpts
}

func (m *messageCryptoServiceImpl) Sign(msg []byte) ([]byte, error) {
	digest, err := m.bccsp.Hash(msg, m.hashOpts)
	if err != nil {
		return nil, err
	}
	return m.bccsp.Sign(m.sk, digest, m.signerOpts)
}

func (m *messageCryptoServiceImpl) Verify(pk, msg, signature []byte) (bool, error) {
	key, err := m.bccsp.KeyImport(pk, m.keyImportOpts)
	if err != nil {
		return false, err
	}
	digest, err := m.bccsp.Hash(msg, m.hashOpts)
	return m.bccsp.Verify(key, signature, digest, m.signerOpts)
}
