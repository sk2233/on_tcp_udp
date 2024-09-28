/*
@author: sk
@date: 2024/9/16
*/
package https

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"my_tcp/utils"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestHttps(t *testing.T) {
	server := NewHttpsServer("127.0.0.1:8080", utils.BasePath+"https/cert.pem",
		utils.BasePath+"https/key.pem")
	server.RegisterHandler("/", GoIndex)
	server.Listen()
}

func GoIndex(_ *HttpsReq, conn net.Conn) {
	resp := NewHttpsResp(200, "OK")
	bs, err := os.ReadFile(utils.BasePath + "https/index.html")
	utils.HandleErr(err)
	resp.Header["Content-Length"] = strconv.Itoa(len(bs))
	resp.Header["Content-Type"] = "text/html; charset=utf-8"
	resp.Data = bs
	WriteResp(conn, resp)
	conn.Close()
}

// 生成证书与公钥
func TestGenPem(t *testing.T) {
	max0 := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max0)
	subject := pkix.Name{
		Organization:       []string{"com.msr"},
		OrganizationalUnit: []string{"better"},
		CommonName:         "go web",
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	pk, _ := rsa.GenerateKey(rand.Reader, 2048)
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)
	certOut, _ := os.Create(utils.BasePath + "https/cert.pem")
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, _ := os.Create(utils.BasePath + "https/key.pem")
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	keyOut.Close()
}
