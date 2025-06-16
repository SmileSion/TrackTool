package controller

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qwy-tacking/config"
	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/model"
	"github.com/qwy-tacking/storage"
)

// 解密 AES ECB 模式 + PKCS#7 padding
func decryptAES_ECB(cipherData []byte) ([]byte, error) {
	// 获取 AES 密钥
	key, err := hex.DecodeString(config.Conf.AES.Key)
	if err != nil {
		return nil, fmt.Errorf("无效的 AES 密钥格式: %v", err)
	}
	// fmt.Println("AES密钥（hex解码后）:", key)
	if len(key) != 16 {
		return nil, fmt.Errorf("无效的 AES 密钥，密钥长度应为 16 字节")
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建 AES cipher 失败: %v", err)
	}

	// 校验密文长度
	if len(cipherData)%aes.BlockSize != 0 {
		return nil, errors.New("密文长度必须为 AES blockSize 的倍数")
	}

	// 解密
	plain := make([]byte, len(cipherData))
	for bs, be := 0, aes.BlockSize; bs < len(cipherData); bs, be = bs+aes.BlockSize, be+aes.BlockSize {
		block.Decrypt(plain[bs:be], cipherData[bs:be])
	}

	// 去除 PKCS#7 padding
	paddingLen := int(plain[len(plain)-1])
	if paddingLen == 0 || paddingLen > aes.BlockSize {
		return nil, errors.New("padding 格式错误")
	}

	return plain[:len(plain)-paddingLen], nil
}

// 时间戳差值验证，10 分钟有效期
func isTimestampValid(eventTime int64) bool {
	return time.Now().Unix()-eventTime <= 600
}

// TrackHandler 接收加密埋点数据，解密校验入 Redis
func TrackHandler(c *gin.Context) {
	var req struct {
		Data string `json:"data"`
	}

	// 参数校验
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	// fmt.Println("接收到的原始数据:", req.Data)

	// Base64 解码
	if len(req.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data 字段不能为空"})
		return
	}
	rawData, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "base64解码失败"})
		return
	}

	// AES 解密
	decrypted, err := decryptAES_ECB(rawData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "解密失败: " + err.Error()})
		return
	}

	// 解密后的数据结构
	var event model.Event
	if err := json.Unmarshal(decrypted, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON解析失败: " + err.Error()})
		return
	}

	// 校验时间戳
	if !isTimestampValid(event.TimeStamp) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "数据过期"})
		return
	}

	// 存入 Redis
	if err := storage.SaveEventToRedis(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis错误: " + err.Error()})
		return
	}

	middleware.Logger.Printf("收到事件: %+v\n", event)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
