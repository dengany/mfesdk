package mfesdk

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log"
	"os"

	"golang.org/x/crypto/pkcs12"
)

type MFECONF struct {
	IsProd             bool            // 是否生产环境 true:生产环境 false:测试环境
	AgencyNo           string          // 机构号
	PublicKey          *rsa.PublicKey  // 公钥
	PrivateKey         *rsa.PrivateKey // 私钥
	PrivateKeyPassword string          // 私钥密码
}
type MfeOption struct {
	IsProd   bool   // 是否生产环境| 生产环境:true 测试环境:false
	AgencyNo string // 机构号
	PubPath  string // 公钥
	PriPath  string // 私钥
	PriPwd   string // 私钥密码
}

func NewMfe(op *MfeOption) *MFECONF {

	if op.PriPath == "" || op.PubPath == "" {
		log.Println("公钥或私钥路径不能为空")
		return nil
	}

	public_key, err := getPublicKey(op.PubPath)
	if err != nil {
		log.Println("获取公钥失败:", err)
		return nil
	}
	private_key, err := getPrivateKey(op.PriPath, op.PriPwd)
	if err != nil {
		// 打印中文日志
		log.Println("获取私钥失败:", err)
		return nil
	}
	return &MFECONF{
		IsProd:             op.IsProd,
		AgencyNo:           op.AgencyNo,
		PublicKey:          public_key,
		PrivateKey:         private_key,
		PrivateKeyPassword: op.PriPwd,
	}
}

func getPrivateKey(pfxPath string, passphrase string) (*rsa.PrivateKey, error) {
	rawData, err := os.ReadFile(pfxPath)
	if err != nil {
		return nil, err
	}
	blocks, err := pkcs12ToPem(rawData, passphrase)
	if err != nil {
		return nil, err
	}

	_, rest := pem.Decode(blocks[0].Bytes)

	pk, err := x509.ParsePKCS1PrivateKey(rest)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

func pkcs12ToPem(rawData []byte, passphrase string) ([]*pem.Block, error) {
	blocks, err := pkcs12.ToPEM(rawData, passphrase)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func getPublicKey(cerPath string) (*rsa.PublicKey, error) {
	rawData, err := os.ReadFile(cerPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(rawData)
	if block == nil {
		return nil, errors.New("PEM块中的数据格式不正确,无法被正确解析。")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("无法提取RSA公钥")
	}
	return pubKey, nil
}
func encrypt(data string, mfecfg *MFECONF) ([]byte, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, mfecfg.PublicKey, []byte(data))
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	return []byte(encoded), nil
}

func decrypt(cipher string, cfg *MFECONF) (string, error) {

	decoded, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return "", err
	}
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, cfg.PrivateKey, decoded)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

func signature(data []byte, cfg *MFECONF) (string, error) {

	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, cfg.PrivateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(signature)
	return encoded, nil
}

func verify(data []byte, sign []byte, cfg *MFECONF) (bool, error) {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	decoded, err := base64.StdEncoding.DecodeString(string(sign))
	if err != nil {
		return false, err
	}
	err = rsa.VerifyPKCS1v15(cfg.PublicKey, crypto.SHA256, hashed, decoded)
	if err != nil {
		return false, err
	}
	return true, nil
}

func fileencrypt(data string, pub *rsa.PublicKey) ([]byte, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(data))
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}
