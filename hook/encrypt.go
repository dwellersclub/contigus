package hook

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"io"
)

/*
Encryptor Interface to encrypt/decrypt data
*/
type Encryptor interface {
	Encrypt(data io.Reader) (string, []byte, error)
	EncryptWithID(data io.Reader, id string) ([]byte, error)
	Decrypt(data io.Reader, id string) ([]byte, error)
}

//NewEncryptor Create a new encryptor
func NewEncryptor() Encryptor {
	return &dynamicEncryptor{}
}

type encryptKey struct {
	value []byte
	id    string
}

type dynamicEncryptor struct{}

func (de *dynamicEncryptor) getRandomKey() encryptKey {
	return encryptKey{}
}
func (de *dynamicEncryptor) getKeyByID(ID string) encryptKey {
	return encryptKey{}
}

func (de *dynamicEncryptor) Encrypt(data io.Reader) (string, []byte, error) {
	return de.encrypt(data, de.getRandomKey())
}

func (de *dynamicEncryptor) EncryptWithID(data io.Reader, id string) ([]byte, error) {
	_, reader, err := de.encrypt(data, de.getKeyByID(id))
	return reader, err
}

func (de *dynamicEncryptor) encrypt(data io.Reader, key encryptKey) (string, []byte, error) {

	block, err := aes.NewCipher(key.value)
	if err != nil {
		return "", nil, err
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	var out bytes.Buffer
	writer := &cipher.StreamWriter{S: stream, W: &out}

	_, err = io.Copy(writer, data)

	if err != nil {
		return "", nil, err
	}

	return key.id, out.Bytes(), nil
}

func (de *dynamicEncryptor) Decrypt(data io.Reader, id string) ([]byte, error) {
	return nil, nil
}
