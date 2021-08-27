package pkg

// Encrypt 配置文件加解密方式
type Encrypt interface {
	// IsCipherText 是否为加密数据
	IsCipherText([]byte) bool
	// Encrypt 数据加密
	Encrypt([]byte) []byte
	// Decode 数据解密
	Decode([]byte) []byte
}
